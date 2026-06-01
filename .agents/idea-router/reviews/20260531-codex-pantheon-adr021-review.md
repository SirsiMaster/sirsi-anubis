# Codex Review — ADR-021 Deities Must Not Assume Single-Repo

## Decision

Codex concurs with the canon principle and the Osiris direction, with one required amendment before implementation: for menubar/dashboard use, CTR registry must be the primary high-signal source, but configured development roots cannot remain merely optional if CTR is empty or cold after reboot.

## Answers

1. **Concur with the principle as canon?**

   Yes. Workstation-resident deities must not source scope from process cwd. `RepoDir: "."` is a CLI-era assumption and is wrong for LaunchAgent, menubar, daemon, and dashboard contexts.

2. **Concur with CTR registry / `sirsi thread discover` over filesystem `.git` walk?**

   Yes as the primary source. It is much safer and more LEAN than walking `$HOME` or `~/Development`. However, CTR alone can be cold immediately after reboot or miss dirty repos with no active agent thread. For Osiris risk, that blind spot matters. The ADR should make the fallback explicit:

   - Primary: CTR thread registry + `sirsi thread discover`.
   - Secondary: user-configured dev roots/repo roots, bounded and cached.
   - No unbounded filesystem walk.
   - Zero known repos degrades to benign `n/a`, not failed.

3. **Objection to Osiris aggregating across all active-thread repos for menubar/dashboard?**

   No objection. This is the right model for Horus-as-workstation-lord. Menubar should show a rollup and worst offender, not a false single-repo status.

## Required Implementation Guardrails

- Do not pin Osiris to `sirsi-pantheon` or any single repo.
- Do not scan `$HOME` or recursively hunt for `.git` by default.
- Cache/cap repo status collection so menubar refresh cannot become expensive.
- Treat non-git and zero-repo states as benign `n/a`, not `failed`.
- Add tests for launchd-style cwd `/`, empty CTR, CTR with multiple repos, configured-root fallback, and non-git dirs.

## Status

ADR-021 is approved as a canon direction once the fallback wording is tightened. Code implementation is still gated by the user's approval.
