package guard

import (
	"strings"
	"testing"

	"github.com/SirsiMaster/sirsi-pantheon/internal/platform"
)

// ── Helpers ──────────────────────────────────────────────────────────────

// healthyMock returns a Mock platform simulating a healthy macOS system:
// 16 GB RAM, low usage, no swap, 50% disk, small processes, no pantheon procs.
func healthyMock() *platform.Mock {
	return &platform.Mock{
		NameStr: "mock",
		CommandResults: map[string]string{
			// 16 GB total RAM
			"sysctl -n hw.memsize": "17179869184",

			// vm_stat: ~30% used (active + wired)
			// page size 16384
			// active:  200000 pages  = 3.2 GB
			// wired:   100000 pages  = 1.6 GB  => total used ~4.8 GB / 16 GB = 30%
			// free:    400000 pages
			// compressed: 50000 pages
			"vm_stat": `Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free:                              400000.
Pages active:                            200000.
Pages inactive:                          100000.
Pages speculative:                        50000.
Pages throttled:                              0.
Pages wired down:                        100000.
Pages purgeable:                          20000.
"Translation faults":                  12345678.
Pages copy-on-write:                    1234567.
Pages zero filled:                      9876543.
Pages reactivated:                        12345.
Pages purged:                              6789.
File-backed pages:                       150000.
Anonymous pages:                         200000.
Pages stored in compressor:               50000.
Pages occupied by compressor:             25000.`,

			// No swap
			"sysctl -n vm.swapusage": "total = 0.00M  used = 0.00M  free = 0.00M  (encrypted)",

			// Disk at 50%
			"df -h /": `Filesystem     Size   Used  Avail Capacity  iused ifree %iused  Mounted on
/dev/disk3s1  460Gi  230Gi  230Gi    50%  1234567 9876543    11%   /`,

			// ps for getProcessListWith (used by checkTopMemoryProcesses)
			"ps -axo pid,rss,vsz,%cpu,user,comm": `  PID   RSS    VSZ  %CPU USER     COMM
  100  51200  102400  1.0 user     /usr/bin/node
  200  30720   61440  0.5 user     /usr/local/bin/gopls
  300  20480   40960  0.2 user     /Applications/Safari.app/Contents/MacOS/Safari
  400  10240   20480  0.1 user     /usr/bin/vim
  500   5120   10240  0.0 user     /bin/zsh`,

			// ps for checkPantheonProcesses
			"ps -axo pid,rss,comm": `  PID   RSS COMM
  100  51200 /usr/bin/node
  200  30720 /usr/local/bin/gopls`,
		},
	}
}

// ── TestDoctorWith_HealthySystem ─────────────────────────────────────────

func TestDoctorWith_HealthySystem(t *testing.T) {
	m := healthyMock()
	report, err := DoctorWith(m)
	if err != nil {
		t.Fatalf("DoctorWith() error = %v", err)
	}

	// Note: checkRecentCrashLogs reads the real filesystem (not mocked),
	// so the score and severity of crash/jetsam findings depend on the host.
	// We only assert on the mock-controlled checks.

	// RAM should be OK
	ramFinding := findByCheck(report.Findings, "RAM Pressure")
	if ramFinding == nil {
		t.Fatal("missing RAM Pressure finding")
	}
	if ramFinding.Severity != SeverityOK {
		t.Errorf("RAM Pressure severity = %v, want OK", ramFinding.Severity)
	}

	// Swap should be OK
	swapFinding := findByCheck(report.Findings, "Swap Usage")
	if swapFinding == nil {
		t.Fatal("missing Swap Usage finding")
	}
	if swapFinding.Severity != SeverityOK {
		t.Errorf("Swap Usage severity = %v, want OK", swapFinding.Severity)
	}
	if !strings.Contains(swapFinding.Message, "No swap") {
		t.Errorf("Swap message = %q, want 'No swap' substring", swapFinding.Message)
	}

	// Disk should be OK
	diskFinding := findByCheck(report.Findings, "Disk Space")
	if diskFinding == nil {
		t.Fatal("missing Disk Space finding")
	}
	if diskFinding.Severity != SeverityOK {
		t.Errorf("Disk Space severity = %v, want OK", diskFinding.Severity)
	}

	// Pantheon processes should be Info (none running)
	pantheonFinding := findByCheck(report.Findings, "Pantheon Processes")
	if pantheonFinding == nil {
		t.Fatal("missing Pantheon Processes finding")
	}
	if pantheonFinding.Severity != SeverityInfo {
		t.Errorf("Pantheon Processes severity = %v, want Info", pantheonFinding.Severity)
	}

	// Duration should be populated
	if report.Duration == "" {
		t.Error("Duration is empty")
	}
}

