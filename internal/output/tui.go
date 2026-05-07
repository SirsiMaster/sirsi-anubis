// Package output — Pantheon TUI
//
// The primary interface for Pantheon. When the user types `pantheon` with no
// subcommand, this TUI launches. It is a persistent session: commands execute
// inside the TUI, output streams into a viewport, and the input bar re-enables
// when the command completes. The user stays in Pantheon until they explicitly quit.
package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/notify"
	"github.com/SirsiMaster/sirsi-pantheon/internal/stele"
)

// ── Deity Definitions ─────────────────────────────────────────────────

type deityInfo struct {
	Key   string
	Glyph string
	Name  string
	Role  string // Short role (shown in roster)
}

// Canonical deity roster — ordered by hierarchy (Rule D6).
// Two-word roles: verb/adjective + noun. Fits in a 30-col grid cell.
var deityRoster = []deityInfo{
	{"horus", "𓂀", "Horus", "Workstation Lord"},
	{"ra", "𓇶", "Ra", "Fleet Orchestrator"},
	{"net", "𓁯", "Net", "Universal Weaver"},
	{"thoth", "𓁟", "Thoth", "Local Memory"},
	{"maat", "𓆄", "Ma'at", "Quality Gate"},
	{"isis", "𓁐", "Isis", "Health & Remedy"},
	{"seshat", "𓁆", "Seshat", "Local Knowledge"},
	{"anubis", "𓃣", "Anubis", "System Jackal"},
	{"seba", "𓇽", "Seba", "Infra & Hardware"},
	{"osiris", "𓁹", "Osiris", "State Keeper"},
}

// intentKeywords maps natural-language keywords to deity keys for routing.
var intentKeywords = map[string][]string{
	"ra":     {"deploy", "orchestrate", "sprint", "agent", "watch", "command center"},
	"net":    {"scope", "weave", "context", "canon", "align", "tile", "drift"},
	"thoth":  {"memory", "sync", "compact", "journal", "remember", "persist"},
	"maat":   {"quality", "audit", "coverage", "test", "lint", "feather", "gate", "qa"},
	"isis":   {"fix", "heal", "remediate", "repair", "auto-fix", "guard", "watchdog", "monitor", "ram", "cpu", "doctor", "process", "network", "dns", "wifi", "firewall", "tls", "vpn", "security"},
	"seshat": {"knowledge", "graft", "ingest", "notes", "gemini", "notebooklm"},
	"anubis": {"scan", "waste", "clean", "judge", "purge", "hygiene", "infrastructure", "dedup", "duplicate", "mirror", "ghost", "dead", "remnant", "uninstall", "residual", "haunt"},
	"seba":   {"architecture", "topology", "diagram", "map", "dependency", "graph", "network map", "network topology", "fleet", "subnet", "container", "docker", "kubernetes", "k8s", "pod", "gpu", "vram", "hardware", "accelerator", "ane", "cuda", "metal", "npu", "profile"},
	"osiris": {"checkpoint", "state", "preserve", "restore", "uncommitted", "risk", "drift", "snapshot", "commit status"},
	"horus":  {"code graph", "symbols", "outline", "declarations", "code index"},
}

// Top-level CLI aliases that bypass intent matching.
// These map user shorthand to the deity that owns the verb.
var cliAliases = map[string]string{
	"scan":    "anubis",
	"ghosts":  "anubis",
	"dedup":   "anubis",
	"guard":   "isis",
	"doctor":  "isis",
	"version": "version",
}

// ── TUI State ─────────────────────────────────────────────────────────

type tuiMode int

const (
	modeIdle tuiMode = iota
	modeRunning
)

// deityRunState tracks the last-run outcome for a deity.
type deityRunState int

const (
	stateNeverRun  deityRunState = iota
	stateSucceeded               // last run completed successfully
	stateFailed                  // last run had an error
	stateHasData                 // has actionable data (e.g. Anubis findings)
)

type TUIModel struct {
	width  int
	height int

	input    textinput.Model
	viewport viewport.Model

	outputLines  []string
	mode         tuiMode
	runningDeity string
	runningCmd   string
	runningArgs  []string  // dispatched CLI args (e.g. ["anubis", "weigh"])
	cmdStartTime time.Time // when the current command started
	spinner      spinner.Model
	history      []historyEntry

	// View stack for back-navigation (esc pops, commands push)
	viewStack []viewFrame

	// Post-run command suggestions for tab-cycling
	postRunCmds []string // suggested commands after last run
	tabIdx      int      // -1 = not cycling; 0..len-1 = position

	// Inline predictions + history recall
	cmdHistory   []string // deduplicated command strings for up-arrow
	historyIdx   int      // -1 = not browsing; 0..len-1 = position
	historySaved string   // input text saved when user starts browsing

	activeDeity map[string]bool
	deityState  map[string]deityRunState // tracks last-run outcome per deity
	steleReader *stele.Reader
	quitting    bool

	// Streaming command output
	streamCh chan string // receives lines from running commands

	// Notification awareness
	notifyStore       *notify.Store
	recentNotify      []notify.Notification
	notifyRefreshTime time.Time
}

type historyEntry struct {
	deity, command, output string
}

// viewFrame captures a snapshot of the viewport for back-navigation.
type viewFrame struct {
	outputLines []string
	placeholder string
}

// ── Messages ──────────────────────────────────────────────────────────

type refreshMsg time.Time

// cmdBatchMsg carries the output of a completed command.
type cmdBatchMsg struct {
	lines []string
	err   error
}

// streamLineMsg carries a single line from the streaming channel.
// An empty line with done=true signals command completion.
type streamLineMsg struct {
	line string
	done bool
	err  error
}

// elapsedTickMsg triggers elapsed time updates during running commands.
type elapsedTickMsg time.Time

func refreshTick() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg { return refreshMsg(t) })
}

func elapsedTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return elapsedTickMsg(t) })
}

// waitForStreamLine returns a tea.Cmd that blocks until a line arrives on
// the stream channel. Closed channel signals done.
func waitForStreamLine(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return streamLineMsg{done: true}
		}
		return streamLineMsg{line: line}
	}
}

// ── Constructor ───────────────────────────────────────────────────────

func NewTUIModel() TUIModel {
	ti := textinput.New()
	ti.Placeholder = "scan my dev environment for ghost processes and dead symlinks"
	ti.Focus()
	ti.CharLimit = 256
	ti.SetWidth(76)
	ti.Prompt = "𓉴 "
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(Gold).Bold(true)
	styles.Focused.Text = lipgloss.NewStyle().Foreground(White)
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	styles.Focused.Suggestion = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	ti.SetStyles(styles)

	// Fish-shell-style inline predictions
	ti.ShowSuggestions = true
	ti.KeyMap.AcceptSuggestion = key.NewBinding(key.WithKeys("right"))
	ti.KeyMap.NextSuggestion = key.NewBinding() // unbind — Up is for history
	ti.KeyMap.PrevSuggestion = key.NewBinding() // unbind — Down is for history
	ti.SetSuggestions(topLevelCommands)

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
		mode:        modeIdle,
		historyIdx:  -1,
		tabIdx:      -1,
		activeDeity: make(map[string]bool),
		deityState:  make(map[string]deityRunState),
		streamCh:    make(chan string, 100),
		steleReader: stele.NewReader("tui"),
	}
	m.refreshActive()
	return m
}

func (m TUIModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, refreshTick())
}

