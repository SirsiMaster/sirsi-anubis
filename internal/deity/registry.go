// Package deity provides the canonical deity registry for the Sirsi Pantheon.
// All surfaces (TUI, menubar, dashboard, CLI) import from here to ensure
// glyphs, names, and roles are consistent. See ADR-025 (Deity Registry).
package deity

// Info describes a single deity in the Pantheon hierarchy.
type Info struct {
	Key   string // lowercase identifier used in CLI args and state keys
	Glyph string // Egyptian hieroglyph for display
	Name  string // display name (PascalCase)
	Role  string // two-word role description
}

// Roster is the canonical deity list, ordered by hierarchy (Rule D6).
// Horus → Ra → Net → Thoth → Ma'at → Isis → Seshat → Anubis → Seba → Osiris
var Roster = []Info{
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

// Lookup returns the Info for a deity key, or a fallback with the key as name.
func Lookup(key string) Info {
	for _, d := range Roster {
		if d.Key == key {
			return d
		}
	}
	return Info{Key: key, Glyph: "⚙", Name: key, Role: ""}
}

// Display returns (glyph, name) for a deity key.
func Display(key string) (string, string) {
	d := Lookup(key)
	return d.Glyph, d.Name
}

// Keys returns all deity keys in roster order.
func Keys() []string {
	keys := make([]string, len(Roster))
	for i, d := range Roster {
		keys[i] = d.Key
	}
	return keys
}
