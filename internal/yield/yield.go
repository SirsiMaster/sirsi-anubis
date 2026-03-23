// Package yield implements self-aware resource governance for Pantheon.
// Before running heavy operations (scans, dedup, ghost hunting), Pantheon
// checks system load and defers if the machine is already under pressure.
//
// ADR-006: "First, Do No Harm" — Pantheon tools MUST NOT make things worse.
package yield

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/SirsiMaster/sirsi-pantheon/internal/logging"
)

// SystemLoad represents the current system resource pressure.
type SystemLoad struct {
	// LoadAvg1 is the 1-minute load average.
	LoadAvg1 float64
	// LoadAvg5 is the 5-minute load average.
	LoadAvg5 float64
	// CPUCount is the number of logical CPUs.
	CPUCount int
	// LoadRatio is LoadAvg1 / CPUCount (>0.8 = heavy, >1.0 = overloaded).
	LoadRatio float64
	// Verdict is the recommendation: "healthy", "caution", "yield".
	Verdict string
}

// Severity levels for the system load.
const (
	VerdictHealthy = "healthy" // Load ratio < 0.6 — run freely
	VerdictCaution = "caution" // Load ratio 0.6-0.85 — warn but proceed
	VerdictYield   = "yield"   // Load ratio > 0.85 — defer heavy operations
)

// Check reads the current system load and returns a recommendation.
// This is lightweight (single syscall) and safe to run at any time.
func Check() (*SystemLoad, error) {
	cpus := runtime.NumCPU()

	load1, load5, err := getLoadAverage()
	if err != nil {
		return nil, fmt.Errorf("load average: %w", err)
	}

	ratio := load1 / float64(cpus)

	var verdict string
	switch {
	case ratio > 0.85:
		verdict = VerdictYield
	case ratio > 0.6:
		verdict = VerdictCaution
	default:
		verdict = VerdictHealthy
	}

	return &SystemLoad{
		LoadAvg1:  load1,
		LoadAvg5:  load5,
		CPUCount:  cpus,
		LoadRatio: ratio,
		Verdict:   verdict,
	}, nil
}

// ShouldYield returns true if the system is under enough pressure that
// heavy Pantheon operations should be deferred.
func ShouldYield() bool {
	load, err := Check()
	if err != nil {
		logging.Debug("yield check failed, proceeding", "error", err)
		return false // If we can't check, don't block
	}

	if load.Verdict == VerdictYield {
		logging.Warn("System under load, recommending yield",
			"load1", load.LoadAvg1,
			"cpus", load.CPUCount,
			"ratio", fmt.Sprintf("%.0f%%", load.LoadRatio*100),
		)
		return true
	}

	return false
}

// WarnIfHeavy prints a user-facing warning if the system is under caution
// or yield levels. Returns true if the command should abort (yield level).
// The --force flag overrides yield.
func WarnIfHeavy(commandName string, force bool) bool {
	load, err := Check()
	if err != nil {
		return false
	}

	switch load.Verdict {
	case VerdictYield:
		fmt.Fprintf(os.Stderr,
			"\n⚠️  System under heavy load (CPU: %.0f%% of capacity)\n"+
				"   Load average: %.1f across %d cores\n"+
				"   Running '%s' may degrade IDE and other tools.\n",
			load.LoadRatio*100, load.LoadAvg1, load.CPUCount, commandName)

		if force {
			fmt.Fprintf(os.Stderr, "   Proceeding anyway (--force).\n\n")
			return false
		}

		fmt.Fprintf(os.Stderr,
			"   💡 Recommendation: Close unused IDE conversations or restart your editor.\n"+
				"   Use --force to override this check.\n\n")
		return true

	case VerdictCaution:
		fmt.Fprintf(os.Stderr,
			"⚡ System moderately loaded (CPU: %.0f%% of capacity). Proceeding with '%s'.\n\n",
			load.LoadRatio*100, commandName)
		return false

	default:
		return false
	}
}

// getLoadAverage reads the 1-min and 5-min load averages.
func getLoadAverage() (float64, float64, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		return getLoadAverageUnix()
	default:
		// Windows: no simple load average; return 0 (don't block)
		return 0, 0, nil
	}
}

// getLoadAverageUnix uses sysctl on macOS or /proc/loadavg on Linux.
func getLoadAverageUnix() (float64, float64, error) {
	var out []byte
	var err error

	if runtime.GOOS == "darwin" {
		out, err = exec.Command("sysctl", "-n", "vm.loadavg").Output()
	} else {
		out, err = os.ReadFile("/proc/loadavg")
	}
	if err != nil {
		return 0, 0, err
	}

	// macOS: "{ 3.45 2.12 1.89 }"  Linux: "3.45 2.12 1.89 2/345 12345"
	s := strings.TrimSpace(string(out))
	s = strings.Trim(s, "{ }")
	fields := strings.Fields(s)
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("unexpected loadavg format: %q", s)
	}

	load1, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse load1: %w", err)
	}
	load5, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse load5: %w", err)
	}

	return load1, load5, nil
}
