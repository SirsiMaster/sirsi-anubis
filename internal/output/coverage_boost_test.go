package output

import (
	"strings"
	"testing"

	"github.com/SirsiMaster/sirsi-pantheon/internal/deity"
)

// ── Tab Definitions ──────────────────────────────────────────────────

func TestTabsExist(t *testing.T) {
	t.Parallel()
	if len(tabs) != 5 {
		t.Errorf("expected 5 tabs, got %d", len(tabs))
	}
	names := []string{"Scan", "Health", "Quality", "Intel", "Status"}
	for i, want := range names {
		if tabs[i].Name != want {
			t.Errorf("tab %d: got %q, want %q", i, tabs[i].Name, want)
		}
	}
}

func TestTabActionsNotEmpty(t *testing.T) {
	t.Parallel()
	for _, tab := range tabs {
		if len(tab.Actions) == 0 {
			t.Errorf("tab %q has no actions", tab.Name)
		}
		for _, a := range tab.Actions {
			if a.Label == "" || a.Desc == "" || len(a.Args) == 0 {
				t.Errorf("tab %q action %q: incomplete (label=%q desc=%q args=%v)",
					tab.Name, a.Label, a.Label, a.Desc, a.Args)
			}
		}
	}
}

// ── Model ────────────────────────────────────────────────────────────

func TestNewTUIModel(t *testing.T) {
	t.Parallel()
	m := NewTUIModel()
	if m.mode != viewTabs {
		t.Errorf("initial mode should be viewTabs, got %d", m.mode)
	}
	if m.activeTab != 0 {
		t.Errorf("initial tab should be 0, got %d", m.activeTab)
	}
	if m.deityState == nil {
		t.Error("deityState map should be initialized")
	}
}

// ── Tab Bar Rendering ────────────────────────────────────────────────

func TestRenderTabBar(t *testing.T) {
	t.Parallel()
	m := NewTUIModel()
	m.width = 120
	bar := m.renderTabBar()

	// Tab bar contains tab names
	for _, tab := range tabs {
		if !strings.Contains(bar, tab.Name) {
			t.Errorf("tab bar should contain name %q", tab.Name)
		}
	}
	// Active tab should have indicator
	if !strings.Contains(bar, "▸") {
		t.Error("tab bar should contain active indicator")
	}
}

// ── Tab Page Rendering ──────────────────────────────────────────────

func TestRenderTabPage(t *testing.T) {
	t.Parallel()
	m := NewTUIModel()
	m.width = 120
	m.height = 40

	for i, tab := range tabs {
		if tab.Name == "Status" {
			continue // Status uses special bento grid renderer
		}
		m.activeTab = i
		page := m.renderTabPage()
		if !strings.Contains(page, tab.Tagline) {
			t.Errorf("tab %q page should contain tagline %q", tab.Name, tab.Tagline)
		}
	}
}

func TestRenderTabPageActions(t *testing.T) {
	t.Parallel()
	m := NewTUIModel()
	m.width = 120
	m.height = 40
	m.activeTab = 0 // Scan tab

	page := m.renderTabPage()
	for _, action := range tabs[0].Actions {
		if !strings.Contains(page, action.Label) {
			t.Errorf("Scan tab should contain action %q", action.Label)
		}
	}
}

// ── Bottom Hints ────────────────────────────────────────────────────

func TestRenderBottomHints(t *testing.T) {
	t.Parallel()
	m := NewTUIModel()

	// Tab mode
	m.mode = viewTabs
	hints := m.renderBottomHints()
	if !strings.Contains(hints, "switch tabs") {
		t.Errorf("tab hints should contain 'switch tabs', got %q", hints)
	}

	// Running mode
	m.mode = viewRunning
	hints = m.renderBottomHints()
	if !strings.Contains(hints, "cancel") {
		t.Errorf("running hints should contain 'cancel', got %q", hints)
	}

	// Done mode
	m.mode = viewDone
	hints = m.renderBottomHints()
	if !strings.Contains(hints, "esc back") {
		t.Errorf("done hints should contain 'esc back', got %q", hints)
	}
}