// ── Update ────────────────────────────────────────────────────────────

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.SetWidth(min(msg.Width-8, 80))
		m.recalcViewportHeight()
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case streamLineMsg:
		return m.handleStreamLine(msg)

	case cmdBatchMsg:
		return m.handleBatchOutput(msg)

	case elapsedTickMsg:
		// Re-render to update the elapsed time display while running
		if m.mode == modeRunning {
			return m, elapsedTick()
		}
		return m, nil

	case refreshMsg:
		m.refreshActive()
		return m, refreshTick()

	case spinner.TickMsg:
		if m.mode == modeRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	if m.mode == modeIdle {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m TUIModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		if m.mode == modeRunning {
			return m, nil
		}
		// Pop the view stack if there's a previous view to restore
		if len(m.viewStack) > 0 {
			prev := m.viewStack[len(m.viewStack)-1]
			m.viewStack = m.viewStack[:len(m.viewStack)-1]
			m.outputLines = prev.outputLines
			m.input.Placeholder = prev.placeholder
			if len(m.outputLines) > 0 {
				m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
			} else {
				m.viewport.SetContent("")
			}
			m.recalcViewportHeight()
			return m, nil
		}
		// No stack — clear output or quit
		if len(m.outputLines) > 0 {
			m.outputLines = nil
			m.viewport.SetContent("")
			m.input.Placeholder = "scan my dev environment for ghost processes and dead symlinks"
			return m, nil
		}
		m.quitting = true
		return m, tea.Quit

	case "enter":
		if m.mode == modeRunning {
			return m, nil
		}
		raw := strings.TrimSpace(m.input.Value())
		if raw == "" {
			return m, nil
		}
		if raw == "q" || raw == "quit" || raw == "exit" {
			m.quitting = true
			return m, tea.Quit
		}
		if raw == "clear" {
			m.outputLines = nil
			m.viewport.SetContent("")
			m.input.Reset()
			return m, nil
		}
		if raw == "help" || raw == "?" {
			m.pushView()
			return m.showHelp()
		}
		if raw == "scan" || raw == "findings" {
			m.pushView()
			return m.showFindings("")
		}
		if strings.HasPrefix(raw, "findings ") {
			m.pushView()
			return m.showFindings(strings.TrimPrefix(raw, "findings "))
		}
		// Allow bare category names as shortcuts after a scan
		if m.isScanCategory(raw) {
			m.pushView()
			return m.showFindings(raw)
		}
		// Quick action shortcuts — map number keys to suggested actions
		if raw == "1" || raw == "2" || raw == "3" {
			if resolved := m.resolveQuickAction(raw); resolved != "" {
				raw = resolved
			}
		}
		m.historyIdx = -1
		return m.executeCommand(raw)

	case "up":
		if m.mode == modeRunning {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		// History recall — walk backward
		if len(m.cmdHistory) == 0 {
			return m, nil
		}
		if m.historyIdx == -1 {
			m.historySaved = m.input.Value()
			m.historyIdx = len(m.cmdHistory)
		}
		if m.historyIdx > 0 {
			m.historyIdx--
			m.input.SetValue(m.cmdHistory[m.historyIdx])
			m.input.CursorEnd()
			m.updateSuggestionList()
		}
		return m, nil

	case "down":
		if m.mode == modeRunning {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		if m.historyIdx >= 0 {
			m.historyIdx++
			if m.historyIdx >= len(m.cmdHistory) {
				m.historyIdx = -1
				m.input.SetValue(m.historySaved)
			} else {
				m.input.SetValue(m.cmdHistory[m.historyIdx])
			}
			m.input.CursorEnd()
			m.updateSuggestionList()
		}
		return m, nil

	case "tab":
		if m.mode != modeIdle || len(m.postRunCmds) == 0 {
			return m, nil
		}
		// Cycle through post-run suggested commands
		m.tabIdx++
		if m.tabIdx >= len(m.postRunCmds) {
			m.tabIdx = 0
		}
		m.input.SetValue(m.postRunCmds[m.tabIdx])
		m.input.CursorEnd()
		return m, nil

	case "right":
		if m.mode != modeIdle {
			return m, nil
		}
		// Only accept suggestion when cursor is at end of input
		cursorAtEnd := m.input.Position() >= len([]rune(m.input.Value()))
		if !cursorAtEnd {
			// Cursor not at end — move cursor, don't accept suggestion
			saved := m.input.KeyMap.AcceptSuggestion
			m.input.KeyMap.AcceptSuggestion = key.NewBinding()
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			m.input.KeyMap.AcceptSuggestion = saved
			return m, cmd
		}
		// Cursor at end — let bubbles accept the suggestion
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m.updateSuggestionList()
		return m, cmd
	}

	if m.mode == modeIdle {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m.historyIdx = -1 // reset history on any typed input
		m.tabIdx = -1     // reset tab cycling on any typed input
		m.updateSuggestionList()
		return m, cmd
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// updateSuggestionList refreshes inline predictions based on current input.
func (m *TUIModel) updateSuggestionList() {
	suggestions := buildSuggestions(m.input.Value(), m.cmdHistory)
	m.input.SetSuggestions(suggestions)
}

// pushView saves the current viewport state so esc can restore it.
func (m *TUIModel) pushView() {
	if len(m.outputLines) > 0 {
		m.viewStack = append(m.viewStack, viewFrame{
			outputLines: append([]string{}, m.outputLines...),
			placeholder: m.input.Placeholder,
		})
	}
}

// ── Command Execution ─────────────────────────────────────────────────

func (m TUIModel) executeCommand(raw string) (TUIModel, tea.Cmd) {
	deity, args, intentMatched := m.dispatch(raw)

	m.mode = modeRunning
	m.runningDeity = deity
	m.runningCmd = raw
	m.runningArgs = args
	m.cmdStartTime = time.Now()
	m.streamCh = make(chan string, 100) // fresh channel per command
	m.outputLines = nil
	m.input.Blur()
	m.input.Reset()

	glyph, name := deityDisplay(deity)
	if deity != "" {
		m.outputLines = append(m.outputLines,
			lipgloss.NewStyle().Foreground(Gold).Bold(true).Render(
				fmt.Sprintf("  %s %s", glyph, name)))
		if intentMatched {
			// Show the user what their natural language was interpreted as
			m.outputLines = append(m.outputLines,
				lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(
					fmt.Sprintf("  \"%s\" → %s", raw, strings.Join(args, " "))))
		}
		m.outputLines = append(m.outputLines, "")
	}
	m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
	m.recalcViewportHeight()

	exe, _ := os.Executable()
	cmd := exec.Command(exe, args...)

	return m, tea.Batch(m.spinner.Tick, elapsedTick(), m.runCommandStreaming(cmd))
}

// dispatch routes user input to a deity. Returns (deity, args, intentMatched).
// intentMatched is true when the input was natural language matched via keywords
// rather than a direct deity name or CLI alias.
func (m *TUIModel) dispatch(raw string) (string, []string, bool) {
	lower := strings.ToLower(raw)
	tokens := strings.Fields(lower)
	rawTokens := strings.Fields(raw)

	if len(tokens) == 0 {
		return "", nil, false
	}

	for _, d := range deityRoster {
		if tokens[0] == d.Key {
			return d.Key, rawTokens, false
		}
	}

	if target, ok := cliAliases[tokens[0]]; ok {
		return target, rawTokens, false
	}

	bestDeity := ""
	bestScore := 0
	for deity, keywords := range intentKeywords {
		score := 0
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			bestDeity = deity
		}
	}

	if bestDeity != "" {
		args := inferSubcommand(bestDeity, lower)
		return bestDeity, args, true
	}

	return "", rawTokens, false
}

// inferSubcommand maps a deity + natural language input to the most likely
// CLI args. Without this, intent matches dispatch to bare deity names which
// just show help text.
func inferSubcommand(deity, lower string) []string {
	type rule struct {
		keywords   []string
		subcommand []string
	}

	// Order matters — first match wins within a deity.
	deityRules := map[string][]rule{
		"isis": {
			{[]string{"network", "dns", "wifi", "firewall", "tls", "vpn", "security"}, []string{"isis", "network"}},
			{[]string{"doctor", "health", "diagnostic"}, []string{"doctor"}},
			{[]string{"heal", "remediate", "fix", "repair"}, []string{"maat", "heal"}},
			{[]string{"guard", "monitor", "ram", "cpu", "process"}, []string{"guard"}},
		},
		"anubis": {
			{[]string{"ghost", "dead", "remnant", "haunt", "uninstall"}, []string{"anubis", "ka"}},
			{[]string{"duplicate", "dedup", "mirror"}, []string{"anubis", "mirror"}},
			{[]string{"clean", "judge", "purge"}, []string{"anubis", "judge", "--dry-run"}},
			{[]string{"scan", "waste", "hygiene"}, []string{"anubis", "weigh"}},
			{[]string{"apps", "installed", "applications"}, []string{"anubis", "apps"}},
		},
		"thoth": {
			{[]string{"sync", "memory"}, []string{"thoth", "sync"}},
			{[]string{"compact", "persist"}, []string{"thoth", "compact"}},
			{[]string{"init"}, []string{"thoth", "init"}},
			{[]string{"brain", "neural", "weights"}, []string{"thoth", "brain"}},
			{[]string{"status"}, []string{"thoth", "status"}},
		},
		"maat": {
			{[]string{"audit", "quality", "qa"}, []string{"maat", "audit"}},
			{[]string{"coverage", "test", "lint"}, []string{"maat", "pulse"}},
			{[]string{"scales", "policy", "enforce"}, []string{"maat", "scales"}},
			{[]string{"heal", "remediate"}, []string{"maat", "heal"}},
		},
		"seshat": {
			{[]string{"ingest", "graft", "knowledge"}, []string{"seshat", "ingest"}},
			{[]string{"notebooklm", "notebook"}, []string{"seshat", "notebooklm"}},
			{[]string{"list", "browse"}, []string{"seshat", "list"}},
			{[]string{"export"}, []string{"seshat", "export"}},
			{[]string{"adapters", "sources"}, []string{"seshat", "adapters"}},
			{[]string{"mcp", "context server"}, []string{"seshat", "mcp"}},
			{[]string{"auth", "authenticate"}, []string{"seshat", "auth", "google"}},
			{[]string{"chrome", "profile"}, []string{"seshat", "profiles", "chrome"}},
		},
		"seba": {
			{[]string{"gpu", "vram", "cuda", "metal", "ane", "npu"}, []string{"seba", "hardware"}},
			{[]string{"hardware", "accelerator", "profile"}, []string{"seba", "hardware"}},
			{[]string{"diagram", "graph"}, []string{"seba", "diagram"}},
			{[]string{"architecture", "topology", "map"}, []string{"seba", "scan"}},
			{[]string{"fleet", "subnet", "container", "docker", "kubernetes"}, []string{"seba", "fleet"}},
			{[]string{"book", "registry", "projects"}, []string{"seba", "book"}},
			{[]string{"compute", "tokenize", "ane"}, []string{"seba", "compute"}},
		},
		"ra": {
			{[]string{"status"}, []string{"ra", "status"}},
			{[]string{"deploy", "sprint"}, []string{"ra", "deploy"}},
			{[]string{"health"}, []string{"ra", "health"}},
			{[]string{"test"}, []string{"ra", "test"}},
			{[]string{"lint"}, []string{"ra", "lint"}},
			{[]string{"nightly", "ci"}, []string{"ra", "nightly"}},
			{[]string{"broadcast"}, []string{"ra", "broadcast"}},
			{[]string{"watch", "logs"}, []string{"ra", "watch"}},
			{[]string{"kill", "stop"}, []string{"ra", "kill"}},
			{[]string{"collect"}, []string{"ra", "collect"}},
			{[]string{"pipeline"}, []string{"ra", "pipeline"}},
		},
		"net": {
			{[]string{"align", "drift"}, []string{"net", "align"}},
			{[]string{"scope", "status"}, []string{"net", "status"}},
		},
		"osiris": {
			{[]string{"assess", "checkpoint", "uncommitted", "risk", "drift"}, []string{"osiris", "assess"}},
			{[]string{"status", "summary"}, []string{"osiris", "status"}},
		},
		"horus": {
			{[]string{"scan", "index", "graph"}, []string{"horus", "scan"}},
			{[]string{"outline", "declarations"}, []string{"horus", "outline"}},
			{[]string{"symbols", "search"}, []string{"horus", "symbols"}},
			{[]string{"context"}, []string{"horus", "context"}},
			{[]string{"stats", "statistics"}, []string{"horus", "stats"}},
		},
	}

	if rules, ok := deityRules[deity]; ok {
		for _, r := range rules {
			for _, kw := range r.keywords {
				if strings.Contains(lower, kw) {
					return r.subcommand
				}
			}
		}
	}

	// Fallback: bare deity name
	return []string{deity}
}

// runCommandStreaming runs a command and sends output lines to streamCh
// one at a time. The channel is closed when the command completes. A
// goroutine handles the pipe reading so the TUI stays responsive.
func (m TUIModel) runCommandStreaming(cmd *exec.Cmd) tea.Cmd {
	ch := m.streamCh
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

		// Goroutine: scan lines and push to channel, then close.
		go func() {
			scanner := bufio.NewScanner(combined)
			for scanner.Scan() {
				ch <- scanner.Text()
			}
			_ = cmd.Wait()
			close(ch)
		}()

		// Return first line (or done if command exits immediately).
		line, ok := <-ch
		if !ok {
			return streamLineMsg{done: true}
		}
		return streamLineMsg{line: line}
	}
}

func (m TUIModel) handleStreamLine(msg streamLineMsg) (TUIModel, tea.Cmd) {
	if msg.done {
		// Command finished — finalize like handleBatchOutput does.
		m.mode = modeIdle
		m.input.Focus()

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
				lipgloss.NewStyle().Foreground(Red).Render(fmt.Sprintf("  ✗ %v", msg.err)))
			m.outputLines = append(m.outputLines, m.postRunErrorGuidance(msg.err)...)
			m.input.Placeholder = "doctor · help  (diagnose or see all commands)"
			if m.runningDeity != "" {
				m.deityState[m.runningDeity] = stateFailed
			}
		} else {
			m.outputLines = append(m.outputLines, "",
				lipgloss.NewStyle().Foreground(Green).Render("  ✓ Done"))
			m.outputLines = append(m.outputLines, m.postRunSuggestions()...)
			m.input.Placeholder = m.postRunPlaceholder()
			m.postRunCmds = m.postRunCommandList()
			m.tabIdx = -1
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
		m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
		m.viewport.GotoBottom()
		m.savePersistedState()
		m.runningDeity = ""
		m.runningCmd = ""
		m.runningArgs = nil
		return m, nil
	}

	// Append the new line and keep listening for more.
	m.outputLines = append(m.outputLines, "  "+msg.line)
	m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
	m.viewport.GotoBottom()
	return m, waitForStreamLine(m.streamCh)
}

