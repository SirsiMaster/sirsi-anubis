// Package output — Pantheon TUI
//
// Tab-based interface inspired by Mole (mole.fit). Each deity group gets
// its own page, its own job. No split panes, no REPL. Guided navigation
// with numbered actions. Press a number to act. Press esc to go back.
package output

import (
	"bufio"
	"context"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/deity"
	"github.com/SirsiMaster/sirsi-pantheon/internal/guard"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal/rules"
	"github.com/SirsiMaster/sirsi-pantheon/internal/ka"
	"github.com/SirsiMaster/sirsi-pantheon/internal/notify"
	"github.com/SirsiMaster/sirsi-pantheon/internal/osiris"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seba"
	"github.com/SirsiMaster/sirsi-pantheon/internal/stele"
	"github.com/SirsiMaster/sirsi-pantheon/internal/suggest"
	"github.com/SirsiMaster/sirsi-pantheon/internal/vitals"
)

// ── Tab Definitions ──────────────────────────────────────────────────
// Five tabs, like Mole's five planets. Each has a purpose, a voice,
// and numbered actions. The user never types a command — they press
// a number.

// nativeResult is returned by native deity calls.
type nativeResult struct {
	lines    []string // rendered output lines
	deityKey string   // which deity ran
	fixCmds  []string // actionable fix commands (override suggest engine)
	err      error
}

type nativeResultMsg nativeResult

// tabAction defines one action on a tab. Native functions return
// (rendered lines, deityKey, fixCmds, error). fixCmds override
// the suggest engine — these become the numbered "What's Next" items.
type tabAction struct {
	Label  string
	Desc   string
	Args   []string                                   // CLI args (fallback if Native is nil)
	Native func() ([]string, string, []string, error) // (lines, deityKey, fixCmds, err)
}

type tabDef struct {
	Name    string
	Glyph   string
	Tagline string
	Actions []tabAction
}

var tabs = []tabDef{
	{
		Name:    "Scan",
		Glyph:   "𓃣",
		Tagline: "Weigh what lingers. Purge what wastes.",
		Actions: []tabAction{
			{"Scan", "Find infrastructure waste on this machine", []string{"anubis", "weigh"}, nativeScan},
			{"Ghosts", "Hunt remnants of uninstalled apps", []string{"anubis", "ka"}, nativeGhosts},
			{"Clean", "Review and remove safe items", []string{"anubis", "judge", "--dry-run"}, nil},
			{"Duplicates", "Find duplicate files across directories", []string{"anubis", "mirror"}, nil},
		},
	},
	{
		Name:    "Health",
		Glyph:   "𓁐",
		Tagline: "Every system breaks. Not every system heals.",
		Actions: []tabAction{
			{"Doctor", "Full system health diagnostic", []string{"doctor"}, nativeDoctor},
			{"Network", "Network security posture audit", []string{"isis", "network"}, nativeNetworkAudit},
			{"Fix Network", "Auto-fix DNS, firewall, and security", []string{"isis", "network", "--fix"}, nativeNetworkFix},
			{"Guard", "Monitor processes and RAM pressure", []string{"guard"}, nil},
		},
	},
	{
		Name:    "Quality",
		Glyph:   "𓆄",
		Tagline: "The feather weighs against the heart.",
		Actions: []tabAction{
			{"Audit", "Governance and code quality scan", []string{"maat", "audit"}, nil},
			{"Risk", "Uncommitted work risk assessment", []string{"osiris", "assess"}, nativeRisk},
			{"Lint", "Run linters across the codebase", []string{"ra", "lint"}, nil},
			{"Test", "Run test suites fleet-wide", []string{"ra", "test"}, nil},
		},
	},
	{
		Name:    "Intel",
		Glyph:   "𓇽",
		Tagline: "Map the terrain before you march.",
		Actions: []tabAction{
			{"Hardware", "Accelerator and architecture profile", []string{"seba", "hardware"}, nativeHardware},
			{"Diagram", "Generate architecture diagrams", []string{"seba", "diagram"}, nil},
			{"Knowledge", "Ingest knowledge from sources", []string{"seshat", "ingest"}, nil},
			{"Memory", "Sync project memory state", []string{"thoth", "sync"}, nil},
		},
	},
	{
		Name:    "Status",
		Glyph:   "𓂀",
		Tagline: "It never closes its eyes. Every heartbeat, in its light.",
		Actions: []tabAction{
			{"Refresh", "Refresh system vitals", []string{"doctor"}, nil},
			{"Ra Status", "Fleet orchestrator status", []string{"ra", "status"}, nil},
			{"Code Graph", "Build code symbol index", []string{"horus", "scan"}, nil},
		},
	},
}

