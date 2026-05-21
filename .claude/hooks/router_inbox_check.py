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
