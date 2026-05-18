package output

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/deity"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/notify"
	"github.com/SirsiMaster/sirsi-pantheon/internal/suggest"
)

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

		// Stream per-step progress for scan and doctor
		isScan := len(action.Args) >= 2 && action.Args[0] == "anubis" && action.Args[1] == "scan"
		isDoctor := len(action.Args) >= 2 && action.Args[0] == "isis" && action.Args[1] == "diagnose"
		if isScan || isDoctor {
			ch := make(chan string, 100)
			m.streamCh = ch
			if isScan {
				scanProgressCh = ch
			} else {
				doctorProgressCh = ch
			}

			label := "Scanning..."
			if isDoctor {
				label = "Diagnosing..."
			}
			m.outputLines = []string{
				"",
				"  " + lipgloss.NewStyle().Foreground(Gold).Bold(true).Render(label),
				"",
			}
			m.viewport.SetContent(strings.Join(m.outputLines, "\n"))

			fn := action.Native
			var streamErr error
			return m, tea.Batch(m.spinner.Tick, elapsedTick(), func() tea.Msg {
				go func() {
					res := fn()
					streamErr = res.err
					close(ch)
				}()
				line, ok := <-ch
				if !ok {
					return streamLineMsg{done: true, err: streamErr}
				}
				return streamLineMsg{line: line}
			})
		}

		fn := action.Native
		return m, tea.Batch(m.spinner.Tick, elapsedTick(), func() tea.Msg {
			res := fn()
			return nativeResultMsg(res)
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
		"scan": "anubis", "ghosts": "anubis", "clean": "anubis",
		"duplicates": "anubis", "purge": "anubis", "analyze": "anubis",
		"diagnose": "isis", "network": "isis", "fix": "isis", "monitor": "isis",
		"audit": "maat", "risk": "osiris",
		"hardware": "seba", "diagram": "seba",
		"learn": "seshat", "fleet": "ra", "index": "horus",
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
		m.lastDeity = m.runningDeity

		if m.runningDeity != "" {
			m.activeDeity[m.runningDeity] = true
		}
		m.history = append(m.history, historyEntry{
			deity: m.runningDeity, command: m.runningCmd,
			output: strings.Join(m.outputLines, "\n"),
		})
		m.cmdHistory = deduplicateHistory(m.history)

		if msg.err != nil {
			m.lastCommand = m.runningCmd
			m.lastSummary = "Failed"
			m.outputLines = append(m.outputLines, "",
				lipgloss.NewStyle().Foreground(Red).Render("  ✗ "+msg.err.Error()))
			if m.runningDeity != "" {
				m.deityState[m.runningDeity] = stateFailed
			}
		} else {
			m.lastCommand = m.runningCmd
			m.lastSummary = "Completed"
			if m.runningDeity != "" {
				state := stateSucceeded
				if m.runningDeity == "anubis" {
					// Replace streaming progress with final rendered scan result
					if scan, loadErr := jackal.LoadLatest(); loadErr == nil && len(scan.Findings) > 0 {
						state = stateHasData
						res := &jackal.ScanResult{
							Findings:   make([]jackal.Finding, len(scan.Findings)),
							TotalSize:  scan.TotalSize,
							RulesRan:   scan.RulesRan,
							ByCategory: make(map[jackal.Category]jackal.CategorySummary),
						}
						for i, f := range scan.Findings {
							res.Findings[i] = jackal.Finding{
								Description: f.Description, Path: f.Path,
								SizeBytes: f.SizeBytes, Severity: f.Severity,
								Category: f.Category, Advisory: f.Advisory,
								CanFix: f.CanFix, Remediation: f.Remediation,
							}
						}
						for cat, s := range scan.ByCategory {
							res.ByCategory[cat] = s
						}
						m.outputLines = RenderScanResult(res)
						// Offer cleanup
						safeCount := 0
						for _, f := range res.Findings {
							if f.Severity == jackal.SeveritySafe && f.CanFix {
								safeCount++
							}
						}
						if safeCount > 0 {
							m.postRunCmds = []string{"anubis clean --dry-run"}
							m.postRunActions = []suggest.Action{{
								Command:     "anubis clean --dry-run",
								Description: "Preview safe items to clean",
							}}
						}
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
	// If the result carries an analyze result, enter analyze mode.
	if msg.analyzeRes != nil && msg.err == nil {
		m.mode = viewAnalyze
		m.analyzePath = msg.analyzeRes.Path
		m.analyzeEntries = msg.analyzeRes.Entries
		m.analyzeTotal = msg.analyzeRes.TotalSize
		m.analyzeCursor = 0
		m.analyzeHistory = nil
		m.runningDeity = ""
		m.runningCmd = ""
		m.runningArgs = nil
		return m, nil
	}

	// If the result carries a select request, enter checkbox mode.
	if msg.selectReq != nil && msg.err == nil {
		m.mode = viewSelect
		m.selectTitle = msg.selectReq.title
		m.selectItems = msg.selectReq.items
		m.selectCursor = 0
		m.selectOnConfirm = msg.selectReq.onConfirm
		m.runningDeity = ""
		m.runningCmd = ""
		m.runningArgs = nil
		return m, nil
	}

	m.mode = viewDone
	m.lastDeity = msg.deityKey
	m.runningDeity = msg.deityKey

	if msg.err != nil {
		m.lastCommand = m.runningCmd
		m.lastSummary = "Failed"
		m.outputLines = []string{
			"",
			"  " + lipgloss.NewStyle().Foreground(Red).Render("✗ "+msg.err.Error()),
		}
		if msg.deityKey != "" {
			m.deityState[msg.deityKey] = stateFailed
		}
	} else {
		m.lastCommand = m.runningCmd
		m.lastSummary = "Completed"
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
	// Known fix descriptions
	fixDescs := map[string]string{
		"anubis clean --dry-run": "Preview safe items to clean",
		"anubis clean --confirm": "Clean safe items (move to Trash)",
		"anubis scan":            "Run a fresh scan",
		"isis fix":               "Auto-fix DNS, firewall, security",
		"isis diagnose":          "Full system health check",
	}

	if len(msg.fixCmds) > 0 {
		m.postRunCmds = msg.fixCmds
		m.postRunActions = nil
		for _, cmd := range msg.fixCmds {
			desc := fixDescs[cmd]
			if desc == "" {
				desc = "Run " + cmd
			}
			m.postRunActions = append(m.postRunActions, suggest.Action{
				Command:     cmd,
				Description: desc,
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