func (m TUIModel) handleBatchOutput(msg cmdBatchMsg) (TUIModel, tea.Cmd) {
	for _, line := range msg.lines {
		m.outputLines = append(m.outputLines, "  "+line)
	}

	m.mode = modeIdle
	m.input.Focus()

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
			lipgloss.NewStyle().Foreground(Red).Render(fmt.Sprintf("  ✗ %v", msg.err)))
		m.outputLines = append(m.outputLines, m.postRunErrorGuidance(msg.err)...)
		m.input.Placeholder = "doctor · help  (diagnose or see all commands)"
		if m.runningDeity != "" {
			m.deityState[m.runningDeity] = stateFailed
		}
	} else {
		m.outputLines = append(m.outputLines, "",
			lipgloss.NewStyle().Foreground(Green).Render("  ✓ Done"))
		// Append deity-specific next steps so the user isn't left at a dead end.
		m.outputLines = append(m.outputLines, m.postRunSuggestions()...)
		m.input.Placeholder = m.postRunPlaceholder()
		m.postRunCmds = m.postRunCommandList()
		m.tabIdx = -1
		if m.runningDeity != "" {
			// Check if this deity produced actionable data
			state := stateSucceeded
			if m.runningDeity == "anubis" {
				if scan, err := jackal.LoadLatest(); err == nil && len(scan.Findings) > 0 {
					state = stateHasData
				}
			}
			m.deityState[m.runningDeity] = state
		}
	}
	m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
	m.viewport.GotoBottom()
	m.savePersistedState()
	m.runningDeity = ""
	m.runningCmd = ""
	m.runningArgs = nil
	return m, nil
}

// ── Active Deity Detection ────────────────────────────────────────────

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
			deity := strings.ToLower(e.Deity)
			if !strings.Contains(deity, ":") {
				m.activeDeity[deity] = true
			}
		}
	}

	m.refreshNotifications()

	home, _ := os.UserHomeDir()
	pidDir := filepath.Join(home, ".config", "ra", "pids")
	pidEntries, _ := os.ReadDir(pidDir)
	for _, f := range pidEntries {
		if f.IsDir() {
			continue
		}
		name := strings.TrimSuffix(f.Name(), ".pid")
		for _, d := range deityRoster {
			if strings.Contains(strings.ToLower(name), d.Key) {
				m.activeDeity[d.Key] = true
			}
		}
	}
}

// refreshNotifications loads the latest notifications from the store.
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

// ── Layout ────────────────────────────────────────────────────────────

const leftPaneWidth = 42