// ── Status Page ─────────────────────────────────────────────────────

func TestRenderStatusPage(t *testing.T) {
	t.Parallel()
	m := NewTUIModel()
	m.width = 120
	m.height = 40
	m.activeTab = 4 // Status

	page := m.renderTabPage()
	if !strings.Contains(page, "CPU") {
		t.Error("status page should contain CPU section")
	}
	if !strings.Contains(page, "Memory") {
		t.Error("status page should contain Memory section")
	}
	if !strings.Contains(page, "Health") {
		t.Error("status page should contain Health score")
	}
	if !strings.Contains(page, "Deities") {
		t.Error("status page should contain Deities section")
	}
}

// ── Deity State ─────────────────────────────────────────────────────

func TestDeityStateDisplay(t *testing.T) {
	t.Parallel()
	m := NewTUIModel()
	m.width = 120
	m.height = 40
	m.activeTab = 4

	m.deityState["anubis"] = stateSucceeded
	m.deityState["isis"] = stateFailed

	page := m.renderTabPage()
	if !strings.Contains(page, "✓") {
		t.Error("status page should show ✓ for succeeded deity")
	}
	if !strings.Contains(page, "✗") {
		t.Error("status page should show ✗ for failed deity")
	}
}

// ── Pluralize ───────────────────────────────────────────────────────

func TestPluralize(t *testing.T) {
	t.Parallel()
	if got := pluralize("deity", 1); got != "deity" {
		t.Errorf("pluralize(deity, 1) = %q", got)
	}
	if got := pluralize("deity", 2); got != "deities" {
		t.Errorf("pluralize(deity, 2) = %q", got)
	}
	if got := pluralize("test", 3); got != "tests" {
		t.Errorf("pluralize(test, 3) = %q", got)
	}
}

// ── Deity Registry Integration ──────────────────────────────────────

func TestDeityRegistryUsed(t *testing.T) {
	t.Parallel()
	// Verify the TUI uses the shared registry
	if len(deity.Roster) != 10 {
		t.Errorf("expected 10 deities in shared registry, got %d", len(deity.Roster))
	}
	g, n := deity.Display("anubis")
	if g != "𓃣" || n != "Anubis" {
		t.Errorf("deity.Display(anubis) = (%q, %q)", g, n)
	}
}

// ── History Dedup ───────────────────────────────────────────────────

func TestDeduplicateHistory(t *testing.T) {
	t.Parallel()
	history := []historyEntry{
		{command: "scan"},
		{command: "doctor"},
		{command: "scan"},
	}
	result := deduplicateHistory(history)
	if len(result) != 2 {
		t.Errorf("expected 2 unique commands, got %d: %v", len(result), result)
	}
}

func TestTUIModelPersistedStateRoundTripIncludesRecommendations(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	m := NewTUIModel()
	m.deityState["anubis"] = stateHasData
	m.lastCommand = "scan"
	m.lastSummary = "Completed"
	m.postRunCmds = []string{"clean --dry-run", "diagnose"}
	m.savePersistedState()

	restored := NewTUIModel()
	restored.loadPersistedState()

	if restored.deityState["anubis"] != stateHasData {
		t.Fatalf("restored anubis state = %v, want %v", restored.deityState["anubis"], stateHasData)
	}
	if restored.lastCommand != "scan" {
		t.Fatalf("restored lastCommand = %q, want scan", restored.lastCommand)
	}
	if restored.lastSummary != "Completed" {
		t.Fatalf("restored lastSummary = %q, want Completed", restored.lastSummary)
	}
	if strings.Join(restored.postRunCmds, ",") != "clean --dry-run,diagnose" {
		t.Fatalf("restored postRunCmds = %#v", restored.postRunCmds)
	}
}
