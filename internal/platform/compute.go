// Package platform — compute.go
//
// Hardware compute capability detection for Apple Silicon, NVIDIA, and generic x86.
// Detects CPU topology (P-cores, E-cores, SMT), GPU, ANE (Apple Neural Engine),
// and memory bandwidth for optimal thread pool sizing and accelerator routing.
package platform

import (
	"runtime"
	"strconv"
	"strings"
	"sync"
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
	return DetectComputeWith(Current())
}

// DetectComputeWith is the injectable version of DetectCompute.
func DetectComputeWith(p Platform) *ComputeCapability {
	cc := &ComputeCapability{
		LogicalCores: runtime.NumCPU(),
	}

	switch p.Name() {
	case "darwin":
		detectDarwinCompute(p, cc)
	case "linux":
		detectLinuxCompute(p, cc)
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
// All system queries run concurrently on dedicated OS threads.
func detectDarwinCompute(p Platform, cc *ComputeCapability) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Helper to run sysctl and parse int result
	sysctlInt := func(key string, target *int) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			if out, err := p.Command("sysctl", "-n", key); err == nil {
				val, _ := strconv.Atoi(strings.TrimSpace(string(out)))
				mu.Lock()
				*target = val
				mu.Unlock()
			}
		}()
	}

	// CPU model (string)
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		if out, err := p.Command("sysctl", "-n", "machdep.cpu.brand_string"); err == nil {
			mu.Lock()
			cc.CPUModel = strings.TrimSpace(string(out))
			mu.Unlock()
		}
	}()

	// Total RAM (int64)
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		if out, err := p.Command("sysctl", "-n", "hw.memsize"); err == nil {
			val, _ := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
			mu.Lock()
			cc.TotalRAMBytes = val
			mu.Unlock()
		}
	}()

	// Integer sysctl probes — all concurrent
	sysctlInt("hw.physicalcpu", &cc.PhysicalCores)
	sysctlInt("hw.perflevel0.logicalcpu", &cc.PCores)
	sysctlInt("hw.perflevel1.logicalcpu", &cc.ECores)

	// ANE detection (lightweight sysctl probe — replaces ioreg -l -w0 which dumped 8MB+)
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		if out, err := p.Command("sysctl", "-n", "hw.optional.ane"); err == nil {
			val := strings.TrimSpace(string(out))
			if val == "1" {
				mu.Lock()
				cc.ANEAvailable = true
				cc.ANECores = 16
				mu.Unlock()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		if out, err := p.Command("system_profiler", "SPDisplaysDataType", "-detailLevel", "mini"); err == nil {
			outStr := string(out)
			for _, line := range strings.Split(outStr, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "Total Number of Cores:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						val, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
						mu.Lock()
						cc.GPUCores = val
						mu.Unlock()
					}
				}
				if strings.HasPrefix(line, "Chipset Model:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						mu.Lock()
						cc.GPUModel = strings.TrimSpace(parts[1])
						mu.Unlock()
					}
				}
			}
		}
	}()

	wg.Wait()

	// Post-processing (depends on CPU model being set)
	if !cc.ANEAvailable && strings.Contains(cc.CPUModel, "Apple") {
		cc.ANEAvailable = true
		cc.ANECores = 16
	}

	if strings.Contains(cc.CPUModel, "Apple") {
		cc.UnifiedMemory = true
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
func detectLinuxCompute(p Platform, cc *ComputeCapability) {
	cc.PhysicalCores = runtime.NumCPU()
	// On Linux, read /proc/cpuinfo for model
	if out, err := p.Command("grep", "-m1", "model name", "/proc/cpuinfo"); err == nil {
		parts := strings.SplitN(string(out), ":", 2)
		if len(parts) == 2 {
			cc.CPUModel = strings.TrimSpace(parts[1])
		}
	}
	// Linux RAM
	if out, err := p.Command("grep", "MemTotal", "/proc/meminfo"); err == nil {
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
