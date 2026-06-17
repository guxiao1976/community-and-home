package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := "rpc/internal/logic/user"
	importLine := "\t\"github.com/guxiao1976/community-common/v2/pkg/responsex\""

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") || e.Name() == "helpers_test.go" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, _ := os.ReadFile(path)
		content := string(data)

		if !strings.Contains(content, "responsex.NewBaseResp") {
			continue
		}
		if strings.Contains(content, "community-common/v2/pkg/responsex") {
			continue
		}

		// Add before closing paren of import
		// Strategy: find "\n)" after "import (", replace with importLine + "\n)"
		importIdx := strings.Index(content, "import (")
		if importIdx < 0 {
			fmt.Println("no import block:", path)
			continue
		}
		closeIdx := strings.Index(content[importIdx:], "\n)")
		if closeIdx < 0 {
			fmt.Println("no close paren:", path)
			continue
		}
		closeIdx += importIdx
		newContent := content[:closeIdx] + "\n" + importLine + content[closeIdx:]
		os.WriteFile(path, []byte(newContent), 0644)
		fmt.Println("import added:", path)
	}
	fmt.Println("done")
}