// nativeCommands maps suggest command strings to native functions.
// When a post-run suggestion matches one of these, it runs natively
// instead of shelling out to a subprocess.
var nativeCommands = map[string]func() ([]string, string, []string, error){
	"anubis weigh":       nativeScan,
	"anubis ka":          nativeGhosts,
	"seba hardware":      nativeHardware,
	"osiris assess":      nativeRisk,
	"findings":           nativeFindings,
	"scan":               nativeScan,
	"doctor":             nativeDoctor,
	"isis network":       nativeNetworkAudit,
	"isis network --fix": nativeNetworkFix,
}

// ── Native Deity Functions ───────────────────────────────────────────

func nativeScan() ([]string, string, []string, error) {
	engine := jackal.DefaultEngine()
	engine.RegisterAll(rules.AllRules()...)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	res, err := engine.Scan(ctx, jackal.ScanOptions{})
	if err != nil {
		return nil, "anubis", nil, err
	}
	jackal.EnrichAdvisory(res)
	_ = jackal.Persist(res, 0)
	return RenderScanResult(res), "anubis", nil, nil
}

func nativeGhosts() ([]string, string, []string, error) {
	scanner := ka.NewScanner()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ghosts, err := scanner.Scan(ctx, false)
	if err != nil {
		return nil, "anubis", nil, err
	}
	return RenderGhostResult(ghosts), "anubis", nil, nil
}

func nativeHardware() ([]string, string, []string, error) {
	hw, err := seba.DetectHardware()
	if err != nil {
		return nil, "seba", nil, err
	}
	return RenderHardwareProfile(hw), "seba", nil, nil
}

func nativeFindings() ([]string, string, []string, error) {
	scan, err := jackal.LoadLatest()
	if err != nil {
		return []string{"", "  No scan results found. Press esc and run Scan first."}, "anubis", nil, nil
	}
	res := &jackal.ScanResult{
		Findings:   make([]jackal.Finding, len(scan.Findings)),
		TotalSize:  scan.TotalSize,
		RulesRan:   scan.RulesRan,
		ByCategory: make(map[jackal.Category]jackal.CategorySummary),
	}
	for i, f := range scan.Findings {
		res.Findings[i] = jackal.Finding{
			Description: f.Description,
			Path:        f.Path,
			SizeBytes:   f.SizeBytes,
			Severity:    f.Severity,
			Category:    f.Category,
			Advisory:    f.Advisory,
			CanFix:      f.CanFix,
			Remediation: f.Remediation,
		}
	}
	for cat, s := range scan.ByCategory {
		res.ByCategory[cat] = s
	}
	return RenderScanResult(res), "anubis", nil, nil
}

func nativeNetworkAudit() ([]string, string, []string, error) {
	report, err := guard.NetworkAudit()
	if err != nil {
		return nil, "isis", nil, err
	}
	lines, fixCmds := RenderNetworkAudit(report)
	return lines, "isis", fixCmds, nil
}

func nativeNetworkFix() ([]string, string, []string, error) {
	report, err := guard.NetworkAuditFix()
	if err != nil {
		return nil, "isis", nil, err
	}
	lines, _ := RenderNetworkAudit(report)
	return lines, "isis", nil, nil
}

func nativeDoctor() ([]string, string, []string, error) {
	report, err := guard.Doctor()
	if err != nil {
		return nil, "isis", nil, err
	}
	lines, fixCmds := RenderDoctorReport(report)
	return lines, "isis", fixCmds, nil
}

func nativeRisk() ([]string, string, []string, error) {
	cp, err := osiris.Assess(".")
	if err != nil {
		return nil, "osiris", nil, err
	}
	return RenderRiskAssessment(cp), "osiris", nil, nil
}

// ── View Mode ────────────────────────────────────────────────────────

type viewMode int

const (
	viewTabs    viewMode = iota // Showing a tab landing page
	viewRunning                 // Command executing
	viewDone                    // Command finished, output + next actions
	viewPrompt                  // Power-user command prompt (: key)
)

// ── Model ────────────────────────────────────────────────────────────

type TUIModel struct {
	width  int
	height int

	activeTab int      // 0-4 index into tabs
	mode      viewMode // current view state

	// Command execution
	input        textinput.Model
	viewport     viewport.Model
	outputLines  []string
	runningDeity string
	runningCmd   string
	runningArgs  []string
	cmdStartTime time.Time
	spinner      spinner.Model
	streamCh     chan string
	runningProc  *atomic.Pointer[os.Process]

	// Post-run suggestions
	postRunCmds    []string
	postRunActions []suggest.Action // full actions with descriptions
	tabIdx         int

	// History
	history    []historyEntry
	cmdHistory []string
	historyIdx int

	// State
	activeDeity map[string]bool
	deityState  map[string]deityRunState
	steleReader *stele.Reader
	quitting    bool

	// Notifications
	notifyStore       *notify.Store
	recentNotify      []notify.Notification
	notifyRefreshTime time.Time

	// System vitals
	vitals systemVitals
}

