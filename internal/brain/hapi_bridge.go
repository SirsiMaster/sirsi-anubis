package brain

import (
	"github.com/SirsiMaster/sirsi-pantheon/internal/logging"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seba"
)

// HardwareBridge connects accelerator detection to the inference layer.
type HardwareBridge struct {
	profile *seba.HardwareProfile
}

// NewHardwareBridge initializes hardware detection and returns a bridge instance.
func NewHardwareBridge() (*HardwareBridge, error) {
	profile, err := seba.DetectHardware()
	if err != nil {
		logging.Error("🧠 Brain: hardware detection failed", "error", err)
		return nil, err
	}

	logging.Info("🧠 Brain: Accelerator detected",
		"type", profile.GPU.Type,
		"vram", profile.GPU.VRAM,
		"cores", profile.CPUCores)

	return &HardwareBridge{profile: profile}, nil
}

// BackendPreference returns the optimal inference backend based on hardware.
func (b *HardwareBridge) BackendPreference() string {
	if b.profile == nil {
		return "stub"
	}

	switch b.profile.GPU.Type {
	case seba.GPUAppleMetal:
		return "coreml"
	case seba.GPUNVIDIA:
		return "onnx-cuda"
	case seba.GPUAMD:
		return "onnx-rocm"
	default:
		return "onnx-cpu"
	}
}

// Profile returns the underlying hardware profile.
func (b *HardwareBridge) Profile() *seba.HardwareProfile {
	return b.profile
}

// Deprecated aliases for backward compatibility.
type HapiBridge = HardwareBridge

func NewHapiBridge() (*HardwareBridge, error) { return NewHardwareBridge() }
