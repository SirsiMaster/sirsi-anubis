package tui

import "fmt"

// Command registry (docs/TUI_DESIGN_PROOF.md §3.4, §7 delta 2).
//
// v0.22's fatal trick was DECLARING key bindings that dispatched nowhere. Here
// every binding resolves through this registry, and — the structural guarantee
// — the status-bar hints are GENERATED from registered commands. A hint cannot
// exist for an unwired key, because the hint is a projection of the registry,
// not a hand-written string. ValidateHints turns that guarantee into a test.

// CommandID is a stable command identifier. Where a command mirrors a CLI verb
// it uses the same id, so there is no parallel TUI verb list to drift (delta 5).
type CommandID string

// Canonical command IDs. The verb subset mirrors the live cobra tree
// (scan, clean, status, audit, risk, hardware, ra, thread, router, maat) so the
// palette is a projection of the real command set; navigation/meta IDs are
// console-local.
const (
	CmdScan      CommandID = "scan"
	CmdClean     CommandID = "clean" // destructive — confirm modal required
	CmdStatus    CommandID = "status"
	CmdAudit     CommandID = "audit"
	CmdRisk      CommandID = "risk"
	CmdRaDeploy  CommandID = "ra.deploy"
	CmdRaStatus  CommandID = "ra.status"
	CmdRaKill    CommandID = "ra.kill" // destructive — confirm modal required
	CmdRouterAck CommandID = "router.ack"

	// Console-local navigation / meta commands.
	CmdInspect   CommandID = "inspect"
	CmdFilter    CommandID = "filter"
	CmdRefresh   CommandID = "refresh"
	CmdBack      CommandID = "back"
	CmdPalette   CommandID = "palette"
	CmdHelp      CommandID = "help"
	CmdQuit      CommandID = "quit"
	CmdFocusNext CommandID = "focus.next"
	CmdMoveUp    CommandID = "move.up"
	CmdMoveDown  CommandID = "move.down"
	CmdTop       CommandID = "top"
	CmdBottom    CommandID = "bottom"
)

// Command is a single wired action. Key is the zero-keystroke binding shown in
// the status bar (empty means palette-only). Hint is the terse verb shown
// beside the key. Destructive commands never execute from the keystroke alone —
// the reducer routes them to a confirm modal (§4, Rule A1).
type Command struct {
	ID          CommandID
	Title       string // palette / fuzzy-search name
	Key         string // status-bar key, e.g. "enter", "c", "/"; "" = palette-only
	Hint        string // status-bar verb, e.g. "inspect"
	Destructive bool
}

// Registry is the single source of truth for wired commands. It is the backing
// that every rendered affordance (status hint, palette entry) projects from.
type Registry struct {
	byID  map[CommandID]Command
	byKey map[string]CommandID
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		byID:  make(map[CommandID]Command),
		byKey: make(map[string]CommandID),
	}
}

// Register wires a command. A duplicate id or a key already bound to a
// different command is a programming error and returns an error so wiring
// mistakes surface in tests rather than as silent dead keys.
func (r *Registry) Register(c Command) error {
	if c.ID == "" {
		return fmt.Errorf("tui: command with empty id")
	}
	if _, dup := r.byID[c.ID]; dup {
		return fmt.Errorf("tui: command %q already registered", c.ID)
	}
	if c.Key != "" {
		if existing, clash := r.byKey[c.Key]; clash {
			return fmt.Errorf("tui: key %q already bound to %q", c.Key, existing)
		}
		r.byKey[c.Key] = c.ID
	}
	r.byID[c.ID] = c
	return nil
}

// Lookup returns the command for id.
func (r *Registry) Lookup(id CommandID) (Command, bool) {
	c, ok := r.byID[id]
	return c, ok
}

// ResolveKey maps a keypress to its wired command, the only path by which a key
// produces an action.
func (r *Registry) ResolveKey(key string) (Command, bool) {
	id, ok := r.byKey[key]
	if !ok {
		return Command{}, false
	}
	return r.byID[id], true
}

// IDs returns every registered command id (palette source).
func (r *Registry) IDs() []CommandID {
	out := make([]CommandID, 0, len(r.byID))
	for id := range r.byID {
		out = append(out, id)
	}
	return out
}

// DefaultRegistry wires the canonical console command set once. Views reference
// these ids by name; they never re-register, so there is a single source of
// truth for every wired key. Destructive verbs (clean, ra.kill) are flagged so
// the reducer routes them through a confirm modal rather than firing on a key.
func DefaultRegistry() (*Registry, error) {
	reg := NewRegistry()
	cmds := []Command{
		{ID: CmdMoveUp, Title: "Move up", Key: "up", Hint: "move"},
		{ID: CmdMoveDown, Title: "Move down", Key: "down", Hint: "move"},
		{ID: CmdInspect, Title: "Inspect selection", Key: "enter", Hint: "inspect"},
		{ID: CmdFilter, Title: "Filter table", Key: "/", Hint: "filter"},
		{ID: CmdRefresh, Title: "Refresh view", Key: "r", Hint: "refresh"},
		{ID: CmdBack, Title: "Back / dismiss", Key: "esc", Hint: "back"},
		{ID: CmdPalette, Title: "Command palette", Key: "ctrl+k", Hint: "palette"},
		{ID: CmdHelp, Title: "Help", Key: "?", Hint: "help"},
		{ID: CmdQuit, Title: "Quit", Key: "q", Hint: "quit"},
		{ID: CmdFocusNext, Title: "Focus next pane", Key: "tab", Hint: "pane"},
		{ID: CmdTop, Title: "Top of table", Key: "g", Hint: "top"},
		{ID: CmdBottom, Title: "Bottom of table", Key: "G", Hint: "bottom"},
		{ID: CmdScan, Title: "Scan workstation", Hint: "scan"},
		{ID: CmdClean, Title: "Clean findings", Key: "c", Hint: "clean", Destructive: true},
		{ID: CmdStatus, Title: "System status", Hint: "status"},
		{ID: CmdAudit, Title: "Quality audit", Hint: "audit"},
		{ID: CmdRisk, Title: "Risk assessment", Hint: "risk"},
		{ID: CmdRaDeploy, Title: "Ra deploy scope", Hint: "deploy"},
		{ID: CmdRaStatus, Title: "Ra fleet status", Hint: "fleet"},
		{ID: CmdRaKill, Title: "Ra kill node", Hint: "kill", Destructive: true},
		{ID: CmdRouterAck, Title: "Ack router item", Key: "a", Hint: "ack"},
	}
	for _, c := range cmds {
		if err := reg.Register(c); err != nil {
			return nil, err
		}
	}
	return reg, nil
}

// Hint is a rendered status-bar affordance: a key and the verb it triggers.
type Hint struct {
	Key   string
	Label string
}

// Hints projects the given command ids into status-bar hints, preserving order.
// It returns an error if any id is not registered or has no key — making
// "a visible hint references an unregistered/unwired command" a test failure,
// exactly the §7 delta-2 guarantee. Only registry-backed hints can be rendered.
func (r *Registry) Hints(ids []CommandID) ([]Hint, error) {
	hints := make([]Hint, 0, len(ids))
	for _, id := range ids {
		c, ok := r.byID[id]
		if !ok {
			return nil, fmt.Errorf("tui: hint references unregistered command %q", id)
		}
		if c.Key == "" {
			return nil, fmt.Errorf("tui: hint references palette-only command %q (no key)", id)
		}
		hints = append(hints, Hint{Key: c.Key, Label: c.Hint})
	}
	return hints, nil
}
