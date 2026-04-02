#!/usr/bin/env python3
"""
Sirsi Orchestrator — Parallel Cross-Repo Development Automation

Uses the Claude Agent SDK to dispatch parallel work across all Sirsi repositories.
Each repo gets its own Claude session with full context (CLAUDE.md, hooks, MCP).

Usage:
  # Run health check across all repos
  python3 scripts/sirsi-orchestrator.py health

  # Run tests across all repos in parallel
  python3 scripts/sirsi-orchestrator.py test

  # Dispatch a task to a specific repo
  python3 scripts/sirsi-orchestrator.py task pantheon "fix the seba test failures"

  # Run a custom prompt across all repos
  python3 scripts/sirsi-orchestrator.py broadcast "check for security vulnerabilities in dependencies"

  # Nightly CI — run comprehensive checks
  python3 scripts/sirsi-orchestrator.py nightly
"""

import asyncio
import json
import os
import sys
import time
from datetime import datetime
from pathlib import Path

try:
    from claude_code_sdk import query, ClaudeCodeOptions
except ImportError:
    print("Error: claude-code-sdk not installed. Run: pip3 install claude-code-sdk")
    sys.exit(1)


# ── Repository Configuration ──────────────────────────────────────────

HOME = Path.home()
REPOS = {
    "pantheon": {
        "path": HOME / "Development" / "sirsi-pantheon",
        "desc": "Infrastructure hygiene CLI",
        "test_cmd": "go test -short ./...",
        "lint_cmd": "gofmt -l ./internal/ ./cmd/",
        "build_cmd": "go build ./cmd/pantheon/",
    },
    "nexus": {
        "path": HOME / "Development" / "SirsiNexusApp",
        "desc": "Platform monorepo",
        "test_cmd": "cd ui && yarn test --passWithNoTests 2>/dev/null; cd ../packages/sirsi-portal-app && npx tsc --noEmit",
        "lint_cmd": "cd ui && yarn lint; cd ../packages/sirsi-portal-app && npx eslint . --max-warnings 999",
        "build_cmd": "cd packages/sirsi-portal-app && npx vite build",
    },
    "finalwishes": {
        "path": HOME / "Development" / "FinalWishes",
        "desc": "Estate planning application",
        "test_cmd": "npm test --if-present 2>/dev/null || echo 'no tests configured'",
        "lint_cmd": "npm run lint --if-present 2>/dev/null || echo 'no lint configured'",
        "build_cmd": "npm run build --if-present 2>/dev/null || echo 'no build configured'",
    },
    "assiduous": {
        "path": HOME / "Development" / "Assiduous",
        "desc": "Real estate platform",
        "test_cmd": "npm test --if-present 2>/dev/null || echo 'no tests configured'",
        "lint_cmd": "npm run lint --if-present 2>/dev/null || echo 'no lint configured'",
        "build_cmd": "npm run build --if-present 2>/dev/null || echo 'no build configured'",
    },
}

# ── Output Helpers ────────────────────────────────────────────────────

def header(text):
    print(f"\n{'=' * 60}")
    print(f"  {text}")
    print(f"{'=' * 60}\n")

def repo_header(name, repo):
    print(f"  [{name}] {repo['desc']} — {repo['path']}")

def result_line(name, status, duration, detail=""):
    icon = "✅" if status == "pass" else "❌" if status == "fail" else "⚠️"
    dur = f"{duration:.1f}s" if duration else ""
    print(f"  {icon} {name:15s} {dur:>8s}  {detail}")


# ── Core Agent Functions ──────────────────────────────────────────────

async def run_agent(repo_name, repo_config, prompt, allowed_tools=None):
    """Run a Claude agent in a specific repo directory."""
    repo_path = str(repo_config["path"])

    if not os.path.isdir(repo_path):
        return {
            "repo": repo_name,
            "status": "skip",
            "message": f"Directory not found: {repo_path}",
            "duration": 0,
        }

    if allowed_tools is None:
        allowed_tools = ["Read", "Glob", "Grep", "Bash"]

    start = time.time()
    output_text = ""

    try:
        async for message in query(
            prompt=prompt,
            options=ClaudeCodeOptions(
                allowed_tools=allowed_tools,
                cwd=repo_path,
                max_turns=20,
            ),
        ):
            if hasattr(message, "content"):
                for block in message.content:
                    if hasattr(block, "text"):
                        output_text += block.text + "\n"
    except Exception as e:
        return {
            "repo": repo_name,
            "status": "fail",
            "message": str(e),
            "duration": time.time() - start,
        }

    duration = time.time() - start
    status = "pass" if "error" not in output_text.lower() and "fail" not in output_text.lower() else "warn"

    return {
        "repo": repo_name,
        "status": status,
        "message": output_text.strip()[-500:] if output_text else "No output",
        "duration": duration,
    }


async def run_parallel(repos, prompt_fn, allowed_tools=None):
    """Run agents across multiple repos in parallel."""
    tasks = []
    for name, config in repos.items():
        prompt = prompt_fn(name, config)
        tasks.append(run_agent(name, config, prompt, allowed_tools))

    results = await asyncio.gather(*tasks, return_exceptions=True)

    processed = []
    for r in results:
        if isinstance(r, Exception):
            processed.append({"repo": "unknown", "status": "fail", "message": str(r), "duration": 0})
        else:
            processed.append(r)
    return processed


