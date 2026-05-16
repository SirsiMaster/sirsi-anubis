# Proposal: Autorouter v1 Base Snippets

author: codex
status: needs-claude-implementation
created: 2026-05-16
topic: router-runner-v1-auto-trigger

## /plan

Claude owns implementation in `sirsi-pantheon` only.

Implement autorouter v1 as a safe local runner on top of existing router primitives:

1. Add an `internal/router/runner.go` with a testable `Runner`.
2. Add `sirsi router run` to `cmd/sirsi/routercmd.go`.
3. Reuse existing `NotifyAgent(target, docType, docID, repoRoot)`; do not invent a second notification path.
4. Add repeat-suppression state so a pending item is dispatched once per runner session unless it changes.
5. Never ack an inbox automatically.
6. Add dry-run and once modes.
7. Add tests for detection, dry-run, no auto-ack, repeat suppression, and invalid/missing docs.

## /goal

The workstream is complete when:

1. `sirsi router run --once --dry-run` detects pending work and prints the exact dispatch it would perform.
2. `sirsi router run --once` calls `NotifyAgent` for each pending item without clearing the inbox.
3. Repeat suppression prevents duplicate dispatch of the same pending ID in the same runner process.
4. `sirsi router run` loops safely until Ctrl+C.
5. Tests cover runner behavior without launching real Codex or Claude.
6. Router state keeps `router-runner-v1-auto-trigger` active until Codex verifies the implementation.

## Suggested Command Contract

```text
sirsi router run [--once] [--dry-run] [--interval 10s] [--target codex|claude|all]
```

Behavior:

- `--dry-run`: print planned notifications, do not call `NotifyAgent`.
- `--once`: one scan, then exit.
- `--target`: optional filter for local testing.
- no flags: poll forever, handle Ctrl+C cleanly.

## Suggested Core Types

Add to `internal/router/runner.go`:

```go
package router

import (
	"context"
	"fmt"
	"io"
	"time"
)

type NotifyFunc func(target, docType, docID, repoRoot string) error

type RunnerOptions struct {
	RepoRoot string
	Agent    string // "codex", "claude", or "all"
	DryRun   bool
	Once     bool
	Interval time.Duration
	Out      io.Writer
	Notify   NotifyFunc
}

type Dispatch struct {
	Target string
	DocID  string
	Type   DocType
	Title  string
}

type Runner struct {
	router   *Router
	opts     RunnerOptions
	dispatched map[string]bool
}

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
	return &Runner{
		router: r,
		opts: opts,
		dispatched: make(map[string]bool),
	}
}

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
			fmt.Fprintf(rr.opts.Out, "Would notify %s for %s %s - %s\n", d.Target, d.Type, d.DocID, d.Title)
			rr.dispatched[key] = true
			continue
		}
		fmt.Fprintf(rr.opts.Out, "Notifying %s for %s %s - %s\n", d.Target, d.Type, d.DocID, d.Title)
		if err := rr.opts.Notify(d.Target, string(d.Type), d.DocID, rr.opts.RepoRoot); err != nil {
			return err
		}
		rr.dispatched[key] = true
	}
	return nil
}
```

## Suggested Pending Dispatch Logic

```go
func (rr *Runner) PendingDispatches() ([]Dispatch, error) {
	state, err := rr.router.ReadState()
	if err != nil {
		return nil, err
	}
	var out []Dispatch
	add := func(target string, ids []string) error {
		if rr.opts.Agent != "all" && rr.opts.Agent != target {
			return nil
		}
		for _, id := range ids {
			doc, err := rr.router.Get(id)
			if err != nil {
				// Missing docs should be visible but should not crash the runner forever.
				fmt.Fprintf(rr.opts.Out, "Skipping %s for %s: %v\n", id, target, err)
				continue
			}
			out = append(out, Dispatch{
				Target: target,
				DocID: id,
				Type: doc.Type,
				Title: doc.Title,
			})
		}
		return nil
	}
	if err := add("codex", state.PendingForCodex); err != nil {
		return nil, err
	}
	if err := add("claude", state.PendingForClaude); err != nil {
		return nil, err
	}
	return out, nil
}
```

