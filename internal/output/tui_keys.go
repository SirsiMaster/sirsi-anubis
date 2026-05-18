// Package output — tui_keys.go
//
// All key handling methods extracted from tui.go.
// Handles keyboard input dispatch for each view mode.
package output

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
)

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
	case viewSelect:
		return m.handleSelectKey(key)
	case viewAnalyze:
		return m.handleAnalyzeKey(key)
	}

	return m, nil
}

func (m TUIModel) handleTabKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		if m.activeTab > 0 {
			m.activeTab--
		}
		if m.activeTab == 4 {
			return m, liveTick()
		}
		return m, nil
	case "right", "l":
		if m.activeTab < len(tabs)-1 {
			m.activeTab++
		}
		if m.activeTab == 4 {
			return m, liveTick()
		}
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7":
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
			if i == 4 {
				return m, liveTick()
			}
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

func (m TUIModel) handleSelectKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.selectCursor > 0 {
			m.selectCursor--
		}
		return m, nil
	case "down", "j":
		if m.selectCursor < len(m.selectItems)-1 {
			m.selectCursor++
		}
		return m, nil
	case " ":
		if m.selectCursor < len(m.selectItems) {
			m.selectItems[m.selectCursor].Selected = !m.selectItems[m.selectCursor].Selected
		}
		return m, nil
	case "a":
		allSelected := true
		for _, item := range m.selectItems {
			if !item.Selected {
				allSelected = false
				break
			}
		}
		for i := range m.selectItems {
			m.selectItems[i].Selected = !allSelected
		}
		return m, nil
	case "enter":
		var selected []selectItem
		for _, item := range m.selectItems {
			if item.Selected {
				selected = append(selected, item)
			}
		}
		if len(selected) == 0 {
			m.mode = viewTabs
			return m, nil
		}
		if m.selectOnConfirm != nil {
			m.mode = viewRunning
			m.cmdStartTime = time.Now()
			m.outputLines = nil
			m.postRunCmds = nil
			onConfirm := m.selectOnConfirm
			return m, tea.Batch(m.spinner.Tick, elapsedTick(), func() tea.Msg {
				lines, deityKey, fixCmds, err := onConfirm(selected)
				return nativeResultMsg{lines: lines, deityKey: deityKey, fixCmds: fixCmds, err: err}
			})
		}
		m.mode = viewTabs
		return m, nil
	case "esc":
		m.mode = viewTabs
		m.selectItems = nil
		m.selectOnConfirm = nil
		return m, nil
	}
	return m, nil
}

func (m TUIModel) handleAnalyzeKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.analyzeCursor > 0 {
			m.analyzeCursor--
		}
		return m, nil
	case "down", "j":
		if m.analyzeCursor < len(m.analyzeEntries)-1 {
			m.analyzeCursor++
		}
		return m, nil
	case "enter", "right", "l":
		if m.analyzeCursor < len(m.analyzeEntries) {
			entry := m.analyzeEntries[m.analyzeCursor]
			if entry.IsDir {
				m.analyzeHistory = append(m.analyzeHistory, analyzeLevel{
					path: m.analyzePath, entries: m.analyzeEntries,
					total: m.analyzeTotal, cursor: m.analyzeCursor,
				})
				childPath := entry.Path
				m.mode = viewRunning
				m.runningCmd = "analyze " + ShortenPath(childPath)
				m.runningDeity = "anubis"
				m.cmdStartTime = time.Now()
				m.outputLines = nil
				m.postRunCmds = nil
				return m, tea.Batch(m.spinner.Tick, elapsedTick(), func() tea.Msg {
					res, err := jackal.Analyze(childPath, 0)
					if err != nil {
						return analyzeResultMsg{err: err}
					}
					return analyzeResultMsg{result: res}
				})
			}
		}
		return m, nil
	case "esc", "left", "h":
		if len(m.analyzeHistory) > 0 {
			last := m.analyzeHistory[len(m.analyzeHistory)-1]
			m.analyzeHistory = m.analyzeHistory[:len(m.analyzeHistory)-1]
			m.analyzePath = last.path
			m.analyzeEntries = last.entries
			m.analyzeTotal = last.total
			m.analyzeCursor = last.cursor
		} else {
			m.mode = viewTabs
			m.analyzeEntries = nil
			m.analyzeHistory = nil
		}
		return m, nil
	case "q":
		m.mode = viewTabs
		m.analyzeEntries = nil
		m.analyzeHistory = nil
		return m, nil
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
