# Pantheon Hardening Roadmap v3.1.0

## 🎯 Target: 99.0% Weighted Coverage
**Status**: 90.1% (Session 16b Log)  
**Objective**: Enforce **Rule A16 (Interface Injection)** project-wide to eliminate all "Untouchable" paths.

---

## 🛤 Phase 1: Deep Interface Injection (Current)
- [x] **Platform Interface Expansion**: Added `MoveToTrash`, `PickFolder`, `OpenBrowser`, `ReadDir`.
- [x] **Mirror Module**: Refactored `Server` to use injected `Platform`. (Coverage: 65.9%)
- [x] **Sight Module**: Stabilized `parseLSRegisterDump` and `Fix` logic. (Coverage: 91.7%)
- [ ] **Scarab Module**: Refactor `internal/scarab/containers.go` to use `Platform.Command`.
- [ ] **Guard Module**: Refactor `internal/guard/audit.go` to use `Platform.Command` and `Platform.Processes`.
- [ ] **Standardization**: Replace all remaining `os.UserHomeDir` and `os.Getwd` calls with `Platform` methods.

## 🛤 Phase 2: Cross-Platform "Truth" Testing
- [ ] **Windows Support**: 
    - Implement `internal/platform/windows.go` (Tasklist, PowerShell RecycleBin).
    - Update `internal/platform/detect.go` to recognize `windows`.
- [ ] **Linux Verification (Scarab Bridge)**:
    - Use `Scarab` to spin up Linux containers for automated "Truth" verification of `ReadDir` and `Trash`.
- [ ] **CI Pipeline Hardening**:
    - [ ] Add `windows-latest` runner to `.github/workflows/ci.yml`.
    - [ ] Add `macos-14` (Apple Silicon) runner for M1/M2 specific detection tests.

## 🛤 Phase 3: Rule A17 (Ma'at QA Sovereign)
- [ ] Run full Ma'at assessment on all 22 modules.
- [ ] Achieve **Feather Weight 90+** across the entire ecosystem.
- [ ] **Rule A18 Integration**: Sign all decision logs with SHA-256 + Timestamp (ADR-009).

---

## 📋 Verification Checklist
- [ ] `go test -race ./...` passes on macOS, Linux, and Windows.
- [ ] Zero instances of `runtime.GOOS` outside of `internal/platform`.
- [ ] Zero instances of `os` or `exec` side-effects outside of `internal/platform`.
