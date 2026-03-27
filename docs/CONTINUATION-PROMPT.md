# 𓂀 Pantheon — Continuation Prompt
# Read this FIRST in a new session. Then read `.thoth/memory.yaml`.
# Last updated: 2026-03-26T23:20:00-04:00

---

## Session 24 — Priorities

### P0: Sideload Extension + Verify Crashpad Monitor
1. Build and sideload the v0.7.0 VSIX:
   ```bash
   cd extensions/vscode && npm run package
   antigravity --install-extension sirsi-pantheon-0.7.0.vsix
   ```
2. Open Command Palette → **"Pantheon: Crashpad Stability Report"** → verify webview opens.
3. Check status bar — if 34 pending dumps are still there, should show 🔴 critical status.
4. Run **"Pantheon: Show System Metrics"** → verify Crashpad option appears in quick picker.
5. Run **"Pantheon: Thoth Accountability Report"** → verify webview still works.

### P1: Clear the Crashpad + Establish Baseline
- Use the Crashpad Monitor's "Clear Pending Dumps" to reset the 34 stale dumps.
- Monitor over next few sessions — new dumps = chronic issue worth investigating.
- If Extension Host crashes reappear → disable extensions one by one (AG Monitor is already disabled).

### P2: OpenVSX Publish v0.7.0
- Publish updated VSIX to OpenVSX (open-vsx.org).
- Requires SirsiMaster Chrome profile (Rule A20).
- After publish: install from marketplace instead of sideloading.

### P3: Deploy Updated Site
- Deploy updated `build-log.html` and `case-studies/` to Firebase Hosting.
- Update Sekhmet deity page with Crashpad Monitor feature.
- Deploy deity registry index with updated stats.

---

## Context Pointers
- **Crashpad Monitor**: `extensions/vscode/src/crashpadMonitor.ts` (370+ lines)
- **Thoth Accountability Engine**: `extensions/vscode/src/thothAccountability.ts` (645 lines)
- **Extension entry point**: `extensions/vscode/src/extension.ts`
- **Commands**: `extensions/vscode/src/commands.ts` (10 commands registered)
- **Package manifest**: `extensions/vscode/package.json` (v0.7.0)
- **Case Study 011**: `docs/case-studies/session-23-extension-host-crash-forensics.md`
- **Case Study 012**: `docs/case-studies/session-23-crashpad-monitor.md`
- **Journal**: `.thoth/journal.md` (Entry 020-021)
- **Rule A19**: ABSOLUTE PROHIBITION — `PANTHEON_RULES.md` §2.16

## Extension Sideload Location
- Antigravity: `~/Desktop/.antigravity/extensions/sirsimaster.sirsi-pantheon-0.7.0/`
- Disabled: `~/Desktop/.antigravity/extensions/shivangtanwar.ag-monitor-pro-1.0.0.disabled/`
- **DO NOT PATCH**: Git extension or Antigravity extension in `/Applications/Antigravity.app/`

## Session 23 Stats
- **Files created**: 3 (crashpadMonitor.ts, case-study-011, case-study-012)
- **Files modified**: 10+ (extension.ts, commands.ts, package.json, RULES, CLAUDE, GEMINI, journal, memory, changelog, build-log.html, VERSION)
- **Commits**: 3 (`59d6d12` forensics, `bfb5463` crashpad monitor, canonization)
- **Version**: 0.6.0-alpha → **0.7.0-alpha**
- **Extension commands**: 8 → **10**
- **Extension modules**: 6 → **7**
