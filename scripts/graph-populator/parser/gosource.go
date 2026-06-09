package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GoNodes holds all node types parsed from Go sources.
type GoNodes struct {
	Services []map[string]any
	Structs  []map[string]any
	Fields   []map[string]any
	Routes   []map[string]any
	Tables   []map[string]any
	Columns  []map[string]any
}

// ParseGoSourcesResult groups parsed Go nodes and relationships.
type ParseGoSourcesResult struct {
	Nodes GoNodes
	Rels  []RelDef
}

// ParseGoSources scans services/*/ for Go structs, routes, and models.
func ParseGoSources(projectRoot string) (*ParseGoSourcesResult, error) {
	result := &ParseGoSourcesResult{}
	servicesDir := filepath.Join(projectRoot, "services")

	// Discover service directories
	serviceDirs, err := os.ReadDir(servicesDir)
	if err != nil {
		return nil, fmt.Errorf("list services dir: %w", err)
	}

	for _, dir := range serviceDirs {
		if !dir.IsDir() {
			continue
		}
		serviceName := dir.Name()
		svcPath := filepath.Join(servicesDir, serviceName)

		// Create a Service node with port info from YAML configs
		serviceID := serviceName
		svcProps := map[string]any{
			"id":       serviceID,
			"name":     serviceName,
			"language": "go",
		}
		parseServicePorts(svcPath, svcProps)
		result.Nodes.Services = append(result.Nodes.Services, svcProps)

		// Parse routes from handler/routes.go
		routes, routeRels := parseGoRoutes(svcPath, serviceName)
		result.Nodes.Routes = append(result.Nodes.Routes, routes...)
		result.Rels = append(result.Rels, routeRels...)

		// Parse API types from api/internal/types/types.go
		structs, fields, structRels := parseGoTypes(svcPath, serviceName)
		result.Nodes.Structs = append(result.Nodes.Structs, structs...)
		result.Nodes.Fields = append(result.Nodes.Fields, fields...)
		result.Rels = append(result.Rels, structRels...)

		// Parse model definitions from model/*_gen.go
		modelStructs, modelTables, modelColumns, modelRels := parseGoModels(svcPath, serviceName)
		result.Nodes.Structs = append(result.Nodes.Structs, modelStructs...)
		result.Nodes.Tables = append(result.Nodes.Tables, modelTables...)
		result.Nodes.Columns = append(result.Nodes.Columns, modelColumns...)
		result.Rels = append(result.Rels, modelRels...)
	}

	return result, nil
}



// parseGoRoutes extracts REST routes from a service's handler/routes.go.
// Handles rest.WithPrefix() by extracting it from each AddRoutes block.
func parseGoRoutes(svcPath, serviceName string) ([]map[string]any, []RelDef) {
	routesPath := filepath.Join(svcPath, "api", "internal", "handler", "routes.go")
	data, err := os.ReadFile(routesPath)
	if err != nil {
		return nil, nil
	}
	content := string(data)

	var routes []map[string]any
	var rels []RelDef

	// Find all server.AddRoutes( blocks with balanced paren matching
	addRoutesRe := regexp.MustCompile(`server\.AddRoutes\(`)
	addRoutesLocs := addRoutesRe.FindAllStringIndex(content, -1)

	seen := make(map[string]bool)

	for _, loc := range addRoutesLocs {
		start := loc[1] // position right after "(" in "server.AddRoutes("

		// Find matching closing parenthesis
		depth := 1
		pos := start
		for pos < len(content) && depth > 0 {
			switch content[pos] {
			case '(':
				depth++
			case ')':
				depth--
			}
			pos++
		}
		if depth != 0 {
			continue
		}
		blockContent := content[start : pos-1]

		// Extract prefix if present: rest.WithPrefix("/api/xxx")
		prefix := ""
		prefixRe := regexp.MustCompile(`rest\.WithPrefix\("([^"]+)"\)`)
		if prefixMatch := prefixRe.FindStringSubmatch(blockContent); len(prefixMatch) >= 2 {
			prefix = prefixMatch[1]
		}

		// Extract routes from this AddRoutes block
		routeRe := regexp.MustCompile(`Method:\s*(http\.\w+),\s*Path:\s*"([^"]+)"`)
		matches := routeRe.FindAllStringSubmatch(blockContent, -1)

		for _, m := range matches {
			method := strings.TrimPrefix(m[1], "http.Method")
			path := m[2]
			if prefix != "" {
				path = prefix + path
			}
			routeID := fmt.Sprintf("%s:%s:%s", serviceName, method, path)
			if seen[routeID] {
				continue
			}
			seen[routeID] = true

			routes = append(routes, map[string]any{
				"id":      routeID,
				"method":  method,
				"path":    path,
				"service": serviceName,
			})

			rels = append(rels, RelDef{
				FromID: serviceName,
				Type:   "EXPOSES",
				ToID:   routeID,
			})
		}
	}

	return routes, rels
}
// parseGoTypes extracts struct definitions from api/internal/types/types.go.
func parseGoTypes(svcPath, serviceName string) ([]map[string]any, []map[string]any, []RelDef) {
	typesPath := filepath.Join(svcPath, "api", "internal", "types", "types.go")
	data, err := os.ReadFile(typesPath)
	if err != nil {
		return nil, nil, nil
	}
	content := string(data)

	var structs []map[string]any
	var fields []map[string]any
	var rels []RelDef

	// Find all struct definitions: type Name struct {
	structRe := regexp.MustCompile(`(?m)^type\s+(\w+)\s+struct\s*\{`)
	structLocs := structRe.FindAllStringSubmatchIndex(content, -1)

	// Determine module path from go.mod
	modulePath := resolveGoModule(svcPath)
	pkgPrefix := modulePath

	for _, loc := range structLocs {
		structName := content[loc[2]:loc[3]]
		structStart := loc[1] // position after '{'

		// Find matching closing brace
		depth := 1
		pos := structStart
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
		structBlock := content[structStart : pos-1]

		structID := fmt.Sprintf("%s/%s", pkgPrefix, structName)

		structs = append(structs, map[string]any{
			"id":      structID,
			"name":    structName,
			"package": "types",
			"file":    fmt.Sprintf("services/%s/api/internal/types/types.go", serviceName),
		})

		// Parse fields within this struct
		structFields := parseGoStructFields(structBlock, structID, "types", structName)
		fields = append(fields, structFields...)
		for _, f := range structFields {
			rels = append(rels, RelDef{
				FromID: structID,
				Type:   "HAS_FIELD",
				ToID:   f["id"].(string),
			})
		}
	}

	return structs, fields, rels
}

