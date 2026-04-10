package mobile

import (
	"github.com/SirsiMaster/sirsi-pantheon/internal/seba"
)

// SebaDetectHardware returns full device hardware profile as JSON.
// Returns Response JSON with HardwareProfile data.
func SebaDetectHardware() string {
	profile, err := seba.DetectHardware()
	if err != nil {
		return errorJSON("hardware detection failed: " + err.Error())
	}

	return successJSON(profile)
}

// SebaDetectAccelerators probes available compute accelerators.
// Returns Response JSON with AcceleratorProfile data.
// On iOS this detects: Apple Neural Engine, Metal GPU, CPU cores.
func SebaDetectAccelerators() string {
	profile := seba.DetectAccelerators()
	return successJSON(profile)
}
