package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TsNodes holds all node types parsed from TypeScript sources.
type TsNodes struct {
	Interfaces []map[string]any
	Fields     []map[string]any
	ApiCalls   []map[string]any
}

// ParseTypeScriptResult groups parsed TypeScript nodes and relationships.
type ParseTypeScriptResult struct {
	Nodes TsNodes
	Rels  []RelDef
}

// ParseTypeScript scans web directories for TypeScript interfaces, types, and API call definitions.
func ParseTypeScript(projectRoot string) (*ParseTypeScriptResult, error) {
	result := &ParseTypeScriptResult{}

	// Parse interfaces from web/common/types/
	commonTypesDir := filepath.Join(projectRoot, "web", "common", "types")
	if entries, err := os.ReadDir(commonTypesDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".d.ts") && !strings.HasSuffix(entry.Name(), ".ts")) {
				continue
			}
			filePath := filepath.Join(commonTypesDir, entry.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			content := string(data)
			sourceID := fmt.Sprintf("web:common/types:%s", entry.Name())

			ifaces, fields, ifaceRels := parseTsInterfaces(content, sourceID, "common/types")
			result.Nodes.Interfaces = append(result.Nodes.Interfaces, ifaces...)
			result.Nodes.Fields = append(result.Nodes.Fields, fields...)
			result.Rels = append(result.Rels, ifaceRels...)
		}
	}

	// Parse API calls from web/pc/src/api/ and web/mobile/src/api/
	for _, apiDir := range []string{
		filepath.Join(projectRoot, "web", "pc", "src", "api"),
		filepath.Join(projectRoot, "web", "mobile", "src", "api"),
	} {
		if entries, err := os.ReadDir(apiDir); err == nil {
			prefix := "web/pc"
			if strings.Contains(apiDir, "mobile") {
				prefix = "web/mobile"
			}
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ts") {
					continue
				}
				filePath := filepath.Join(apiDir, entry.Name())
				data, err := os.ReadFile(filePath)
				if err != nil {
					continue
				}
				content := string(data)

				apiCalls, apiRels := parseTsApiCalls(content, prefix, entry.Name())
				result.Nodes.ApiCalls = append(result.Nodes.ApiCalls, apiCalls...)
				result.Rels = append(result.Rels, apiRels...)
			}
		}
	}

	return result, nil
}

// parseTsInterfaces extracts interface definitions from a TypeScript file content.
func parseTsInterfaces(content, sourceID, fileLabel string) ([]map[string]any, []map[string]any, []RelDef) {
	var interfaces []map[string]any
	var fields []map[string]any
	var rels []RelDef

	// Match: export interface Xxx { ... } or interface Xxx { ... }
	ifaceRe := regexp.MustCompile(`(?m)^\s*(export\s+)?interface\s+(\w+)\s*(?:extends\s+[\w<>,\s]+)?\{`)
	ifaceLocs := ifaceRe.FindAllStringSubmatchIndex(content, -1)

	for _, loc := range ifaceLocs {
		ifaceName := content[loc[4]:loc[5]]
		ifaceStart := loc[1] // position after '{'

		// Find the closing brace (track depth)
		depth := 1
		pos := ifaceStart
		for pos < len(content) && depth > 0 {
			switch content[pos] {
			case '{':
				depth++
			case '}':
				depth--
			}
			pos++
		}
		if depth != 0 {
			continue
		}
		ifaceBlock := content[ifaceStart : pos-1]

		ifaceID := fmt.Sprintf("%s:%s", sourceID, ifaceName)

		interfaces = append(interfaces, map[string]any{
			"id":   ifaceID,
			"name": ifaceName,
			"file": fileLabel,
		})

		// Parse fields within this interface
		ifaceFields := parseTsInterfaceFields(ifaceBlock, ifaceID, ifaceName)
		fields = append(fields, ifaceFields...)
		for _, f := range ifaceFields {
			rels = append(rels, RelDef{
				FromID: ifaceID,
				Type:   "HAS_FIELD",
				ToID:   f["id"].(string),
			})
		}
	}

	return interfaces, fields, rels
}

// parseTsInterfaceFields extracts fields from a TypeScript interface block.
func parseTsInterfaceFields(block, ifaceID, ifaceName string) []map[string]any {
	var fields []map[string]any

	// Pattern: fieldName: type; or fieldName?: type;
	fieldRe := regexp.MustCompile(`(?m)^\s+(\w+)\??:\s*([^;]+);`)
	matches := fieldRe.FindAllStringSubmatch(block, -1)

	seen := make(map[string]bool)
	for _, m := range matches {
		fieldName := m[1]
		if seen[fieldName] {
			continue
		}
		seen[fieldName] = true

		fieldType := strings.TrimSpace(m[2])

		fieldID := fmt.Sprintf("%s.%s", ifaceID, fieldName)
		fields = append(fields, map[string]any{
			"id":   fieldID,
			"name": fieldName,
			"type": fieldType,
		})
	}

	return fields
}

// parseTsApiCalls extracts API function definitions from TypeScript API files.
func parseTsApiCalls(content, prefix, fileName string) ([]map[string]any, []RelDef) {
	var apiCalls []map[string]any
	var rels []RelDef

	// Match exported functions that make API calls (both `export const` and `export async function` patterns)
	funcRe := regexp.MustCompile(`(?m)^export\s+(?:(?:async\s+)?function\s+|const\s+)(\w+)\s*[:=\(]`)
	funcLocs := funcRe.FindAllStringSubmatchIndex(content, -1)

	for _, loc := range funcLocs {
		funcStart := loc[1]
		funcName := content[loc[2]:loc[3]]

		// Read from funcStart to the end of the function (semicolon or next export)
		endPos := len(content)
		nextFuncLoc := funcRe.FindStringIndex(content[funcStart+len(funcName)+1:])
		if nextFuncLoc != nil {
			endPos = funcStart + 1 + nextFuncLoc[0]
		}

		// Find URL strings in the function body
		funcBody := content[funcStart:endPos]
		urlRe := regexp.MustCompile(`['"](/api/[^'"]+)['"]`)
		urlMatches := urlRe.FindAllStringSubmatch(funcBody, -1)

		// Find HTTP method calls
		methodRe := regexp.MustCompile(`request\.(get|post|put|delete|patch)\b`)
		methodMatch := methodRe.FindStringSubmatch(funcBody)

		method := "GET"
		if methodMatch != nil {
			method = strings.ToUpper(methodMatch[1])
		}

		if len(urlMatches) == 0 {
			continue
		}

		for _, urlMatch := range urlMatches {
			url := urlMatch[1]
			apiCallID := fmt.Sprintf("%s:%s", prefix, funcName)

			apiCalls = append(apiCalls, map[string]any{
				"id":     apiCallID,
				"name":   funcName,
				"method": method,
				"url":    url,
				"file":   fmt.Sprintf("%s/src/api/%s", prefix, fileName),
			})
		}
	}

	return apiCalls, rels
}
