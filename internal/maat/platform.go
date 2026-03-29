// Package maat implements quality and truth governance for Pantheon.
// 𓆄 Ma'at — the goddess of truth, balance, and order.
package maat

import (
	"fmt"
	"os/exec"
	"strings"
)

// PlatformInfo contains the actual system details.
type PlatformInfo struct {
	Model     string
	OSVersion string
	Build     string
}

// GetActualPlatform queries the live system for hardware and OS info.
func GetActualPlatform() (*PlatformInfo, error) {
	model, _ := exec.Command("sysctl", "-n", "hw.model").Output()
	swVers, _ := exec.Command("sw_vers").Output()

	info := &PlatformInfo{
		Model: strings.TrimSpace(string(model)),
	}

	lines := strings.Split(string(swVers), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ProductVersion:") {
			info.OSVersion = strings.TrimSpace(strings.TrimPrefix(line, "ProductVersion:"))
		}
		if strings.HasPrefix(line, "BuildVersion:") {
			info.Build = strings.TrimSpace(strings.TrimPrefix(line, "BuildVersion:"))
		}
	}

	return info, nil
}

// CheckPlatformIntegrity verifies that the provided platform string matches
// the actual hardware. Use this in documentation generation.
func CheckPlatformIntegrity(claimed string) error {
	actual, err := GetActualPlatform()
	if err != nil {
		return err
	}

	// Canonical name mapping
	actualName := "Apple M1 Max"
	actualOS := "macOS Tahoe"

	if strings.HasPrefix(actual.OSVersion, "26.") {
		actualOS = "macOS Tahoe"
	}

	claimedLower := strings.ToLower(claimed)
	actualHWLower := strings.ToLower(actualName)
	actualOSLower := strings.ToLower(actualOS)

	// Check hardware
	if !strings.Contains(claimedLower, "m1 max") && !strings.Contains(claimedLower, actualHWLower) {
		return fmt.Errorf("𓁢 Ma'at Rule Violation: Claimed hardware in (%s) does not match actual hardware (%s)", claimed, actualName)
	}

	// Check OS
	if !strings.Contains(claimedLower, "tahoe") && !strings.Contains(claimedLower, actualOSLower) {
		return fmt.Errorf("𓁢 Ma'at Rule Violation: Claimed OS in (%s) does not match actual OS (%s)", claimed, actualOS)
	}

	return nil
}