type systemVitals = vitals.Snapshot

type deityRunState = deity.RunState

const (
	stateNeverRun  = deity.StateNeverRun
	stateSucceeded = deity.StateSucceeded
	stateFailed    = deity.StateFailed
	stateHasData   = deity.StateHasData
)

type historyEntry struct {
	deity, command, output string
}

// ── Messages ─────────────────────────────────────────────────────────

type refreshMsg time.Time
type elapsedTickMsg time.Time

type streamLineMsg struct {
	line string
	done bool
	err  error
}

func refreshTick() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg { return refreshMsg(t) })
}

func elapsedTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return elapsedTickMsg(t) })
}

func waitForStreamLine(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return streamLineMsg{done: true}
		}
		return streamLineMsg{line: line}
	}
}

// ── Constructor ──────────────────────────────────────────────────────

func NewTUIModel() TUIModel {
	ti := textinput.New()
	ti.Placeholder = "type a command..."
	ti.CharLimit = 256
	ti.Prompt = "𓉴 "
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(Gold).Bold(true)
	styles.Focused.Text = lipgloss.NewStyle().Foreground(White)
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	ti.SetStyles(styles)

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(Gold)

	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(10))

	m := TUIModel{
		input:       ti,
		viewport:    vp,
		spinner:     sp,
		width:       100,
		height:      40,
		mode:        viewTabs,
		activeTab:   0,
		historyIdx:  -1,
		tabIdx:      -1,
		activeDeity: make(map[string]bool),
		deityState:  make(map[string]deityRunState),
		streamCh:    make(chan string, 100),
		runningProc: &atomic.Pointer[os.Process]{},
		steleReader: stele.NewReader("tui"),
	}
	m.refreshActive()
	return m
}

func (m TUIModel) Init() tea.Cmd {
	return tea.Batch(refreshTick())
}

// ── Update ───────────────────────────────────────────────────────────

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcViewport()
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case nativeResultMsg:
		return m.handleNativeResult(msg)

	case streamLineMsg:
		return m.handleStreamLine(msg)

	case elapsedTickMsg:
		if m.mode == viewRunning {
			return m, elapsedTick()
		}
		return m, nil

	case refreshMsg:
		m.refreshActive()
		return m, refreshTick()

	case spinner.TickMsg:
		if m.mode == viewRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// Pass through to active component
	if m.mode == viewPrompt {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

// ── Key Handling ─────────────────────────────────────────────────────

func (m TUIModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys
	switch key {
	case "ctrl+c":
		if m.mode == viewRunning {
			if proc := m.runningProc.Load(); proc != nil {
				_ = proc.Kill()
				m.runningProc.Store(nil)
				return m, nil
			}
		}
		m.quitting = true
		return m, tea.Quit

	case "q":
		if m.mode == viewTabs {
			m.quitting = true
			return m, tea.Quit
		}
	}

	switch m.mode {
	case viewTabs:
		return m.handleTabKey(key)
	case viewRunning:
		// Scroll output while running
		switch key {
		case "up", "pgup":
			m.viewport.PageUp()
		case "down", "pgdown":
			m.viewport.PageDown()
		}
		return m, nil
	case viewDone:
		return m.handleDoneKey(key)
	case viewPrompt:
		return m.handlePromptKey(key, msg)
	}

	return m, nil
}

func (m TUIModel) handleTabKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		if m.activeTab > 0 {
			m.activeTab--
		}
		return m, nil
	case "right", "l":
		if m.activeTab < len(tabs)-1 {
			m.activeTab++
		}
		return m, nil
	case "1", "2", "3", "4", "5":
		idx := int(key[0]-'0') - 1
		tab := tabs[m.activeTab]
		if idx < len(tab.Actions) {
			return m.executeAction(tab.Actions[idx])
		}
		return m, nil
	case ":":
		// Power-user command prompt
		m.mode = viewPrompt
		m.input.Focus()
		m.input.Reset()
		return m, textinput.Blink
	case "esc":
		m.quitting = true
		return m, tea.Quit
	}

	// Tab switching by first letter
	for i, tab := range tabs {
		if strings.EqualFold(key, tab.Name[:1]) {
			m.activeTab = i
			return m, nil
		}
	}

	return m, nil
}

