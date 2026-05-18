// Package output — tui_view.go
//
// All View-related rendering methods extracted from tui.go.
// Main View(), tab bar, tab pages, status dashboard, running/done screens,
// bottom hints, health score, cards, and layout helpers.
package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/deity"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/suggest"
	"github.com/SirsiMaster/sirsi-pantheon/internal/vitals"
)

// ── View ─────────────────────────────────────────────────────────────

func (m TUIModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	var b strings.Builder
	maxW := min(m.width-2, 120)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
	dimText := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
	gold := lipgloss.NewStyle().Foreground(Gold)
	heavyDiv := dim.Render(strings.Repeat("━", maxW))
	lightDiv := dim.Render(strings.Repeat("─", maxW))

	// ── Persistent Header — "This is Pantheon" ──
	ver := readVersionFile()
	titleLine := "  " + gold.Bold(true).Render("𓉴 PANTHEON")
	if ver != "" {
		pad := maxW - visibleLen(titleLine) - len(ver) - 2
		if pad < 1 {
			pad = 1
		}
		titleLine += strings.Repeat(" ", pad) + dimText.Render(ver)
	}
	tagline := "  " + dimText.Render("Unified DevOps Intelligence")
	urlPad := maxW - visibleLen(tagline) - len("sirsi.ai") - 2
	if urlPad < 1 {
		urlPad = 1
	}
	tagline += strings.Repeat(" ", urlPad) + dimText.Render("sirsi.ai")

	b.WriteString("\n")
	b.WriteString(titleLine + "\n")
	b.WriteString(tagline + "\n")
	b.WriteString(" " + heavyDiv + "\n")

	// ── Tab bar ──
	b.WriteString(m.renderTabBar())
	b.WriteString(" " + lightDiv + "\n")

	// ── Content ──
	switch m.mode {
	case viewTabs:
		b.WriteString(m.renderTabPage())
	case viewRunning:
		b.WriteString(m.renderRunning())
	case viewDone:
		b.WriteString(m.renderDone())
	case viewPrompt:
		b.WriteString(m.renderTabPage())
	case viewSelect:
		b.WriteString(m.renderSelect())
	case viewAnalyze:
		b.WriteString(m.renderAnalyze())
	}

	// ── Bottom bar ──
	b.WriteString(" " + lightDiv + "\n")
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
	// Footer: faint brand + uptime if available
	footerRight := ""
	if m.vitals.UptimeStr != "" {
		footerRight = dimText.Render("up " + m.vitals.UptimeStr)
	}
	footerLeft := dimText.Render(" 𓉴 sirsi-pantheon")
	footPad := maxW - visibleLen(footerLeft) - visibleLen(footerRight)
	if footPad < 1 {
		footPad = 1
	}
	content += footerLeft + strings.Repeat(" ", footPad) + footerRight

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// readVersionFile reads the VERSION file from the repo root or ~/.config.
func readVersionFile() string {
	for _, p := range []string{"VERSION", filepath.Join(os.Getenv("HOME"), ".config", "sirsi", "VERSION")} {
		if data, err := os.ReadFile(p); err == nil {
			v := strings.TrimSpace(string(data))
			if v != "" {
				return "v" + v
			}
		}
	}
	return ""
}

// renderTabBar draws the horizontal tab switcher.
// Tab names only — no hieroglyphs here (inconsistent terminal widths).
// Glyphs live in the tab landing pages where alignment doesn't matter.
func (m TUIModel) renderTabBar() string {
	active := lipgloss.NewStyle().Foreground(Gold).Bold(true)
	inactive := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	dot := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render("·")

	var parts []string
	for i, tab := range tabs {
		if i == m.activeTab {
			parts = append(parts, active.Render("▸ "+tab.Name))
		} else {
			parts = append(parts, inactive.Render("  "+tab.Name))
		}
	}

	return "  " + strings.Join(parts, "  "+dot+"  ") + "\n"
}

// renderTabPage draws the landing page for the active tab.
func (m TUIModel) renderTabPage() string {
	tab := tabs[m.activeTab]
	gold := lipgloss.NewStyle().Foreground(Gold)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	body := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	var b strings.Builder

	if tab.Name == "Status" {
		b.WriteString(m.renderStatusPage(gold, dim))
	} else {
		b.WriteString("\n")
		b.WriteString("  " + gold.Bold(true).Render(tab.Glyph+"  "+tab.Name) + "\n")
		b.WriteString("  " + lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#555555")).
			Render(tab.Tagline) + "\n")
		b.WriteString("\n")

		for i, action := range tab.Actions {
			keyBadge := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0F0F0F")).
				Background(Gold).
				Bold(true).
				Padding(0, 1).
				Render(fmt.Sprintf("%d", i+1))
			b.WriteString("  " + keyBadge +
				"  " + body.Bold(true).Render(action.Label) + "\n")
			b.WriteString("       " + dim.Render(action.Desc) + "\n\n")
		}

		if m.lastCommand != "" {
			b.WriteString("  " + dim.Render("Last run: ") + body.Render(m.lastCommand))
			if m.lastSummary != "" {
				b.WriteString(dim.Render(" — " + m.lastSummary))
			}
			b.WriteString("\n")
			if len(m.postRunCmds) > 0 {
				b.WriteString("  " + dim.Render("Recommended next: ") + body.Render(strings.Join(firstN(m.postRunCmds, 3), "  · ")) + "\n")
			}
		}
	}

	return b.String()
}

// visibleLen estimates the visible length of a string, stripping ANSI escapes.
func visibleLen(s string) int {
	n := 0
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		n++
	}
	return n
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

	b.WriteString("\n")
	b.WriteString(m.viewport.View() + "\n")

	// ── Completion banner — decree from the deity that ran ──
	bannerMsg := "Judgment Complete"
	bannerStyle := green
	if m.lastDeity != "" {
		glyph, name := deity.Display(m.lastDeity)
		if m.deityState[m.lastDeity] == stateFailed {
			bannerMsg = glyph + " " + name + " — Failed"
			bannerStyle = lipgloss.NewStyle().Foreground(Red)
		} else {
			bannerMsg = glyph + " " + name + " — Complete"
		}
	}
	b.WriteString("\n")
	b.WriteString("  " + ResultBanner(bannerMsg, bannerStyle, 50) + "\n")
	b.WriteString("\n")

	// ── Numbered next actions ──
	if len(m.postRunCmds) > 0 {
		b.WriteString("  " + dim.Render("What's next?") + "\n\n")
	}
	shown := 0
	for i, cmd := range m.postRunCmds {
		if i >= 3 {
			break
		}
		desc := ""
		if i < len(m.postRunActions) {
			desc = m.postRunActions[i].Description
		}
		keyBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F0F0F")).
			Background(Gold).
			Bold(true).
			Padding(0, 1).
			Render(fmt.Sprintf("%d", i+1))
		line := "   " + keyBadge + "  " + body.Render(cmd)
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
		n := len(tabs[m.activeTab].Actions)
		hints = []string{"←/→ switch tabs", fmt.Sprintf("1-%d act", n), ": command", "q quit"}
	case viewRunning:
		hints = []string{"↑/↓ scroll", "ctrl+c cancel"}
	case viewDone:
		if len(m.postRunCmds) > 0 {
			hints = []string{"1-3 next action", "↑/↓ scroll", ": command", "esc back"}
		} else {
			hints = []string{"↑/↓ scroll", ": command", "esc back"}
		}
	case viewSelect:
		hints = []string{"↑/↓ move", "space toggle", "a all", "enter confirm", "esc cancel"}
	case viewAnalyze:
		hints = []string{"↑/↓ navigate", "enter drill-down", "esc back", "q quit"}
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

func (m *TUIModel) refreshVitals() {
	m.vitals = vitals.Collect()
}

const maxHistory = 60

func (m *TUIModel) appendHistory() {
	m.cpuHistory = appendCapped(m.cpuHistory, m.vitals.CPUPercent)
	m.memHistory = appendCapped(m.memHistory, m.vitals.RAMPercent)
	downNorm := m.vitals.NetDownBps / (10 * 1024 * 1024) * 100
	upNorm := m.vitals.NetUpBps / (10 * 1024 * 1024) * 100
	m.netDownHist = appendCapped(m.netDownHist, downNorm)
	m.netUpHist = appendCapped(m.netUpHist, upNorm)
}

func appendCapped(buf []float64, val float64) []float64 {
	buf = append(buf, val)
	if len(buf) > maxHistory {
		buf = buf[len(buf)-maxHistory:]
	}
	return buf
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

func firstN(values []string, n int) []string {
	if len(values) <= n {
		return values
	}
	return values[:n]
}
