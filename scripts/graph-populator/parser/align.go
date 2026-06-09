package parser

import (
	"fmt"
	"strings"
)

// AlignCrossSource creates relationships across parsed sources:
// - Proto -> Go code generation (ProtoMessage GENERATES GoStruct, ProtoField GENERATES GoField)
// - Go JSON serialization alignment (GoField -> TsField via json tags)
// - Go service -> Proto RPC implementation
func AlignCrossSource(protoRes *ParseProtosResult, goRes *ParseGoSourcesResult, tsRes *ParseTypeScriptResult) []RelDef {
	var rels []RelDef

	// 1. Proto -> Go: Match proto message/field names to Go struct/field names
	rels = append(rels, alignProtoToGo(protoRes, goRes)...)

	// 2. Go JSON -> TypeScript: Match Go json tags to TS field names
	rels = append(rels, alignGoToTs(goRes, tsRes)...)

	// 3. Service -> ProtoRPC: Match service names to proto RPC implementations
	rels = append(rels, alignServiceToProtoRpc(goRes, protoRes)...)

	// 4. ApiCall -> ApiRoute: Match frontend API calls to backend routes by URL+method
	rels = append(rels, alignApiCallToRoute(goRes, tsRes)...)

	return rels
}

// alignProtoToGo creates GENERATES relationships between proto messages/fields and Go structs/fields.
func alignProtoToGo(protoRes *ParseProtosResult, goRes *ParseGoSourcesResult) []RelDef {
	var rels []RelDef

	// Build a map of Go struct name -> struct ID for matching
	goStructMap := make(map[string]string) // lowercased name -> id
	for _, s := range goRes.Nodes.Structs {
		name, ok := s["name"].(string)
		if ok {
			goStructMap[strings.ToLower(name)] = s["id"].(string)
		}
	}

	// Build a map of Go field name -> field info for matching
	goFieldMap := make(map[string][]map[string]any) // lowercased struct name -> fields
	for _, f := range goRes.Nodes.Fields {
		fieldID := f["id"].(string)
		fieldName := f["name"].(string)
		// Extract struct name from field ID like "module.StructName.FieldName"
		parts := strings.Split(fieldID, ".")
		if len(parts) >= 2 {
			structName := strings.ToLower(parts[len(parts)-2])
			goFieldMap[structName] = append(goFieldMap[structName], map[string]any{
				"name": fieldName,
				"id":   fieldID,
			})
		}
	}

	// Match proto messages -> Go structs by name (case-insensitive)
	for _, msg := range protoRes.Nodes.Messages {
		msgName, _ := msg["name"].(string)
		msgID, _ := msg["id"].(string)

		if goStructID, ok := goStructMap[strings.ToLower(msgName)]; ok {
			rels = append(rels, RelDef{
				FromID: msgID,
				Type:   "GENERATES",
				ToID:   goStructID,
			})
		}
	}

	// Match proto fields -> Go fields within matched structs
	for _, pf := range protoRes.Nodes.Fields {
		fieldID, _ := pf["id"].(string)
		fieldName, _ := pf["name"].(string)
		// Proto field ID format: "pkg.MessageName.field"
		parts := strings.Split(fieldID, ".")
		if len(parts) < 3 {
			continue
		}
		msgName := strings.ToLower(parts[len(parts)-2])

		if goFields, ok := goFieldMap[msgName]; ok {
			for _, gf := range goFields {
				if strings.EqualFold(gf["name"].(string), fieldName) {
					rels = append(rels, RelDef{
						FromID: fieldID,
						Type:   "GENERATES",
						ToID:   gf["id"].(string),
					})
				}
			}
		}
	}

	return rels
}

