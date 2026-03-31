// Package hapi provides hardware detection and resource optimization.
// Implementation moved to internal/seba; this file provides backward compatibility.
package hapi

import "github.com/SirsiMaster/sirsi-pantheon/internal/seba"

// Type aliases for backward compatibility.
type GPUType = seba.GPUType
type GPUInfo = seba.GPUInfo
type HardwareProfile = seba.HardwareProfile

// GPU type constants.
const (
	GPUAppleMetal GPUType = seba.GPUAppleMetal
	GPUNVIDIA     GPUType = seba.GPUNVIDIA
	GPUAMD        GPUType = seba.GPUAMD
	GPUIntel      GPUType = seba.GPUIntel
	GPUNone       GPUType = seba.GPUNone
)

// DetectHardware delegates to seba.DetectHardware.
func DetectHardware() (*HardwareProfile, error) {
	return seba.DetectHardware()
}

// FormatGPUType delegates to seba.FormatGPUType.
func FormatGPUType(t GPUType) string {
	return seba.FormatGPUType(t)
}

// FormatBytes delegates to seba.FormatBytes.
func FormatBytes(b int64) string {
	return seba.FormatBytes(b)
}
