package seba

import (
	"testing"
)

// ── parseDisplayProfile Tests ───────────────────────────────────────────

func TestParseDisplayProfile_AppleSilicon(t *testing.T) {
	sampleOutput := `Graphics/Displays:

    Apple M1 Max:

      Chipset Model: Apple M1 Max
      Type: GPU
      Bus: Built-In
      Total Number of Cores: 32
      Vendor: Apple (0x106b)
      Metal Family: Supported, Metal GPUFamily Apple 8
`

	info := parseDisplayProfile(sampleOutput)
	if info.Name != "Apple M1 Max" {
		t.Errorf("Name = %q, want %q", info.Name, "Apple M1 Max")
	}
	if info.Type != GPUAppleMetal {
		t.Errorf("Type = %q, want %q", info.Type, GPUAppleMetal)
	}
	if info.MetalFamily == "" {
		t.Error("MetalFamily should be populated")
	}
}

func TestParseDisplayProfile_Empty(t *testing.T) {
	info := parseDisplayProfile("")
	if info.Name != "Unknown GPU" {
		t.Errorf("Name = %q, want %q for empty input", info.Name, "Unknown GPU")
	}
	if info.Type != GPUNone {
		t.Errorf("Type = %q, want %q for empty input", info.Type, GPUNone)
	}
}

// ── parseNvidiaSmi Tests ────────────────────────────────────────────────

func TestParseNvidiaSmi_FullOutput(t *testing.T) {
	sampleOutput := "NVIDIA GeForce RTX 4090, 24564, 550.54.14, 8.9"
	info := parseNvidiaSmi(sampleOutput)

	if info.Type != GPUNVIDIA {
		t.Errorf("Type = %q, want %q", info.Type, GPUNVIDIA)
	}
	if info.Name != "NVIDIA GeForce RTX 4090" {
		t.Errorf("Name = %q, want %q", info.Name, "NVIDIA GeForce RTX 4090")
	}
	if info.VRAM != 24564*1024*1024 {
		t.Errorf("VRAM = %d, want %d", info.VRAM, 24564*1024*1024)
	}
	if info.DriverVer != "550.54.14" {
		t.Errorf("DriverVer = %q, want %q", info.DriverVer, "550.54.14")
	}
	if info.Compute != "8.9" {
		t.Errorf("Compute = %q, want %q", info.Compute, "8.9")
	}
}

func TestParseNvidiaSmi_PartialOutput(t *testing.T) {
	info := parseNvidiaSmi("GTX 1080")
	if info.Name != "GTX 1080" {
		t.Errorf("Name = %q, want %q", info.Name, "GTX 1080")
	}
	if info.VRAM != 0 {
		t.Errorf("VRAM = %d, want 0 for partial output", info.VRAM)
	}
}

// ── detectLinuxHardware (smoke test) ────────────────────────────────────

func TestDetectLinuxHardware(t *testing.T) {
	profile := &HardwareProfile{}
	detectLinuxHardware(profile)
	// On macOS, /proc/cpuinfo and `free` don't exist — values will be empty.
	// Just verify the code path runs without panic.
	t.Logf("Linux hardware: CPU=%q, RAM=%d, GPU=%q",
		profile.CPUModel, profile.TotalRAM, profile.GPU.Name)
}

// ── DetectHardware Integration ──────────────────────────────────────────

func TestDetectHardware_ReturnsProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live hardware detection in short mode")
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
}

// ── Tokenize via Accelerator Tests ──────────────────────────────────────

func TestANETokenize(t *testing.T) {
	t.Parallel()
	ane := &AppleANEAccelerator{cores: 16}
	tokens, err := ane.Tokenize("test tokenize")
	if err != nil {
		t.Fatalf("ANE.Tokenize error: %v", err)
	}
	if len(tokens) == 0 {
		t.Error("ANE.Tokenize should return tokens")
	}
}

func TestMetalTokenize(t *testing.T) {
	t.Parallel()
	metal := &MetalAccelerator{cores: 18, model: "M3 Pro"}
	tokens, err := metal.Tokenize("metal shader test")
	if err != nil {
		t.Fatalf("Metal.Tokenize error: %v", err)
	}
	if len(tokens) == 0 {
		t.Error("Metal.Tokenize should return tokens")
	}
}

func TestCUDATokenize(t *testing.T) {
	t.Parallel()
	cuda := &CUDAAccelerator{model: "RTX 4090", vram: 24 * 1024 * 1024 * 1024}
	tokens, err := cuda.Tokenize("cuda kernel")
	if err != nil {
		t.Fatalf("CUDA.Tokenize error: %v", err)
	}
	if len(tokens) == 0 {
		t.Error("CUDA.Tokenize should return tokens")
	}
}

func TestROCmTokenize(t *testing.T) {
	t.Parallel()
	rocm := &ROCmAccelerator{model: "RX 7900 XTX"}
	tokens, err := rocm.Tokenize("rocm test")
	if err != nil {
		t.Fatalf("ROCm.Tokenize error: %v", err)
	}
	if len(tokens) == 0 {
		t.Error("ROCm.Tokenize should return tokens")
	}
}

func TestCPUTokenize(t *testing.T) {
	t.Parallel()
	cpu := &CPUAccelerator{cores: 8}
	tokens, err := cpu.Tokenize("cpu fallback test")
	if err != nil {
		t.Fatalf("CPU.Tokenize error: %v", err)
	}
	if len(tokens) == 0 {
		t.Error("CPU.Tokenize should return tokens")
	}
}

// ── Tokenize Consistency ─────────────────────────────────────────────────

func TestTokenize_Consistency(t *testing.T) {
	t.Parallel()
	text := "consistency check across all backends"

	ane := &AppleANEAccelerator{cores: 16}
	metal := &MetalAccelerator{cores: 18}
	cuda := &CUDAAccelerator{model: "RTX 4090"}
	rocm := &ROCmAccelerator{model: "RX 7900"}
	cpu := &CPUAccelerator{cores: 8}

	t1, _ := ane.Tokenize(text)
	t2, _ := metal.Tokenize(text)
	t3, _ := cuda.Tokenize(text)
	t4, _ := rocm.Tokenize(text)
	t5, _ := cpu.Tokenize(text)

	all := [][]int{t1, t2, t3, t4, t5}
	for i := 1; i < len(all); i++ {
		if len(all[0]) != len(all[i]) {
			t.Errorf("accelerator[%d] produced %d tokens, accelerator[0] produced %d", i, len(all[i]), len(all[0]))
		}
	}
}
