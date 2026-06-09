package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProtoNodes holds all node types parsed from proto files.
type ProtoNodes struct {
	Messages []map[string]any
	Fields   []map[string]any
	Rpcs     []map[string]any
}

// ParseProtosResult groups parsed proto nodes and relationships.
type ParseProtosResult struct {
	Nodes ProtoNodes
	Rels  []RelDef
}

// ParseProtos scans api-proto/api/ for .proto files and returns all parsed nodes and relationships.
func ParseProtos(projectRoot string) (*ParseProtosResult, error) {
	apiProtoDir := filepath.Join(projectRoot, "api-proto", "api")
	result := &ParseProtosResult{}

	// Track message -> Rpc usage for cross-references
	type rpcUsage struct {
		rpcName string
		input   string
		output  string
	}
	var rpcUsages []rpcUsage

	err := filepath.Walk(apiProtoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".proto") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read proto file %s: %w", path, err)
		}
		content := string(data)

		relPath, _ := filepath.Rel(apiProtoDir, path)
		relPath = filepath.ToSlash(relPath)
		pkg := extractProtoPackage(content)
		serviceName := extractProtoServiceName(content)

		// Parse messages
		messages := parseProtoMessages(content, pkg, relPath)
		for _, msg := range messages {
			msgID := fmt.Sprintf("%s.%s", pkg, msg["name"])
			msg["id"] = msgID

			msgName := msg["name"].(string)
			msgIdx := msgName // original name (without full package)

			// Parse fields within each message
			fields := parseProtoFields(content, pkg, msgName, relPath)
			result.Nodes.Fields = append(result.Nodes.Fields, fields...)

			// Add HAS_FIELD relationships
			for _, f := range fields {
				result.Rels = append(result.Rels, RelDef{
					FromID: msgID,
					Type:   "HAS_FIELD",
					ToID:   f["id"].(string),
				})
			}

			result.Nodes.Messages = append(result.Nodes.Messages, msg)

			// Track message name for Rpc usage lookup
			_ = msgIdx
		}

		// Parse RPCs
		rpcs := parseProtoRpcs(content, pkg, serviceName, relPath)
		for _, rpc := range rpcs {
			rpcID := fmt.Sprintf("%s.%s.%s", pkg, serviceName, rpc["name"])
			rpc["id"] = rpcID
			result.Nodes.Rpcs = append(result.Nodes.Rpcs, rpc)

			inputType := rpc["inputType"].(string)
			outputType := rpc["outputType"].(string)

			result.Rels = append(result.Rels, RelDef{
				FromID: rpcID,
				Type:   "USES",
				ToID:   fmt.Sprintf("%s.%s", pkg, inputType),
			})
			result.Rels = append(result.Rels, RelDef{
				FromID: rpcID,
				Type:   "USES",
				ToID:   fmt.Sprintf("%s.%s", pkg, outputType),
			})

			rpcUsages = append(rpcUsages, rpcUsage{
				rpcName: rpcID,
				input:   inputType,
				output:  outputType,
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk proto dir: %w", err)
	}

	return result, nil
}

// extractProtoPackage extracts the package declaration from proto content.
func extractProtoPackage(content string) string {
	re := regexp.MustCompile(`package\s+([a-zA-Z0-9_.]+)\s*;`)
	matches := re.FindStringSubmatch(content)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// extractProtoServiceName extracts the service name from proto content.
func extractProtoServiceName(content string) string {
	re := regexp.MustCompile(`service\s+(\w+)\s*\{`)
	matches := re.FindStringSubmatch(content)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// parseProtoMessages extracts message definitions from proto content.
func parseProtoMessages(content, pkg, relPath string) []map[string]any {
	var messages []map[string]any

	// Match message definitions: message Xxx {
	msgRe := regexp.MustCompile(`(?m)^\s*message\s+(\w+)\s*\{`)
	matches := msgRe.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)
	for _, m := range matches {
		name := m[1]
		if seen[name] {
			continue
		}
		seen[name] = true
		messages = append(messages, map[string]any{
			"name":      name,
			"package":   pkg,
			"protoFile": relPath,
		})
	}
	return messages
}

// parseProtoFields extracts field definitions from a specific message.
// It scans the content for `type fieldName = fieldNumber;` patterns.
func parseProtoFields(content, pkg, msgName, relPath string) []map[string]any {
	var fields []map[string]any

	// Find the message block boundaries
	msgBlock := extractMessageBlock(content, msgName)
	if msgBlock == "" {
		return fields
	}

	// Pattern: <type> <name> = <number> [options];
	// Types include: string, int64, int32, bool, uint64, sint32, fixed32, sfixed32, etc.
	fieldRe := regexp.MustCompile(`(?m)^\s+(repeated\s+|optional\s+)?([a-zA-Z0-9_.]+)\s+(\w+)\s*=\s*(\d+)`)
	fieldMatches := fieldRe.FindAllStringSubmatch(msgBlock, -1)

	seen := make(map[string]bool)
	for _, m := range fieldMatches {
		fullType := m[2]
		fieldName := m[3]
		fieldNum := m[4]

		if seen[fieldName] {
			continue
		}
		seen[fieldName] = true

		fieldID := fmt.Sprintf("%s.%s.%s", pkg, msgName, fieldName)
		fields = append(fields, map[string]any{
			"id":     fieldID,
			"name":   fieldName,
			"type":   fullType,
			"number": fieldNum,
		})
	}

	return fields
}

// extractMessageBlock extracts the content within a message block for a given message name.
func extractMessageBlock(content, msgName string) string {
	// Find start of message block
	re := regexp.MustCompile(`(?m)^\s*message\s+` + regexp.QuoteMeta(msgName) + `\s*\{`)
	loc := re.FindStringIndex(content)
	if loc == nil {
		return ""
	}

	start := loc[1] // position right after '{'
	depth := 1
	pos := start
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
		return ""
	}
	return content[start : pos-1]
}

// parseProtoRpcs extracts RPC method definitions from proto content.
func parseProtoRpcs(content, pkg, serviceName, relPath string) []map[string]any {
	var rpcs []map[string]any

	if serviceName == "" {
		return rpcs
	}

	// Pattern: rpc MethodName(InputType) returns (OutputType);
	rpcRe := regexp.MustCompile(`rpc\s+(\w+)\s*\(\s*([a-zA-Z0-9_.]+)\s*\)\s*returns\s*\(\s*([a-zA-Z0-9_.]+)\s*\)`)
	matches := rpcRe.FindAllStringSubmatch(content, -1)

	seen := make(map[string]bool)
	for _, m := range matches {
		name := m[1]
		if seen[name] {
			continue
		}
		seen[name] = true
		rpcs = append(rpcs, map[string]any{
			"name":       name,
			"service":    serviceName,
			"package":    pkg,
			"inputType":  m[2],
			"outputType": m[3],
		})
	}

	return rpcs
}