func (m TUIModel) handleDoneKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		// Back to tab
		m.mode = viewTabs
		m.outputLines = nil
		m.postRunCmds = nil
		m.recalcViewport()
		return m, nil
	case "up", "pgup":
		m.viewport.PageUp()
		return m, nil
	case "down", "pgdown":
		m.viewport.PageDown()
		return m, nil
	case "1", "2", "3":
		idx := int(key[0]-'0') - 1
		if idx < len(m.postRunCmds) {
			cmd := m.postRunCmds[idx]
			// Check for native handlers first
			if fn, ok := nativeCommands[cmd]; ok {
				action := tabAction{
					Label:  cmd,
					Args:   strings.Fields(cmd),
					Native: fn,
				}
				return m.executeAction(action)
			}
			// Fallback to subprocess
			args := strings.Fields(cmd)
			if len(args) > 0 {
				return m.executeArgs(args)
			}
		}
		return m, nil
	case ":":
		m.mode = viewPrompt
		m.input.Focus()
		m.input.Reset()
		return m, textinput.Blink
	}
	return m, nil
}

func (m TUIModel) handlePromptKey(key string, msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.mode = viewTabs
		m.input.Blur()
		return m, nil
	case "enter":
		raw := strings.TrimSpace(m.input.Value())
		if raw == "" {
			m.mode = viewTabs
			m.input.Blur()
			return m, nil
		}
		m.input.Blur()
		args := strings.Fields(raw)
		return m.executeArgs(args)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// ── Command Execution ────────────────────────────────────────────────

func (m TUIModel) executeAction(action tabAction) (TUIModel, tea.Cmd) {
	if action.Native != nil {
		m.mode = viewRunning
		m.runningCmd = strings.Join(action.Args, " ")
		m.runningArgs = action.Args
		m.runningDeity = ""
		if len(action.Args) > 0 {
			m.runningDeity = action.Args[0]
		}
		m.cmdStartTime = time.Now()
		m.outputLines = nil
		m.postRunCmds = nil

		fn := action.Native
		return m, tea.Batch(m.spinner.Tick, elapsedTick(), func() tea.Msg {
			lines, deityKey, fixCmds, err := fn()
			return nativeResultMsg{lines: lines, deityKey: deityKey, fixCmds: fixCmds, err: err}
		})
	}
	return m.executeArgs(action.Args)
}

func (m TUIModel) executeArgs(args []string) (TUIModel, tea.Cmd) {
	m.mode = viewRunning
	m.runningCmd = strings.Join(args, " ")
	m.runningArgs = args
	m.cmdStartTime = time.Now()
	m.streamCh = make(chan string, 100)
	m.outputLines = nil
	m.postRunCmds = nil

	// Determine deity from first arg
	m.runningDeity = ""
	for _, d := range deity.Roster {
		if len(args) > 0 && args[0] == d.Key {
			m.runningDeity = d.Key
			break
		}
	}
	// Check CLI aliases
	aliases := map[string]string{
		"scan": "anubis", "ghosts": "anubis", "dedup": "anubis",
		"guard": "isis", "doctor": "isis",
	}
	if m.runningDeity == "" && len(args) > 0 {
		if d, ok := aliases[args[0]]; ok {
			m.runningDeity = d
		}
	}

	m.recalcViewport()

	exe, _ := os.Executable()
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), "SIRSI_TUI=1")

	return m, tea.Batch(m.spinner.Tick, elapsedTick(), m.runCommandStreaming(cmd))
}

func (m TUIModel) runCommandStreaming(cmd *exec.Cmd) tea.Cmd {
	ch := m.streamCh
	procPtr := m.runningProc
	return func() tea.Msg {
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			close(ch)
			return streamLineMsg{done: true, err: err}
		}
		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			close(ch)
			return streamLineMsg{done: true, err: err}
		}
		combined := io.MultiReader(stdoutPipe, stderrPipe)
		if err := cmd.Start(); err != nil {
			close(ch)
			return streamLineMsg{done: true, err: err}
		}
		procPtr.Store(cmd.Process)
		go func() {
			scanner := bufio.NewScanner(combined)
			for scanner.Scan() {
				ch <- scanner.Text()
			}
			_ = cmd.Wait()
			close(ch)
		}()
		line, ok := <-ch
		if !ok {
			return streamLineMsg{done: true}
		}
		return streamLineMsg{line: line}
	}
}

