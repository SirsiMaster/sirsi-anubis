package scarab

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// Injectable command runners for testability (B11 pattern).
var (
	runDockerInfo = func() error { return exec.Command("docker", "info").Run() }
	runDockerPS   = func() ([]byte, error) {
		return exec.Command("docker", "ps", "-a", "--format", "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}").Output()
	}
	runDockerImages = func() ([]byte, error) { return exec.Command("docker", "images", "-f", "dangling=true", "-q").Output() }
	runDockerVols   = func() ([]byte, error) {
		return exec.Command("docker", "volume", "ls", "-f", "dangling=true", "-q").Output()
	}
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
// Docker PS, Images, and Volumes run concurrently on dedicated OS threads.
func AuditContainers() (*ContainerAudit, error) {
	audit := &ContainerAudit{}

	// Check Docker is running
	if err := runDockerInfo(); err != nil {
		return audit, nil // Docker not running — not an error
	}
	audit.DockerRunning = true

	// Run all Docker queries concurrently on dedicated threads.
	var psOut, imgOut, volOut []byte
	var psErr, imgErr, volErr error
	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		psOut, psErr = runDockerPS()
	}()
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		imgOut, imgErr = runDockerImages()
	}()
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		volOut, volErr = runDockerVols()
	}()
	wg.Wait()

	// Process results
	if psErr == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(psOut)), "\n") {
			c := splitContainerLine(line)
			if c == nil {
				continue
			}
			if c.Running {
				audit.RunningCount++
			} else {
				audit.StoppedCount++
			}
			audit.Containers = append(audit.Containers, *c)
		}
	}

	if imgErr == nil {
		audit.DanglingImages = countNonEmptyLines(strings.TrimSpace(string(imgOut)))
	}

	if volErr == nil {
		audit.UnusedVolumes = countNonEmptyLines(strings.TrimSpace(string(volOut)))
	}

	return audit, nil
}

// splitContainerLine parses a single tab-delimited docker ps output line.
func splitContainerLine(line string) *Container {
	if line == "" {
		return nil
	}
	parts := strings.SplitN(line, "\t", 5)
	if len(parts) < 4 {
		return nil
	}
	c := &Container{
		ID:      parts[0],
		Name:    parts[1],
		Image:   parts[2],
		Status:  parts[3],
		Running: strings.HasPrefix(parts[3], "Up"),
	}
	if len(parts) >= 5 {
		c.Ports = strings.TrimSpace(parts[4])
	}
	return c
}

// countNonEmptyLines counts non-blank lines in a string.
func countNonEmptyLines(s string) int {
	if s == "" {
		return 0
	}
	count := 0
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

// FormatContainerStatus returns a styled status string.
func FormatContainerStatus(c Container) string {
	if c.Running {
		return fmt.Sprintf("🟢 %s", c.Status)
	}
	return fmt.Sprintf("🔴 %s", c.Status)
}
