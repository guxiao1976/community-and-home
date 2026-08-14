package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// InfraNodes holds all node types parsed from infra config files.
type InfraNodes struct {
	Services []map[string]any
}

// ParseInfraResult groups parsed infra nodes and relationships.
type ParseInfraResult struct {
	Nodes InfraNodes
	Rels  []RelDef
}

// ParseInfra scans docker-compose.yml, go.mod files, and vite.config.ts for infrastructure metadata.
func ParseInfra(projectRoot string) (*ParseInfraResult, error) {
	result := &ParseInfraResult{}

	// 1. Parse docker-compose.yml for middleware services
	dockerServices, err := parseDockerCompose(filepath.Join(projectRoot, "docker-compose.yml"))
	if err == nil {
		result.Nodes.Services = append(result.Nodes.Services, dockerServices...)
	}

	// 2. Parse vite.config.ts for service proxy mappings (service -> port)
	viteServices, viteRels, err := parseViteConfig(filepath.Join(projectRoot, "web", "pc", "vite.config.ts"))
	if err == nil {
		result.Nodes.Services = append(result.Nodes.Services, viteServices...)
		result.Rels = append(result.Rels, viteRels...)
	}

	// 3. Parse go.mod files for module paths and dependencies
	goModServices, goModRels, err := parseGoMods(projectRoot)
	if err == nil {
		// Merge into existing services, preferring first-seen
		seen := make(map[string]bool)
		for _, s := range result.Nodes.Services {
			if id, ok := s["id"]; ok {
				seen[id.(string)] = true
			}
		}
		for _, s := range goModServices {
			if id, ok := s["id"]; ok && !seen[id.(string)] {
				result.Nodes.Services = append(result.Nodes.Services, s)
				seen[id.(string)] = true
			}
		}
		result.Rels = append(result.Rels, goModRels...)
	}

	// 4. Detect cross-service dependencies from Go source file proto imports
	protoDeps, err := detectProtoDependencies(projectRoot)
	if err == nil {
		result.Rels = append(result.Rels, protoDeps...)
	}

	return result, nil
}

// parseDockerCompose extracts service info from docker-compose.yml.
func parseDockerCompose(path string) ([]map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(data)

	var services []map[string]any

	// Find service blocks: ^  service_name:
	// We look for top-level service entries (indented by 2 spaces, ending with colon)
	// that are NOT x- prefixed (those are anchors)
	re := regexp.MustCompile(`(?m)^\s{2}([a-zA-Z][a-zA-Z0-9_-]+):\s*$`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, m := range matches {
		svcName := m[1]
		if strings.HasPrefix(svcName, "x-") {
			continue
		}

		svc := map[string]any{
			"id":       fmt.Sprintf("middleware-%s", svcName),
			"name":     svcName,
			"language": "infra",
		}

		// Extract port
		portRe := regexp.MustCompile(svcName + `:\s*\n(?s:.*?)ports:\s*\["(\d+):`)
		if portMatch := portRe.FindStringSubmatch(content); len(portMatch) >= 2 {
			svc["port"] = portMatch[1]
		}

		// Extract IP
		ipRe := regexp.MustCompile(svcName + `:\s*\n(?s:.*?)ipv4_address:\s*([\d.]+)`)
		if ipMatch := ipRe.FindStringSubmatch(content); len(ipMatch) >= 2 {
			svc["ip"] = ipMatch[1]
		}

		services = append(services, svc)
	}

	return services, nil
}

// parseViteConfig extracts proxy mappings from vite.config.ts.
func parseViteConfig(path string) ([]map[string]any, []RelDef, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	content := string(data)

	var services []map[string]any
	var rels []RelDef

	// Pattern: '/api/xxx': { target: 'http://127.0.0.1:PORT', ... }
	proxyRe := regexp.MustCompile(`'/(api/[^']+)':\s*\{\s*target:\s*'http://127\.0\.0\.1:(\d+)'`)
	matches := proxyRe.FindAllStringSubmatch(content, -1)

	serviceNames := map[string]string{
		"8881": "auth-service",
		"8882": "user-service",
		"8883": "permission-service",
		"8884": "file-service",
		"8886": "monitoring-service",
		"8889": "master-data-service",
		"8890": "moderation-service",
		"8891": "ai-model-service",
	}

	for _, m := range matches {
		apiPath := m[1]
		port := m[2]
		svcName := serviceNames[port]
		if svcName == "" {
			continue
		}

		// Add proxy mapping relationship: api path -> service
		rels = append(rels, RelDef{
			FromID: fmt.Sprintf("vite-proxy:%s", apiPath),
			Type:   "PROXIES_TO",
			ToID:   svcName,
		})
	}

	return services, rels, nil
}

// discoverServices scans services/*/go.mod for module paths -> service names and service dir names.
// 替代硬编码清单（此前缺 monitoring-service 等导致图数据不完整）；新增服务自动覆盖。
func discoverServices(projectRoot string) (map[string]string, []string) {
	known := make(map[string]string)
	var dirs []string
	servicesDir := filepath.Join(projectRoot, "services")
	entries, err := os.ReadDir(servicesDir)
	if err != nil {
		return known, dirs
	}
	modRe := regexp.MustCompile(`^module\s+(\S+)`)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		dirs = append(dirs, name)
		data, err := os.ReadFile(filepath.Join(servicesDir, name, "go.mod"))
		if err != nil {
			continue
		}
		if m := modRe.FindStringSubmatch(string(data)); m != nil {
			known[m[1]] = name
		}
	}
	return known, dirs
}