func (m TUIModel) handleStreamLine(msg streamLineMsg) (TUIModel, tea.Cmd) {
	if msg.done {
		m.mode = viewDone
		m.runningProc.Store(nil)

		if m.runningDeity != "" {
			m.activeDeity[m.runningDeity] = true
		}
		m.history = append(m.history, historyEntry{
			deity: m.runningDeity, command: m.runningCmd,
			output: strings.Join(m.outputLines, "\n"),
		})
		m.cmdHistory = deduplicateHistory(m.history)

		if msg.err != nil {
			m.outputLines = append(m.outputLines, "",
				lipgloss.NewStyle().Foreground(Red).Render("  ✗ "+msg.err.Error()))
			if m.runningDeity != "" {
				m.deityState[m.runningDeity] = stateFailed
			}
		} else {
			// Don't add "✓ Done" here — renderDone handles it
			if m.runningDeity != "" {
				state := stateSucceeded
				if m.runningDeity == "anubis" {
					if scan, err := jackal.LoadLatest(); err == nil && len(scan.Findings) > 0 {
						state = stateHasData
					}
				}
				m.deityState[m.runningDeity] = state
			}
		}

		// Build post-run suggestions
		ctx := m.buildSuggestContext()
		if msg.err != nil {
			ctx.Err = msg.err
		}
		m.postRunCmds = suggest.Commands(ctx)
		m.postRunActions = suggest.After(ctx)
		if msg.err != nil {
			m.postRunActions = suggest.OnError(ctx)
		}

		m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
		m.recalcViewport() // re-fit viewport to actual content size
		m.viewport.GotoTop()
		m.savePersistedState()

		// Record notification
		if m.notifyStore != nil && m.runningDeity != "" {
			sev := notify.SeveritySuccess
			summary := fmt.Sprintf("%s completed", m.runningCmd)
			if msg.err != nil {
				sev = notify.SeverityError
				summary = fmt.Sprintf("%s failed", m.runningCmd)
			}
			_ = m.notifyStore.Record(notify.Notification{
				Source: m.runningDeity, Action: m.runningCmd,
				Severity: sev, Summary: summary,
			})
			m.notifyRefreshTime = time.Time{}
			m.refreshNotifications()
		}

		m.runningDeity = ""
		m.runningCmd = ""
		m.runningArgs = nil
		return m, nil
	}

	m.outputLines = append(m.outputLines, "  "+msg.line)
	m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
	m.viewport.GotoBottom()
	return m, waitForStreamLine(m.streamCh)
}

// handleNativeResult processes results from native deity function calls.
func (m TUIModel) handleNativeResult(msg nativeResultMsg) (TUIModel, tea.Cmd) {
	m.mode = viewDone
	m.runningDeity = msg.deityKey

	if msg.err != nil {
		m.outputLines = []string{
			"",
			"  " + lipgloss.NewStyle().Foreground(Red).Render("✗ "+msg.err.Error()),
		}
		if msg.deityKey != "" {
			m.deityState[msg.deityKey] = stateFailed
		}
	} else {
		m.outputLines = msg.lines
		if msg.deityKey != "" {
			state := stateSucceeded
			if msg.deityKey == "anubis" {
				if scan, err := jackal.LoadLatest(); err == nil && len(scan.Findings) > 0 {
					state = stateHasData
				}
			}
			m.deityState[msg.deityKey] = state
		}
	}

	// Use fix commands from the renderer if provided (actionable results).
	// Otherwise fall back to the generic suggest engine.
	if len(msg.fixCmds) > 0 {
		m.postRunCmds = msg.fixCmds
		m.postRunActions = nil
		for _, cmd := range msg.fixCmds {
			m.postRunActions = append(m.postRunActions, suggest.Action{
				Command:     cmd,
				Description: "Fix detected issues",
			})
		}
	} else {
		ctx := m.buildSuggestContext()
		ctx.Deity = msg.deityKey
		if msg.err != nil {
			ctx.Err = msg.err
			m.postRunActions = suggest.OnError(ctx)
		} else {
			m.postRunActions = suggest.After(ctx)
		}
		m.postRunCmds = suggest.Commands(ctx)
		if msg.deityKey == "anubis" {
			if scan, loadErr := jackal.LoadLatest(); loadErr == nil {
				ctx.FindingsCount = len(scan.Findings)
			}
			m.postRunActions = suggest.After(ctx)
			m.postRunCmds = suggest.Commands(ctx)
		}
	}

	m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
	m.recalcViewport()
	m.viewport.GotoTop()
	m.savePersistedState()

	// Record notification
	if m.notifyStore != nil && msg.deityKey != "" {
		sev := notify.SeveritySuccess
		summary := fmt.Sprintf("%s completed", m.runningCmd)
		if msg.err != nil {
			sev = notify.SeverityError
			summary = fmt.Sprintf("%s failed", m.runningCmd)
		}
		_ = m.notifyStore.Record(notify.Notification{
			Source: msg.deityKey, Action: m.runningCmd,
			Severity: sev, Summary: summary,
		})
		m.notifyRefreshTime = time.Time{}
		m.refreshNotifications()
	}

	m.runningDeity = ""
	m.runningCmd = ""
	m.runningArgs = nil
	return m, nil
}

// ── View ─────────────────────────────────────────────────────────────