// parseGoModels extracts model struct definitions from model/*_gen.go files.
// Returns: structs (GoStruct), tables (DbTable), columns (DbColumn), rels.
func parseGoModels(svcPath, serviceName string) ([]map[string]any, []map[string]any, []map[string]any, []RelDef) {
	modelDir := filepath.Join(svcPath, "model")
	entries, err := os.ReadDir(modelDir)
	if err != nil {
		return nil, nil, nil, nil
	}

	modulePath := resolveGoModule(svcPath)
	pkgPrefix := modulePath

	var structs []map[string]any
	var tables []map[string]any
	var columns []map[string]any
	var rels []RelDef

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || entry.Name() == "vars.go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		modelPath := filepath.Join(modelDir, entry.Name())
		data, err := os.ReadFile(modelPath)
		if err != nil {
			continue
		}
		content := string(data)

		// Find struct definitions - both standalone `type Xxx struct {` and inside `type (...)` blocks
		structRe := regexp.MustCompile(`(?m)^\s*type\s+(\w+)\s+struct\s*\{`)
		structLocs := structRe.FindAllStringSubmatchIndex(content, -1)

		// Also match indented structs inside `type (...)` blocks: `\t\tXxx struct {`
		indentedRe := regexp.MustCompile(`(?m)^\s{2,}(\w+)\s+struct\s*\{`)
		indentedLocs := indentedRe.FindAllStringSubmatchIndex(content, -1)
		structLocs = append(structLocs, indentedLocs...)

		for _, loc := range structLocs {
			structName := content[loc[2]:loc[3]]
			structStart := loc[1]

			// Find matching closing brace
			depth := 1
			pos := structStart
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
			structBlock := content[structStart : pos-1]

			// Only process structs with db:"..." tags
			if !strings.Contains(structBlock, `db:"`) {
				continue
			}

			structID := fmt.Sprintf("%s/model.%s", pkgPrefix, structName)

			// Determine table name from struct field or just use the one we found
			structTableName := extractStructTableName(content, structName)

			// Strip backticks from table name if present
			structTableName = strings.Trim(structTableName, "`")

			// Add GoStruct node for this model struct
			structs = append(structs, map[string]any{
				"id":      structID,
				"name":    structName,
				"package": "model",
				"file":    fmt.Sprintf("services/%s/model/%s", serviceName, entry.Name()),
			})

			// Add DbTable node
			tables = append(tables, map[string]any{
				"id":      structTableName,
				"name":    structTableName,
				"service": serviceName,
			})

			// MAPS_TO from GoStruct to DbTable
			rels = append(rels, RelDef{
				FromID: structID,
				Type:   "MAPS_TO",
				ToID:   structTableName,
			})

			// Parse columns from model struct fields
			modelColumns := parseGoModelFields(structBlock, structID, structTableName)
			columns = append(columns, modelColumns...)
			for _, c := range modelColumns {
				rels = append(rels, RelDef{
					FromID: structTableName,
					Type:   "HAS_COLUMN",
					ToID:   c["id"].(string),
				})
			}
		}
	}

	return structs, tables, columns, rels
}