func (m TUIModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	hasOutput := len(m.outputLines) > 0
	maxW := min(m.width-2, 90)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
	divider := dim.Render(strings.Repeat("─", maxW))

	header := lipgloss.NewStyle().Foreground(Gold).Bold(true).Render("𓉴  Sirsi Pantheon")
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("#999999")).
		Render("DevOps intelligence for developers and infrastructure teams")
	signage := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).
		Render(" Sirsi Technologies, Inc. 2026 (Apache 2.0)")

	var b strings.Builder
	usedLines := 0

	if !hasOutput {
		// ── Single-pane: full roster
		b.WriteString("\n")
		b.WriteString(" " + header + "\n")
		b.WriteString(" " + desc + "\n")
		b.WriteString(" " + divider + "\n")
		usedLines += 4

		roster := m.renderRosterColumns(false)
		b.WriteString(roster)
		usedLines += strings.Count(roster, "\n")

		status := m.renderStatusLine()
		b.WriteString(status)
		usedLines += strings.Count(status, "\n")

		// Context-aware quick actions (onboarding + suggestions)
		actions := m.renderQuickActions()
		b.WriteString(actions)
		usedLines += strings.Count(actions, "\n")

		b.WriteString(" " + divider + "\n")
		b.WriteString(" " + m.input.View() + "\n")
		b.WriteString(m.renderHints(false) + "\n")
		usedLines += 3
	} else if m.width < 70 {
		// ── Narrow terminal: stack vertically instead of split-pane
		b.WriteString("\n")
		b.WriteString(" " + header + "\n")
		b.WriteString(" " + divider + "\n")
		usedLines += 3

		if m.mode == modeRunning {
			glyph, name := deityDisplay(m.runningDeity)
			elapsed := time.Since(m.cmdStartTime).Truncate(time.Second)
			elapsedStr := ""
			if elapsed >= time.Second {
				elapsedStr = fmt.Sprintf(" (%s)", elapsed)
			}
			b.WriteString(" " + m.spinner.View() + " " +
				lipgloss.NewStyle().Foreground(Gold).Render(glyph+" "+name) +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(" running..."+elapsedStr) + "\n")
			usedLines++
		}
		b.WriteString(m.viewport.View() + "\n")
		usedLines += m.viewport.Height()

		b.WriteString(" " + divider + "\n")
		b.WriteString(" " + m.input.View() + "\n")
		b.WriteString(m.renderHints(true) + "\n")
		usedLines += 3
	} else {
		// ── Split-pane: left roster | right output
		left := m.renderLeftPane()
		right := m.renderRightPane()

		leftStyle := lipgloss.NewStyle().
			Width(leftPaneWidth).
			BorderRight(true).
			BorderStyle(lipgloss.Border{Right: "│"}).
			BorderForeground(lipgloss.Color("#333333")).
			PaddingRight(1)

		rightWidth := m.width - leftPaneWidth - 3
		if rightWidth < 20 {
			rightWidth = 20
		}
		rightStyle := lipgloss.NewStyle().Width(rightWidth).PaddingLeft(1)

		panes := lipgloss.JoinHorizontal(lipgloss.Top,
			leftStyle.Render(left),
			rightStyle.Render(right),
		)

		b.WriteString("\n")
		b.WriteString(" " + header + "\n")
		b.WriteString(" " + divider + "\n")
		usedLines += 3

		b.WriteString(panes + "\n")
		usedLines += strings.Count(panes, "\n") + 1

		b.WriteString(" " + divider + "\n")
		b.WriteString(" " + m.input.View() + "\n")
		b.WriteString(m.renderHints(true) + "\n")
		usedLines += 3
	}

	// Pad to push signage to the bottom — exactly once
	remaining := m.height - usedLines - 2
	if remaining > 0 {
		b.WriteString(strings.Repeat("\n", remaining))
	}
	b.WriteString(signage)

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

// renderRosterColumns renders deities in a column grid that fits the available width.
// In compact mode (split-pane left pane), uses leftPaneWidth instead of terminal width.
// 3 columns if width >= 90, 2 columns if >= 60, single column otherwise.
func (m TUIModel) renderRosterColumns(compact bool) string {
	var b strings.Builder

	availWidth := m.width
	if compact {
		availWidth = leftPaneWidth
	}

	cols := 3
	if availWidth < 90 {
		cols = 2
	}
	if availWidth < 60 {
		cols = 1
	}

	rows := (len(deityRoster) + cols - 1) / cols
	colWidth := (availWidth - 2) / cols
	if colWidth > 34 {
		colWidth = 34
	}

	for r := 0; r < rows; r++ {
		var rowParts []string
		for c := 0; c < cols; c++ {
			idx := c*rows + r // column-major: fill down then across
			if idx < len(deityRoster) {
				rowParts = append(rowParts, m.renderDeityCell(deityRoster[idx], colWidth))
			}
		}
		b.WriteString(" " + strings.Join(rowParts, "") + "\n")
	}

	return b.String()
}

// renderDeityCell renders one deity as a fixed-width cell for the grid.
// Avoids lipgloss Width/MaxWidth for layout — Egyptian glyphs have
// unpredictable terminal widths. Instead we measure with lipgloss.Width()
// and pad with real spaces so the error model is consistent.
func (m TUIModel) renderDeityCell(d deityInfo, width int) string {
	active := m.activeDeity[d.Key]
	isRunning := m.mode == modeRunning && m.runningDeity == d.Key
	state := m.deityState[d.Key]

	var nameColor, roleColor color.Color
	if isRunning {
		nameColor = Gold
		roleColor = lipgloss.Color("#CCCCCC")
	} else if active {
		nameColor = Gold
		roleColor = lipgloss.Color("#CCCCCC")
	} else if state == stateFailed {
		nameColor = lipgloss.Color("#FF6666")
		roleColor = lipgloss.Color("#996666")
	} else if state == stateHasData {
		nameColor = lipgloss.Color("#BBBBBB")
		roleColor = lipgloss.Color("#777777")
	} else {
		nameColor = lipgloss.Color("#BBBBBB")
		roleColor = lipgloss.Color("#777777")
	}

	// State-aware indicator dot
	var dot string
	switch {
	case isRunning:
		dot = m.spinner.View()
	case state == stateFailed:
		dot = lipgloss.NewStyle().Foreground(Red).Render("✗")
	case state == stateHasData:
		dot = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00")).Render("◆")
	case state == stateSucceeded:
		dot = lipgloss.NewStyle().Foreground(Green).Render("✓")
	case active:
		dot = lipgloss.NewStyle().Foreground(Gold).Render("●")
	default:
		dot = lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render("·")
	}

	glyph := lipgloss.NewStyle().Foreground(nameColor).Render(d.Glyph)
	name := lipgloss.NewStyle().Bold(true).Foreground(nameColor).Render(d.Name)
	role := lipgloss.NewStyle().Foreground(roleColor).Render(d.Role)

	// Pad name to a fixed visual column so roles align across rows.
	prefix := dot + " " + glyph + " " + name
	prefixW := lipgloss.Width(prefix)
	const nameEnd = 14 // target column where role text starts
	if prefixW < nameEnd {
		prefix += strings.Repeat(" ", nameEnd-prefixW)
	}

	cell := prefix + role
	cellW := lipgloss.Width(cell)
	if cellW < width {
		cell += strings.Repeat(" ", width-cellW)
	}
	return cell
}

func (m TUIModel) renderStatusLine() string {
	activeCount := 0
	for _, d := range deityRoster {
		if m.activeDeity[d.Key] {
			activeCount++
		}
	}
	if activeCount > 0 {
		return lipgloss.NewStyle().Foreground(Green).
			Render(fmt.Sprintf(" %d %s active", activeCount, pluralize("deity", activeCount))) + "\n"
	}
	return ""
}

func (m TUIModel) renderLeftPane() string {
	var b strings.Builder
	b.WriteString(m.renderRosterColumns(true))
	b.WriteString(m.renderStatusLine())
	b.WriteString(m.renderNotifications())
	return b.String()
}

// renderNotifications shows recent notifications in the left pane.
func (m TUIModel) renderNotifications() string {
	if len(m.recentNotify) == 0 {
		return ""
	}

	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(Gold).Render(" 🔔 Recent") + "\n")

	for _, n := range m.recentNotify {
		icon := notify.SeverityIcon(n.Severity)
		src := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BBBBBB")).Render(n.Source)
		summary := n.Summary
		if len(summary) > 28 {
			summary = summary[:25] + "…"
		}
		b.WriteString(fmt.Sprintf(" %s %s %s\n", icon, src, dim.Render(summary)))
	}
	return b.String()
}

func (m TUIModel) renderRightPane() string {
	var b strings.Builder
	if m.mode == modeRunning {
		glyph, name := deityDisplay(m.runningDeity)
		elapsed := time.Since(m.cmdStartTime).Truncate(time.Second)
		elapsedStr := fmt.Sprintf(" (%s)", elapsed)
		if elapsed < time.Second {
			elapsedStr = ""
		}
		b.WriteString(m.spinner.View() + " " +
			lipgloss.NewStyle().Foreground(Gold).Render(glyph+" "+name) +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(" running..."+elapsedStr) + "\n\n")
	}
	b.WriteString(m.viewport.View())
	return b.String()
}