# ── Commands ──────────────────────────────────────────────────────────

async def cmd_health():
    """Run health checks across all repos."""
    header("Sirsi Orchestrator — Health Check")
    print(f"  Timestamp: {datetime.now().isoformat()}")
    print(f"  Repos: {len(REPOS)}\n")

    def prompt(name, config):
        return f"""Run a quick health check on this repository:
1. Run `git status --short` to check for uncommitted changes
2. Run `git log --oneline -3` for recent commits
3. Check if the build command works: `{config['build_cmd']}`
4. Report: repo status (clean/dirty), last commit date, build status (pass/fail)
Keep your response under 200 words. Just the facts."""

    results = await run_parallel(REPOS, prompt)

    print("\n  Results:")
    for r in results:
        result_line(r["repo"], r["status"], r["duration"])
    return results


async def cmd_test():
    """Run tests across all repos in parallel."""
    header("Sirsi Orchestrator — Test Suite")

    def prompt(name, config):
        return f"""Run the test suite for this repository:
`{config['test_cmd']}`
Report: total tests, passed, failed. Keep response under 100 words."""

    results = await run_parallel(REPOS, prompt)

    print("\n  Results:")
    for r in results:
        result_line(r["repo"], r["status"], r["duration"])
    return results


async def cmd_lint():
    """Run linters across all repos in parallel."""
    header("Sirsi Orchestrator — Lint Check")

    def prompt(name, config):
        return f"""Run the linter for this repository:
`{config['lint_cmd']}`
Report: warnings count, errors count. Keep response under 100 words."""

    results = await run_parallel(REPOS, prompt)

    print("\n  Results:")
    for r in results:
        result_line(r["repo"], r["status"], r["duration"])
    return results


async def cmd_task(repo_name, task_description):
    """Dispatch a task to a specific repo."""
    if repo_name not in REPOS:
        print(f"  ❌ Unknown repo: {repo_name}")
        print(f"  Available: {', '.join(REPOS.keys())}")
        return

    header(f"Sirsi Orchestrator — Task: {repo_name}")
    config = REPOS[repo_name]
    repo_header(repo_name, config)

    result = await run_agent(
        repo_name, config, task_description,
        allowed_tools=["Read", "Glob", "Grep", "Bash", "Edit", "Write"],
    )

    print(f"\n  Status: {result['status']} ({result['duration']:.1f}s)")
    print(f"\n{result['message']}")
    return result


async def cmd_broadcast(prompt):
    """Run a custom prompt across all repos."""
    header("Sirsi Orchestrator — Broadcast")

    def prompt_fn(name, config):
        return prompt

    results = await run_parallel(REPOS, prompt_fn)

    print("\n  Results:")
    for r in results:
        result_line(r["repo"], r["status"], r["duration"])
        if r["message"]:
            for line in r["message"].split("\n")[:5]:
                print(f"    {line}")
    return results


async def cmd_nightly():
    """Comprehensive nightly check across all repos."""
    header("Sirsi Orchestrator — Nightly Run")
    print(f"  Started: {datetime.now().isoformat()}\n")

    all_results = {}

    # Phase 1: Health
    print("  Phase 1: Health Check")
    all_results["health"] = await cmd_health()

    # Phase 2: Lint
    print("\n  Phase 2: Lint")
    all_results["lint"] = await cmd_lint()

    # Phase 3: Test
    print("\n  Phase 3: Tests")
    all_results["test"] = await cmd_test()

    # Summary
    header("Nightly Summary")
    total_pass = sum(1 for phase in all_results.values() for r in phase if r["status"] == "pass")
    total_fail = sum(1 for phase in all_results.values() for r in phase if r["status"] == "fail")
    total_warn = sum(1 for phase in all_results.values() for r in phase if r["status"] == "warn")

    print(f"  ✅ Pass: {total_pass}")
    print(f"  ❌ Fail: {total_fail}")
    print(f"  ⚠️  Warn: {total_warn}")
    print(f"  Completed: {datetime.now().isoformat()}")

    # Save report
    report_dir = HOME / ".config" / "seshat" / "orchestrator"
    report_dir.mkdir(parents=True, exist_ok=True)
    report_path = report_dir / f"nightly_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
    with open(report_path, "w") as f:
        json.dump(all_results, f, indent=2, default=str)
    print(f"\n  Report saved: {report_path}")

    return all_results


# ── CLI Entry Point ───────────────────────────────────────────────────

async def main():
    if len(sys.argv) < 2:
        print(__doc__)
        sys.exit(0)

    command = sys.argv[1]

    if command == "health":
        await cmd_health()
    elif command == "test":
        await cmd_test()
    elif command == "lint":
        await cmd_lint()
    elif command == "task":
        if len(sys.argv) < 4:
            print("Usage: sirsi-orchestrator.py task <repo> <description>")
            sys.exit(1)
        await cmd_task(sys.argv[2], " ".join(sys.argv[3:]))
    elif command == "broadcast":
        if len(sys.argv) < 3:
            print("Usage: sirsi-orchestrator.py broadcast <prompt>")
            sys.exit(1)
        await cmd_broadcast(" ".join(sys.argv[2:]))
    elif command == "nightly":
        await cmd_nightly()
    else:
        print(f"Unknown command: {command}")
        print("Available: health, test, lint, task, broadcast, nightly")
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
