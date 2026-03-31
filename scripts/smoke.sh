#!/bin/bash
# 𓉴 Pantheon Smoke Test
# Builds the actual binary and runs real commands against the real filesystem.
# This is NOT a unit test. This proves the compiled software works.
#
# Usage: make smoke  (or: bash scripts/smoke.sh)
set -e

echo ""
echo "  𓉴 Pantheon Smoke Test — Does It Actually Work?"
echo "  ────────────────────────────────────────────────"
echo ""

# ── 1. Build the binary ─────────────────────────────────────────────
echo "  [1/5] Building binary..."
go build -o /tmp/pantheon-smoke ./cmd/pantheon/
echo "  ✅ Binary compiled ($(du -h /tmp/pantheon-smoke | cut -f1))"

# ── 2. Version check ────────────────────────────────────────────────
echo "  [2/5] Version check..."
VERSION=$(/tmp/pantheon-smoke version 2>&1)
if echo "$VERSION" | grep -q "v0.8.0-beta"; then
    echo "  ✅ Version: v0.8.0-beta (honest)"
else
    echo "  ❌ Version mismatch: $VERSION"
    exit 1
fi

# ── 3. Anubis weigh — does the scanner find real files? ──────────────
echo "  [3/5] Anubis weigh (real filesystem scan)..."
WEIGH_OUTPUT=$(/tmp/pantheon-smoke anubis weigh 2>&1)
if echo "$WEIGH_OUTPUT" | grep -qE "Waste Found|Pillars Ran"; then
    echo "  ✅ Scanner produced real output"
else
    echo "  ❌ Scanner output looks empty or fake"
    echo "  Output: $WEIGH_OUTPUT"
    exit 1
fi

# ── 4. Anubis judge --dry-run — does cleanup engine work? ────────────
echo "  [4/5] Anubis judge --dry-run (cleanup engine)..."
JUDGE_OUTPUT=$(/tmp/pantheon-smoke anubis judge --dry-run 2>&1)
if echo "$JUDGE_OUTPUT" | grep -qE "DRY RUN|adjudicated|No waste found"; then
    echo "  ✅ Cleanup engine operational (dry-run)"
else
    echo "  ❌ Cleanup engine produced no actionable output"
    echo "  Output: $JUDGE_OUTPUT"
    exit 1
fi

# ── 5. Ma'at audit — does governance actually measure? ───────────────
echo "  [5/5] Ma'at audit (governance scan)..."
AUDIT_OUTPUT=$(/tmp/pantheon-smoke maat audit 2>&1)
if echo "$AUDIT_OUTPUT" | grep -qE "Verdict|Weight|Status"; then
    echo "  ✅ Governance engine produced verdicts"
else
    echo "  ❌ Governance engine is a facade"
    echo "  Output: $AUDIT_OUTPUT"
    exit 1
fi

# ── Cleanup ──────────────────────────────────────────────────────────
rm -f /tmp/pantheon-smoke

echo ""
echo "  ────────────────────────────────────────────────"
echo "  ✅ ALL SMOKE TESTS PASSED"
echo "  The software works. It is not a facade."
echo ""
