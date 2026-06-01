#!/usr/bin/env python3
"""Router supervisor hook (ADR-024).

SessionStart/prompt: register the thread (handshake), heartbeat, and — keyed on
OS truth — re-assert the ONE prescribed watcher by injecting the router's
`arm_instruction` into context when no live watcher exists for this thread.
Stop: optional backstop that blocks until the watcher is detected alive.

ADR-024 retires the per-claude caffeinator and the auto fs-watcher: the `/loop`
Monitor is the single liveness/wake mechanism for a claude surface. `register`
no longer spawns a watcher; it RETURNS the spec, and this hook hands the
agent the spec's `arm_instruction`.
"""

from __future__ import annotations

import json
import os
import subprocess
import sys
from pathlib import Path


def find_repo_root(start: Path) -> Path | None:
    for candidate in [start, *start.parents]:
        if (candidate / ".agents" / "idea-router" / "state.json").is_file():
            return candidate
    return None


def load_json(path: Path) -> dict:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def claude_agent_id(repo_root: Path, agents: dict) -> str:
    repo_root_resolved = repo_root.resolve()
    matches: list[str] = []
    for agent_id, config in agents.get("agents", {}).items():
        if config.get("type") != "claude":
            continue
        cwd = config.get("cwd")
        if cwd and Path(cwd).expanduser().resolve() == repo_root_resolved:
            matches.append(agent_id)
    return sorted(matches)[0] if matches else "claude-pantheon"


def pending_items(state: dict, agent_id: str) -> list[str]:
    pending = state.get("pending") or {}
    items = list(pending.get(agent_id) or [])
    if agent_id == "claude-pantheon":
        for item in state.get("pending_for_claude") or []:
            if item not in items:
                items.append(item)
    return items


def pull_model_open_items(router_root: Path, agent_id: str) -> list[str]:
    """Count open items addressed to agent_id under items/ — the one inbox
    (ADR-024 §5). reviews/ and decisions/ are NOT polled."""
    items_dir = router_root / "items"
    if not items_dir.is_dir():
        return []
    matches: list[str] = []
    for path in items_dir.glob("*.md"):
        try:
            text = path.read_text(encoding="utf-8")
        except OSError:
            continue
        if not text.startswith("---\n"):
            continue
        end = text.find("\n---\n", 4)
        if end < 0:
            continue
        frontmatter = text[4:end]
        to_val = status_val = ""
        for line in frontmatter.splitlines():
            key, sep, value = line.partition(":")
            if not sep:
                continue
            key = key.strip()
            value = value.strip().strip('"')
            if key == "to":
                to_val = value
            elif key == "status":
                status_val = value
        if status_val == "open" and to_val == agent_id:
            matches.append(path.stem)
    return matches


def claude_session_pid() -> int | None:
    """Grandparent of this script (claude → shell → python3) is the CLI process."""
    try:
        shell_pid = os.getppid()
        out = subprocess.run(
            ["ps", "-p", str(shell_pid), "-o", "ppid="],
            capture_output=True, text=True, timeout=2,
        )
        return int(out.stdout.strip()) if out.returncode == 0 else None
    except (subprocess.TimeoutExpired, ValueError):
        return None


def supervisor_mode(env: dict | None = None) -> str:
    """Resolve the supervisor mode (ADR-024 §4).

      "off"     — SIRSI_SUPERVISOR=0: suppress managed arming + Stop-gate.
                  The register spec stays visible (handled by `register` itself).
      "enforce" — SIRSI_SUPERVISOR=enforce: arming injection + the Stop-gate
                  backstop (off by default).
      "on"      — default: arming injection on, Stop-gate off.
    """
    env = env if env is not None else os.environ
    v = (env.get("SIRSI_SUPERVISOR") or "").strip().lower()
    if v == "0":
        return "off"
    if v == "enforce":
        return "enforce"
    return "on"


def watcher_armed(thread_id: str, runner=subprocess.run) -> bool:
    """OS-truth liveness check (ADR-024 §3, F2): is a watcher process alive for
    THIS thread? Keys on `pgrep -f thr-<thread_id>` — the same (agent_id, pid)
    identity ADR-022 reaps on — NEVER the harness TaskList (it falsely reports
    empty) and NEVER the shared `DIR=` loop body (it matches OTHER agents'
    loops on a shared host, a false 'already armed')."""
    if not thread_id:
        return False
    try:
        out = runner(["pgrep", "-f", thread_id], capture_output=True, text=True, timeout=2)
    except (FileNotFoundError, subprocess.TimeoutExpired):
        return False
    return out.returncode == 0 and bool((out.stdout or "").strip())


def should_arm(thread_id: str, mode: str, runner=subprocess.run) -> bool:
    """Check-then-arm decision (ADR-024 §3). Arm iff the supervisor is not off,
    we have a thread, and NO live watcher exists for it (OS truth). This is the
    single gate behind F1 (re-assert when gone) and F2 (never duplicate when an
    OS watcher is alive — keyed on pgrep, never TaskList)."""
    if mode == "off" or not thread_id:
        return False
    return not watcher_armed(thread_id, runner=runner)


