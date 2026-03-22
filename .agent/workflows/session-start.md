---
description: How to start a new session on a Sirsi project using the Thoth knowledge system
---

# 𓁟 Thoth Session Start

// turbo-all

## Step 1: Read the Thoth memory file (ALWAYS do this first)
Read `.thoth/memory.yaml` in the project root. This is the compressed project state:
- Identity, version, stats
- Architecture quick reference
- Critical design decisions
- Known limitations and recent changes
- File map of important paths

This replaces reading thousands of lines of source code.

## Step 2: Read the engineering journal (when reasoning matters)
Read `.thoth/journal.md` for timestamped decision entries:
- What was happening (context)
- What we discovered (insight)
- What we chose and why (decision)
- What happened (result)

## Step 3: Check current state
```bash
cat VERSION
```

```bash
go build ./cmd/anubis/ && go test -race -count=1 ./... 2>&1 | grep -E '^(ok|FAIL)' && echo "✓ Tests passing"
```

```bash
~/go/bin/golangci-lint run --timeout=5m 2>&1 && echo "✓ Lint clean"
```

## Step 4: After making significant changes — update Thoth
1. Update `.thoth/memory.yaml` with new version, stats, decisions, limitations
2. Add a journal entry to `.thoth/journal.md` with context, insight, decision
3. Commit Thoth files with your code changes

## Step 5: For deep dives — check artifacts
Look in `.thoth/artifacts/` for benchmark results, platform audits, security reviews, and design documents relevant to your current task.

## Step 6: Before every release — update build-in-public artifacts (ADR-003)
This is mandatory per [ADR-003](docs/ADR-003-BUILD-IN-PUBLIC.md). Every release must update:
1. `VERSION` — bump the version
2. `CHANGELOG.md` — technical changes, migration notes, breaking changes
3. `docs/BUILD_LOG.md` — sprint narrative with real data (what broke, benchmarks, honest gaps)
4. `docs/build-log.html` — update the public HTML page stats and timeline
5. `.thoth/memory.yaml` and `.thoth/journal.md` — already covered in Step 4
6. New `docs/ADR-XXX-*.md` if the change involves a structural decision

**Voice rules**: Use direct verbs (built, fixed, refactored). Never "the user wanted." Include mistakes, real benchmarks, and honest disclaimers about what isn't ready.

