package router

import (
	"context"
	"fmt"
	"io"
	"time"
)

// NotifyFunc is the function signature for agent notification.
type NotifyFunc func(target, docType, docID, repoRoot string) error

// RunnerOptions configures the autorouter dispatch loop.
type RunnerOptions struct {
	RepoRoot string
	Agent    string // "codex", "claude", or "all"
	DryRun   bool
	Once     bool
	Interval time.Duration
	Out      io.Writer
	Notify   NotifyFunc
}

// Dispatch represents a single pending notification to deliver.
type Dispatch struct {
	Target string
	DocID  string
	Type   DocType
	Title  string
}

// Runner polls the idea-router state and dispatches notifications
// for pending inbox items. It does NOT acknowledge items — the target
// agent must ack after reading.
type Runner struct {
	router     *Router
	opts       RunnerOptions
	dispatched map[string]bool
}

// NewRunner creates a Runner with the given options.
func NewRunner(r *Router, opts RunnerOptions) *Runner {
	if opts.Interval == 0 {
		opts.Interval = 10 * time.Second
	}
	if opts.Agent == "" {
		opts.Agent = "all"
	}
	if opts.Notify == nil {
		opts.Notify = NotifyAgent
	}
	if opts.Out == nil {
		opts.Out = io.Discard
	}
	return &Runner{
		router:     r,
		opts:       opts,
		dispatched: make(map[string]bool),
	}
}

// Run executes the dispatch loop until context is cancelled or --once completes.
func (rr *Runner) Run(ctx context.Context) error {
	for {
		if err := rr.Tick(ctx); err != nil {
			return err
		}
		if rr.opts.Once {
			return nil
		}
		timer := time.NewTimer(rr.opts.Interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}

// Tick performs a single scan of pending dispatches and notifies as needed.
func (rr *Runner) Tick(ctx context.Context) error {
	dispatches, err := rr.PendingDispatches()
	if err != nil {
		return err
	}
	if len(dispatches) == 0 {
		fmt.Fprintln(rr.opts.Out, "No pending dispatches.")
		return nil
	}
	for _, d := range dispatches {
		key := d.Target + ":" + d.DocID
		if rr.dispatched[key] {
			continue
		}
		if rr.opts.DryRun {
			fmt.Fprintf(rr.opts.Out, "[dry-run] Would notify %s for %s %s — %s\n", d.Target, d.Type, d.DocID, d.Title)
			rr.dispatched[key] = true
			continue
		}
		fmt.Fprintf(rr.opts.Out, "Notifying %s for %s %s — %s\n", d.Target, d.Type, d.DocID, d.Title)
		if err := rr.opts.Notify(d.Target, string(d.Type), d.DocID, rr.opts.RepoRoot); err != nil {
			fmt.Fprintf(rr.opts.Out, "  Warning: notification failed: %v\n", err)
			// Don't mark as dispatched so it retries next tick
			continue
		}
		rr.dispatched[key] = true
	}
	return nil
}

// PendingDispatches reads the router state and returns items that need dispatch.
func (rr *Runner) PendingDispatches() ([]Dispatch, error) {
	state, err := rr.router.ReadState()
	if err != nil {
		return nil, err
	}
	var out []Dispatch
	add := func(target string, ids []string) {
		if rr.opts.Agent != "all" && rr.opts.Agent != target {
			return
		}
		for _, id := range ids {
			doc, err := rr.router.Get(id)
			if err != nil {
				fmt.Fprintf(rr.opts.Out, "Skipping %s for %s: %v\n", id, target, err)
				continue
			}
			out = append(out, Dispatch{
				Target: target,
				DocID:  id,
				Type:   doc.Type,
				Title:  doc.Title,
			})
		}
	}
	add("codex", state.PendingForCodex)
	add("claude", state.PendingForClaude)
	return out, nil
}