// ── TestDoctorWith_HighRAMPressure ───────────────────────────────────────

func TestDoctorWith_HighRAMPressure(t *testing.T) {
	m := healthyMock()

	// Override vm_stat: active + wired > 90% of 16 GB
	// 16 GB = 17179869184 bytes / 16384 page_size = 1048576 total pages
	// active 600000 pages = 9.83 GB, wired 400000 pages = 6.55 GB => 16.38 GB / 16 GB = ~96%
	m.CommandResults["vm_stat"] = `Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free:                               10000.
Pages active:                            600000.
Pages inactive:                           20000.
Pages speculative:                         5000.
Pages throttled:                              0.
Pages wired down:                        400000.
Pages purgeable:                           1000.
"Translation faults":                  12345678.
Pages copy-on-write:                    1234567.
Pages zero filled:                      9876543.
Pages reactivated:                        12345.
Pages purged:                              6789.
File-backed pages:                       150000.
Anonymous pages:                         200000.
Pages stored in compressor:               80000.
Pages occupied by compressor:             40000.`

	report, err := DoctorWith(m)
	if err != nil {
		t.Fatalf("DoctorWith() error = %v", err)
	}

	ramFinding := findByCheck(report.Findings, "RAM Pressure")
	if ramFinding == nil {
		t.Fatal("missing RAM Pressure finding")
	}
	if ramFinding.Severity != SeverityCritical {
		t.Errorf("RAM Pressure severity = %v, want CRITICAL", ramFinding.Severity)
	}
	if !strings.Contains(ramFinding.Message, "critically high") {
		t.Errorf("RAM message = %q, want 'critically high' substring", ramFinding.Message)
	}

	// Score should be penalized
	if report.Score > 80 {
		t.Errorf("high RAM pressure score = %d, want <= 80", report.Score)
	}
}

// ── TestDoctorWith_SwapActive ────────────────────────────────────────────

func TestDoctorWith_SwapActive(t *testing.T) {
	tests := []struct {
		name          string
		swapOutput    string
		wantSeverity  DiagnosticSeverity
		wantSubstring string
	}{
		{
			name:          "moderate swap with small total",
			swapOutput:    "total = 512.00M  used = 256.00M  free = 256.00M  (encrypted)",
			wantSeverity:  SeverityWarn,
			wantSubstring: "Swap active",
		},
		{
			// The parser checks ALL numeric tokens in the line for > 1000.
			// "total = 2048.00M" alone triggers CRITICAL since 2048 > 1000.
			name:          "heavy swap — total exceeds 1000M",
			swapOutput:    "total = 4096.00M  used = 2048.00M  free = 2048.00M  (encrypted)",
			wantSeverity:  SeverityCritical,
			wantSubstring: "Heavy swapping",
		},
		{
			name:          "no swap",
			swapOutput:    "total = 0.00M  used = 0.00M  free = 0.00M  (encrypted)",
			wantSeverity:  SeverityOK,
			wantSubstring: "No swap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := healthyMock()
			m.CommandResults["sysctl -n vm.swapusage"] = tt.swapOutput

			report, err := DoctorWith(m)
			if err != nil {
				t.Fatalf("DoctorWith() error = %v", err)
			}

			swapFinding := findByCheck(report.Findings, "Swap Usage")
			if swapFinding == nil {
				t.Fatal("missing Swap Usage finding")
			}
			if swapFinding.Severity != tt.wantSeverity {
				t.Errorf("Swap severity = %v, want %v", swapFinding.Severity, tt.wantSeverity)
			}
			if !strings.Contains(swapFinding.Message, tt.wantSubstring) {
				t.Errorf("Swap message = %q, want %q substring", swapFinding.Message, tt.wantSubstring)
			}
		})
	}
}

// ── TestDoctorWith_DiskFull ──────────────────────────────────────────────

