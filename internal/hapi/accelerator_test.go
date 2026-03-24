package hapi

import (
	"crypto/sha256"
	"testing"
)

// ── Accelerator Interface Tests ─────────────────────────────────────────

func TestAppleANEAccelerator(t *testing.T) {
	ane := &AppleANEAccelerator{cores: 16}

	if ane.Type() != AccelAppleANE {
		t.Errorf("Type = %q, want %q", ane.Type(), AccelAppleANE)
	}
	if ane.Vendor() != "apple" {
		t.Errorf("Vendor = %q, want 'apple'", ane.Vendor())
	}
	if !ane.SupportsEmbedding() {
		t.Error("ANE should support embedding")
	}
	if ane.SupportsHashing() {
		t.Error("ANE should NOT support hashing")
	}
	if !ane.SupportsClassification() {
		t.Error("ANE should support classification")
	}
	if !ane.Available() {
		t.Error("ANE with 16 cores should be available")
	}

	// ComputeHash returns zero on ANE (not supported)
	hash := ane.ComputeHash([]byte("test"))
	if hash != [32]byte{} {
		t.Error("ANE ComputeHash should return zero hash")
	}

	// Zero cores = not available
	aneZero := &AppleANEAccelerator{cores: 0}
	if aneZero.Available() {
		t.Error("ANE with 0 cores should NOT be available")
	}
}

func TestMetalAccelerator(t *testing.T) {
	metal := &MetalAccelerator{cores: 18, model: "Apple M3 Pro"}

	if metal.Type() != AccelMetal {
		t.Errorf("Type = %q, want %q", metal.Type(), AccelMetal)
	}
	if metal.Vendor() != "apple" {
		t.Errorf("Vendor = %q, want 'apple'", metal.Vendor())
	}
	if metal.SupportsEmbedding() {
		t.Error("Metal should NOT support embedding")
	}
	if !metal.SupportsHashing() {
		t.Error("Metal should support hashing")
	}
	if metal.SupportsClassification() {
		t.Error("Metal should NOT support classification")
	}
	if !metal.Available() {
		t.Error("Metal with 18 cores should be available")
	}

	// ComputeHash should produce real SHA-256 (CPU fallback)
	data := []byte("hello accelerator")
	expected := sha256.Sum256(data)
	got := metal.ComputeHash(data)
	if got != expected {
		t.Error("Metal ComputeHash should match CPU SHA-256")
	}

	// Zero cores = not available
	metalZero := &MetalAccelerator{cores: 0}
	if metalZero.Available() {
		t.Error("Metal with 0 cores should NOT be available")
	}
}

func TestCUDAAccelerator(t *testing.T) {
	cuda := &CUDAAccelerator{model: "RTX 4090", vram: 24 * 1024 * 1024 * 1024}

	if cuda.Type() != AccelCUDA {
		t.Errorf("Type = %q, want %q", cuda.Type(), AccelCUDA)
	}
	if cuda.Vendor() != "nvidia" {
		t.Errorf("Vendor = %q, want 'nvidia'", cuda.Vendor())
	}
	if !cuda.SupportsEmbedding() {
		t.Error("CUDA should support embedding")
	}
	if !cuda.SupportsHashing() {
		t.Error("CUDA should support hashing")
	}
	if !cuda.SupportsClassification() {
		t.Error("CUDA should support classification")
	}
	if !cuda.Available() {
		t.Error("CUDA with model should be available")
	}

	data := []byte("cuda test")
	expected := sha256.Sum256(data)
	got := cuda.ComputeHash(data)
	if got != expected {
		t.Error("CUDA ComputeHash should match CPU SHA-256")
	}

	// Empty model = not available
	cudaEmpty := &CUDAAccelerator{}
	if cudaEmpty.Available() {
		t.Error("CUDA with empty model should NOT be available")
	}
}

func TestROCmAccelerator(t *testing.T) {
	rocm := &ROCmAccelerator{model: "Radeon RX 7900 XTX"}

	if rocm.Type() != AccelROCm {
		t.Errorf("Type = %q, want %q", rocm.Type(), AccelROCm)
	}
	if rocm.Vendor() != "amd" {
		t.Errorf("Vendor = %q, want 'amd'", rocm.Vendor())
	}
	if !rocm.SupportsEmbedding() {
		t.Error("ROCm should support embedding")
	}
	if !rocm.SupportsHashing() {
		t.Error("ROCm should support hashing")
	}
	if rocm.SupportsClassification() {
		t.Error("ROCm should NOT support classification")
	}
	if !rocm.Available() {
		t.Error("ROCm with model should be available")
	}

	data := []byte("rocm test")
	expected := sha256.Sum256(data)
	got := rocm.ComputeHash(data)
	if got != expected {
		t.Error("ROCm ComputeHash should match CPU SHA-256")
	}

	rocmEmpty := &ROCmAccelerator{}
	if rocmEmpty.Available() {
		t.Error("ROCm with empty model should NOT be available")
	}
}