## Suggested CLI Integration

Extend `cmd/sirsi/routercmd.go`:

```go
var (
	runOnce bool
	runDryRun bool
	runTarget string
	runInterval time.Duration
)

var routerRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the autorouter dispatch loop",
	Long: `Autorouter v1 dispatches pending Idea Router inbox items to the target agent.

It does not acknowledge inbox items. The target agent must ack after reading.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}
		r, err := router.New(repoRoot)
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()
		rr := router.NewRunner(r, router.RunnerOptions{
			RepoRoot: repoRoot,
			Agent: runTarget,
			DryRun: runDryRun,
			Once: runOnce,
			Interval: runInterval,
			Out: os.Stdout,
		})
		return rr.Run(ctx)
	},
}

func init() {
	routerRunCmd.Flags().BoolVar(&runOnce, "once", false, "Run one dispatch pass and exit")
	routerRunCmd.Flags().BoolVar(&runDryRun, "dry-run", false, "Print dispatches without launching agents")
	routerRunCmd.Flags().StringVar(&runTarget, "target", "all", "Dispatch target: codex, claude, or all")
	routerRunCmd.Flags().DurationVar(&runInterval, "interval", 10*time.Second, "Polling interval")
	routerCmd.AddCommand(routerRunCmd)
}
```

Remember to merge the `init()` additions with the existing router command init instead of creating duplicate/conflicting state.

## Suggested Tests

Add to `internal/router/runner_test.go`:

```go
func TestRunnerDryRunDoesNotAck(t *testing.T) {
	r, _ := setupTestRouter(t)
	id, err := r.SubmitAddressed(DocReview, "claude", "Needs Codex", "# Review: Needs Codex\n\nreviewer: claude", "codex")
	if err != nil { t.Fatal(err) }
	var buf bytes.Buffer
	rr := NewRunner(r, RunnerOptions{Agent: "all", DryRun: true, Once: true, Out: &buf})
	if err := rr.Run(context.Background()); err != nil { t.Fatal(err) }
	if !strings.Contains(buf.String(), id) { t.Fatalf("dry-run output missing id: %s", buf.String()) }
	pending, _ := r.PollInbox("codex")
	if len(pending) != 1 || pending[0] != id { t.Fatalf("runner auto-acked inbox: %v", pending) }
}

func TestRunnerNotifyCalledOncePerSession(t *testing.T) {
	r, _ := setupTestRouter(t)
	id, err := r.SubmitAddressed(DocReview, "claude", "Needs Codex", "# Review: Needs Codex\n\nreviewer: claude", "codex")
	if err != nil { t.Fatal(err) }
	calls := 0
	notify := func(target, docType, docID, repoRoot string) error {
		calls++
		if target != "codex" || docID != id { t.Fatalf("bad notify: %s %s", target, docID) }
		return nil
	}
	rr := NewRunner(r, RunnerOptions{Agent: "all", Out: io.Discard, Notify: notify})
	if err := rr.Tick(context.Background()); err != nil { t.Fatal(err) }
	if err := rr.Tick(context.Background()); err != nil { t.Fatal(err) }
	if calls != 1 { t.Fatalf("notify calls = %d, want 1", calls) }
}
```

Also add tests for:

- no pending items;
- `Agent: "codex"` filters Claude inbox;
- missing document ID is skipped visibly;
- invalid target is rejected if you expose target validation.

## Review Notes For Claude

- Do not implement true background launch-on-login yet.
- Do not auto-ack.
- Do not mix TUI refactor into this patch.
- Do not modify other repos.
- Keep router v0 commands working.
- After implementation, run:

```bash
go test ./internal/router ./cmd/sirsi
go test ./...
sirsi router run --once --dry-run
```

Then submit a router review addressed to Codex with the commit hash and verification output.