func (m TUIModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	var b strings.Builder
	maxW := min(m.width-2, 120)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
	divider := dim.Render(strings.Repeat("─", maxW))

	// ── Tab bar (always visible) ──
	b.WriteString("\n")
	b.WriteString(m.renderTabBar())
	b.WriteString(" " + divider + "\n")

	switch m.mode {
	case viewTabs:
		b.WriteString(m.renderTabPage())
	case viewRunning:
		b.WriteString(m.renderRunning())
	case viewDone:
		b.WriteString(m.renderDone())
	case viewPrompt:
		b.WriteString(m.renderTabPage())
		// Prompt overlay at bottom handled below
	}

	// ── Bottom bar ──
	b.WriteString(" " + divider + "\n")
	if m.mode == viewPrompt {
		b.WriteString(" " + m.input.View() + "\n")
	} else {
		b.WriteString(m.renderBottomHints() + "\n")
	}

	// Push footer to bottom
	content := b.String()
	lines := strings.Count(content, "\n")
	remaining := m.height - lines - 2
	if remaining > 0 {
		content += strings.Repeat("\n", remaining)
	}
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).
		Render(" sirsi.ai")

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// renderTabBar draws the horizontal tab switcher, Mole-style.
func (m TUIModel) renderTabBar() string {
	var parts []string
	for i, tab := range tabs {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
		if i == m.activeTab {
			style = lipgloss.NewStyle().
				Foreground(Gold).
				Bold(true).
				Underline(true)
		}
		parts = append(parts, style.Render(tab.Glyph+" "+tab.Name))
	}

	bar := "  " + lipgloss.NewStyle().Foreground(Gold).Bold(true).Render("𓉴") +
		"    " + strings.Join(parts, "    ")
	return bar + "\n"
}

// renderTabPage draws the landing page for the active tab.
func (m TUIModel) renderTabPage() string {
	tab := tabs[m.activeTab]
	gold := lipgloss.NewStyle().Foreground(Gold)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	body := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	num := lipgloss.NewStyle().Foreground(Gold).Bold(true)

	var b strings.Builder

	if tab.Name == "Status" {
		// Status tab: bento grid
		b.WriteString(m.renderStatusPage(gold, dim))
	} else {
		// Deity tab: tagline + numbered actions
		b.WriteString("\n")
		b.WriteString("  " + gold.Render(tab.Glyph+"  "+tab.Name) + "\n")
		b.WriteString("  " + lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#888888")).
			Render(tab.Tagline) + "\n")
		b.WriteString("\n\n")

		for i, action := range tab.Actions {
			b.WriteString("  " + num.Render(fmt.Sprintf(" %d ", i+1)) +
				"  " + body.Render(action.Label) + "\n")
			b.WriteString("     " + dim.Render(action.Desc) + "\n\n")
		}
	}

	return b.String()
}

// renderStatusPage renders the bento-grid vitals dashboard.
func (m TUIModel) renderStatusPage(gold, dim lipgloss.Style) string {
	var b strings.Builder
	body := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	bigNum := lipgloss.NewStyle().Foreground(White).Bold(true)
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString("\n")

	// ── Vitals cards ──
	colW := (min(m.width, 120) - 8) / 3
	if colW < 25 {
		colW = 25
	}

	// Row 1: RAM | Git | Accelerator
	ramCard := m.renderCard("RAM", fmt.Sprintf("%.0f%%", m.vitals.RAMPercent),
		m.vitals.RAMPressure, m.vitals.RAMIcon, colW)
	gitInfo := m.vitals.GitBranch
	if m.vitals.Uncommitted > 0 {
		gitInfo += fmt.Sprintf(" +%d", m.vitals.Uncommitted)
	}
	gitCard := m.renderCard("GIT", gitInfo, m.vitals.LastCommit, "🌿", colW)
	accelCard := m.renderCard("ACCEL", m.vitals.Accelerator, "", "⚡", colW)

	b.WriteString("  " + ramCard + "  " + gitCard + "  " + accelCard + "\n\n")

	// ── Waste summary ──
	if scan, err := jackal.LoadLatest(); err == nil && scan.TotalSize > 0 {
		age := time.Since(scan.Timestamp)
		ageStr := "just now"
		if age > time.Hour {
			ageStr = fmt.Sprintf("%.0fh ago", age.Hours())
		} else if age > time.Minute {
			ageStr = fmt.Sprintf("%.0fm ago", age.Minutes())
		}
		wasteIcon := "🟢"
		if scan.TotalSize > 10*1024*1024*1024 {
			wasteIcon = "🔴"
		} else if scan.TotalSize > 5*1024*1024*1024 {
			wasteIcon = "🟡"
		}
		b.WriteString("  " + label.Render("WASTE") + "\n")
		b.WriteString("  " + wasteIcon + " " +
			bigNum.Render(jackal.FormatSize(scan.TotalSize)) +
			"  " + dim.Render(fmt.Sprintf("%d findings · scanned %s", len(scan.Findings), ageStr)) + "\n\n")
	}

	// ── Deity Status ──
	b.WriteString("  " + label.Render("DEITIES") + "\n")
	for _, d := range deity.Roster {
		state := m.deityState[d.Key]
		var indicator, status string
		switch state {
		case stateSucceeded:
			indicator = lipgloss.NewStyle().Foreground(Green).Render("✓")
			status = dim.Render("healthy")
		case stateFailed:
			indicator = lipgloss.NewStyle().Foreground(Red).Render("✗")
			status = lipgloss.NewStyle().Foreground(Red).Render("failed")
		case stateHasData:
			indicator = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00")).Render("◆")
			status = gold.Render("has data")
		default:
			indicator = dim.Render("·")
			status = dim.Render("—")
		}
		b.WriteString("  " + indicator + " " +
			body.Render(d.Glyph+" "+d.Name) + "  " + status + "\n")
	}
	b.WriteString("\n")

	// ── Recent Activity ──
	if len(m.recentNotify) > 0 {
		b.WriteString("  " + label.Render("RECENT") + "\n")
		for i, n := range m.recentNotify {
			if i >= 5 {
				break
			}
			icon := notify.SeverityIcon(n.Severity)
			summary := n.Summary
			if len(summary) > 60 {
				summary = summary[:57] + "…"
			}
			b.WriteString(fmt.Sprintf("  %s %s  %s\n",
				icon, gold.Render(n.Source), dim.Render(summary)))
		}
	}

	return b.String()
}