def register_handshake(agent_id: str, repo_path: Path, thread_id: str | None,
                       anchor: int | None, runner=subprocess.run) -> tuple[str | None, str]:
    """Run `thread register --json` (the ADR-024 handshake) and return
    (thread_id, arm_instruction). register no longer spawns a watcher; it
    returns the spec the surface must arm."""
    args = ["sirsi", "thread", "register", "--agent", agent_id,
            "--surface", "claude", "--repo", str(repo_path), "--json"]
    if thread_id:
        args += ["--thread", thread_id]
    if anchor:
        args += ["--anchor-pid", str(anchor)]
    try:
        out = runner(args, capture_output=True, text=True, timeout=3)
        data = json.loads(out.stdout) if out.returncode == 0 and out.stdout.strip() else {}
    except (FileNotFoundError, subprocess.TimeoutExpired, json.JSONDecodeError):
        return thread_id, ""
    tid = data.get("thread_id") or thread_id
    arm = (data.get("watcher") or {}).get("arm_instruction", "")
    return tid, arm


def adopt_or_register(agent_id: str, repo_path: Path, runner=subprocess.run) -> tuple[str | None, str]:
    """Adopt a fresh active thread for agent_id if one exists, else register a
    new one. Returns (thread_id, arm_instruction). No watcher is spawned here —
    register is a pure handshake (ADR-024 Decision 2)."""
    try:
        out = runner(["sirsi", "thread", "list", "--json"], capture_output=True, text=True, timeout=2)
        threads = json.loads(out.stdout) if out.returncode == 0 and out.stdout.strip() else []
    except (FileNotFoundError, subprocess.TimeoutExpired, json.JSONDecodeError):
        threads = []

    anchor = claude_session_pid()
    fresh = [
        t for t in threads
        if (t.get("thread") or {}).get("agent_id") == agent_id
        and (t.get("thread") or {}).get("status") == "active"
        and t.get("idle_seconds", 1e9) < 300
    ]
    existing = None
    if fresh:
        fresh.sort(key=lambda t: t.get("idle_seconds", 1e9))
        existing = (fresh[0].get("thread") or {}).get("thread_id")
    # Idempotent on (agent_id, pid): reuses `existing` if passed, else mints one.
    return register_handshake(agent_id, repo_path, existing, anchor, runner=runner)


def heartbeat(thread_id: str, runner=subprocess.run) -> None:
    if not thread_id:
        return
    try:
        runner(["sirsi", "thread", "heartbeat", "--thread", thread_id, "--quiet"],
               capture_output=True, timeout=2)
    except (FileNotFoundError, subprocess.TimeoutExpired):
        pass


def main() -> int:
    mode = sys.argv[1] if len(sys.argv) > 1 else "session"
    repo_override = os.environ.get("SIRSI_ROUTER_REPO_ROOT")
    repo_root = Path(repo_override).expanduser().resolve() if repo_override else find_repo_root(Path.cwd())
    if repo_root is None:
        return 0

    router_root = repo_root / ".agents" / "idea-router"
    try:
        state = load_json(router_root / "state.json")
        agents = load_json(router_root / "agents.json")
    except (OSError, json.JSONDecodeError):
        return 0

    agent_id = claude_agent_id(repo_root, agents)
    sup = supervisor_mode()

    # Register handshake + heartbeat (no caffeinator, no fs-watcher — ADR-024).
    thread_id, arm_instruction = adopt_or_register(agent_id, repo_root)
    heartbeat(thread_id)

    # Stop-gate backstop (off by default; only under SIRSI_SUPERVISOR=enforce).
    if mode == "stop":
        if sup == "enforce" and thread_id and not watcher_armed(thread_id):
            print(f"router-supervisor: thread {thread_id} has no live /loop watcher — "
                  f"arm it before stopping (SIRSI_SUPERVISOR=enforce).", file=sys.stderr)
            return 2
        return 0

    # Check-then-arm (F1): re-assert the ONE watcher every SessionStart/wakeup.
    # Keyed on OS truth (pgrep), so we never duplicate (F2) and never collide
    # with other agents' loops. Suppressed only when SIRSI_SUPERVISOR=0.
    if arm_instruction and should_arm(thread_id, sup):
        print(f"router-supervisor:{agent_id} arm your watcher (thread {thread_id}): {arm_instruction}")

    legacy = pending_items(state, agent_id)
    pull = pull_model_open_items(router_root, agent_id)
    total = len(legacy) + len(pull)
    if total == 0:
        return 0

    prefix = "router-inbox" if mode == "prompt" else "router"
    noun = "item" if total == 1 else "items"
    parts = []
    if legacy:
        parts.append(f"{len(legacy)} legacy")
    if pull:
        parts.append(f"{len(pull)} pull-model")
    print(f"{prefix}:{agent_id} has {total} pending inbox {noun} ({' + '.join(parts)})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