func TestDoctorWith_DiskFull(t *testing.T) {
	tests := []struct {
		name         string
		dfOutput     string
		wantSeverity DiagnosticSeverity
		wantSubstr   string
	}{
		{
			name: "critically full 97%",
			dfOutput: `Filesystem     Size   Used  Avail Capacity  iused ifree %iused  Mounted on
/dev/disk3s1  460Gi  447Gi   13Gi    97%  1234567 9876543    11%   /`,
			wantSeverity: SeverityCritical,
			wantSubstr:   "critically full",
		},
		{
			name: "high usage 90%",
			dfOutput: `Filesystem     Size   Used  Avail Capacity  iused ifree %iused  Mounted on
/dev/disk3s1  460Gi  414Gi   46Gi    90%  1234567 9876543    11%   /`,
			wantSeverity: SeverityWarn,
			wantSubstr:   "high",
		},
		{
			name: "healthy 50%",
			dfOutput: `Filesystem     Size   Used  Avail Capacity  iused ifree %iused  Mounted on
/dev/disk3s1  460Gi  230Gi  230Gi    50%  1234567 9876543    11%   /`,
			wantSeverity: SeverityOK,
			wantSubstr:   "healthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := healthyMock()
			m.CommandResults["df -h /"] = tt.dfOutput

			report, err := DoctorWith(m)
			if err != nil {
				t.Fatalf("DoctorWith() error = %v", err)
			}

			diskFinding := findByCheck(report.Findings, "Disk Space")
			if diskFinding == nil {
				t.Fatal("missing Disk Space finding")
			}
			if diskFinding.Severity != tt.wantSeverity {
				t.Errorf("Disk severity = %v, want %v", diskFinding.Severity, tt.wantSeverity)
			}
			if !strings.Contains(diskFinding.Message, tt.wantSubstr) {
				t.Errorf("Disk message = %q, want %q substring", diskFinding.Message, tt.wantSubstr)
			}
		})
	}
}

// ── TestCalculateScore ───────────────────────────────────────────────────

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name     string
		findings []DiagnosticFinding
		want     int
	}{
		{
			name:     "no findings",
			findings: nil,
			want:     100,
		},
		{
			name: "all OK",
			findings: []DiagnosticFinding{
				{Severity: SeverityOK},
				{Severity: SeverityOK},
				{Severity: SeverityOK},
			},
			want: 100,
		},
		{
			name: "one info",
			findings: []DiagnosticFinding{
				{Severity: SeverityInfo},
			},
			want: 98,
		},
		{
			name: "one warn",
			findings: []DiagnosticFinding{
				{Severity: SeverityWarn},
			},
			want: 90,
		},
		{
			name: "one critical",
			findings: []DiagnosticFinding{
				{Severity: SeverityCritical},
			},
			want: 80,
		},
		{
			name: "mixed severities",
			findings: []DiagnosticFinding{
				{Severity: SeverityOK},
				{Severity: SeverityInfo},
				{Severity: SeverityWarn},
				{Severity: SeverityCritical},
			},
			want: 68, // 100 - 0 - 2 - 10 - 20
		},
		{
			name: "floors at zero",
			findings: []DiagnosticFinding{
				{Severity: SeverityCritical},
				{Severity: SeverityCritical},
				{Severity: SeverityCritical},
				{Severity: SeverityCritical},
				{Severity: SeverityCritical},
				{Severity: SeverityCritical}, // 6 * 20 = 120, exceeds 100
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateScore(tt.findings)
			if got != tt.want {
				t.Errorf("calculateScore() = %d, want %d", got, tt.want)
			}
		})
	}
}

// ── TestDiagnosticSeverity_Icon ──────────────────────────────────────────

func TestDiagnosticSeverity_Icon(t *testing.T) {
	tests := []struct {
		severity DiagnosticSeverity
		want     string
	}{
		{SeverityOK, "🟢"},
		{SeverityInfo, "🔵"},
		{SeverityWarn, "🟡"},
		{SeverityCritical, "🔴"},
		{DiagnosticSeverity(99), "⚪"}, // unknown
	}

	for _, tt := range tests {
		t.Run(tt.severity.String(), func(t *testing.T) {
			got := tt.severity.Icon()
			if got != tt.want {
				t.Errorf("DiagnosticSeverity(%d).Icon() = %q, want %q", tt.severity, got, tt.want)
			}
		})
	}
}

// ── TestDiagnosticSeverity_String ────────────────────────────────────────

