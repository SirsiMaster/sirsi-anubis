// Package hapi — accelerator backward compatibility wrapper.
// Implementation moved to internal/seba/accel.go.
package hapi

import "github.com/SirsiMaster/sirsi-pantheon/internal/seba"

// Type aliases for accelerator types.
type AcceleratorType = seba.AcceleratorType
type Accelerator = seba.Accelerator
type AcceleratorProfile = seba.AcceleratorProfile

// Accelerator type constants.
const (
	AccelCPU      AcceleratorType = seba.AccelCPU
	AccelAppleANE AcceleratorType = seba.AccelAppleANE
	AccelMetal    AcceleratorType = seba.AccelMetal
	AccelCUDA     AcceleratorType = seba.AccelCUDA
	AccelROCm     AcceleratorType = seba.AccelROCm
	AccelOneAPI   AcceleratorType = seba.AccelOneAPI
)

// DetectAccelerators delegates to seba.DetectAccelerators.
func DetectAccelerators() *AcceleratorProfile {
	return seba.DetectAccelerators()
}

// FastTokenize delegates to seba.FastTokenize.
func FastTokenize(text string) []int {
	return seba.FastTokenize(text)
}
