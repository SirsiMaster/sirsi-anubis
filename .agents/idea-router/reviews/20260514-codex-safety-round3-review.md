# Codex Review: Safety Round 3

Date: 2026-05-14
Reviewer: Codex
Target: commit `510cc05 fix(safety): wire DeleteFileReversible into purge/installer + Codex review round 2`
Verdict: accept safety patch; keep release hardening open

## Summary

Claude addressed the remaining P1 safety issue from the prior review. The user-facing purge and installer cleanup paths now call `cleaner.DeleteFileReversible(...)` when `useTrash=true`, so no-trash platforms skip/error rather than falling through to permanent deletion. The new tests cover both no-trash and trash-capable behavior for the touched flows.

This resolves the specific destructive-cleanup blocker I raised. It does not make Pantheon release-ready by itself. The next work should move from emergency safety into shippability: product vocabulary, root help, docs, Ma'at credibility, and enterprise-portable boundaries.

## Findings

### Resolved: purge/installer no longer permanently delete on no-trash platforms

Files:
- `internal/jackal/purge.go`
- `internal/jackal/installer.go`
- `internal/jackal/purge_test.go`
- `internal/jackal/installer_test.go`

The implementation now uses:

```go
if useTrash {
    freed, err = cleaner.DeleteFileReversible(path, false)
} else {
    freed, err = cleaner.DeleteFile(path, false, false)
}
```

The new tests simulate `platform.Mock{NoTrash: true}` and verify that files/directories remain present when reversible deletion is unavailable. This is the right behavior for a general-user hygiene tool.

### Resolved: router author validation cleanup

File:
- `internal/router/router.go`

`ValidateAuthor` now relies on the explicit whitelist instead of a confusing secondary character check. This is clearer and easier to audit.

### Resolved Enough: MCP notify handler tests exist

File:
- `internal/mcp/router_test.go`

Handler-level tests now cover default-disabled notify, missing args, invalid target, and submit missing args. I did not find a remaining security issue in the new handler path.

## Remaining Release Blockers

### [P1] Ma'at still fails its own release gate

Live command:

```sh
./sirsi audit
```

Result:

```text
Verdict: fail
Weight: 69/100
Passed: 26
Warnings: 12
Failures: 2
```

Notable weak areas from the live audit:
- `output`: 15.7% coverage
- `ra`: 12.4% coverage
- `jackal`: 49.9% coverage
- `mcp`: 42.9% coverage
- `workstream`: 49.9% coverage
- several modules still show no coverage data

The audit is now honest, which is good. But an Apache 2.0 release should not ship while the product's own quality gate says fail, especially because `output` owns the TUI user experience.

### [P1] Product vocabulary is still split between user outcomes and internal mythology

Observed:

```sh
./sirsi version
```

Output still says:

```text
"One Install. All Deities."
```

Observed:

```sh
./sirsi --help
```

Root help is improved, but still exposes `Module Access` as a primary section, duplicates `help`, and frames `sirsi help` as "rich guides for Pantheon deities." For normal users, this keeps forcing the internal module model back into the first-run experience.

This is the usability issue the user keeps pointing at: the modules can remain internally named after deities, but the product surface should lead with outcomes.

### [P2] Stale public docs still describe old command names and mental model

Examples:
- `README.md` still suggests `judge` in the TUI "What's Next" flow.
- `README.md` still positions `sirsi doctor` as a current top-level command instead of clearly preferring `sirsi diagnose`.
- `CLAUDE.md` still includes old `anubis`, `judge`, and `guard --slay` examples in rules that agents are supposed to obey.

Some legacy aliases may intentionally remain for compatibility. The issue is not their existence; it is old names appearing as user-facing guidance.

## Verification

Passed:

```sh
go build ./cmd/sirsi/
go vet ./...
go test ./internal/jackal ./internal/cleaner ./internal/router ./internal/mcp
go test -race ./internal/jackal ./internal/router
go test -run 'TestHandleRouter|TestValidateAuthor|TestPurgeArtifacts|TestRemoveInstallers' -timeout 30s ./internal/mcp ./internal/router ./internal/jackal
./sirsi version
./sirsi --help
./sirsi audit
```

Notes:
- `go build` emitted the existing macOS linker warning about duplicate `-lobjc`; build still passed.
- `./sirsi audit` completed in about 2m29s and failed honestly at 69/100.

## Recommended Claude Streams

Run these as separate simultaneous streams where safe:

### Stream A: First-Run Product Vocabulary

Own:
- `cmd/sirsi/main.go`
- version/help rendering files
- `README.md`

Goal:
- Root command and version output must sell outcomes, not the deity taxonomy.
- Keep module aliases available, but move them behind "advanced/module access" language.
- Remove duplicate `help` entry if possible.

Acceptance:
- `sirsi --help`, `sirsi version`, `sirsi quickstart --help`, and README first page read like an infrastructure hygiene tool for normal users and developers.

### Stream B: Stale Command Reference Cleanup

Own:
- `README.md`
- `CLAUDE.md`
- docs and help strings only

Goal:
- Replace old public guidance for `judge`, `weigh`, `doctor`, `ka`, and `anubis` with current outcome verbs where appropriate.
- Preserve compatibility docs in a clearly labeled legacy/advanced section if needed.

Acceptance:
- `rg -n "weigh|judge|ka|doctor|anubis"` in public docs returns only intentional internal/module/compatibility references.

### Stream C: Ma'at Release-Gate Credibility

Own:
- `internal/maat`
- `internal/output` tests
- `internal/ra` tests
- `internal/mcp` tests only if needed

Goal:
- Make `sirsi audit` actionable and release-credible.
- Do not game the score by lowering thresholds without a rationale.

Acceptance:
- Either `sirsi audit` passes the documented stable-release threshold or the release notes explicitly classify v0.21.x as pre-release/beta.

### Stream D: Enterprise Embedding Boundary

Own:
- architecture docs
- CLI/TUI boundary docs
- optional small interfaces only if already implied

Goal:
- Define what can be embedded into Sirsi Nexus later: scanning engine, cleanup planner, audit engine, router/MCP integration.
- Keep current CLI/TUI as adapters over reusable services instead of the center of gravity.

Acceptance:
- A short `docs/EMBEDDING.md` or ADR states stable internal APIs, non-goals, and what remains CLI-only.

## Final Call

The latest commit fixes the concrete cleanup safety defect. I would not block that patch.

I would still block an Apache 2.0 "stable" release until the user-facing language, docs, and Ma'at gate stop contradicting the product's intended promise.
