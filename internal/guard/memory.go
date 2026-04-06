package guard

import (
	"runtime"

	"github.com/SirsiMaster/sirsi-pantheon/internal/seba"
)

// ResourceStats contains memory and system metrics for the Anubis dashboard.
type ResourceStats struct {
	UsedMemory    string
	TotalMemory   string
	PressureLevel string
}

// GetStats returns the current workstation resource utilization.
func GetStats() (*ResourceStats, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	profile, _ := seba.DetectHardware()

	return &ResourceStats{
		UsedMemory:    seba.FormatBytes(int64(m.Alloc)),
		TotalMemory:   seba.FormatBytes(profile.TotalRAM),
		PressureLevel: "Normal", // Simplified for now
	}, nil
}