// alignGoToTs creates JSON_SERIALIZED_AS relationships between Go struct fields and TS interface fields.
func alignGoToTs(goRes *ParseGoSourcesResult, tsRes *ParseTypeScriptResult) []RelDef {
	var rels []RelDef

	// Build TS field index by field name
	tsFieldIndex := make(map[string][]map[string]any) // field name -> list of ts field infos
	for _, tf := range tsRes.Nodes.Fields {
		name, _ := tf["name"].(string)
		tsFieldIndex[strings.ToLower(name)] = append(tsFieldIndex[strings.ToLower(name)], tf)
	}

	for _, gf := range goRes.Nodes.Fields {
		jsonTag, _ := gf["jsonTag"].(string)
		if jsonTag == "" {
			continue
		}
		// Get first part of json tag (before comma)
		jsonKey := strings.Split(jsonTag, ",")[0]
		if jsonKey == "" || jsonKey == "-" {
			continue
		}

		fieldID, _ := gf["id"].(string)

		// Try matching by json tag key
		if tsFields, ok := tsFieldIndex[strings.ToLower(jsonKey)]; ok {
			for _, tf := range tsFields {
				rels = append(rels, RelDef{
					FromID: fieldID,
					Type:   "JSON_SERIALIZED_AS",
					ToID:   tf["id"].(string),
				})
			}
		}
	}

	return rels
}

// alignServiceToProtoRpc creates IMPLEMENTS relationships between services and proto RPCs.
func alignServiceToProtoRpc(goRes *ParseGoSourcesResult, protoRes *ParseProtosResult) []RelDef {
	var rels []RelDef

	// Map service names from directory names
	serviceNames := make(map[string]string)
	for _, svc := range goRes.Nodes.Services {
		id, _ := svc["id"].(string)
		name, _ := svc["name"].(string)
		if name == "" {
			name = id
		}
		serviceNames[strings.ToLower(name)] = id
	}

	// Map proto packages to service names
	packageToService := map[string]string{
		"user.v1":       "user-service",
		"auth.v1":       "auth-service",
		"permission.v1": "permission-service",
		"masterdata.v1": "master-data-service",
		"file.v1":       "file-service",
		"moderation.v1": "moderation-service",
		"aimodel.v1":    "ai-model-service",
		"common.v1":     "common",
		"community.v1":  "community-hub-service",
	}

	for _, rpc := range protoRes.Nodes.Rpcs {
		rpcID, _ := rpc["id"].(string)
		pkg, _ := rpc["package"].(string)

		if svcName, ok := packageToService[pkg]; ok {
			if svcID, ok := serviceNames[svcName]; ok {
				rels = append(rels, RelDef{
					FromID: svcID,
					Type:   "IMPLEMENTS",
					ToID:   rpcID,
				})
			}
		}
	}

	return rels
}

// alignApiCallToRoute creates PROXIES_TO relationships between frontend API calls and backend REST routes.
func alignApiCallToRoute(goRes *ParseGoSourcesResult, tsRes *ParseTypeScriptResult) []RelDef {
	var rels []RelDef

	// Build route index: method+path -> route ID
	routeIndex := make(map[string]string)
	for _, r := range goRes.Nodes.Routes {
		method, _ := r["method"].(string)
		path, _ := r["path"].(string)
		routeID, _ := r["id"].(string)
		key := fmt.Sprintf("%s:%s", method, path)
		routeIndex[key] = routeID
	}

	for _, apiCall := range tsRes.Nodes.ApiCalls {
		method, _ := apiCall["method"].(string)
		url, _ := apiCall["url"].(string)
		apiCallID, _ := apiCall["id"].(string)

		// Try exact match first
		key := fmt.Sprintf("%s:%s", method, url)
		if routeID, ok := routeIndex[key]; ok {
			rels = append(rels, RelDef{
				FromID: apiCallID,
				Type:   "PROXIES_TO",
				ToID:   routeID,
			})
			continue
		}

		// Try matching just by path (any method)
		for routeKey, routeID := range routeIndex {
			routeParts := strings.SplitN(routeKey, ":", 2)
			if len(routeParts) == 2 && routeParts[1] == url {
				rels = append(rels, RelDef{
					FromID: apiCallID,
					Type:   "PROXIES_TO",
					ToID:   routeID,
				})
				break
			}
		}
	}

	return rels
}
