// +build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := "rpc/internal/logic/user"
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("read dir error:", err)
		os.Exit(1)
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("read error:", path, err)
			continue
		}
		content := string(data)
		old1 := "newBaseResp()"
		new1 := "responsex.NewBaseResp()"
		old2 := "errResp("
		new2 := "responsex.NewBaseRespWithError("
		if strings.Contains(content, old1) || strings.Contains(content, old2) {
			content = strings.ReplaceAll(content, old1, new1)
			content = strings.ReplaceAll(content, old2, new2)
			err = os.WriteFile(path, []byte(content), 0644)
			if err != nil {
				fmt.Println("write error:", path, err)
			} else {
				fmt.Println("updated:", path)
			}
		}
	}
	fmt.Println("done")
}
