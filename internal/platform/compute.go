// Package platform — compute.go
//
// Hardware compute capability detection for Apple Silicon, NVIDIA, and generic x86.
// Detects CPU topology (P-cores, E-cores, SMT), GPU, ANE (Apple Neural Engine),
// and memory bandwidth for optimal thread pool sizing and accelerator routing.
package platform

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// ComputeCapability describes the hardware compute resources available.
type ComputeCapability struct {
	// CPU
	LogicalCores  int    `json:"logical_cores"`  // Total logical cores (including SMT)
	PhysicalCores int    `json:"physical_cores"` // Physical cores
	PCores        int    `json:"p_cores"`        // Performance cores (Apple Silicon)
	ECores        int    `json:"e_cores"`        // Efficiency cores (Apple Silicon)
	CPUModel      string `json:"cpu_model"`      // e.g., "Apple M3 Pro"

	// GPU
	GPUCores int    `json:"gpu_cores,omitempty"` // GPU cores (Apple Silicon integrated)
	GPUModel string `json:"gpu_model,omitempty"` // GPU model name

	// ANE (Apple Neural Engine)
	ANEAvailable bool `json:"ane_available"` // Whether ANE is present
	ANECores     int  `json:"ane_cores"`     // Neural Engine cores (16 on M1+)

	// Memory
	TotalRAMBytes   int64 `json:"total_ram_bytes"`
	MemoryBandwidth int   `json:"memory_bandwidth_gbps,omitempty"` // GB/s
	UnifiedMemory   bool  `json:"unified_memory"`                  // Apple Silicon unified memory

	// Derived recommendations
	OptimalWorkers    int `json:"optimal_workers"`     // Recommended goroutine pool size
	OptimalIOWorkers  int `json:"optimal_io_workers"`  // Recommended I/O-bound pool size
	OptimalCPUWorkers int `json:"optimal_cpu_workers"` // Recommended CPU-bound pool size
}

// DetectCompute probes the hardware and returns compute capabilities.
// On Apple Silicon, uses sysctl for detailed core topology and ANE detection.
// On Linux/x86, falls back to /proc/cpuinfo and runtime.NumCPU().
func DetectCompute() *ComputeCapability {
	cc := &ComputeCapability{
		LogicalCores: runtime.NumCPU(),
	}

	switch runtime.GOOS {
	case "darwin":
		detectDarwinCompute(cc)
	case "linux":
		detectLinuxCompute(cc)
	default:
		cc.PhysicalCores = runtime.NumCPU()
	}

	// Derive optimal worker counts
	cc.OptimalWorkers = cc.LogicalCores
	if cc.OptimalWorkers < 2 {
		cc.OptimalWorkers = 2
	}

	// CPU-bound: use P-cores only (or physical cores)
	cc.OptimalCPUWorkers = cc.PCores
	if cc.OptimalCPUWorkers == 0 {
		cc.OptimalCPUWorkers = cc.PhysicalCores
	}
	if cc.OptimalCPUWorkers == 0 {
		cc.OptimalCPUWorkers = runtime.NumCPU()
	}

	// I/O-bound: use all logical cores (SMT helps for I/O wait)
	cc.OptimalIOWorkers = cc.LogicalCores
	if cc.OptimalIOWorkers < 4 {
		cc.OptimalIOWorkers = 4
	}

	return cc
}