func (m TUIModel) renderHints(splitMode bool) string {
	var hints []string
	if m.mode == modeRunning {
		hints = append(hints, "↑/↓ scroll")
	} else {
		if len(m.postRunCmds) > 0 {
			hints = append(hints, "tab cycle suggestions")
		}
		hints = append(hints, "→ accept", "↑ history", "help")
		if len(m.viewStack) > 0 {
			hints = append(hints, fmt.Sprintf("esc back (%d)", len(m.viewStack)))
		} else if splitMode {
			hints = append(hints, "esc back")
		}
	}
	hints = append(hints, "ctrl+c quit")
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).
		Render(" " + strings.Join(hints, " · "))
}

// renderQuickActions shows context-aware suggested actions.
// For new users: starting points. After commands: next logical steps.
type quickAction struct {
	desc string
	cmd  string
}

// computeQuickActions returns up to 3 context-aware suggested actions.
func (m TUIModel) computeQuickActions() []quickAction {
	if len(m.history) == 0 {
		return []quickAction{
			{"Scan for infrastructure waste", "scan"},
			{"Check network security posture", "isis network"},
			{"Full system health diagnostic", "doctor"},
		}
	}

	hasRun := make(map[string]bool)
	for _, h := range m.history {
		hasRun[h.deity] = true
	}

	var actions []quickAction

	if !hasRun["anubis"] {
		actions = append(actions, quickAction{"Scan for waste — find what's eating your disk", "scan"})
	} else if m.deityState["anubis"] == stateHasData {
		actions = append(actions, quickAction{"Review findings and clean up", "findings"})
	}

	if !hasRun["isis"] {
		actions = append(actions, quickAction{"Network security audit", "isis network"})
	}

	if !hasRun["maat"] && hasRun["anubis"] {
		actions = append(actions, quickAction{"Quality assessment", "maat audit"})
	}

	if !hasRun["seba"] {
		actions = append(actions, quickAction{"Hardware and architecture profile", "seba hardware"})
	}

	if !hasRun["osiris"] {
		actions = append(actions, quickAction{"Check uncommitted work risk", "osiris assess"})
	}

	if !hasRun["thoth"] {
		actions = append(actions, quickAction{"Sync project memory", "thoth sync"})
	}

	// Failed deities get retry suggestions
	for deity, state := range m.deityState {
		if state == stateFailed {
			glyph, name := deityDisplay(deity)
			actions = append(actions, quickAction{fmt.Sprintf("Retry %s %s (failed last run)", glyph, name), deity})
		}
	}

	if len(actions) > 3 {
		actions = actions[:3]
	}
	return actions
}

// resolveQuickAction maps a number key ("1", "2", "3") to the corresponding
// suggested action command. Returns empty string if no match.
func (m TUIModel) resolveQuickAction(key string) string {
	actions := m.computeQuickActions()
	idx := 0
	switch key {
	case "1":
		idx = 0
	case "2":
		idx = 1
	case "3":
		idx = 2
	default:
		return ""
	}
	if idx < len(actions) {
		return actions[idx].cmd
	}
	return ""
}

// renderQuickActions shows context-aware suggested actions.
func (m TUIModel) renderQuickActions() string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	gold := lipgloss.NewStyle().Foreground(Gold)

	actions := m.computeQuickActions()
	if len(actions) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n")
	if len(m.history) == 0 {
		b.WriteString(dim.Render("  Try one of these to get started:") + "\n")
	} else {
		b.WriteString(dim.Render("  Suggested:") + "\n")
	}
	for i, a := range actions {
		b.WriteString("   " + gold.Render(fmt.Sprintf("%d", i+1)) + dim.Render("  "+a.desc) + "\n")
	}
	b.WriteString("\n")
	return b.String()
}

// showHelp renders an in-TUI help panel listing all available commands.
func (m TUIModel) showHelp() (TUIModel, tea.Cmd) {
	gold := lipgloss.NewStyle().Foreground(Gold).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	body := lipgloss.NewStyle().Foreground(White)

	var lines []string
	lines = append(lines,
		gold.Render("  Pantheon Commands"),
		"",
		body.Render("  Core:"),
		dim.Render("    scan                  Scan for infrastructure waste"),
		dim.Render("    ghosts                Detect remnants of uninstalled apps"),
		dim.Render("    dedup [dirs]          Find duplicate files"),
		dim.Render("    doctor                System health diagnostic"),
		dim.Render("    guard                 Monitor system resources"),
		"",
		body.Render("  Deities:"),
		dim.Render("    ra status             Orchestrator status"),
		dim.Render("    ra deploy             Deploy task to repos"),
		dim.Render("    ra test / ra lint     Run tests or linters fleet-wide"),
		dim.Render("    net align             Cross-module consistency check"),
		dim.Render("    thoth sync            Sync project memory"),
		dim.Render("    thoth compact         Persist state for continuations"),
		dim.Render("    maat audit            Governance and quality scan"),
		dim.Render("    maat heal             Auto-remediate quality issues"),
		dim.Render("    isis network          Network security audit"),
		dim.Render("    isis heal             Auto-remediate failures"),
		dim.Render("    seshat ingest         Ingest knowledge from sources"),
		dim.Render("    seshat list           Browse knowledge items"),
		dim.Render("    anubis weigh          Scan for waste"),
		dim.Render("    anubis apps           List apps and ghost residuals"),
		dim.Render("    seba hardware         Hardware and accelerator profile"),
		dim.Render("    seba diagram          Architecture diagram generation"),
		dim.Render("    seba fleet            Network and container discovery"),
		dim.Render("    horus scan            Build code symbol graph"),
		dim.Render("    horus outline <file>  File declaration outline"),
		dim.Render("    osiris assess         Uncommitted work risk assessment"),
		"",
		body.Render("  After Scan:"),
		dim.Render("    findings              Full breakdown with advisories"),
		dim.Render("    findings <category>   Drill into dev, ai, cloud, etc."),
		dim.Render("    clean                 Remove safe items (Trash)"),
		dim.Render("    judge                 Policy-based review before cleanup"),
		"",
		body.Render("  Natural Language:"),
		dim.Render("    Type what you want in plain English and Pantheon"),
		dim.Render("    will route to the right deity automatically."),
		"",
		body.Render("  Navigation:"),
		dim.Render("    →         Accept inline suggestion"),
		dim.Render("    ↑/↓       Browse command history / scroll output"),
		dim.Render("    esc       Clear output pane"),
		dim.Render("    clear     Reset display"),
		dim.Render("    ctrl+c    Quit"),
	)

	m.outputLines = lines
	m.viewport.SetContent(strings.Join(lines, "\n"))
	m.recalcViewportHeight()
	m.viewport.GotoTop()
	return m, nil
}

// ── Post-Run Suggestions ─────────────────────────────────────────

// postRunPlaceholder returns a contextual input placeholder based on what
// command just finished, guiding the user to the most likely next action.
func (m TUIModel) postRunPlaceholder() string {
	sub := ""
	if len(m.runningArgs) >= 2 {
		sub = m.runningArgs[1]
	}

	switch m.runningDeity {
	case "anubis":
		switch sub {
		case "weigh":
			return "findings · clean · judge  (or type a category like dev, ai, cloud)"
		case "judge":
			return "findings · scan  (verify cleanup with a fresh scan)"
		case "ka":
			return "findings · clean  (remove ghost residuals)"
		case "mirror":
			return "scan · clean  (full scan or reclaim space)"
		case "apps":
			return "ghosts · scan · clean  (check residuals or scan waste)"
		}
	case "isis":
		switch sub {
		case "network":
			return "heal · doctor  (remediate issues or full health check)"
		default:
			return "isis network · doctor · heal  (audit, diagnose, or fix)"
		}
	case "maat":
		switch sub {
		case "audit":
			return "maat pulse · heal  (quick summary or auto-remediate)"
		case "pulse":
			return "maat audit · heal  (full audit or auto-remediate)"
		case "heal":
			return "maat audit · maat pulse  (verify fixes)"
		default:
			return "maat audit · maat pulse · heal"
		}
	case "ra":
		switch sub {
		case "deploy":
			return "ra status · ra health  (check progress or health)"
		case "status":
			return "ra deploy · ra health · ra test  (deploy, check health, or run tests)"
		case "health":
			return "ra deploy · ra test · ra lint  (deploy or run checks)"
		case "test", "lint":
			return "ra status · heal  (check results or auto-remediate)"
		default:
			return "ra status · ra deploy · ra health"
		}
	case "net":
		switch sub {
		case "align":
			return "net status · maat audit  (check alignment or run QA)"
		default:
			return "net align · maat audit  (check drift or run QA)"
		}
	case "thoth":
		switch sub {
		case "sync":
			return "thoth compact · maat audit  (persist state or check quality)"
		case "compact":
			return "thoth sync · osiris assess  (sync memory or check risk)"
		case "init":
			return "thoth sync  (populate memory from source files)"
		default:
			return "thoth sync · thoth compact"
		}
	case "seshat":
		switch sub {
		case "ingest":
			return "seshat list · seshat export  (review or export knowledge)"
		case "notebooklm":
			return "seshat ingest · seshat list  (ingest more or browse)"
		default:
			return "seshat ingest · seshat list · seshat export"
		}
	case "seba":
		switch sub {
		case "hardware":
			return "seba diagram · seba scan · scan  (visualize, map, or scan waste)"
		case "diagram":
			return "seba scan · seba hardware  (system map or hardware profile)"
		case "fleet":
			return "seba diagram · isis network  (visualize fleet or audit network)"
		case "scan":
			return "seba diagram · seba hardware  (visualize or profile hardware)"
		default:
			return "seba scan · seba diagram · seba hardware"
		}
	case "osiris":
		switch sub {
		case "assess":
			return "osiris status · thoth sync  (quick status or sync memory)"
		default:
			return "osiris assess · thoth sync  (full assessment or sync state)"
		}
	case "horus":
		switch sub {
		case "scan":
			return "horus outline · horus symbols · horus stats  (explore the code graph)"
		default:
			return "horus scan · horus outline · horus stats"
		}
	}

	return "What next?"
}