// extractStructTableName extracts the table name associated with a struct.
func extractStructTableName(content, structName string) string {
	// Look for patterns like: table: "xxx" near the struct name
	re := regexp.MustCompile(`table:\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(content)
	if len(matches) >= 2 {
		return matches[1]
	}
	return strings.ToLower(structName)
}

// parseGoStructFields parses fields within a Go struct block for types.go.
func parseGoStructFields(block, structID, pkg, structName string) []map[string]any {
	var fields []map[string]any

	// Match: FieldName Type `json:"name,options"` or `form:"name,options"`
	fieldRe := regexp.MustCompile(`(?m)^\s+(\w+)\s+([^*]\S+?)\s+` + "`" + `([^` + "`" + `]+)` + "`")
	matches := fieldRe.FindAllStringSubmatch(block, -1)

	seen := make(map[string]bool)
	for _, m := range matches {
		fieldName := m[1]
		if seen[fieldName] {
			continue
		}
		seen[fieldName] = true

		fieldType := strings.TrimSpace(m[2])
		tags := m[3]

		jsonTag := extractGoTag(tags, "json")
		dbTag := extractGoTag(tags, "db")

		fieldID := fmt.Sprintf("%s.%s", structID, fieldName)
		fields = append(fields, map[string]any{
			"id":      fieldID,
			"name":    fieldName,
			"type":    fieldType,
			"jsonTag": jsonTag,
			"dbTag":   dbTag,
		})
	}

	return fields
}

// parseGoModelFields parses fields within a Go model struct block (with db tags).
func parseGoModelFields(block, structID, tableName string) []map[string]any {
	var fields []map[string]any

	fieldRe := regexp.MustCompile(`(?m)^\s+(\w+)\s+([^*]\S+?)\s+` + "`" + `([^` + "`" + `]+)` + "`")
	matches := fieldRe.FindAllStringSubmatch(block, -1)

	seen := make(map[string]bool)
	for _, m := range matches {
		fieldName := m[1]
		if seen[fieldName] {
			continue
		}
		seen[fieldName] = true

		fieldType := strings.TrimSpace(m[2])
		tags := m[3]

		dbTag := extractGoTag(tags, "db")
		if dbTag == "" {
			continue // not a DB column
		}

		fieldID := fmt.Sprintf("%s.%s", tableName, dbTag)
		fields = append(fields, map[string]any{
			"id":    fieldID,
			"name":  dbTag,
			"type":  mapGoTypeToDbType(fieldType),
			"table": tableName,
		})
	}

	return fields
}

// extractGoTag extracts a specific tag value from a Go struct tag string.
func extractGoTag(tags, key string) string {
	re := regexp.MustCompile(key + `:"([^"]*)"`)
	matches := re.FindStringSubmatch(tags)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// mapGoTypeToDbType maps Go types to approximate SQL column types.
func mapGoTypeToDbType(goType string) string {
	goType = strings.TrimPrefix(goType, "*") // pointer types
	switch goType {
	case "string":
		return "varchar"
	case "int64", "int", "int32":
		return "bigint"
	case "float64", "float32":
		return "decimal"
	case "bool":
		return "tinyint"
	case "time.Time":
		return "datetime"
	default:
		if strings.Contains(goType, "Null") || strings.Contains(goType, "sql.") {
			return "nullable"
		}
		if strings.Contains(goType, "Time") {
			return "datetime"
		}
		return "varchar"
	}
}

// resolveGoModule reads the go.mod file to determine the module path.
func resolveGoModule(svcPath string) string {
	goModPath := filepath.Join(svcPath, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Sprintf("github.com/guxiao1976/%s", filepath.Base(svcPath))
	}
	re := regexp.MustCompile(`^module\s+(\S+)`)
	matches := re.FindStringSubmatch(string(data))
	if len(matches) >= 2 {
		return matches[1]
	}
	return fmt.Sprintf("github.com/guxiao1976/%s", filepath.Base(svcPath))
}

// parseServicePorts scans YAML config files for port numbers and adds them to svcProps.
func parseServicePorts(svcPath string, svcProps map[string]any) {
	// Scan rpc config: services/*/rpc/etc/*.yaml — pattern ListenOn: 0.0.0.0:PORT
	rpcEtcDir := filepath.Join(svcPath, "rpc", "etc")
	if entries, err := os.ReadDir(rpcEtcDir); err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(rpcEtcDir, e.Name()))
			if err != nil {
				continue
			}
			re := regexp.MustCompile(`ListenOn:\s*[^:]+:(\d+)`)
			if m := re.FindStringSubmatch(string(data)); len(m) >= 2 {
				svcProps["port"] = m[1]
			}
		}
	}

	// Scan api config: services/*/api/etc/*.yaml — pattern Port: PORT
	apiEtcDir := filepath.Join(svcPath, "api", "etc")
	if entries, err := os.ReadDir(apiEtcDir); err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(apiEtcDir, e.Name()))
			if err != nil {
				continue
			}
			re := regexp.MustCompile(`Port:\s*(\d+)`)
			if m := re.FindStringSubmatch(string(data)); len(m) >= 2 {
				svcProps["apiPort"] = m[1]
			}
		}
	}
}
