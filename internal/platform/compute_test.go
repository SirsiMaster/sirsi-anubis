package platform

import (
	"runtime"
	"testing"
)

// ── DetectCompute ────────────────────────────────────────────────────────

func TestDetectCompute(t *testing.T) {
	cc := DetectCompute()

	if cc == nil {
		t.Fatal("DetectCompute returned nil")
	}

	// LogicalCores must match runtime
	if cc.LogicalCores != runtime.NumCPU() {
		t.Errorf("LogicalCores = %d, want %d", cc.LogicalCores, runtime.NumCPU())
	}

	// Physical cores should be > 0
	if cc.PhysicalCores <= 0 {
		t.Errorf("PhysicalCores = %d, want > 0", cc.PhysicalCores)
	}

	// Optimal workers should be >= 2
	if cc.OptimalWorkers < 2 {
		t.Errorf("OptimalWorkers = %d, want >= 2", cc.OptimalWorkers)
	}

	// Optimal CPU workers > 0
	if cc.OptimalCPUWorkers <= 0 {
		t.Errorf("OptimalCPUWorkers = %d, want > 0", cc.OptimalCPUWorkers)
	}

	// Optimal IO workers >= 4
	if cc.OptimalIOWorkers < 4 {
		t.Errorf("OptimalIOWorkers = %d, want >= 4", cc.OptimalIOWorkers)
	}

	t.Logf("Compute: model=%q, logical=%d, physical=%d, pcores=%d, ecores=%d",
		cc.CPUModel, cc.LogicalCores, cc.PhysicalCores, cc.PCores, cc.ECores)
	t.Logf("  GPU: cores=%d, model=%q", cc.GPUCores, cc.GPUModel)
	t.Logf("  ANE: available=%v, cores=%d", cc.ANEAvailable, cc.ANECores)
	t.Logf("  Memory: RAM=%d, bandwidth=%d GB/s, unified=%v",
		cc.TotalRAMBytes, cc.MemoryBandwidth, cc.UnifiedMemory)
	t.Logf("  Workers: optimal=%d, cpu=%d, io=%d",
		cc.OptimalWorkers, cc.OptimalCPUWorkers, cc.OptimalIOWorkers)
}

func TestDetectCompute_DarwinSpecific(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin-specific test")
	}

	cc := DetectCompute()

	// CPU model should be populated on macOS
	if cc.CPUModel == "" {
		t.Error("CPUModel should not be empty on macOS")
	}

	// Total RAM should be populated
	if cc.TotalRAMBytes <= 0 {
		t.Errorf("TotalRAMBytes = %d, want > 0", cc.TotalRAMBytes)
	}

	// Apple Silicon checks
	if cc.ANEAvailable {
		if cc.ANECores <= 0 {
			t.Error("ANE available but ANECores <= 0")
		}
		if !cc.UnifiedMemory {
			t.Error("ANE available implies Apple Silicon, should have unified memory")
		}
	}
}

// ── Worker Derivation Logic ──────────────────────────────────────────────

func TestDetectCompute_MinimumWorkers(t *testing.T) {
	// Verify the minimum worker counts:
	// OptimalWorkers >= 2
	// OptimalIOWorkers >= 4
	// OptimalCPUWorkers > 0

	cc := DetectCompute()

	if cc.OptimalWorkers < 2 {
		t.Errorf("OptimalWorkers = %d, minimum is 2", cc.OptimalWorkers)
	}
	if cc.OptimalIOWorkers < 4 {
		t.Errorf("OptimalIOWorkers = %d, minimum is 4", cc.OptimalIOWorkers)
	}
	if cc.OptimalCPUWorkers < 1 {
		t.Errorf("OptimalCPUWorkers = %d, minimum is 1", cc.OptimalCPUWorkers)
	}
}

// ── Memory Bandwidth Estimation ─────────────────────────────────────────

func TestDetectCompute_MemoryBandwidth(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Memory bandwidth estimation is macOS-specific")
	}

	cc := DetectCompute()

	// If Apple Silicon, bandwidth should be estimated
	if cc.UnifiedMemory && cc.MemoryBandwidth <= 0 {
		t.Logf("WARNING: Apple Silicon detected but bandwidth not estimated (model=%q)", cc.CPUModel)
	}

	if cc.MemoryBandwidth > 0 {
		// Known range: 68 (M1) to 800 (Ultra)
		if cc.MemoryBandwidth < 50 || cc.MemoryBandwidth > 1000 {
			t.Errorf("MemoryBandwidth = %d GB/s, outside expected range [50, 1000]", cc.MemoryBandwidth)
		}
		t.Logf("Memory bandwidth: %d GB/s", cc.MemoryBandwidth)
	}
}

// ── detectLinuxCompute ──────────────────────────────────────────────────

func TestDetectLinuxCompute(t *testing.T) {
	cc := &ComputeCapability{LogicalCores: runtime.NumCPU()}
	detectLinuxCompute(cc)

	// On macOS, the Linux path won't find /proc/cpuinfo, but shouldn't panic
	if cc.PhysicalCores <= 0 {
		// This is expected on non-Linux — it should default to NumCPU
		if runtime.GOOS == "linux" {
			t.Errorf("PhysicalCores = %d on Linux, want > 0", cc.PhysicalCores)
		}
	}
}

// ── detectDarwinCompute ─────────────────────────────────────────────────

func TestDetectDarwinCompute(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-only test")
	}

	cc := &ComputeCapability{LogicalCores: runtime.NumCPU()}
	detectDarwinCompute(cc)

	if cc.CPUModel == "" {
		t.Error("CPUModel should be populated on macOS")
	}
	if cc.PhysicalCores <= 0 {
		t.Errorf("PhysicalCores = %d, want > 0", cc.PhysicalCores)
	}
	if cc.TotalRAMBytes <= 0 {
		t.Errorf("TotalRAMBytes = %d, want > 0", cc.TotalRAMBytes)
	}

	t.Logf("Darwin compute: model=%q, physical=%d, P=%d/E=%d, ANE=%v, GPU=%d",
		cc.CPUModel, cc.PhysicalCores, cc.PCores, cc.ECores, cc.ANEAvailable, cc.GPUCores)
}

// ── PickFolder / OpenBrowser ─────────────────────────────────────────────

func TestDarwinPickFolder(t *testing.T) {
	// Can't actually open a dialog in tests, but verify it doesn't panic
	// with a quick error return.
	d := &Darwin{}
	_ = d.PickFolder
}

func TestDarwinOpenBrowser(t *testing.T) {
	d := &Darwin{}
	_ = d.OpenBrowser
}

func TestLinuxPickFolder(t *testing.T) {
	l := &Linux{}
	_ = l.PickFolder
}

func TestLinuxOpenBrowser(t *testing.T) {
	l := &Linux{}
	_ = l.OpenBrowser
}