// postRunCommandList returns the raw command strings for tab-cycling
// after a command completes. These mirror what postRunSuggestions shows
// but without styling — just the commands.
func (m TUIModel) postRunCommandList() []string {
	sub := ""
	if len(m.runningArgs) >= 2 {
		sub = m.runningArgs[1]
	}

	switch m.runningDeity {
	case "anubis":
		switch sub {
		case "weigh":
			return []string{"findings", "clean", "judge"}
		case "judge":
			return []string{"findings", "scan"}
		case "ka":
			return []string{"findings", "clean"}
		case "mirror":
			return []string{"scan"}
		}
	case "isis":
		return []string{"heal", "doctor"}
	case "maat":
		return []string{"maat pulse", "maat audit", "heal"}
	case "ra":
		switch sub {
		case "deploy":
			return []string{"ra status", "ra health", "ra collect"}
		case "status":
			return []string{"ra deploy", "ra health", "ra test"}
		default:
			return []string{"ra status", "ra deploy"}
		}
	case "net":
		return []string{"net status", "net align", "maat audit"}
	case "thoth":
		return []string{"thoth sync", "thoth compact", "osiris assess"}
	case "seshat":
		return []string{"seshat list", "seshat export", "seshat ingest"}
	case "seba":
		return []string{"seba diagram", "seba scan", "seba hardware"}
	case "osiris":
		return []string{"osiris assess", "osiris status", "thoth sync"}
	case "horus":
		return []string{"horus scan", "horus outline", "horus stats"}
	}
	return nil
}

// postRunSuggestions returns deity-aware next-step hints based on what
// command just finished. This prevents the "Done → dead terminal" problem
// by surfacing actionable follow-ups in context.
func (m TUIModel) postRunSuggestions() []string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	gold := lipgloss.NewStyle().Foreground(Gold)
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	// Determine the subcommand from dispatched args.
	sub := ""
	if len(m.runningArgs) >= 2 {
		sub = m.runningArgs[1]
	} else if len(m.runningArgs) == 1 {
		sub = m.runningArgs[0]
	}

	var lines []string

	switch m.runningDeity {
	case "anubis":
		switch sub {
		case "weigh":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("findings")+"      "+hint.Render("View full breakdown with advisories"),
				"  "+gold.Render("findings dev")+"  "+hint.Render("Drill into a specific category"),
				"  "+gold.Render("clean")+"         "+hint.Render("Remove safe items (moves to Trash)"),
				"  "+gold.Render("judge")+"         "+hint.Render("Policy-based review before cleanup"),
				"",
			)
		case "judge":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("findings")+"      "+hint.Render("Review remaining findings"),
				"  "+gold.Render("scan")+"          "+hint.Render("Run a fresh scan to verify cleanup"),
				"",
			)
		case "ka":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("findings")+"      "+hint.Render("View all findings including ghosts"),
				"  "+gold.Render("clean")+"         "+hint.Render("Remove ghost residuals"),
				"",
			)
		case "mirror":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("scan")+"          "+hint.Render("Run full waste scan"),
				"",
			)
		}
	case "isis":
		lines = append(lines,
			"",
			dim.Render("  ── What's Next ──────────────────────────"),
			"",
			"  "+gold.Render("heal")+"          "+hint.Render("Auto-remediate failed checks"),
			"  "+gold.Render("doctor")+"        "+hint.Render("Full system health diagnostic"),
			"",
		)
	case "maat":
		lines = append(lines,
			"",
			dim.Render("  ── What's Next ──────────────────────────"),
			"",
			"  "+gold.Render("maat pulse")+"    "+hint.Render("Quick coverage summary"),
			"  "+gold.Render("heal")+"          "+hint.Render("Auto-remediate quality issues"),
			"",
		)
	case "seba":
		switch sub {
		case "hardware":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("seba diagram")+"  "+hint.Render("Generate architecture diagram from scan"),
				"  "+gold.Render("seba scan")+"     "+hint.Render("Full infrastructure topology map"),
				"  "+gold.Render("scan")+"          "+hint.Render("Scan for infrastructure waste"),
				"",
			)
		case "diagram":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("seba scan")+"     "+hint.Render("Full infrastructure topology map"),
				"  "+gold.Render("seba hardware")+" "+hint.Render("Hardware and accelerator profile"),
				"",
			)
		case "fleet":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("seba diagram")+"  "+hint.Render("Visualize fleet architecture"),
				"  "+gold.Render("isis network")+"  "+hint.Render("Network security audit"),
				"",
			)
		default:
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("seba diagram")+"  "+hint.Render("Generate architecture diagram"),
				"  "+gold.Render("seba hardware")+" "+hint.Render("Hardware and accelerator profile"),
				"  "+gold.Render("scan")+"          "+hint.Render("Scan for infrastructure waste"),
				"",
			)
		}
	case "ra":
		switch sub {
		case "deploy":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("ra status")+"     "+hint.Render("Check deployment progress"),
				"  "+gold.Render("ra health")+"     "+hint.Render("Health check across all repos"),
				"  "+gold.Render("ra collect")+"    "+hint.Render("Collect logs from agents"),
				"",
			)
		case "status":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("ra deploy")+"     "+hint.Render("Deploy a task to repos"),
				"  "+gold.Render("ra health")+"     "+hint.Render("Health check across all repos"),
				"  "+gold.Render("ra test")+"       "+hint.Render("Run tests across all repos"),
				"",
			)
		case "health":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("ra deploy")+"     "+hint.Render("Deploy a task to repos"),
				"  "+gold.Render("ra test")+"       "+hint.Render("Run tests across all repos"),
				"  "+gold.Render("ra lint")+"       "+hint.Render("Run linters across all repos"),
				"  "+gold.Render("heal")+"          "+hint.Render("Auto-remediate failures"),
				"",
			)
		case "test", "lint":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("ra status")+"     "+hint.Render("Check overall repo status"),
				"  "+gold.Render("heal")+"          "+hint.Render("Auto-remediate failures"),
				"  "+gold.Render("maat audit")+"    "+hint.Render("Full quality assessment"),
				"",
			)
		default:
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("ra status")+"     "+hint.Render("Check orchestrator status"),
				"  "+gold.Render("ra deploy")+"     "+hint.Render("Deploy a task to repos"),
				"  "+gold.Render("ra health")+"     "+hint.Render("Health check across all repos"),
				"",
			)
		}
	case "net":
		switch sub {
		case "align":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("net status")+"    "+hint.Render("Check plan alignment score"),
				"  "+gold.Render("maat audit")+"    "+hint.Render("Run governance quality check"),
				"",
			)
		default:
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("net align")+"     "+hint.Render("Validate cross-module consistency"),
				"  "+gold.Render("maat audit")+"    "+hint.Render("Run governance quality check"),
				"",
			)
		}
	case "thoth":
		switch sub {
		case "sync":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("thoth compact")+" "+hint.Render("Persist state before context compression"),
				"  "+gold.Render("osiris assess")+" "+hint.Render("Check uncommitted work risk"),
				"  "+gold.Render("maat audit")+"    "+hint.Render("Run quality assessment"),
				"",
			)
		case "compact":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("thoth sync")+"    "+hint.Render("Sync memory from source files"),
				"  "+gold.Render("osiris assess")+" "+hint.Render("Check uncommitted work risk"),
				"",
			)
		case "init":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("thoth sync")+"    "+hint.Render("Populate memory from source + git history"),
				"  "+gold.Render("scan")+"          "+hint.Render("Scan for infrastructure waste"),
				"",
			)
		default:
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("thoth sync")+"    "+hint.Render("Sync project memory"),
				"  "+gold.Render("thoth compact")+" "+hint.Render("Persist state for continuations"),
				"",
			)
		}
	case "seshat":
		switch sub {
		case "ingest":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("seshat list")+"   "+hint.Render("Browse ingested knowledge items"),
				"  "+gold.Render("seshat export")+" "+hint.Render("Export knowledge to a target"),
				"  "+gold.Render("seshat notebooklm")+" "+hint.Render("Export to Google NotebookLM"),
				"",
			)
		case "notebooklm":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("seshat ingest")+" "+hint.Render("Ingest from more sources"),
				"  "+gold.Render("seshat list")+"   "+hint.Render("Browse knowledge items"),
				"",
			)
		default:
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("seshat ingest")+" "+hint.Render("Ingest knowledge from sources"),
				"  "+gold.Render("seshat list")+"   "+hint.Render("Browse knowledge items"),
				"  "+gold.Render("seshat export")+" "+hint.Render("Export knowledge to a target"),
				"",
			)
		}
	case "osiris":
		switch sub {
		case "assess":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("osiris status")+" "+hint.Render("Quick one-line risk summary"),
				"  "+gold.Render("thoth sync")+"    "+hint.Render("Sync memory before committing"),
				"  "+gold.Render("scan")+"          "+hint.Render("Scan for infrastructure waste"),
				"",
			)
		default:
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("osiris assess")+" "+hint.Render("Full checkpoint assessment"),
				"  "+gold.Render("thoth sync")+"    "+hint.Render("Sync project memory"),
				"",
			)
		}
	case "horus":
		switch sub {
		case "scan":
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("horus outline <file>")+" "+hint.Render("Print file declaration outline"),
				"  "+gold.Render("horus symbols")+"       "+hint.Render("Search symbols in the graph"),
				"  "+gold.Render("horus stats")+"         "+hint.Render("Graph statistics"),
				"",
			)
		default:
			lines = append(lines,
				"",
				dim.Render("  ── What's Next ──────────────────────────"),
				"",
				"  "+gold.Render("horus scan")+"    "+hint.Render("Build code symbol graph"),
				"  "+gold.Render("horus stats")+"   "+hint.Render("Graph statistics"),
				"  "+gold.Render("seba diagram")+"  "+hint.Render("Architecture diagram"),
				"",
			)
		}
	}

	return lines
}

