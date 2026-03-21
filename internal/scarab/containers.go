package scarab

import (
	"fmt"
	"os/exec"
	"strings"
)

// Container represents a Docker container.
type Container struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	Size    string `json:"size,omitempty"`
	Ports   string `json:"ports,omitempty"`
	Running bool   `json:"running"`
}

// ContainerAudit contains Docker/container scan results.
type ContainerAudit struct {
	Containers     []Container `json:"containers"`
	DanglingImages int         `json:"dangling_images"`
	StoppedCount   int         `json:"stopped_count"`
	RunningCount   int         `json:"running_count"`
	UnusedVolumes  int         `json:"unused_volumes"`
	DockerRunning  bool        `json:"docker_running"`
}

// AuditContainers scans the local Docker environment.
func AuditContainers() (*ContainerAudit, error) {
	audit := &ContainerAudit{}

	// Check Docker is running
	if err := exec.Command("docker", "info").Run(); err != nil {
		return audit, nil // Docker not running — not an error
	}
	audit.DockerRunning = true

	// List all containers
	out, err := exec.Command("docker", "ps", "-a",
		"--format", "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}").Output()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, "\t", 5)
			if len(parts) < 4 {
				continue
			}
			c := Container{
				ID:      parts[0],
				Name:    parts[1],
				Image:   parts[2],
				Status:  parts[3],
				Running: strings.HasPrefix(parts[3], "Up"),
			}
			if len(parts) >= 5 {
				c.Ports = parts[4]
			}
			if c.Running {
				audit.RunningCount++
			} else {
				audit.StoppedCount++
			}
			audit.Containers = append(audit.Containers, c)
		}
	}

	// Count dangling images
	out, err = exec.Command("docker", "images", "-f", "dangling=true", "-q").Output()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if line != "" {
				audit.DanglingImages++
			}
		}
	}

	// Count unused volumes
	out, err = exec.Command("docker", "volume", "ls", "-f", "dangling=true", "-q").Output()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if line != "" {
				audit.UnusedVolumes++
			}
		}
	}

	return audit, nil
}

// FormatContainerStatus returns a styled status string.
func FormatContainerStatus(c Container) string {
	if c.Running {
		return fmt.Sprintf("🟢 %s", c.Status)
	}
	return fmt.Sprintf("🔴 %s", c.Status)
}