// renderCard renders a small bento card with a label, big value, and subtitle.
func (m TUIModel) renderCard(labelText, value, subtitle, icon string, width int) string {
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	bigNum := lipgloss.NewStyle().Foreground(White).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	card := label.Render(icon+" "+labelText) + "\n" +
		"  " + bigNum.Render(value) + "\n"
	if subtitle != "" {
		card += "  " + dim.Render(subtitle) + "\n"
	}

	return lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		Padding(0, 1).
		Render(card)
}

// renderRunning shows the command execution screen.
func (m TUIModel) renderRunning() string {
	var b strings.Builder
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	glyph, name := deity.Display(m.runningDeity)
	elapsed := time.Since(m.cmdStartTime).Truncate(time.Second)
	elapsedStr := ""
	if elapsed >= time.Second {
		elapsedStr = " " + dim.Render(fmt.Sprintf("(%s)", elapsed))
	}

	b.WriteString("\n")
	b.WriteString("  " + m.spinner.View() + " " +
		lipgloss.NewStyle().Foreground(Gold).Bold(true).Render(glyph+" "+name) +
		"  " + dim.Render(m.runningCmd) + elapsedStr + "\n")
	b.WriteString("\n")
	b.WriteString(m.viewport.View() + "\n")

	return b.String()
}

// renderDone shows command output + numbered next actions.
func (m TUIModel) renderDone() string {
	var b strings.Builder
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	green := lipgloss.NewStyle().Foreground(Green)
	body := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	num := lipgloss.NewStyle().Foreground(Gold).Bold(true)

	b.WriteString("\n")
	b.WriteString(m.viewport.View() + "\n")

	// ── Completion ──
	b.WriteString("\n")
	b.WriteString("  " + green.Render("✓ Done") + "\n")
	b.WriteString("\n")

	// ── Numbered next actions ──
	shown := 0
	for i, cmd := range m.postRunCmds {
		if i >= 3 {
			break
		}
		desc := ""
		if i < len(m.postRunActions) {
			desc = m.postRunActions[i].Description
		}
		line := "   " + num.Render(fmt.Sprintf("%d", i+1)) + "  " + body.Render(cmd)
		if desc != "" {
			line += "  " + dim.Render(desc)
		}
		b.WriteString(line + "\n")
		shown++
	}

	b.WriteString("\n")
	if shown > 0 {
		b.WriteString("   " + dim.Render(fmt.Sprintf("press 1-%d to continue  ·  esc back  ·  : command", shown)) + "\n")
	} else {
		b.WriteString("   " + dim.Render("esc back  ·  : command") + "\n")
	}

	return b.String()
}

