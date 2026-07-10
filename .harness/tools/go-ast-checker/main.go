package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// CheckResult represents a single check result
type CheckResult struct {
	Check    string `json:"check"`
	Status   string `json:"status"` // PASS, FAIL, WARN
	Detail   string `json:"detail"`
	Location string `json:"location,omitempty"`
	Why      string `json:"why,omitempty"`
	Fix      string `json:"fix,omitempty"`
	Example  string `json:"example,omitempty"`
}

// Checker performs AST-based checks on Go code
type Checker struct {
	results []CheckResult
}

// CheckSnowflakeIDTag checks if int64 ID fields have json:",string" tag
func (c *Checker) CheckSnowflakeIDTag(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Iterate through struct fields
		for _, field := range structType.Fields.List {
			if field.Tag == nil {
				continue
			}

			// Check each field name
			for _, name := range field.Names {
				fieldName := name.Name

				// Check if field name ends with Id or ID
				if !strings.HasSuffix(fieldName, "Id") && !strings.HasSuffix(fieldName, "ID") {
					continue
				}

				// Check if type is int64
				ident, ok := field.Type.(*ast.Ident)
				if !ok || ident.Name != "int64" {
					continue
				}

				// Extract tag value
				tagValue := field.Tag.Value
				// Remove backticks
				tagValue = strings.Trim(tagValue, "`")

				// Check for json tag with string option
				hasJSONTag := false
				hasStringOption := false

				// Parse struct tags
				tags := parseStructTag(tagValue)
				if jsonTag, ok := tags["json"]; ok {
					hasJSONTag = true
					// Check if contains "string" option
					parts := strings.Split(jsonTag, ",")
					for _, part := range parts {
						if strings.TrimSpace(part) == "string" {
							hasStringOption = true
							break
						}
					}
				}

				// Also check for path/form/header/db tags (these don't need string option)
				skipCheck := false
				for _, skipTag := range []string{"path", "form", "header", "db"} {
					if _, ok := tags[skipTag]; ok {
						skipCheck = true
						break
					}
				}

				if skipCheck {
					continue
				}

				// Report if json tag exists but missing string option
				if hasJSONTag && !hasStringOption {
					pos := fset.Position(field.Pos())
					c.results = append(c.results, CheckResult{
						Check:    "json_string_tag",
						Status:   "FAIL",
						Detail:   fmt.Sprintf("%s.%s (int64) has json tag but missing 'string' option", typeSpec.Name, fieldName),
						Location: fmt.Sprintf("%s:%d", filename, pos.Line),
						Why:      "Snowflake IDs exceed JavaScript Number.MAX_SAFE_INTEGER, must be transmitted as strings",
						Fix:      fmt.Sprintf("Add 'string' option: json:\"%s,string\"", getJSONFieldName(tags["json"])),
						Example:  fmt.Sprintf("%s int64 `json:\"%s,string\"`", fieldName, getJSONFieldName(tags["json"])),
					})
				}
			}
		}
		return true
	})

	return nil
}

// CheckCrossServiceImport checks if code imports another service's model package
func (c *Checker) CheckCrossServiceImport(filename, currentService string, serviceModules map[string]string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	currentModule := serviceModules[currentService]
	if currentModule == "" {
		return nil // Skip if current service not found
	}

	// Check imports
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		// Skip if importing own module's packages
		if currentModule != "" && strings.HasPrefix(importPath, currentModule) {
			continue
		}

		// Check if importing another service's model package
		for svcName, svcModule := range serviceModules {
			if svcName == currentService {
				continue
			}

			// Skip if module is empty (e.g., Python services)
			if svcModule == "" {
				continue
			}

			// Check if import is from another service's internal packages
			if strings.HasPrefix(importPath, svcModule) &&
				(strings.Contains(importPath, "/model") ||
					strings.Contains(importPath, "/internal")) {
				pos := fset.Position(imp.Pos())
				c.results = append(c.results, CheckResult{
					Check:    "cross_service_import",
					Status:   "FAIL",
					Detail:   fmt.Sprintf("Cross-service import detected: %s imports %s", currentService, importPath),
					Location: fmt.Sprintf("%s:%d", filename, pos.Line),
					Why:      "Services should communicate via gRPC, not direct package imports",
					Fix:      fmt.Sprintf("Remove import and use gRPC client for %s", svcName),
					Example:  fmt.Sprintf("client := %sClient.New(...)", svcName),
				})
			}
		}
	}

	return nil
}

