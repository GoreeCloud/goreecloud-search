# SPDX-License-Identifier: AGPL-3.0-or-later
"""Deterministic tests for deep first-Stable visual and Browser review evidence audit."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest


MODULE_PATH = Path(__file__).parents[2] / "goreecloud" / "final_review_evidence_audit.py"
SPEC = importlib.util.spec_from_file_location("final_review_evidence_audit", MODULE_PATH)
assert SPEC is not None and SPEC.loader is not None
AUDIT = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(AUDIT)

SOURCE = "1" * 40
IMAGE = "ghcr.io/goreecloud/goreecloud-search@sha256:" + ("a" * 64)
BROWSER_SOURCE = "2" * 40
ARTIFACT_DIGEST = "sha256:" + ("c" * 64)
NOW = "2026-08-22T16:00:00Z"
VISUAL_CASES = ("compact_light", "compact_dark", "expanded_light", "expanded_dark")
BROWSER_BEHAVIORS = (
    "search_only_default_provider",
    "address_bar_routed_through_search",
    "new_tab_routed_through_search",
    "dedicated_search_field_routed_through_search",
    "no_external_browser_fallback",
    "search_unavailability_state_verified",
    "recovery_after_search_reachability_verified",
)


def write_json(path: Path, value: dict) -> None:
    path.write_text(json.dumps(value, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def sha256(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def review(reference: str) -> dict:
    return {"passed": True, "evidence_reference": reference}


def write_release(path: Path) -> None:
    write_json(
        path,
        {
            "schema_version": 1,
            "product": "GoreeCloud Search",
            "candidate": {
                "source_revision": SOURCE,
                "image": IMAGE,
                "isolated_runtime_acceptance": "passed",
            },
            "rollback_scope": {
                "image_level_rehearsal": "passed",
                "production_cutover_authorized": False,
            },
        },
    )


def write_visual(path: Path) -> dict:
    value = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "generated_at": NOW,
        "candidate": {"source_revision": SOURCE, "image": IMAGE},
        "glaze_ui_version": "1.1.0",
        "review_artifact": {"reference": "visual-artifact", "digest": ARTIFACT_DIGEST},
        "reviews": {case: review(f"visual-{case}") for case in VISUAL_CASES},
        "physical_android_preferences_review": review("android-preferences"),
        "desktop_regression_review": review("desktop-regression"),
        "persisted_theme_preference_review": review("persisted-theme"),
        "scope": {
            "exact_candidate_visual_artifact_verified": True,
            "manual_visual_acceptance_verified": True,
            "production_cutover_authorized": False,
            "statement": "Exact candidate visual and device review completed.",
        },
    }
    write_json(path, value)
    return value


def write_browser(path: Path) -> dict:
    value = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "generated_at": NOW,
        "search_candidate": {"source_revision": SOURCE, "image": IMAGE},
        "browser_source_revision": BROWSER_SOURCE,
        "runtime_artifact": {"reference": "browser-runtime", "digest": ARTIFACT_DIGEST},
        "behaviors": {behavior: True for behavior in BROWSER_BEHAVIORS},
        "scope": {
            "actual_browser_runtime_verified": True,
            "search_candidate_runtime_verified": True,
            "production_cutover_authorized": False,
            "statement": "Actual Browser runtime verified against the exact Search candidate.",
        },
    }
    write_json(path, value)
    return value


def visual_summary(visual: dict) -> dict:
    return {
        "glaze_ui_version": "1.1.0",
        "review_artifact": visual["review_artifact"],
        "reviews": visual["reviews"],
        "physical_android_preferences_review": visual["physical_android_preferences_review"],
        "desktop_regression_review": visual["desktop_regression_review"],
        "persisted_theme_preference_review": visual["persisted_theme_preference_review"],
    }


def browser_summary(browser: dict) -> dict:
    return {
        "browser_source_revision": BROWSER_SOURCE,
        "runtime_artifact": browser["runtime_artifact"],
        "behaviors": browser["behaviors"],
    }


def write_final(path: Path, visual_path: Path, browser_path: Path, visual: dict, browser: dict) -> None:
    write_json(
        path,
        {
            "schema_version": 2,
            "product": "GoreeCloud Search",
            "generated_at": NOW,
            "candidate": {"source_revision": SOURCE, "image": IMAGE},
            "artifact_bindings": {
                "release_evidence_sha256": "d" * 64,
                "target_runtime_evidence_sha256": "e" * 64,
                "recovery_evidence_sha256": "f" * 64,
                "provider_evidence_sha256": "0" * 64,
                "visual_evidence_sha256": sha256(visual_path),
                "browser_evidence_sha256": sha256(browser_path),
            },
            "visual_acceptance": visual_summary(visual),
            "browser_integration": browser_summary(browser),
            "scope": {
                "glaze_ui_1_1_final_visual_acceptance_verified": True,
                "browser_runtime_integration_verified": True,
                "real_provider_acceptance_verified": True,
                "recovery_evidence_verified": True,
                "final_candidate_acceptance_complete": True,
                "production_cutover_authorized": False,
                "statement": "Complete evidence set remains non-authorizing.",
            },
        },
    )


def build_fixture(root: Path) -> argparse.Namespace:
    release_path = root / "release-evidence.json"
    visual_path = root / "visual-evidence.json"
    browser_path = root / "browser-evidence.json"
    final_path = root / "final-evidence.json"
    write_release(release_path)
    visual = write_visual(visual_path)
    browser = write_browser(browser_path)
    write_final(final_path, visual_path, browser_path, visual, browser)
    return argparse.Namespace(
        release_evidence=str(release_path),
        visual_evidence=str(visual_path),
        browser_evidence=str(browser_path),
        final_evidence=str(final_path),
    )


class FinalReviewEvidenceAuditTests(unittest.TestCase):
    """Exercise deep visual, device, Browser, and final-summary validation."""

    def test_complete_passes(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = build_fixture(Path(tempdir))
            result = AUDIT.audit(args)
            self.assertEqual(result["source_revision"], SOURCE)
            self.assertEqual(result["image"], IMAGE)

    def test_missing_visual_case(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = build_fixture(Path(tempdir))
            path = Path(args.visual_evidence)
            visual = json.loads(path.read_text(encoding="utf-8"))
            del visual["reviews"]["compact_dark"]
            write_json(path, visual)
            with self.assertRaises(AUDIT.AuditError):
                AUDIT.audit(args)

    def test_android_review_false(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = build_fixture(Path(tempdir))
            path = Path(args.visual_evidence)
            visual = json.loads(path.read_text(encoding="utf-8"))
            visual["physical_android_preferences_review"]["passed"] = False
            write_json(path, visual)
            with self.assertRaises(AUDIT.AuditError):
                AUDIT.audit(args)

    def test_browser_behavior_false(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = build_fixture(Path(tempdir))
            path = Path(args.browser_evidence)
            browser = json.loads(path.read_text(encoding="utf-8"))
            browser["behaviors"]["no_external_browser_fallback"] = False
            write_json(path, browser)
            with self.assertRaises(AUDIT.AuditError):
                AUDIT.audit(args)

    def test_final_visual_mismatch(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = build_fixture(Path(tempdir))
            path = Path(args.final_evidence)
            final = json.loads(path.read_text(encoding="utf-8"))
            final["visual_acceptance"]["reviews"]["compact_light"]["evidence_reference"] = "changed"
            write_json(path, final)
            with self.assertRaises(AUDIT.AuditError):
                AUDIT.audit(args)

    def test_final_hash_mismatch(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = build_fixture(Path(tempdir))
            path = Path(args.final_evidence)
            final = json.loads(path.read_text(encoding="utf-8"))
            final["artifact_bindings"]["browser_evidence_sha256"] = "9" * 64
            write_json(path, final)
            with self.assertRaises(AUDIT.AuditError):
                AUDIT.audit(args)

    def test_query_field_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = build_fixture(Path(tempdir))
            path = Path(args.browser_evidence)
            browser = json.loads(path.read_text(encoding="utf-8"))
            browser["query"] = "must-not-be-bound"
            write_json(path, browser)
            with self.assertRaises(AUDIT.AuditError):
                AUDIT.audit(args)


if __name__ == "__main__":
    unittest.main()
