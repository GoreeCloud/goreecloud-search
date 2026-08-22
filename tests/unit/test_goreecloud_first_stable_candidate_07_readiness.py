# SPDX-License-Identifier: AGPL-3.0-or-later
"""Regression tests for candidate #07 first-Stable readiness reporting."""

from __future__ import annotations

import argparse
import importlib.util
import json
from pathlib import Path
import sys
import tempfile
import unittest


MODULE_PATH = Path(__file__).parents[2] / "goreecloud" / "first_stable_candidate_07_readiness.py"
MODULE_DIR = str(MODULE_PATH.parent)
SPEC = importlib.util.spec_from_file_location("first_stable_candidate_07_readiness", MODULE_PATH)
assert SPEC is not None and SPEC.loader is not None
READINESS = importlib.util.module_from_spec(SPEC)
sys.path.insert(0, MODULE_DIR)
sys.modules[SPEC.name] = READINESS
try:
    SPEC.loader.exec_module(READINESS)
finally:
    sys.path.remove(MODULE_DIR)
    sys.modules.pop(SPEC.name, None)


def empty_args(**overrides: str | None) -> argparse.Namespace:
    """Build a readiness namespace with no evidence supplied by default."""
    values = {
        "release_evidence": None,
        "target_runtime_evidence": None,
        "recovery_evidence": None,
        "rollback_baseline": None,
        "provider_evidence": None,
        "visual_evidence": None,
        "browser_evidence": None,
        "final_evidence": None,
        "json_output": False,
    }
    values.update(overrides)
    return argparse.Namespace(**values)


def valid_artifacts(final_status: str) -> dict[str, dict[str, str]]:
    """Return synthetic status-only entries for readiness-state tests."""
    artifacts = {name: {"status": "valid"} for name, _, _ in READINESS.REQUIRED_INPUTS}
    artifacts["final_manifest"] = {"status": final_status}
    return artifacts


class Candidate07ReadinessTests(unittest.TestCase):
    """Keep readiness classification fail-closed and non-authorizing."""

    def test_missing_is_blocked(self) -> None:
        report = READINESS.build_report(empty_args())
        self.assertEqual(report["status"], "blocked")
        self.assertEqual(report["cross_binding"]["status"], "not_run")
        self.assertFalse(report["production_cutover_authorized"])
        self.assertFalse(report["stable_promotion_authorized"])
        for name, _, _ in READINESS.REQUIRED_INPUTS:
            self.assertEqual(report["artifacts"][name]["status"], "missing")
        self.assertEqual(report["artifacts"]["final_manifest"]["status"], "not_supplied")
        self.assertEqual(READINESS.exit_code(report["status"]), 2)

    def test_missing_file_named(self) -> None:
        report = READINESS.build_report(empty_args(release_evidence="does-not-exist.json"))
        release = report["artifacts"]["release"]
        self.assertEqual(release["status"], "missing")
        self.assertEqual(release["path"], "does-not-exist.json")

    def test_manifest_ready_status(self) -> None:
        status = READINESS._readiness_status(  # pylint: disable=protected-access
            valid_artifacts("not_supplied"),
            {"status": "valid"},
        )
        self.assertEqual(status, "ready_for_final_manifest")
        self.assertEqual(READINESS.exit_code(status), 3)

    def test_governance_status(self) -> None:
        status = READINESS._readiness_status(  # pylint: disable=protected-access
            valid_artifacts("valid"),
            {"status": "valid"},
        )
        self.assertEqual(status, "ready_for_governance_review")
        self.assertEqual(READINESS.exit_code(status), 0)

    def test_final_invalid_blocks(self) -> None:
        status = READINESS._readiness_status(  # pylint: disable=protected-access
            valid_artifacts("invalid"),
            {"status": "valid"},
        )
        self.assertEqual(status, "blocked")
        self.assertEqual(READINESS.exit_code(status), 2)

    def test_sensitive_json_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "provider.json"
            path.write_text(json.dumps({"token": "do-not-record"}), encoding="utf-8")
            entry, value, _ = READINESS._load_artifact(  # pylint: disable=protected-access
                str(path), "provider_evidence"
            )
        self.assertEqual(entry["status"], "invalid")
        self.assertIsNone(value)
        self.assertIn("Sensitive", entry["detail"])

    def test_bad_json_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "broken.json"
            path.write_text("{", encoding="utf-8")
            entry, value, _ = READINESS._load_artifact(  # pylint: disable=protected-access
                str(path), "release_evidence"
            )
        self.assertEqual(entry["status"], "invalid")
        self.assertIsNone(value)

    def test_parser_allows_empty(self) -> None:
        args = READINESS.parser().parse_args([])
        self.assertIsNone(args.release_evidence)
        self.assertIsNone(args.final_evidence)

    def test_governance_action(self) -> None:
        action = READINESS._operator_action("ready_for_governance_review")  # pylint: disable=protected-access
        self.assertIn("does not authorize production cutover", action)
        self.assertIn("Stable promotion", action)


if __name__ == "__main__":
    unittest.main()
