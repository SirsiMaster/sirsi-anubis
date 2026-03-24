package hapi

import (
	"testing"
)

func TestDetectLinuxHardware(t *testing.T) {
	profile := &HardwareProfile{}
	detectLinuxHardware(profile)
	// On macOS, /proc/cpuinfo and `free` don't exist — values will be empty.
	// Just verify the code path runs without panic.
	t.Logf("Linux hardware: CPU=%q, RAM=%d, GPU=%q",
		profile.CPUModel, profile.TotalRAM, profile.GPU.Name)
}
