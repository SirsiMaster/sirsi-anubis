// Package vitals provides lightweight system stats collection shared
// across TUI, menubar, and dashboard surfaces.
package vitals

import (
	"fmt"
	"os/exec"
	"strings"
)

// Snapshot holds a point-in-time collection of system vitals.
type Snapshot struct {
	RAMPercent  float64
	RAMPressure string // "low", "medium", "high"
	RAMIcon     string

	GitBranch   string
	Uncommitted int
	LastCommit  string // human-friendly relative time ("2m", "1h", etc.)

	Accelerator string // short label ("M5 Max", "Intel", etc.)
}

// Collect gathers a fresh vitals snapshot. Designed to be fast (< 200ms).
func Collect() Snapshot {
	var s Snapshot
	collectRAM(&s)
	collectGit(&s)
	collectAccelerator(&s)
	return s
}

func collectRAM(s *Snapshot) {
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return
	}
	var total int64
	_, _ = fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &total)
	if total == 0 {
		return
	}

	vmOut, err := exec.Command("vm_stat").Output()
	if err != nil {
		return
	}

	var active, wired int64
	for _, line := range strings.Split(string(vmOut), "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		val := strings.TrimSpace(strings.TrimSuffix(parts[1], "."))
		var n int64
		_, _ = fmt.Sscanf(val, "%d", &n)
		n *= 4096 // pages to bytes (macOS default page size)
		switch {
		case strings.Contains(parts[0], "Pages active"):
			active = n
		case strings.Contains(parts[0], "Pages wired"):
			wired = n
		}
	}

	used := active + wired
	s.RAMPercent = float64(used) / float64(total) * 100

	switch {
	case s.RAMPercent > 85:
		s.RAMPressure = "high"
		s.RAMIcon = "🔴"
	case s.RAMPercent > 65:
		s.RAMPressure = "medium"
		s.RAMIcon = "🟡"
	default:
		s.RAMPressure = "low"
		s.RAMIcon = "🟢"
	}
}

func collectGit(s *Snapshot) {
	if out, err := exec.Command("git", "branch", "--show-current").Output(); err == nil {
		s.GitBranch = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "status", "--porcelain").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) == 1 && lines[0] == "" {
			s.Uncommitted = 0
		} else {
			s.Uncommitted = len(lines)
		}
	}
	if out, err := exec.Command("git", "log", "-1", "--format=%cr").Output(); err == nil {
		s.LastCommit = strings.TrimSpace(string(out))
	}
}

func collectAccelerator(s *Snapshot) {
	out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output()
	if err != nil {
		return
	}
	brand := strings.TrimSpace(string(out))
	s.Accelerator = brand
	// Shorten Apple Silicon to just the chip name
	if strings.Contains(brand, "Apple") {
		parts := strings.Fields(brand)
		if len(parts) >= 2 {
			s.Accelerator = strings.Join(parts[1:], " ") // "M5 Max"
		}
	}
}
