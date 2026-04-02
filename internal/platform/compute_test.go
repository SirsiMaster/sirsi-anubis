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
}

func TestDetectComputeWith_MockDarwin(t *testing.T) {
	m := &Mock{
		NameStr: "darwin",
		CommandResults: map[string]string{
			"sysctl -n machdep.cpu.brand_string":                   "Apple M3 Pro",
			"sysctl -n hw.memsize":                                 "17179869184",
			"sysctl -n hw.physicalcpu":                             "12",
			"sysctl -n hw.perflevel0.logicalcpu":                   "6",
			"sysctl -n hw.perflevel1.logicalcpu":                   "6",
			"sysctl -n hw.optional.ane":                            "1",
			"system_profiler SPDisplaysDataType -detailLevel mini": "Total Number of Cores: 18\nChipset Model: Apple M3 Pro",
		},
	}

	cc := DetectComputeWith(m)

	if cc.CPUModel != "Apple M3 Pro" {
		t.Errorf("CPUModel = %q, want %q", cc.CPUModel, "Apple M3 Pro")
	}
	if cc.PhysicalCores != 12 {
		t.Errorf("PhysicalCores = %d, want 12", cc.PhysicalCores)
	}
	if cc.PCores != 6 {
		t.Errorf("PCores = %d, want 6", cc.PCores)
	}
	if cc.ECores != 6 {
		t.Errorf("ECores = %d, want 6", cc.ECores)
	}
	if !cc.ANEAvailable || cc.ANECores != 16 {
		t.Errorf("ANE: available=%v, cores=%d", cc.ANEAvailable, cc.ANECores)
	}
	if cc.GPUCores != 18 || cc.GPUModel != "Apple M3 Pro" {
		t.Errorf("GPU: cores=%d, model=%q", cc.GPUCores, cc.GPUModel)
	}
	if cc.MemoryBandwidth != 150 { // M3 Pro estimation
		t.Errorf("MemoryBandwidth = %d, want 150", cc.MemoryBandwidth)
	}
	if !cc.UnifiedMemory {
		t.Error("Apple Silicon should have unified memory")
	}
}

func TestDetectComputeWith_MockLinux(t *testing.T) {
	m := &Mock{
		NameStr: "linux",
		CommandResults: map[string]string{
			"grep -m1 model name /proc/cpuinfo": "model name : Intel(R) Core(TM) i9-12900K",
			"grep MemTotal /proc/meminfo":       "MemTotal:       32845600 kB",
		},
	}

	cc := DetectComputeWith(m)

	if cc.CPUModel != "Intel(R) Core(TM) i9-12900K" {
		t.Errorf("CPUModel = %q, want %q", cc.CPUModel, "Intel(R) Core(TM) i9-12900K")
	}
	// 32845600 * 1024 = 33633894400
	if cc.TotalRAMBytes != 33633894400 {
		t.Errorf("TotalRAMBytes = %d, want 33633894400", cc.TotalRAMBytes)
	}
}

// ── detectLinuxCompute ──────────────────────────────────────────────────

func TestDetectLinuxCompute(t *testing.T) {
	m := &Mock{
		CommandResults: map[string]string{
			"grep -m1 model name /proc/cpuinfo": "model name : Test CPU",
			"grep MemTotal /proc/meminfo":       "MemTotal:       1024 kB",
		},
	}
	cc := &ComputeCapability{}
	detectLinuxCompute(m, cc)

	if cc.CPUModel != "Test CPU" {
		t.Errorf("CPUModel = %q, want %q", cc.CPUModel, "Test CPU")
	}
	if cc.TotalRAMBytes != 1024*1024 {
		t.Errorf("TotalRAMBytes = %d, want %d", cc.TotalRAMBytes, 1024*1024)
	}
}

// ── detectDarwinCompute ─────────────────────────────────────────────────

func TestDetectDarwinCompute(t *testing.T) {
	m := &Mock{
		CommandResults: map[string]string{
			"sysctl -n machdep.cpu.brand_string":                   "Apple M1 Max",
			"sysctl -n hw.memsize":                                 "34359738368",
			"sysctl -n hw.physicalcpu":                             "10",
			"sysctl -n hw.perflevel0.logicalcpu":                   "8",
			"sysctl -n hw.perflevel1.logicalcpu":                   "2",
			"sysctl -n hw.optional.ane":                            "1",
			"system_profiler SPDisplaysDataType -detailLevel mini": "Total Number of Cores: 32\nChipset Model: Apple M1 Max",
		},
	}
	cc := &ComputeCapability{}
	detectDarwinCompute(m, cc)

	if cc.CPUModel != "Apple M1 Max" {
		t.Errorf("CPUModel = %q, want %q", cc.CPUModel, "Apple M1 Max")
	}
	if cc.MemoryBandwidth != 400 {
		t.Errorf("MemoryBandwidth = %d, want 400", cc.MemoryBandwidth)
	}
}
