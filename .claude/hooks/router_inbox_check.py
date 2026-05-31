#!/usr/bin/env python3
"""Print a concise Claude Code router inbox reminder when work is pending."""

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
    """Count open items for agent_id under items/ (new pull-model queue).

    Each item is a markdown file with YAML frontmatter. We do a minimal scan
    rather than pulling in a YAML library — only `to:` and `status:` matter.
    """
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
            value = value.strip()
            if key == "to":
                to_val = value
            elif key == "status":
                status_val = value
        if status_val == "open" and to_val == agent_id:
            matches.append(path.stem)
    return matches


def claude_session_pid() -> int | None:
    """Walk up the process tree to find the Claude Code CLI process.

    Hook is invoked as: claude → shell → python3 (this script).
    So our grandparent should be the claude binary.
    """
    try:
        shell_pid = os.getppid()
        out = subprocess.run(
            ["ps", "-p", str(shell_pid), "-o", "ppid="],
            capture_output=True, text=True, timeout=2,
        )
        return int(out.stdout.strip()) if out.returncode == 0 else None
    except (subprocess.TimeoutExpired, ValueError):
        return None


def ensure_router_watcher(thread_id: str, agent_id: str, anchor: int | None) -> None:
    """Idempotent watcher spawn: if no live watcher process exists for
    this thread_id, spawn one anchored to the given PID. Pidfile is
    /tmp/sirsi-router-watch-<thread_id>.pid.

    Solves the adopt-without-watcher gap: when ensure_active_thread()
    adopts an existing fresh thread, the original watcher may have died
    with its original anchor. Re-spawning under our own anchor here keeps
    FSEvents wake alive for the new session.
    """
    if not thread_id or not anchor:
        return
    pidfile = f"/tmp/sirsi-router-watch-{thread_id}.pid"
    try:
        with open(pidfile) as f:
            old_pid = int(f.read().strip())
        os.kill(old_pid, 0)  # raises if dead
        return  # existing watcher alive
    except (FileNotFoundError, ValueError, ProcessLookupError, PermissionError):
        pass

    # Bounce the thread to spawn a watcher: re-register with --thread so
    # the existing thread_id is reused (idempotent) and a fresh watcher
    # gets spawned with our anchor.
    try:
        subprocess.run(
            ["sirsi", "thread", "register",
             "--thread", thread_id,
             "--agent", agent_id,
             "--surface", "claude",
             "--anchor-pid", str(anchor)],
            capture_output=True, timeout=3,
        )
    except (FileNotFoundError, subprocess.TimeoutExpired):
        pass


def ensure_active_thread(agent_id: str, repo_path: Path) -> str | None:
    """Return an active thread_id for agent_id, registering one if none exists.

    Solves the orphan problem: if a session opens and CTR has no active
    thread for our agent (or only stale ones), register fresh instead of
    silently heartbeating someone else's thread.
    """
    try:
        out = subprocess.run(
            ["sirsi", "thread", "list", "--json"],
            capture_output=True, text=True, timeout=2,
        )
        threads = json.loads(out.stdout) if out.returncode == 0 and out.stdout.strip() else []
    except (FileNotFoundError, subprocess.TimeoutExpired, json.JSONDecodeError):
        threads = []

    anchor = claude_session_pid()

    # Active and fresh (< 5 minutes idle) — adopt it AND ensure a live watcher.
    fresh = [
        t for t in threads
        if (t.get("thread") or {}).get("agent_id") == agent_id
        and (t.get("thread") or {}).get("status") == "active"
        and t.get("idle_seconds", 1e9) < 300
    ]
    if fresh:
        fresh.sort(key=lambda t: t.get("idle_seconds", 1e9))
        thread_id = (fresh[0].get("thread") or {}).get("thread_id")
        ensure_router_watcher(thread_id, agent_id, anchor)
        return thread_id

    # No fresh thread — register a new one. `register` itself spawns the
    # watcher; we just supply the right anchor PID.
    args = ["sirsi", "thread", "register",
            "--agent", agent_id, "--surface", "claude", "--repo", str(repo_path)]
    if anchor:
        args += ["--anchor-pid", str(anchor)]
    try:
        out = subprocess.run(args, capture_output=True, text=True, timeout=3)
        # Output contains "thr-XXXXXXXX" — extract it
        for line in (out.stdout + out.stderr).splitlines():
            for tok in line.split():
                if tok.startswith("thr-") and len(tok) > 10:
                    return tok.strip(",.;:")
    except (FileNotFoundError, subprocess.TimeoutExpired):
        pass
    return None


def caffeinate_thread(thread_id: str, claude_pid: int | None) -> None:
    """Spawn a backgrounded heartbeat loop that keeps thread_id fresh until
    the parent claude process exits. Dedup via pidfile so we don't stack
    multiple loops per session.

    This is the 'caffeinate' primitive: thread stays alive while the
    process is alive, no prompt-dependence. Universal pattern any agent
    can adopt — just a shell loop anchored to the parent PID.
    """
    if not thread_id or not claude_pid:
        return
    pidfile = Path(f"/tmp/sirsi-caffeinate-{thread_id}.pid")
    if pidfile.exists():
        try:
            old_pid = int(pidfile.read_text().strip())
            os.kill(old_pid, 0)  # check liveness; raises if dead
            return  # caffeinator already running for this thread
        except (ValueError, ProcessLookupError, PermissionError):
            pass  # stale, replace it

    script = (
        f"while kill -0 {claude_pid} 2>/dev/null; do "
        f"sirsi thread heartbeat --thread {thread_id} --quiet >/dev/null 2>&1; "
        f"sleep 60; "
        f"done; rm -f {pidfile}"
    )
    try:
        p = subprocess.Popen(
            ["bash", "-c", script],
            stdin=subprocess.DEVNULL, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
            start_new_session=True,  # detach from python so it survives our exit
        )
        pidfile.write_text(str(p.pid))
    except OSError:
        pass


def heartbeat_active_thread(agent_id: str, repo_path: Path) -> None:
    """Ensure agent_id has an active thread + immediate heartbeat + caffeinator.

    Order:
      1. Adopt fresh thread or register a new one (no more silent orphan loss)
      2. Immediate heartbeat so this session shows idle=0 right away
      3. Spawn caffeinator (if not already running) for sustained liveness
    """
    thread_id = ensure_active_thread(agent_id, repo_path)
    if not thread_id:
        return
    try:
        subprocess.run(
            ["sirsi", "thread", "heartbeat", "--thread", thread_id, "--quiet"],
            capture_output=True, timeout=2,
        )
    except (FileNotFoundError, subprocess.TimeoutExpired):
        return
    caffeinate_thread(thread_id, claude_session_pid())


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

    # Keep this session's CTR thread caffeinated. Adopts a fresh thread or
    # registers a new one, immediate heartbeat, then spawns a background
    # loop that heartbeats every 60s until the claude process exits.
    heartbeat_active_thread(agent_id, repo_root)

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
    breakdown = " + ".join(parts)
    print(f"{prefix}:{agent_id} has {total} pending inbox {noun} ({breakdown})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