// postRunErrorGuidance returns deity-specific error remediation hints
// so failures don't leave the user stranded.
func (m TUIModel) postRunErrorGuidance(err error) []string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	gold := lipgloss.NewStyle().Foreground(Gold)
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	errStr := ""
	if err != nil {
		errStr = strings.ToLower(err.Error())
	}

	var lines []string
	lines = append(lines, "",
		dim.Render("  ── Troubleshooting ──────────────────────"))

	// Check for common error patterns first (applies to any deity)
	switch {
	case strings.Contains(errStr, "permission denied") || strings.Contains(errStr, "operation not permitted"):
		lines = append(lines,
			"",
			"  "+hint.Render("Permission denied. Some options:"),
			"  "+gold.Render("1")+"  "+hint.Render("Re-run with --sudo if supported"),
			"  "+gold.Render("2")+"  "+hint.Render("Check file/directory ownership"),
			"  "+gold.Render("3")+"  "+hint.Render("Grant Full Disk Access in System Settings → Privacy"),
		)
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "no such file"):
		lines = append(lines,
			"",
			"  "+hint.Render("A required file or command was not found."),
			"  "+gold.Render("doctor")+"  "+hint.Render("Run a health diagnostic to identify missing deps"),
		)
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded"):
		lines = append(lines,
			"",
			"  "+hint.Render("The operation timed out."),
			"  "+gold.Render("1")+"  "+hint.Render("Check network connectivity"),
			"  "+gold.Render("2")+"  "+hint.Render("Try again — transient failures are common"),
		)
	case strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no route"):
		lines = append(lines,
			"",
			"  "+hint.Render("Network connection failed."),
			"  "+gold.Render("isis network")+"  "+hint.Render("Run a network security audit"),
		)
	default:
		// Deity-specific guidance when error pattern isn't recognized
		switch m.runningDeity {
		case "anubis":
			lines = append(lines,
				"",
				"  "+gold.Render("doctor")+"        "+hint.Render("Run system health diagnostic"),
				"  "+gold.Render("scan")+"          "+hint.Render("Try a fresh scan"),
			)
		case "ra":
			lines = append(lines,
				"",
				"  "+gold.Render("ra status")+"     "+hint.Render("Check orchestrator state"),
				"  "+gold.Render("ra health")+"     "+hint.Render("Health check all repos"),
			)
		case "seba":
			lines = append(lines,
				"",
				"  "+gold.Render("seba hardware")+" "+hint.Render("Check hardware compatibility"),
				"  "+gold.Render("doctor")+"        "+hint.Render("Run system health diagnostic"),
			)
		case "seshat":
			lines = append(lines,
				"",
				"  "+gold.Render("seshat adapters")+" "+hint.Render("List available adapters"),
				"  "+gold.Render("doctor")+"         "+hint.Render("Run system health diagnostic"),
			)
		case "thoth":
			lines = append(lines,
				"",
				"  "+gold.Render("thoth status")+"  "+hint.Render("Check memory system health"),
				"  "+gold.Render("thoth init")+"    "+hint.Render("Re-initialize .thoth/ if corrupted"),
			)
		case "maat":
			lines = append(lines,
				"",
				"  "+gold.Render("maat pulse")+"    "+hint.Render("Try a quick pulse check instead"),
				"  "+gold.Render("doctor")+"        "+hint.Render("Run system health diagnostic"),
			)
		case "net":
			lines = append(lines,
				"",
				"  "+gold.Render("net status")+"    "+hint.Render("Check alignment status"),
				"  "+gold.Render("ra status")+"     "+hint.Render("Check orchestrator state"),
			)
		case "osiris":
			lines = append(lines,
				"",
				"  "+gold.Render("osiris status")+" "+hint.Render("Try quick status instead"),
				"  "+gold.Render("doctor")+"        "+hint.Render("Run system health diagnostic"),
			)
		case "horus":
			lines = append(lines,
				"",
				"  "+gold.Render("horus scan")+"    "+hint.Render("Rebuild the code graph"),
				"  "+gold.Render("doctor")+"        "+hint.Render("Run system health diagnostic"),
			)
		default:
			lines = append(lines,
				"",
				"  "+gold.Render("doctor")+"        "+hint.Render("Run system health diagnostic"),
				"  "+gold.Render("help")+"          "+hint.Render("See all available commands"),
			)
		}
	}

	lines = append(lines, "")
	return lines
}

// ── Findings View ────────────────────────────────────────────────

