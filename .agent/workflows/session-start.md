---
description: How to start a new session on the Sirsi Anubis codebase efficiently
---

# Starting a Session on Sirsi Anubis

## Step 1: Read the memory file (ALWAYS do this first)
Read `.anubis-memory.yaml` in the project root. This is a ~100-line structured summary of:
- Current version, test count, module count
- Architecture quick reference (which module does what)
- Critical design decisions
- Known limitations
- Recent audit findings
- File map (most important files)

This saves you from re-reading dozens of source files.

## Step 2: Read the engineering journal (if context matters)
Read `.anubis-journal.md` for the WHY behind decisions. Each entry is timestamped with:
- What happened
- What the insight was
- What decision was made and why
- What the result was

## Step 3: Check current state
```bash
// turbo
cat VERSION
```

```bash
// turbo
go build ./cmd/anubis/ && go test -race -count=1 ./... 2>&1 | grep -E '^(ok|FAIL)' && echo "✓ Tests passing"
```

```bash
// turbo
~/go/bin/golangci-lint run --timeout=5m 2>&1 && echo "✓ Lint clean"
```

## Step 4: Update memory after every major change
After making significant changes, update `.anubis-memory.yaml` with:
- New version, test count, line count
- New design decisions
- New limitations or fixed issues
- Updated file map if new files created

Also add a journal entry to `.anubis-journal.md`.

## Step 5: Commit the memory files
Memory and journal files are tracked in git. Commit them with your code changes.
