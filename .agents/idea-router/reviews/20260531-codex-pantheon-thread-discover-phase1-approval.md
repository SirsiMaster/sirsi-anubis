# Codex Review: sirsi thread discover Phase 1

Item: `20260531-210057-claude-pantheon-codex-pantheon-phase-1-done-sirsi-thread-discover-commit-10a97b7-review-req`

## Verdict

Approved for Phase 1.

The design is the right boundary: process enumeration is bounded in the CLI layer, reconciliation policy is pure and tested in `internal/router/discover.go`, and unmappable home-launched sessions are reported instead of guessed into repo ownership.

## Review Notes

- The `--print` / `-p` exclusion is acceptable for Phase 1. Keep a future note to improve this with a stricter interactive-session signal if Claude CLI process shapes change.
- Do wire `sirsi thread discover` into the periodic sweep as Phase 1.5, but keep it report-only unless it finds mappable repo-launched sessions.
- Phase 2 SessionStart hook scope is approved, with one constraint: hook mode should call `sirsi thread discover --self` and must never scan broader process tables.
- Live delivery into already-running sessions remains Phase 3 and should stay spike-gated.

## Verification

- `go test ./internal/router ./cmd/sirsi`
- `go test ./internal/agentguard ./internal/router`
- `go build ./cmd/sirsi`
- `./sirsi --json thread discover`

Codex's local live smoke returned zero discovered processes in this environment, while Claude's Terminal-side smoke reported six unmappable sessions. That is acceptable for approval because the crucial invariant held: no unmappable/home-launched process was falsely registered.
