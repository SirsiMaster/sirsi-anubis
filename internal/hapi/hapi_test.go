package hapi

import (
	"runtime"
	"testing"
)

// ═══════════════════════════════════════════
// GPU Detection — types and formatting
// ═══════════════════════════════════════════

func TestGPUType_Constants(t *testing.T) {
	// Ensure GPUType constants are distinct
	types := []GPUType{GPUAppleMetal, GPUNVIDIA, GPUAMD, GPUIntel, GPUNone}
	seen := make(map[GPUType]bool)
	for _, gt := range types {
		if seen[gt] {
			t.Errorf("duplicate GPUType constant: %q", gt)
		}
		seen[gt] = true
	}
}

func TestFormatGPUType(t *testing.T) {
	tests := []struct {
		input GPUType
		want  string
	}{
		{GPUAppleMetal, "Apple Metal"},
		{GPUNVIDIA, "NVIDIA CUDA"},
		{GPUAMD, "AMD ROCm"},
		{GPUIntel, "Intel"},
		{GPUNone, "CPU-only"},
		{GPUType("unknown"), "CPU-only"}, // fallback
	}

	for _, tt := range tests {
		got := FormatGPUType(tt.input)
		if got != tt.want {
			t.Errorf("FormatGPUType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
		{34359738368, "32.0 GB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.input)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ═══════════════════════════════════════════
// DetectHardware — smoke test (doesn't crash)
// ═══════════════════════════════════════════

func TestDetectHardware_ReturnsProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live system_profiler call in short mode")
	}
	profile, err := DetectHardware()
	if err != nil {
		t.Fatalf("DetectHardware() error: %v", err)
	}
	if profile == nil {
		t.Fatal("DetectHardware() returned nil profile")
	}
	if profile.CPUCores <= 0 {
		t.Errorf("CPUCores = %d, expected positive", profile.CPUCores)
	}
	if profile.CPUArch == "" {
		t.Error("CPUArch should not be empty")
	}
	if profile.OS == "" {
		t.Error("OS should not be empty")
	}
}

// ═══════════════════════════════════════════
// Snapshots
// ═══════════════════════════════════════════

func TestPruneSnapshot_DryRun(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("snapshot pruning requires macOS")
	}
	// Dry run should return nil without executing anything
	err := PruneSnapshot("com.apple.TimeMachine.2026-03-21-120000.local", true)
	if err != nil {
		t.Errorf("PruneSnapshot dry run should succeed, got: %v", err)
	}
}

func TestListSnapshots_DoesNotCrash(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live tmutil call in short mode")
	}
	// ListSnapshots should return a valid result on any platform
	result, err := ListSnapshots()
	if err != nil {
		t.Fatalf("ListSnapshots() error: %v", err)
	}
	if result == nil {
		t.Fatal("ListSnapshots() returned nil")
	}
	if result.Total != len(result.Snapshots) {
		t.Errorf("Total=%d doesn't match len(Snapshots)=%d", result.Total, len(result.Snapshots))
	}
}

// ═══════════════════════════════════════════
// HardwareProfile struct fields
// ═══════════════════════════════════════════

func TestHardwareProfile_Defaults(t *testing.T) {
	p := HardwareProfile{}
	if p.NeuralEngine {
		t.Error("NeuralEngine should default to false")
	}
	if p.GPU.Type != "" {
		t.Errorf("GPU.Type should default to empty, got %q", p.GPU.Type)
	}
}
