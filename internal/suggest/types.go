// Package suggest provides context-aware action recommendations for all
// Sirsi Pantheon surfaces (TUI, CLI, menubar). It is a pure data layer —
// no rendering, no I/O, no side effects. Consumers format the output
// for their own UI (BubbleTea, lipgloss, systray menu items, etc.).
package suggest

// Action is a single recommended next step.
type Action struct {
	// Command is the CLI command string (e.g. "sirsi anubis judge --confirm").
	Command string

	// Short is a terse label for constrained UIs like menubar items.
	// Example: "Judge & clean"
	Short string

	// Description explains what the command does and why it's suggested now.
	// Example: "Review and clean safe items (dry-run by default)"
	Description string

	// Priority orders suggestions. Lower = more important. 0 is highest.
	Priority int
}

// Context captures what just happened so the engine can recommend next steps.
type Context struct {
	// Deity that ran (e.g. "anubis", "isis", "ra").
	Deity string

	// Subcommand that ran (e.g. "weigh", "judge", "network").
	Subcommand string

	// Err is the error from the command, if any. Nil means success.
	Err error

	// FindingsCount is the number of findings produced (Anubis-specific).
	// Zero means no findings or not applicable.
	FindingsCount int
}