// showFindings loads persisted scan results from disk and renders them
// in the TUI with category breakdown, advisories, and drill-down.
// When filter is non-empty, only findings in that category are shown.
func (m TUIModel) showFindings(filter string) (TUIModel, tea.Cmd) {
	gold := lipgloss.NewStyle().Foreground(Gold).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	warn := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00"))
	body := lipgloss.NewStyle().Foreground(White)
	fixable := lipgloss.NewStyle().Foreground(Green).Render("✓ fixable")
	breakingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6600"))

	scan, err := jackal.LoadLatest()
	if err != nil {
		m.outputLines = []string{
			gold.Render("  𓁢 Scan Findings"),
			"",
			dim.Render("  No scan results found. Run `scan` or `sirsi weigh` first."),
		}
		m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
		m.recalcViewportHeight()
		return m, nil
	}

	filter = strings.TrimSpace(strings.ToLower(filter))
	isFiltered := filter != ""

	var lines []string

	if isFiltered {
		// ── Category drill-down view ──
		catKey := jackal.Category(filter)
		summary, catExists := scan.ByCategory[catKey]

		if !catExists {
			// Try partial match
			for cat, s := range scan.ByCategory {
				if strings.Contains(strings.ToLower(string(cat)), filter) {
					catKey = cat
					summary = s
					catExists = true
					break
				}
			}
		}

		if !catExists {
			lines = append(lines,
				gold.Render(fmt.Sprintf("  𓁢 Findings — \"%s\"", filter)),
				"",
				dim.Render(fmt.Sprintf("  No category matching \"%s\". Available:", filter)),
			)
			for cat := range scan.ByCategory {
				lines = append(lines, "    "+gold.Render(string(cat)))
			}
			lines = append(lines, "",
				dim.Render("  Type `findings` to see all, or `findings <category>` to drill in."))
		} else {
			icon := categoryIcon(string(catKey))
			lines = append(lines,
				gold.Render(fmt.Sprintf("  %s %s — %s  (%d findings)",
					icon, catKey,
					jackal.FormatSize(summary.TotalSize),
					summary.Findings)),
				"",
			)

			// Show ALL findings in this category with full detail
			for _, f := range scan.Findings {
				if f.Category != catKey {
					continue
				}
				lines = append(lines, m.renderFindingDetail(f, body, dim, warn, fixable, breakingStyle)...)
			}

			lines = append(lines, "",
				dim.Render("  ── Actions ──────────────────────────────"),
				"  "+gold.Render("clean")+"         "+dim.Render("Remove safe items in this category"),
				"  "+gold.Render("findings")+"      "+dim.Render("Back to full overview"),
				"",
			)
		}
	} else {
		// ── Full overview ──
		lines = append(lines,
			gold.Render("  𓁢 Scan Findings"),
			dim.Render(fmt.Sprintf("  Scanned %s — %d rules, %s total waste",
				scan.Timestamp.Format("Jan 2 15:04"),
				scan.RulesRan,
				jackal.FormatSize(scan.TotalSize))),
			"",
		)

		// Category breakdown — now interactive (type category name to drill in)
		if len(scan.ByCategory) > 0 {
			lines = append(lines, body.Render("  Category Breakdown:")+" "+dim.Render("(type a name to drill in)"))
			for cat, s := range scan.ByCategory {
				icon := categoryIcon(string(cat))
				lines = append(lines,
					fmt.Sprintf("    %s "+gold.Render("%-14s")+" %s  (%d items)",
						icon, cat,
						jackal.FormatSize(s.TotalSize),
						s.Findings))
			}
			lines = append(lines, "")
		}

		// Top findings with richer detail
		limit := 20
		if len(scan.Findings) < limit {
			limit = len(scan.Findings)
		}
		if limit > 0 {
			lines = append(lines, body.Render(fmt.Sprintf("  Top %d of %d findings:", limit, len(scan.Findings))))
			lines = append(lines, "")
			for _, f := range scan.Findings[:limit] {
				lines = append(lines, m.renderFindingDetail(f, body, dim, warn, fixable, breakingStyle)...)
			}
			if len(scan.Findings) > limit {
				remaining := len(scan.Findings) - limit
				lines = append(lines, "",
					dim.Render(fmt.Sprintf("  ... and %d more. Type a category name to see all findings in it.", remaining)))
			}
		}

		lines = append(lines, "",
			dim.Render("  ── Actions ──────────────────────────────"),
			"  "+gold.Render("clean")+"              "+dim.Render("Remove safe items (Trash)"),
			"  "+gold.Render("judge")+"              "+dim.Render("Policy review before cleanup"),
			"  "+gold.Render("findings <category>")+" "+dim.Render("Drill into one category"),
			"",
		)
	}

	m.outputLines = lines
	m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
	m.recalcViewportHeight()
	m.viewport.GotoTop()

	histCmd := "findings"
	if isFiltered {
		histCmd = "findings " + filter
	}
	m.history = append(m.history, historyEntry{
		deity: "anubis", command: histCmd,
		output: strings.Join(lines, "\n"),
	})
	m.cmdHistory = deduplicateHistory(m.history)

	return m, nil
}

// renderFindingDetail formats a single finding with severity, size,
// description, advisory, remediation status, and path.
func (m TUIModel) renderFindingDetail(
	f jackal.PersistedFinding,
	body, dim, warn lipgloss.Style,
	fixableLabel string,
	breakingStyle lipgloss.Style,
) []string {
	severity := "  "
	switch f.Severity {
	case jackal.SeveritySafe:
		severity = lipgloss.NewStyle().Foreground(Green).Render("safe")
	case jackal.SeverityCaution:
		severity = warn.Render("caution")
	case jackal.SeverityWarning:
		severity = lipgloss.NewStyle().Foreground(Red).Render("warning")
	}

	// Main line: severity + size + description
	lines := []string{
		fmt.Sprintf("  %-9s %-8s %s",
			severity, f.SizeHuman, body.Render(f.Description)),
	}

	// Path (shortened for readability)
	if f.Path != "" {
		lines = append(lines, dim.Render("           "+ShortenPath(f.Path)))
	}

	// Advisory
	if f.Advisory != "" {
		lines = append(lines, dim.Render("           → "+f.Advisory))
	}

	// Remediation + fixability
	if f.CanFix && f.Remediation != "" {
		remLine := "           " + fixableLabel + "  " + dim.Render(f.Remediation)
		if f.Breaking {
			remLine += "  " + breakingStyle.Render("⚠ may affect running services")
		}
		lines = append(lines, remLine)
	}

	lines = append(lines, "") // spacing between findings
	return lines
}

// categoryIcon returns an emoji for a scan category.
func categoryIcon(cat string) string {
	switch cat {
	case "cache":
		return "🗑"
	case "logs":
		return "📋"
	case "build":
		return "🔨"
	case "containers":
		return "🐳"
	case "dev-tools", "dev":
		return "🔧"
	case "packages":
		return "📦"
	case "ai":
		return "🤖"
	case "ides":
		return "💻"
	case "cloud":
		return "☁️"
	case "storage":
		return "💾"
	case "vms":
		return "🖥"
	case "general":
		return "📁"
	default:
		return "📁"
	}
}

// isScanCategory returns true if the input matches a known Anubis scan
// category, enabling bare category names as shortcuts for drill-down.
func (m TUIModel) isScanCategory(raw string) bool {
	// Only treat bare words as categories if we have a recent scan in history.
	hasAnubisHistory := false
	for _, h := range m.history {
		if h.deity == "anubis" {
			hasAnubisHistory = true
			break
		}
	}
	if !hasAnubisHistory {
		return false
	}

	scan, err := jackal.LoadLatest()
	if err != nil {
		return false
	}
	lower := strings.ToLower(strings.TrimSpace(raw))
	for cat := range scan.ByCategory {
		if strings.ToLower(string(cat)) == lower {
			return true
		}
	}
	return false
}

// ── Helpers ───────────────────────────────────────────────────────────

func deityDisplay(key string) (string, string) {
	for _, d := range deityRoster {
		if d.Key == key {
			return d.Glyph, d.Name
		}
	}
	return "⚙", key
}

func (m *TUIModel) recalcViewportHeight() {
	// Reserve: header(2) + divider(1) + input divider(1) + input(1) + hints(1) + padding(2)
	vpHeight := m.height - 8
	if vpHeight < 5 {
		vpHeight = 5
	}
	m.viewport.SetHeight(vpHeight)

	rightWidth := m.width - leftPaneWidth - 5
	if rightWidth < 20 {
		rightWidth = 20
	}
	m.viewport.SetWidth(rightWidth)
}

func pluralize(word string, n int) string {
	if n == 1 {
		return word
	}
	if strings.HasSuffix(word, "y") {
		return word[:len(word)-1] + "ies"
	}
	return word + "s"
}

// ── Launcher ──────────────────────────────────────────────────────────

// LaunchTUI starts the BubbleTea TUI with optional notification awareness.
func LaunchTUI() error {
	return LaunchTUIWithNotify(nil)
}

// LaunchTUIWithNotify starts the TUI with an optional notification store.
// If store is non-nil, recent notifications are shown in the left pane.
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

// tuiPersistedState is the JSON-serializable TUI state saved between sessions.
type tuiPersistedState struct {
	DeityState map[string]deityRunState `json:"deity_state"`
	LastUsed   string                   `json:"last_used"`
}

func tuiStatePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pantheon", "tui-state.json")
}

// loadPersistedState reads saved deity state from disk.
func (m *TUIModel) loadPersistedState() {
	data, err := os.ReadFile(tuiStatePath())
	if err != nil {
		return // no saved state — fine
	}
	var state tuiPersistedState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}
	for k, v := range state.DeityState {
		m.deityState[k] = v
	}
}

// savePersistedState writes deity state to disk.
func (m *TUIModel) savePersistedState() {
	state := tuiPersistedState{
		DeityState: m.deityState,
		LastUsed:   time.Now().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	dir := filepath.Dir(tuiStatePath())
	_ = os.MkdirAll(dir, 0755)
	_ = os.WriteFile(tuiStatePath(), data, 0644)
}