// parseGoMods scans services/*/go.mod and the root go.mod for module paths and dependencies.
func parseGoMods(projectRoot string) ([]map[string]any, []RelDef, error) {
	var services []map[string]any
	var rels []RelDef

	// 自动发现：扫描 services/*/go.mod（替代硬编码 knownServices + svcDirs，消除新增服务漂移）
	knownServices, svcDirs := discoverServices(projectRoot)
	// 补充非 services/ 的项目模块
	knownServices["github.com/guxiao1976/community-common/v2"] = "common"
	knownServices["github.com/guxiao1976/api-proto"] = "api-proto"

	for _, svcDir := range svcDirs {
		goModPath := filepath.Join(projectRoot, "services", svcDir, "go.mod")
		data, err := os.ReadFile(goModPath)
		if err != nil {
			continue
		}
		content := string(data)

		// Extract module path
		modRe := regexp.MustCompile(`^module\s+(\S+)`)
		modMatch := modRe.FindStringSubmatch(content)
		if modMatch == nil {
			continue
		}
		modPath := modMatch[1]

		svcName := svcDir
		if name, ok := knownServices[modPath]; ok {
			svcName = name
		}

		services = append(services, map[string]any{
			"id":       svcName,
			"name":     svcName,
			"module":   modPath,
			"language": "go",
		})

		// Extract replace directives to find project-local dependencies
		replaceRe := regexp.MustCompile(`replace\s+(\S+)\s*=>\s*(\.\./\S+)`)
		replaceMatches := replaceRe.FindAllStringSubmatch(content, -1)

		for _, rm := range replaceMatches {
			depMod := rm[1]
			if depSvcName, ok := knownServices[depMod]; ok {
				rels = append(rels, RelDef{
					FromID: svcName,
					Type:   "DEPENDS_ON",
					ToID:   depSvcName,
				})
			}
		}

		// Also extract require directives for direct project dependencies
		extractGoRequireDirectDeps(content, knownServices, svcName, &rels)
	}

	return services, rels, nil
}

// extractGoRequireDirectDeps finds direct require dependencies that are project services.
func extractGoRequireDirectDeps(content string, knownServices map[string]string, svcName string, rels *[]RelDef) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	inRequire := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "require (") {
			inRequire = true
			continue
		}
		if inRequire {
			if line == ")" {
				break
			}
			// Parse lines like: github.com/guxiao1976/api-proto v0.0.0
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				modPath := parts[0]
				// Skip indirect deps and std lib
				if len(parts) >= 3 && parts[len(parts)-1] == "// indirect" {
					continue
				}
				if depSvcName, ok := knownServices[modPath]; ok && depSvcName != svcName {
					// Check we haven't already added this relationship
					alreadyExists := false
					for _, r := range *rels {
						if r.FromID == svcName && r.ToID == depSvcName && r.Type == "DEPENDS_ON" {
							alreadyExists = true
							break
						}
					}
					if !alreadyExists {
						*rels = append(*rels, RelDef{
							FromID: svcName,
							Type:   "DEPENDS_ON",
							ToID:   depSvcName,
						})
					}
				}
			}
		}
	}
}

// detectProtoDependencies scans Go source files for proto imports and creates DEPENDS_ON relationships.
func detectProtoDependencies(projectRoot string) ([]RelDef, error) {
	var rels []RelDef

	// Map proto package to service name
	protoPkgToService := map[string]string{
		"user":       "user-service",
		"auth":       "auth-service",
		"permission": "permission-service",
		"masterdata": "master-data-service",
		"aimodel":    "ai-model-service",
		"file":       "file-service",
		"moderation": "moderation-service",
		"community":  "community-hub-service",
	}

	// Map service name to its directory name (with prefix mapping)
	serviceNameToDir := map[string]string{
		"user-service":          "user-service",
		"auth-service":          "auth-service",
		"permission-service":    "permission-service",
		"master-data-service":   "master-data-service",
		"ai-model-service":      "ai-model-service",
		"file-service":          "file-service",
		"moderation-service":    "moderation-service",
		"community-hub-service": "community-hub-service",
	}

	servicesDir := filepath.Join(projectRoot, "services")
	serviceDirs, err := os.ReadDir(servicesDir)
	if err != nil {
		return nil, err
	}

	// Proto import pattern: "github.com/guxiao1976/api-proto/gen/go/<pkg>/v1"
	protoImportRe := regexp.MustCompile(`github\.com/guxiao1976/api-proto/gen/go/([a-z]+)/v1`)

	for _, dir := range serviceDirs {
		if !dir.IsDir() {
			continue
		}
		svcName := dir.Name()
		svcDir := filepath.Join(servicesDir, svcName)

		// Track which proto packages are imported by this service
		importedProtos := make(map[string]bool)

		// Walk through all .go files in this service directory
		filepath.Walk(svcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(info.Name(), ".go") {
				return nil
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			matches := protoImportRe.FindAllStringSubmatch(string(data), -1)
			for _, m := range matches {
				importedProtos[m[1]] = true
			}
			return nil
		})

		for pkg := range importedProtos {
			depSvcName, ok := protoPkgToService[pkg]
			if !ok || depSvcName == svcName {
				continue // skip self-dependency and unknown pkgs
			}
			rels = append(rels, RelDef{
				FromID: svcName,
				Type:   "DEPENDS_ON",
				ToID:   depSvcName,
			})
		}

		_ = serviceNameToDir // kept for future use
	}

	// Deduplicate
	seen := make(map[string]bool)
	var deduped []RelDef
	for _, r := range rels {
		key := r.FromID + "|" + r.ToID
		if seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, r)
	}

	return deduped, nil
}