func TestDiagnosticSeverity_String(t *testing.T) {
	tests := []struct {
		severity DiagnosticSeverity
		want     string
	}{
		{SeverityOK, "OK"},
		{SeverityInfo, "INFO"},
		{SeverityWarn, "WARN"},
		{SeverityCritical, "CRITICAL"},
		{DiagnosticSeverity(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.severity.String()
			if got != tt.want {
				t.Errorf("DiagnosticSeverity(%d).String() = %q, want %q", tt.severity, got, tt.want)
			}
		})
	}
}

// ── TestDoctorWith_MemoryHog ─────────────────────────────────────────────

func TestDoctorWith_MemoryHog(t *testing.T) {
	m := healthyMock()

	// Inject a process using > 4 GB RSS (4194304 KB = 4 GB in KB for ps output)
	m.CommandResults["ps -axo pid,rss,vsz,%cpu,user,comm"] = `  PID   RSS    VSZ  %CPU USER     COMM
  100  5242880  10485760  5.0 user     /usr/bin/node
  200  30720   61440  0.5 user     /usr/local/bin/gopls`

	report, err := DoctorWith(m)
	if err != nil {
		t.Fatalf("DoctorWith() error = %v", err)
	}

	memFinding := findByCheck(report.Findings, "Top Memory Consumers")
	if memFinding == nil {
		t.Fatal("missing Top Memory Consumers finding")
	}
	if memFinding.Severity != SeverityWarn {
		t.Errorf("Top Memory severity = %v, want WARN for >4GB process", memFinding.Severity)
	}
	if !strings.Contains(memFinding.Message, "Memory hog") {
		t.Errorf("message = %q, want 'Memory hog' substring", memFinding.Message)
	}
}

// ── TestDoctorWith_PantheonProcesses ─────────────────────────────────────

func TestDoctorWith_PantheonProcesses(t *testing.T) {
	m := healthyMock()

	// Inject pantheon processes in ps output
	m.CommandResults["ps -axo pid,rss,comm"] = `  PID   RSS COMM
  100  51200 /usr/bin/node
  900  20480 /usr/local/bin/pantheon-agent
  901  10240 /usr/local/bin/pantheon-guard`

	report, err := DoctorWith(m)
	if err != nil {
		t.Fatalf("DoctorWith() error = %v", err)
	}

	pantheonFinding := findByCheck(report.Findings, "Pantheon Processes")
	if pantheonFinding == nil {
		t.Fatal("missing Pantheon Processes finding")
	}
	if pantheonFinding.Severity != SeverityOK {
		t.Errorf("Pantheon severity = %v, want OK for small processes", pantheonFinding.Severity)
	}
	if !strings.Contains(pantheonFinding.Message, "2 Pantheon process") {
		t.Errorf("message = %q, want '2 Pantheon process' substring", pantheonFinding.Message)
	}
}

// ── TestDoctorWith_WarnRAM ───────────────────────────────────────────────

func TestDoctorWith_WarnRAM(t *testing.T) {
	m := healthyMock()

	// Set RAM usage to ~80% (between 75-90 triggers WARN)
	// active 400000 + wired 400000 = 800000 pages * 16384 = 13.1 GB / 16 GB = ~82%
	m.CommandResults["vm_stat"] = `Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free:                               50000.
Pages active:                            400000.
Pages inactive:                           20000.
Pages speculative:                         5000.
Pages throttled:                              0.
Pages wired down:                        400000.
Pages purgeable:                           1000.
"Translation faults":                  12345678.
Pages copy-on-write:                    1234567.
Pages zero filled:                      9876543.
Pages reactivated:                        12345.
Pages purged:                              6789.
File-backed pages:                       150000.
Anonymous pages:                         200000.
Pages stored in compressor:               50000.
Pages occupied by compressor:             25000.`

	report, err := DoctorWith(m)
	if err != nil {
		t.Fatalf("DoctorWith() error = %v", err)
	}

	ramFinding := findByCheck(report.Findings, "RAM Pressure")
	if ramFinding == nil {
		t.Fatal("missing RAM Pressure finding")
	}
	if ramFinding.Severity != SeverityWarn {
		t.Errorf("RAM severity = %v, want WARN for ~80%% usage", ramFinding.Severity)
	}
	if !strings.Contains(ramFinding.Message, "elevated") {
		t.Errorf("RAM message = %q, want 'elevated' substring", ramFinding.Message)
	}
}

// ── helpers ──────────────────────────────────────────────────────────────

func findByCheck(findings []DiagnosticFinding, check string) *DiagnosticFinding {
	for i := range findings {
		if findings[i].Check == check {
			return &findings[i]
		}
	}
	return nil
}
