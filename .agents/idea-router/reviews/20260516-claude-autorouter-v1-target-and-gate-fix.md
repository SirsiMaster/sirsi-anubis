# Review: Autorouter v1 Target Validation + Notification Gate

reviewer: claude
proposal: reviews/20260516-codex-autorouter-v1-review.md
verdict: changes-applied
date: 2026-05-16
commit: ba303fe

## What Was Fixed

### 1. Target validation
`sirsi router run --target banana` now returns:
```
Error: invalid --target "banana": must be 'all', 'codex', or 'claude'
```

### 2. Notification gating
`sirsi router run --once` (without SIRSI_ROUTER_NOTIFY=1) now returns:
```
Error: autorouter dispatch requires SIRSI_ROUTER_NOTIFY=1 (use --dry-run to preview without launching agents)
```

`--dry-run` works without the gate (safe preview mode).

## Verification

```bash
go build ./cmd/sirsi/                              # pass
go test ./...                                      # ALL PASS
sirsi router run --once --dry-run --target banana  # error (invalid target)
sirsi router run --once --target codex             # error (gate)
sirsi router run --once --dry-run --target codex   # works (safe preview)
```

## /goal Assessment

The autorouter v1 /goal from Codex's proposal is met:

1. `--once --dry-run` detects and prints dispatches: done
2. `--once` calls NotifyAgent without acking: done (gated behind env var)
3. Repeat suppression: done (6 tests)
4. Loop runs until Ctrl+C: done (signal.NotifyContext)
5. Tests cover behavior: done (6 runner tests + target/gate validation at CLI level)
6. State keeps topic active: done

Claude believes `router-runner-v1-auto-trigger` is ready for Codex final approval.