func TestCPUAccelerator(t *testing.T) {
	cpu := &CPUAccelerator{cores: 12}

	if cpu.Type() != AccelCPU {
		t.Errorf("Type = %q, want %q", cpu.Type(), AccelCPU)
	}
	if cpu.Vendor() != "cpu" {
		t.Errorf("Vendor = %q, want 'cpu'", cpu.Vendor())
	}
	if cpu.SupportsEmbedding() {
		t.Error("CPU should NOT support embedding")
	}
	if !cpu.SupportsHashing() {
		t.Error("CPU should support hashing")
	}
	if cpu.SupportsClassification() {
		t.Error("CPU should NOT support classification")
	}
	if !cpu.Available() {
		t.Error("CPU should always be available")
	}

	data := []byte("cpu fallback test")
	expected := sha256.Sum256(data)
	got := cpu.ComputeHash(data)
	if got != expected {
		t.Error("CPU ComputeHash should match sha256.Sum256")
	}
}

// ── routeWorkload Tests ─────────────────────────────────────────────────

func TestRouteWorkload(t *testing.T) {
	ane := &AppleANEAccelerator{cores: 16}
	metal := &MetalAccelerator{cores: 18}
	cpu := &CPUAccelerator{cores: 12}

	accs := []Accelerator{ane, metal, cpu}

	// Embedding → ANE (first that supports it)
	result := routeWorkload(accs, func(a Accelerator) bool { return a.SupportsEmbedding() })
	if result != string(AccelAppleANE) {
		t.Errorf("embedding route = %q, want apple-ane", result)
	}

	// Hashing → Metal (first that supports it — ANE doesn't)
	result = routeWorkload(accs, func(a Accelerator) bool { return a.SupportsHashing() })
	if result != string(AccelMetal) {
		t.Errorf("hashing route = %q, want apple-metal", result)
	}

	// No accelerator supports → CPU fallback
	result = routeWorkload(accs, func(a Accelerator) bool { return false })
	if result != string(AccelCPU) {
		t.Errorf("no-match route = %q, want cpu", result)
	}
}

func TestRouteWorkload_EmptyList(t *testing.T) {
	result := routeWorkload(nil, func(a Accelerator) bool { return true })
	if result != string(AccelCPU) {
		t.Errorf("empty list route = %q, want cpu", result)
	}
}

func TestRouteWorkload_UnavailableSkipped(t *testing.T) {
	// ANE with 0 cores is unavailable — should skip to CPU
	ane := &AppleANEAccelerator{cores: 0}
	cpu := &CPUAccelerator{cores: 8}

	result := routeWorkload([]Accelerator{ane, cpu}, func(a Accelerator) bool { return a.SupportsEmbedding() })
	// ANE supports embedding but is unavailable, CPU doesn't support embedding → fallback
	if result != string(AccelCPU) {
		t.Errorf("unavailable route = %q, want cpu", result)
	}
}

// ── DetectAccelerators Integration ──────────────────────────────────────

func TestDetectAccelerators(t *testing.T) {
	profile := DetectAccelerators()

	if profile == nil {
		t.Fatal("DetectAccelerators returned nil")
	}
	if profile.CPUCores <= 0 {
		t.Errorf("CPUCores = %d, want > 0", profile.CPUCores)
	}
	if len(profile.All) == 0 {
		t.Error("Should have at least CPU accelerator")
	}
	if profile.Primary == nil {
		t.Error("Primary accelerator should not be nil")
	}
	if profile.Routing == nil {
		t.Error("Routing table should not be nil")
	}

	// Routing table should have entries for all workload types
	for _, workload := range []string{"embedding", "hashing", "classification"} {
		if _, ok := profile.Routing[workload]; !ok {
			t.Errorf("Routing missing entry for %q", workload)
		}
	}

	// Hashing should always be routed (CPU at minimum)
	hashRoute := profile.Routing["hashing"]
	if hashRoute == "" {
		t.Error("Hashing route should not be empty")
	}

	t.Logf("Detected: %d accelerators, primary=%s, gpu=%v, ane=%v",
		len(profile.All), profile.Primary.Type(), profile.HasGPU, profile.HasANE)
}

// ── AcceleratorType Constants ────────────────────────────────────────────

func TestAcceleratorTypeConstants(t *testing.T) {
	types := map[AcceleratorType]string{
		AccelCPU:      "cpu",
		AccelAppleANE: "apple-ane",
		AccelMetal:    "apple-metal",
		AccelCUDA:     "nvidia-cuda",
		AccelROCm:     "amd-rocm",
		AccelOneAPI:   "intel-oneapi",
	}
	for typ, expected := range types {
		if string(typ) != expected {
			t.Errorf("AcceleratorType %q should be %q", typ, expected)
		}
	}
}
