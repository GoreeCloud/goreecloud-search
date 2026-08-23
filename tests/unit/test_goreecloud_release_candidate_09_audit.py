# SPDX-License-Identifier: AGPL-3.0-or-later
"""Regression tests for the current GoreeCloud Search RC 09 evidence contract."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import sys
import tempfile
import unittest


MODULE_PATH = Path(__file__).parents[2] / "goreecloud" / "release_candidate_09_audit.py"
MODULE_DIR = str(MODULE_PATH.parent)
SPEC = importlib.util.spec_from_file_location("release_candidate_09_audit", MODULE_PATH)
assert SPEC is not None and SPEC.loader is not None
AUDIT = importlib.util.module_from_spec(SPEC)
sys.path.insert(0, MODULE_DIR)
try:
    SPEC.loader.exec_module(AUDIT)
finally:
    sys.path.remove(MODULE_DIR)

NOW = "2026-08-23T20:00:00Z"
ARTIFACT_DIGEST = "sha256:" + ("c" * 64)
BROWSER_SOURCE = "3" * 40
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


def write_baseline(path: Path) -> None:
    path.write_text(
        "{\n"
        '  "schema_version": 1,\n'
        '  "environment": "goreecloud-vps-01-production",\n'
        '  "recorded_at": "2026-08-17T18:10:00-05:00",\n'
        '  "source_revision": "3584da535f7ed7c3b4b8dc73cf0424fb4bdf1949",\n'
        '  "image": "ghcr.io/goreecloud/goreecloud-search@sha256:'
        '30ec99e3311fa9dcc934ac267aef123bb2541e0cd0165b360c4a7dc4fa29e3d5",\n'
        '  "purpose": "Known-good production image baseline for isolated candidate rollback '
        'rehearsal. This record does not replace target-host configuration, data, or backup '
        'rollback evidence."\n'
        "}\n",
        encoding="utf-8",
    )


def passed(reference: str) -> dict:
    return {"passed": True, "evidence_reference": reference}


def visual_summary(visual: dict) -> dict:
    return {
        "glaze_ui_version": "1.1.0",
        "review_artifact": visual["review_artifact"],
        "reviews": visual["reviews"],
        "physical_android_preferences_review": visual[
            "physical_android_preferences_review"
        ],
        "desktop_regression_review": visual["desktop_regression_review"],
        "persisted_theme_preference_review": visual[
            "persisted_theme_preference_review"
        ],
    }


def browser_summary(browser: dict) -> dict:
    return {
        "browser_source_revision": BROWSER_SOURCE,
        "runtime_artifact": browser["runtime_artifact"],
        "behaviors": browser["behaviors"],
    }


def build_fixture(root: Path) -> argparse.Namespace:
    paths = {
        name: root / f"{name}-evidence.json"
        for name in ("release", "runtime", "recovery", "provider", "visual", "browser", "final")
    }
    baseline_path = root / "release-baseline.json"
    write_baseline(baseline_path)
    assert sha256(baseline_path) == AUDIT.ROLLBACK_SHA256

    release = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "lifecycle": "Release Candidate",
        "generated_at": NOW,
        "request": {
            "request_id": AUDIT.RC09_REQUEST_ID,
            "candidate_sequence": AUDIT.RC09_SEQUENCE,
            "reviewed_base_revision": AUDIT.RC09_REVIEWED_BASE,
        },
        "candidate": {
            "source_revision": AUDIT.RC09_SOURCE,
            "image": AUDIT.RC09_IMAGE,
            "oci_revision": AUDIT.RC09_SOURCE,
            "oci_version": AUDIT.RC09_OCI_VERSION,
            "registry_digest_pull_verified": True,
            "isolated_runtime_acceptance": "passed",
        },
        "rollback_baseline": {
            "environment": AUDIT.ROLLBACK_ENVIRONMENT,
            "recorded_at": AUDIT.ROLLBACK_RECORDED_AT,
            "source_revision": AUDIT.ROLLBACK_SOURCE,
            "image": AUDIT.ROLLBACK_IMAGE,
            "registry_digest_pull_verified": True,
            "isolated_runtime_acceptance": "passed",
        },
        "rollback_scope": {
            "image_level_rehearsal": "passed",
            "target_environment_configuration_rollback_tested": False,
            "target_environment_data_restore_tested": False,
        },
        "authorization": {
            "production_cutover_authorized": False,
            "stable_release_authorized": False,
            "target_host_change_authorized": False,
        },
        "statement": "RC 09 release identity and image-level rehearsal only.",
    }
    write_json(paths["release"], release)

    runtime = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "generated_at": NOW,
        "target": {
            "base_url": "http://127.0.0.1:8888",
            "container": "goreecloud-search-rc09",
        },
        "http_acceptance": {
            "home_identity": "passed",
            "preferences_identity": "passed",
            "about_identity": "passed",
            "health": "passed",
            "privacy_headers": "passed",
        },
        "providers": "skipped",
        "container_runtime": {
            "status": "verified",
            "running": True,
            "health": "healthy",
            "published_ports": "8080/tcp -> 127.0.0.1:8888",
            "identity_status": "verified",
            "expected_image": AUDIT.RC09_IMAGE,
            "expected_source_revision": AUDIT.RC09_SOURCE,
            "observed_image_reference": AUDIT.RC09_IMAGE,
            "observed_image_id": "sha256:" + ("d" * 64),
            "oci": {
                "title": "GoreeCloud Search",
                "source": "https://github.com/GoreeCloud/goreecloud-search",
                "revision": AUDIT.RC09_SOURCE,
                "version": AUDIT.RC09_OCI_VERSION,
                "licenses": "AGPL-3.0-or-later",
            },
        },
        "scope": {
            "target_runtime_identity_verified": True,
            "target_environment_configuration_rollback_tested": False,
            "target_environment_data_restore_tested": False,
            "backup_restore_tested": False,
            "production_cutover_authorized": False,
            "statement": "Loopback-only target-runtime identity acceptance.",
        },
    }
    write_json(paths["runtime"], runtime)

    recovery = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "environment": "goreecloud-vps-01",
        "generated_at": NOW,
        "candidate": {"source_revision": AUDIT.RC09_SOURCE, "image": AUDIT.RC09_IMAGE},
        "known_good_rollback": {
            "source_revision": AUDIT.ROLLBACK_SOURCE,
            "image": AUDIT.ROLLBACK_IMAGE,
        },
        "artifact_bindings": {
            "release_evidence_sha256": sha256(paths["release"]),
            "target_runtime_evidence_sha256": sha256(paths["runtime"]),
            "rollback_baseline_sha256": AUDIT.ROLLBACK_SHA256,
        },
        "backup": {
            "system": "contract-test-backup",
            "snapshot_reference": "snapshot-rc09",
            "captured_at": NOW,
            "scope": {
                "stack_definition": True,
                "search_settings": True,
                "protected_runtime_configuration_recovery_path": True,
                "caddy_search_route": True,
            },
            "cache_policy": "rebuildable-non-authoritative",
        },
        "restore": {
            "isolated_location": "/tmp/goreecloud-search-rc09-recovery",
            "production_modified": False,
            "stack_definition_restored": True,
            "search_settings_restored": True,
            "protected_runtime_configuration_recovery_verified": True,
            "caddy_search_route_copy_restored": True,
            "compose_validation_passed": True,
            "runtime_recreated": True,
            "runtime_health_passed": True,
            "runtime_identity_passed": True,
            "target_acceptance_passed": True,
        },
        "rollback": {
            "mode": "equivalent-verified-evidence",
            "known_good_image_available": True,
            "previous_runtime_configuration_preserved": True,
            "previous_caddy_route_preserved": True,
            "rollback_procedure_documented": True,
            "isolated_rollback_rehearsal_passed": True,
            "production_route_rollback_tested": False,
            "equivalent_verified_rollback_evidence": True,
        },
        "monitoring": {
            "monitor_identity": "GoreeCloud Search",
            "availability_monitor_verified": True,
            "alert_delivery_verified": True,
        },
        "scope": {
            "application_level_restore_tested": True,
            "monitoring_and_alerting_verified": True,
            "rollback_evidence_verified": True,
            "production_cutover_authorized": False,
        },
    }
    write_json(paths["recovery"], recovery)

    provider = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "generated_at": NOW,
        "candidate": {"source_revision": AUDIT.RC09_SOURCE, "image": AUDIT.RC09_IMAGE},
        "runtime_binding": {
            "verified_before_and_after_requests": True,
            "base_url": "http://127.0.0.1:8888",
            "container": "goreecloud-search-rc09",
            "published_port": "8080/tcp -> 127.0.0.1:8888",
            "observed_image_reference": AUDIT.RC09_IMAGE,
            "observed_image_id": "sha256:" + ("d" * 64),
            "oci_revision": AUDIT.RC09_SOURCE,
        },
        "required_categories": ["general", "images", "videos", "news", "files"],
        "results": [
            {"category": category, "passed": True}
            for category in ("general", "images", "videos", "news", "files")
        ],
        "scope": {
            "real_provider_requests_performed": True,
            "runtime_identity_verified_during_provider_requests": True,
            "all_required_categories_passed": True,
            "query_text_persisted": False,
            "response_content_persisted": False,
            "production_cutover_authorized": False,
        },
    }
    write_json(paths["provider"], provider)

    visual = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "generated_at": NOW,
        "candidate": {"source_revision": AUDIT.RC09_SOURCE, "image": AUDIT.RC09_IMAGE},
        "glaze_ui_version": "1.1.0",
        "review_artifact": {"reference": "rc09-visual", "digest": ARTIFACT_DIGEST},
        "reviews": {case: passed(f"visual-{case}") for case in VISUAL_CASES},
        "physical_android_preferences_review": passed("android-preferences"),
        "desktop_regression_review": passed("desktop-regression"),
        "persisted_theme_preference_review": passed("persisted-theme"),
        "scope": {
            "exact_candidate_visual_artifact_verified": True,
            "manual_visual_acceptance_verified": True,
            "production_cutover_authorized": False,
            "statement": "Exact RC 09 visual and device review completed.",
        },
    }
    write_json(paths["visual"], visual)

    browser = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "generated_at": NOW,
        "search_candidate": {
            "source_revision": AUDIT.RC09_SOURCE,
            "image": AUDIT.RC09_IMAGE,
        },
        "browser_source_revision": BROWSER_SOURCE,
        "runtime_artifact": {"reference": "browser-runtime", "digest": ARTIFACT_DIGEST},
        "behaviors": {behavior: True for behavior in BROWSER_BEHAVIORS},
        "scope": {
            "actual_browser_runtime_verified": True,
            "search_candidate_runtime_verified": True,
            "production_cutover_authorized": False,
            "statement": "Actual Browser runtime verified against immutable RC 09.",
        },
    }
    write_json(paths["browser"], browser)

    final = {
        "schema_version": 2,
        "product": "GoreeCloud Search",
        "generated_at": NOW,
        "candidate": {"source_revision": AUDIT.RC09_SOURCE, "image": AUDIT.RC09_IMAGE},
        "artifact_bindings": {
            "release_evidence_sha256": sha256(paths["release"]),
            "target_runtime_evidence_sha256": sha256(paths["runtime"]),
            "recovery_evidence_sha256": sha256(paths["recovery"]),
            "provider_evidence_sha256": sha256(paths["provider"]),
            "visual_evidence_sha256": sha256(paths["visual"]),
            "browser_evidence_sha256": sha256(paths["browser"]),
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
            "statement": "Complete RC 09 evidence remains non-authorizing.",
        },
    }
    write_json(paths["final"], final)

    return argparse.Namespace(
        release_evidence=str(paths["release"]),
        target_runtime_evidence=str(paths["runtime"]),
        recovery_evidence=str(paths["recovery"]),
        rollback_baseline=str(baseline_path),
        provider_evidence=str(paths["provider"]),
        visual_evidence=str(paths["visual"]),
        browser_evidence=str(paths["browser"]),
        final_evidence=str(paths["final"]),
    )


class RC09AuditTests(unittest.TestCase):
    """Keep the current RC 09 evidence contract exact and non-authorizing."""

    def test_complete(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            result = AUDIT.audit(build_fixture(Path(tempdir)))
            self.assertEqual(result["source_revision"], AUDIT.RC09_SOURCE)
            self.assertEqual(result["image"], AUDIT.RC09_IMAGE)
            self.assertFalse(result["production_cutover_authorized"])
            self.assertFalse(result["stable_promotion_authorized"])

    def test_source_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = build_fixture(Path(tempdir))
            path = Path(args.release_evidence)
            value = json.loads(path.read_text(encoding="utf-8"))
            value["candidate"]["source_revision"] = "1" * 40
            write_json(path, value)
            with self.assertRaises(AUDIT.AuditError):
                AUDIT.audit(args)

    def test_version_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = build_fixture(Path(tempdir))
            path = Path(args.release_evidence)
            value = json.loads(path.read_text(encoding="utf-8"))
            value["candidate"]["oci_version"] = "wrong-version"
            write_json(path, value)
            with self.assertRaises(AUDIT.AuditError):
                AUDIT.audit(args)

    def test_auth_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = build_fixture(Path(tempdir))
            path = Path(args.release_evidence)
            value = json.loads(path.read_text(encoding="utf-8"))
            value["authorization"]["stable_release_authorized"] = True
            write_json(path, value)
            with self.assertRaises(AUDIT.base_audit.AuditError):
                AUDIT.audit(args)

    def test_baseline_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = build_fixture(Path(tempdir))
            path = Path(args.rollback_baseline)
            path.write_text(path.read_text(encoding="utf-8") + "\n", encoding="utf-8")
            with self.assertRaises(AUDIT.AuditError):
                AUDIT.audit(args)

    def test_runtime_version(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = build_fixture(Path(tempdir))
            path = Path(args.target_runtime_evidence)
            value = json.loads(path.read_text(encoding="utf-8"))
            value["container_runtime"]["oci"]["version"] = "wrong-version"
            write_json(path, value)
            with self.assertRaises(AUDIT.AuditError):
                AUDIT.audit(args)

    def test_provider_missing(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = build_fixture(Path(tempdir))
            path = Path(args.provider_evidence)
            value = json.loads(path.read_text(encoding="utf-8"))
            value["results"][0]["passed"] = False
            write_json(path, value)
            with self.assertRaises(AUDIT.base_audit.AuditError):
                AUDIT.audit(args)


if __name__ == "__main__":
    unittest.main()