// detectDarwinCompute uses sysctl to probe Apple Silicon topology.
func detectDarwinCompute(cc *ComputeCapability) {
	// CPU model
	if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
		cc.CPUModel = strings.TrimSpace(string(out))
	}

	// Physical cores
	if out, err := exec.Command("sysctl", "-n", "hw.physicalcpu").Output(); err == nil {
		cc.PhysicalCores, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}

	// P-cores and E-cores (Apple Silicon specific)
	if out, err := exec.Command("sysctl", "-n", "hw.perflevel0.logicalcpu").Output(); err == nil {
		cc.PCores, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}
	if out, err := exec.Command("sysctl", "-n", "hw.perflevel1.logicalcpu").Output(); err == nil {
		cc.ECores, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}

	// Total RAM
	if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
		cc.TotalRAMBytes, _ = strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	}

	// ANE detection — Apple Silicon M1+ has Neural Engine
	// Detected via IORegistry: AppleANE device
	if out, err := exec.Command("ioreg", "-l", "-w0").Output(); err == nil {
		outStr := string(out)
		if strings.Contains(outStr, "appleane") || strings.Contains(strings.ToLower(outStr), "ane") {
			cc.ANEAvailable = true
			cc.ANECores = 16 // All M1/M2/M3/M4 have 16 ANE cores
		}
	}

	// If ioreg takes too long, also check CPU model
	if !cc.ANEAvailable && strings.Contains(cc.CPUModel, "Apple") {
		cc.ANEAvailable = true
		cc.ANECores = 16
	}

	// Unified memory (all Apple Silicon)
	if strings.Contains(cc.CPUModel, "Apple") {
		cc.UnifiedMemory = true
	}

	// GPU cores — detect from system_profiler
	if out, err := exec.Command("system_profiler", "SPDisplaysDataType", "-detailLevel", "mini").Output(); err == nil {
		outStr := string(out)
		for _, line := range strings.Split(outStr, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Total Number of Cores:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					cc.GPUCores, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
				}
			}
			if strings.HasPrefix(line, "Chipset Model:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					cc.GPUModel = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// Memory bandwidth estimation based on chip
	model := strings.ToLower(cc.CPUModel)
	switch {
	case strings.Contains(model, "m4 ultra"):
		cc.MemoryBandwidth = 800
	case strings.Contains(model, "m4 max"):
		cc.MemoryBandwidth = 546
	case strings.Contains(model, "m4 pro"):
		cc.MemoryBandwidth = 273
	case strings.Contains(model, "m4"):
		cc.MemoryBandwidth = 120
	case strings.Contains(model, "m3 ultra"):
		cc.MemoryBandwidth = 800
	case strings.Contains(model, "m3 max"):
		cc.MemoryBandwidth = 400
	case strings.Contains(model, "m3 pro"):
		cc.MemoryBandwidth = 150
	case strings.Contains(model, "m3"):
		cc.MemoryBandwidth = 100
	case strings.Contains(model, "m2 ultra"):
		cc.MemoryBandwidth = 800
	case strings.Contains(model, "m2 max"):
		cc.MemoryBandwidth = 400
	case strings.Contains(model, "m2 pro"):
		cc.MemoryBandwidth = 200
	case strings.Contains(model, "m2"):
		cc.MemoryBandwidth = 100
	case strings.Contains(model, "m1 ultra"):
		cc.MemoryBandwidth = 800
	case strings.Contains(model, "m1 max"):
		cc.MemoryBandwidth = 400
	case strings.Contains(model, "m1 pro"):
		cc.MemoryBandwidth = 200
	case strings.Contains(model, "m1"):
		cc.MemoryBandwidth = 68
	}
}

// detectLinuxCompute reads /proc/cpuinfo for topology.
func detectLinuxCompute(cc *ComputeCapability) {
	cc.PhysicalCores = runtime.NumCPU()
	// On Linux, read /proc/cpuinfo for model
	if out, err := exec.Command("grep", "-m1", "model name", "/proc/cpuinfo").Output(); err == nil {
		parts := strings.SplitN(string(out), ":", 2)
		if len(parts) == 2 {
			cc.CPUModel = strings.TrimSpace(parts[1])
		}
	}
	// Linux RAM
	if out, err := exec.Command("grep", "MemTotal", "/proc/meminfo").Output(); err == nil {
		parts := strings.SplitN(string(out), ":", 2)
		if len(parts) == 2 {
			val := strings.TrimSpace(parts[1])
			val = strings.TrimSuffix(val, " kB")
			if kb, err := strconv.ParseInt(val, 10, 64); err == nil {
				cc.TotalRAMBytes = kb * 1024
			}
		}
	}
}
