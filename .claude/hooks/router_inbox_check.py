#!/usr/bin/env python3
"""Print a concise Claude Code router inbox reminder when work is pending."""

from __future__ import annotations

import json
import os
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
    items = pending_items(state, agent_id)
    if not items:
        return 0

    prefix = "router-inbox" if mode == "prompt" else "router"
    noun = "item" if len(items) == 1 else "items"
    print(f"{prefix}:{agent_id} has {len(items)} pending inbox {noun}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
