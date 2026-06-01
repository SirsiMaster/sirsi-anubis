#!/usr/bin/env python3
"""ADR-024 acceptance tests 4/5/6 for the router supervisor hook.

Run: python3 .claude/hooks/router_inbox_check_test.py
"""

import importlib.util
import types
import unittest
from pathlib import Path

# Load the hyphenless-but-underscored hook module by path.
_spec = importlib.util.spec_from_file_location(
    "router_inbox_check", str(Path(__file__).with_name("router_inbox_check.py"))
)
hook = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(hook)


def fake_runner(returncode: int, stdout: str = ""):
    """Return a subprocess.run stand-in that always yields this result."""
    def _run(*_args, **_kwargs):
        return types.SimpleNamespace(returncode=returncode, stdout=stdout)
    return _run


class TestSupervisorMode(unittest.TestCase):
    def test_off(self):
        self.assertEqual(hook.supervisor_mode({"SIRSI_SUPERVISOR": "0"}), "off")

    def test_enforce(self):
        self.assertEqual(hook.supervisor_mode({"SIRSI_SUPERVISOR": "enforce"}), "enforce")

    def test_default_on(self):
        self.assertEqual(hook.supervisor_mode({}), "on")


class TestWatcherArmedKeysOnPgrep(unittest.TestCase):
    def test_alive(self):
        # pgrep found a matching process (rc 0, a pid on stdout) => armed.
        self.assertTrue(hook.watcher_armed("thr-abc", runner=fake_runner(0, "54321\n")))

    def test_gone(self):
        # pgrep found nothing (rc 1, empty) => not armed.
        self.assertFalse(hook.watcher_armed("thr-abc", runner=fake_runner(1, "")))


class TestShouldArm(unittest.TestCase):
    # Acceptance test 5 (F1): wakeup with the watcher GONE => re-assert exactly one.
    def test_f1_rearm_when_gone(self):
        self.assertTrue(hook.should_arm("thr-abc", "on", runner=fake_runner(1, "")))

    # Acceptance test 6 (F2): an OS watcher is ALIVE (pgrep hit) => do NOT arm a
    # duplicate. The decision consults pgrep only — it never sees a TaskList, so
    # a falsely-empty TaskList cannot cause a duplicate.
    def test_f2_no_duplicate_when_alive(self):
        self.assertFalse(hook.should_arm("thr-abc", "on", runner=fake_runner(0, "54321\n")))

    # Acceptance test 4: SIRSI_SUPERVISOR=0 suppresses managed arming entirely,
    # even when no watcher exists. (The spec stays visible — that's `register`'s
    # job, not this hook's.)
    def test_supervisor_off_suppresses_arming(self):
        self.assertFalse(hook.should_arm("thr-abc", "off", runner=fake_runner(1, "")))

    def test_no_thread_no_arm(self):
        self.assertFalse(hook.should_arm("", "on", runner=fake_runner(1, "")))


if __name__ == "__main__":
    unittest.main(verbosity=2)