// parseStructTag parses a struct tag string into a map
func parseStructTag(tag string) map[string]string {
	result := make(map[string]string)

	// Simple tag parser (handles common cases)
	parts := strings.Fields(tag)
	for _, part := range parts {
		colonIdx := strings.Index(part, ":")
		if colonIdx == -1 {
			continue
		}

		key := part[:colonIdx]
		value := strings.Trim(part[colonIdx+1:], `"`)
		result[key] = value
	}

	return result
}

// getJSONFieldName extracts the field name from json tag value
func getJSONFieldName(jsonTag string) string {
	parts := strings.Split(jsonTag, ",")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return ""
}

func main() {
	var (
		serviceDir   string
		serviceName  string
		jsonOutput   bool
		registryFile string
	)

	flag.StringVar(&serviceDir, "service-dir", "", "Service directory to check")
	flag.StringVar(&serviceName, "service-name", "", "Service name")
	flag.BoolVar(&jsonOutput, "json", false, "Output results as JSON")
	flag.StringVar(&registryFile, "registry", ".harness/registry/services.json", "Service registry file")
	flag.Parse()

	if serviceDir == "" || serviceName == "" {
		fmt.Fprintln(os.Stderr, "Usage: go-ast-checker -service-dir <dir> -service-name <name> [-json] [-registry <file>]")
		os.Exit(1)
	}

	// Load service registry
	serviceModules := make(map[string]string)
	if data, err := os.ReadFile(registryFile); err == nil {
		var registry struct {
			Services []struct {
				Name   string `json:"name"`
				Module string `json:"module"`
			} `json:"services"`
		}
		if err := json.Unmarshal(data, &registry); err == nil {
			for _, svc := range registry.Services {
				serviceModules[svc.Name] = svc.Module
			}
		}
	}

	checker := &Checker{}

	// Walk through all .go files in service directory
	err := filepath.Walk(serviceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Check Snowflake ID tags
		if err := checker.CheckSnowflakeIDTag(path); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to check %s: %v\n", path, err)
		}

		// Check cross-service imports
		if err := checker.CheckCrossServiceImport(path, serviceName, serviceModules); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to check imports in %s: %v\n", path, err)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	// Output results
	if jsonOutput {
		output, _ := json.MarshalIndent(checker.results, "", "  ")
		fmt.Println(string(output))
	} else {
		// Human-readable output
		if len(checker.results) == 0 {
			fmt.Println("✅ All AST checks passed")
		} else {
			for _, result := range checker.results {
				if result.Status == "FAIL" {
					fmt.Printf("❌ FAIL: %s\n", result.Check)
				} else if result.Status == "WARN" {
					fmt.Printf("⚠️  WARN: %s\n", result.Check)
				}
				fmt.Printf("   Detail: %s\n", result.Detail)
				if result.Location != "" {
					fmt.Printf("   Location: %s\n", result.Location)
				}
				if result.Why != "" {
					fmt.Printf("   Why: %s\n", result.Why)
				}
				if result.Fix != "" {
					fmt.Printf("   Fix: %s\n", result.Fix)
				}
				fmt.Println()
			}
		}
	}

	// Exit with non-zero if any failures
	for _, result := range checker.results {
		if result.Status == "FAIL" {
			os.Exit(1)
		}
	}
}
