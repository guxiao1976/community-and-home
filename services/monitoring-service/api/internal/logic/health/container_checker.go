package health

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/guxiao1976/community-monitoring/api/internal/config"
	"github.com/guxiao1976/community-monitoring/api/internal/types"
)

type dockerPSLine struct {
	Names      string `json:"Names"`
	Status     string `json:"Status"`
	RunningFor string `json:"RunningFor"`
	Image      string `json:"Image"`
}

func checkContainers(cfg config.MonitoringConfig) []types.ContainerHealth {
	running := make(map[string]dockerPSLine)
	cmd := exec.Command("docker", "ps", "--format", "{{json .}}", "--all")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			var c dockerPSLine
			if json.Unmarshal([]byte(line), &c) == nil {
				running[c.Names] = c
			}
		}
	}

	results := make([]types.ContainerHealth, 0, len(cfg.Containers))
	for _, want := range cfg.Containers {
		dockerErr := ""
		if err != nil {
			dockerErr = err.Error()
		}

		if c, ok := running[want.Name]; ok {
			status := "unhealthy"
			if strings.Contains(c.Status, "Up") {
				status = "healthy"
			}
			results = append(results, types.ContainerHealth{
				Name:        want.Name,
				DisplayName: want.DisplayName,
				Image:       c.Image,
				Status:      status,
				RunningFor:  c.RunningFor,
				Error:       dockerErr,
			})
		} else {
			results = append(results, types.ContainerHealth{
				Name:        want.Name,
				DisplayName: want.DisplayName,
				Image:       "",
				Status:      "unhealthy",
				RunningFor:  "",
				Error:       dockerErr,
			})
		}
	}
	return results
}
