# Case Study 011 — The Extension Host Crash Chain

**Date**: March 26, 2026  
**Severity**: Critical — full IDE reinstall required  
**Root Cause**: Modifying `/Applications/Antigravity.app/` extension manifests  
**Recovery**: 2 reinstalls + 2 restarts to regain agent functionality  

---

## The Incident

During Session 22, we diagnosed 4 simultaneous extension issues in the Antigravity IDE (a VS Code fork). Two of the fixes required patching `package.json` files inside the application bundle:

1. **Git extension** — Two Antigravity-added commands (`git.antigravityCloneNonInteractive`, `git.antigravityGetRemoteUrl`) were missing the mandatory `title` property.
2. **Antigravity extension** — The `menus.commandPalette` referenced 3 commands (`importAntigravitySettings`, `importAntigravityExtensions`, `importAntigravityKeymap`) that were never declared in the `commands` array.

We patched both manifests, re-signed the bundle with `codesign --force --deep --sign -`, and restarted the IDE.

**The IDE never recovered.**

---

## The Crash Chain (from Crashpad forensics)

Three crash dumps were generated in 59 minutes:

### Crash 1 — 21:46 ET — Extension Host V8 OOM
- **Process**: `Antigravity Helper (Plugin)` — the Extension Host
- **Crash type**: `electron.v8-oom.is_heap_oom`
- **Evidence**: `Ineffective mark-compacts near heap limit`, `allocation failure; GC in old space requested`
- **V8 heap mu**: `average mu = 0.291, current mu = 0.132` (severely degraded GC efficiency)
- **What happened**: The Extension Host validates all extension manifests at startup. With broken command declarations (commands referenced in menus but with no handler), the validation/error-reporting path leaked memory through repeated error cycles until V8 ran out of heap.

### Crash 2 — 22:24 ET — Main Process Killed by macOS
- **Process**: `Electron` (main process, from `/Volumes/Antigravity/`)
- **Crash type**: `libMemoryResourceException.dylib` — macOS Jetsam termination
- **What happened**: After the Extension Host crashed, orphan child processes and leaked memory triggered macOS's kernel-level memory pressure response. The OS killed the entire Electron app.

### Crash 3 — 22:45 ET — Post-Reinstall, Same Kill
- **Process**: `Electron` (main process, now from `/Applications/Antigravity.app/`)
- **Crash type**: Same — `libMemoryResourceException.dylib`
- **What happened**: Fresh install, but the Crashpad `pending/` directory (34 unsubmitted dumps) persisted through the reinstall. The Extension Host started again with stale cached state. Second restart finally cleared the chain.

---

## Why the Patches Were Lethal

The Antigravity extension patch appeared harmless — we only added JSON properties. But the effect was catastrophic:

1. **Manifest semantics, not syntax, matter.** Adding `command` declarations without corresponding `activationEvents` or handler registration creates a state where the Extension Host finds commands it cannot bind. This isn't a graceful warning — it's a repeated validation failure.

2. **Error reporting leaks memory.** VS Code's Extension Host error reporting (particularly for manifest validation) allocates objects for telemetry, error stacks, and retry logic. When validation fails repeatedly for the same commands on every activation attempt, these allocations accumulate without collection.

3. **Code signing is not the only risk.** The original Rule A19 focused on code signing. The real risk is **semantic corruption** — the manifest is valid JSON, it passes schema validation, but it describes a state the Extension Host cannot realize. No codesign check catches this.

4. **Cascade kills are invisible.** The user sees "IDE won't launch." There's no error dialog, no log visible to the user, no indication that the Extension Host is the problem. The only evidence is in `~/Library/Application Support/Antigravity/Crashpad/pending/*.dmp` — which requires forensic extraction.

---

## The Evidence

```
# Crashpad pending directory after the incident
$ ls ~/Library/Application\ Support/Antigravity/Crashpad/pending/ | wc -l
34

# The trigger crash
$ strings <crash_1>.dmp | grep "VSCODE_CRASH_REPORTER_PROCESS_TYPE"
VSCODE_CRASH_REPORTER_PROCESS_TYPE=extensionHost

$ strings <crash_1>.dmp | grep "v8-oom"
electron.v8-oom.location
electron.v8-oom.is_heap_oom

$ strings <crash_1>.dmp | grep "allocation"
0.00 ms (average mu = 0.291, current mu = 0.132) allocation failure; GC in old space requested
```

---

## Lessons Learned

### For Developers Using VS Code Forks
1. **Never patch extension manifests inside `.app` bundles.** Even if the JSON is valid, the Extension Host interprets it semantically. Undeclared handlers for declared commands create a memory leak that crashes the process.
2. **Report upstream.** The missing `title` properties and undeclared commands are bugs in Antigravity's fork of the Git and core extensions. The only safe fix is an upstream patch from the Antigravity team.
3. **Monitor Crashpad.** VS Code/Electron crash dumps accumulate silently. Check `~/Library/Application Support/<IDE>/Crashpad/pending/` periodically. 34 pending dumps is a sign of chronic instability.

### For Pantheon
4. **Rule A19 is absolute.** No exceptions. No "manifest-only" carve-outs. The Session 22 exception was wrong.
5. **Guardian should monitor Crashpad.** The pending dump count is a leading indicator of IDE instability. Guardian could warn the user before the cascade begins.
6. **The fork question is legitimate.** When your IDE has bugs in its core extensions that you can't fix without breaking code signing, you've hit the limits of the vendor relationship.

---

## Rule A19 — Updated

**Before** (Session 22 exception):
> Modification is possible but requires re-signing.

**After** (Session 23 hardening):
> **ABSOLUTE PROHIBITION.** No file inside `/Applications/*.app/` may be modified under any circumstances. Manifest patches cause semantic corruption that crashes the Extension Host via V8 OOM. Even valid JSON modifications create states the host cannot realize. There are no safe exceptions.

---

## Verification Commands

```bash
# Check for pending crash dumps
ls ~/Library/Application\ Support/Antigravity/Crashpad/pending/*.dmp 2>/dev/null | wc -l

# Extract process type from a dump
strings <dump_file> | grep "VSCODE_CRASH_REPORTER_PROCESS_TYPE"

# Check for V8 OOM evidence
strings <dump_file> | grep "electron.v8-oom"

# Verify code signature integrity
codesign --verify --deep --strict /Applications/Antigravity.app 2>&1
```