// renderBottomHints shows context-appropriate key hints.
func (m TUIModel) renderBottomHints() string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	var hints []string

	switch m.mode {
	case viewTabs:
		hints = []string{"←/→ switch tabs", "1-5 act", ": command", "q quit"}
	case viewRunning:
		hints = []string{"↑/↓ scroll", "ctrl+c cancel"}
	case viewDone:
		if len(m.postRunCmds) > 0 {
			hints = []string{"1-3 next action", "↑/↓ scroll", ": command", "esc back"}
		} else {
			hints = []string{"↑/↓ scroll", ": command", "esc back"}
		}
	}

	return " " + dim.Render(strings.Join(hints, "  ·  "))
}

// ── Suggest Context ──────────────────────────────────────────────────

func (m TUIModel) buildSuggestContext() suggest.Context {
	sub := ""
	if len(m.runningArgs) >= 2 {
		sub = m.runningArgs[1]
	}
	ctx := suggest.Context{
		Deity:      m.runningDeity,
		Subcommand: sub,
	}
	if m.runningDeity == "anubis" {
		if scan, err := jackal.LoadLatest(); err == nil {
			ctx.FindingsCount = len(scan.Findings)
		}
	}
	return ctx
}

// ── Background Refresh ───────────────────────────────────────────────

func (m *TUIModel) refreshActive() {
	for k := range m.activeDeity {
		delete(m.activeDeity, k)
	}
	entries, _ := m.steleReader.ReadNew()
	now := time.Now()
	for _, e := range entries {
		ts, err := time.Parse(time.RFC3339, e.TS)
		if err != nil {
			continue
		}
		if now.Sub(ts) < 5*time.Minute {
			dKey := strings.ToLower(e.Deity)
			if !strings.Contains(dKey, ":") {
				m.activeDeity[dKey] = true
			}
		}
	}
	m.refreshNotifications()
	m.refreshVitals()

	home, _ := os.UserHomeDir()
	pidDir := filepath.Join(home, ".config", "ra", "pids")
	pidEntries, _ := os.ReadDir(pidDir)
	for _, f := range pidEntries {
		if f.IsDir() {
			continue
		}
		name := strings.TrimSuffix(f.Name(), ".pid")
		for _, d := range deity.Roster {
			if strings.Contains(strings.ToLower(name), d.Key) {
				m.activeDeity[d.Key] = true
			}
		}
	}
}

func (m *TUIModel) refreshVitals() {
	m.vitals = vitals.Collect()
}

func (m *TUIModel) refreshNotifications() {
	if m.notifyStore == nil {
		return
	}
	now := time.Now()
	if now.Sub(m.notifyRefreshTime) < 5*time.Second {
		return
	}
	m.notifyRefreshTime = now
	items, err := m.notifyStore.Recent(5)
	if err != nil {
		return
	}
	m.recentNotify = items
}

// ── Layout ───────────────────────────────────────────────────────────

func (m *TUIModel) recalcViewport() {
	// Reserve: tab bar(2) + divider(1) + running header(2) + bottom divider(1) + hints(1) + padding(1)
	vpHeight := m.height - 8
	if m.mode == viewDone && len(m.postRunCmds) > 0 {
		shown := min(len(m.postRunCmds), 3)
		vpHeight -= (shown * 3) + 3 // each action ~3 lines + header + spacing
	}
	// Cap viewport to content size — don't waste space with blank lines
	if len(m.outputLines) > 0 && len(m.outputLines) < vpHeight {
		vpHeight = len(m.outputLines) + 1
	}
	if vpHeight < 3 {
		vpHeight = 3
	}
	m.viewport.SetHeight(vpHeight)

	vpWidth := m.width - 4
	if vpWidth < 20 {
		vpWidth = 20
	}
	m.viewport.SetWidth(vpWidth)
}

// ── Helpers ──────────────────────────────────────────────────────────

func pluralize(word string, n int) string {
	if n == 1 {
		return word
	}
	if strings.HasSuffix(word, "y") {
		return word[:len(word)-1] + "ies"
	}
	return word + "s"
}

// ── Launcher ─────────────────────────────────────────────────────────

func LaunchTUI() error {
	return LaunchTUIWithNotify(nil)
}

func LaunchTUIWithNotify(store *notify.Store) error {
	m := NewTUIModel()
	m.notifyStore = store
	m.refreshNotifications()
	m.loadPersistedState()
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

// ── Persistent State ─────────────────────────────────────────────────

func (m *TUIModel) loadPersistedState() {
	state, err := deity.LoadState()
	if err != nil {
		return
	}
	for k, v := range state.DeityState {
		m.deityState[k] = v
	}
}

func (m *TUIModel) savePersistedState() {
	_ = deity.SaveState(deity.PersistedState{DeityState: m.deityState})
}

// ── Unused but required by tests ─────────────────────────────────────
// These are no-ops preserved for test compilation. The old REPL functions
// (showFindings, showHelp, renderRosterColumns, etc.) are removed.

var _ = color.RGBA{} // keep image/color import for lipgloss
