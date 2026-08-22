# SPDX-License-Identifier: AGPL-3.0-or-later
"""Deterministic tests for the GoreeCloud Search first-Stable evidence safety audit."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import shutil
import tempfile
import unittest


MODULE_PATH = Path(__file__).parents[2] / "goreecloud" / "first_stable_evidence_audit.py"
SPEC = importlib.util.spec_from_file_location("first_stable_evidence_audit", MODULE_PATH)
assert SPEC is not None and SPEC.loader is not None
AUDIT = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(AUDIT)

SOURCE = "1" * 40
IMAGE = "ghcr.io/goreecloud/goreecloud-search@sha256:" + ("a" * 64)
ROLLBACK_SOURCE = "2" * 40
ROLLBACK_IMAGE = "ghcr.io/goreecloud/goreecloud-search@sha256:" + ("b" * 64)
ARTIFACT_DIGEST = "sha256:" + ("c" * 64)
NOW = "2026-08-22T16:00:00Z"


def sha256(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write_json(path: Path, value: dict) -> None:
    path.write_text(json.dumps(value, indent=2, sort_keys=True) + "\n", encoding="utf-8")


class EvidenceFixture:  # pylint: disable=too-few-public-methods
    """Build one internally consistent synthetic six-artifact evidence set."""

    def __init__(self, root: Path):
        self.root = root
        self.paths = {
            name: root / f"{name}-evidence.json"
            for name in ("release", "runtime", "recovery", "provider", "visual", "browser", "final")
        }
        self._write_release()
        self._write_runtime()
        self._write_recovery()
        self._write_provider()
        self._write_visual()
        self._write_browser()
        self._write_final()

    def _write_release(self) -> None:
        write_json(
            self.paths["release"],
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

    def _write_runtime(self) -> None:
        write_json(
            self.paths["runtime"],
            {
                "schema_version": 1,
                "product": "GoreeCloud Search",
                "generated_at": NOW,
                "target": {"base_url": "http://127.0.0.1:8888", "container": "goreecloud-search"},
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
                    "expected_image": IMAGE,
                    "expected_source_revision": SOURCE,
                    "observed_image_reference": IMAGE,
                    "observed_image_id": "sha256:" + ("d" * 64),
                    "oci": {
                        "title": "GoreeCloud Search",
                        "source": "https://github.com/GoreeCloud/goreecloud-search",
                        "revision": SOURCE,
                        "version": "2026.8.19-test",
                        "licenses": "AGPL-3.0-or-later",
                    },
                },
                "scope": {
                    "target_runtime_identity_verified": True,
                    "target_environment_configuration_rollback_tested": False,
                    "target_environment_data_restore_tested": False,
                    "backup_restore_tested": False,
                    "production_cutover_authorized": False,
                },
            },
        )

    def _write_recovery(self) -> None:
        write_json(
            self.paths["recovery"],
            {
                "schema_version": 1,
                "product": "GoreeCloud Search",
                "environment": "goreecloud-vps-01",
                "generated_at": NOW,
                "candidate": {"source_revision": SOURCE, "image": IMAGE},
                "known_good_rollback": {"source_revision": ROLLBACK_SOURCE, "image": ROLLBACK_IMAGE},
                "artifact_bindings": {
                    "release_evidence_sha256": sha256(self.paths["release"]),
                    "target_runtime_evidence_sha256": sha256(self.paths["runtime"]),
                    "rollback_baseline_sha256": "e" * 64,
                },
                "backup": {
                    "system": "contract-test-backup",
                    "snapshot_reference": "snapshot-001",
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
                    "isolated_location": "/tmp/goreecloud-search-recovery",
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
            },
        )

    def _write_provider(self) -> None:
        write_json(
            self.paths["provider"],
            {
                "schema_version": 1,
                "product": "GoreeCloud Search",
                "generated_at": NOW,
                "candidate": {"source_revision": SOURCE, "image": IMAGE},
                "runtime_binding": {
                    "verified_before_and_after_requests": True,
                    "base_url": "http://127.0.0.1:8888",
                    "container": "goreecloud-search",
                    "published_port": "8080/tcp -> 127.0.0.1:8888",
                    "observed_image_reference": IMAGE,
                    "observed_image_id": "sha256:" + ("d" * 64),
                    "oci_revision": SOURCE,
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
            },
        )

    def _write_visual(self) -> None:
        write_json(
            self.paths["visual"],
            {
                "schema_version": 1,
                "product": "GoreeCloud Search",
                "generated_at": NOW,
                "candidate": {"source_revision": SOURCE, "image": IMAGE},
                "review_artifact": {"reference": "visual-artifact", "digest": ARTIFACT_DIGEST},
                "scope": {
                    "exact_candidate_visual_artifact_verified": True,
                    "manual_visual_acceptance_verified": True,
                    "production_cutover_authorized": False,
                },
            },
        )

    def _write_browser(self) -> None:
        write_json(
            self.paths["browser"],
            {
                "schema_version": 1,
                "product": "GoreeCloud Search",
                "generated_at": NOW,
                "search_candidate": {"source_revision": SOURCE, "image": IMAGE},
                "browser_source_revision": "3" * 40,
                "runtime_artifact": {"reference": "browser-artifact", "digest": ARTIFACT_DIGEST},
                "scope": {
                    "actual_browser_runtime_verified": True,
                    "search_candidate_runtime_verified": True,
                    "production_cutover_authorized": False,
                },
            },
        )

    def _write_final(self) -> None:
        bindings = {
            "release_evidence_sha256": sha256(self.paths["release"]),
            "target_runtime_evidence_sha256": sha256(self.paths["runtime"]),
            "recovery_evidence_sha256": sha256(self.paths["recovery"]),
            "provider_evidence_sha256": sha256(self.paths["provider"]),
            "visual_evidence_sha256": sha256(self.paths["visual"]),
            "browser_evidence_sha256": sha256(self.paths["browser"]),
        }
        write_json(
            self.paths["final"],
            {
                "schema_version": 2,
                "product": "GoreeCloud Search",
                "generated_at": NOW,
                "candidate": {"source_revision": SOURCE, "image": IMAGE},
                "artifact_bindings": bindings,
                "scope": {
                    "glaze_ui_1_1_final_visual_acceptance_verified": True,
                    "browser_runtime_integration_verified": True,
                    "real_provider_acceptance_verified": True,
                    "recovery_evidence_verified": True,
                    "final_candidate_acceptance_complete": True,
                    "production_cutover_authorized": False,
                },
            },
        )

    def args(self) -> argparse.Namespace:
        return argparse.Namespace(
            release_evidence=str(self.paths["release"]),
            target_runtime_evidence=str(self.paths["runtime"]),
            recovery_evidence=str(self.paths["recovery"]),
            provider_evidence=str(self.paths["provider"]),
            visual_evidence=str(self.paths["visual"]),
            browser_evidence=str(self.paths["browser"]),
            final_evidence=str(self.paths["final"]),
        )


class FirstStableEvidenceAuditTests(unittest.TestCase):
    """Exercise strict cross-artifact safety and identity binding."""

    def setUp(self) -> None:
        self.root = Path(tempfile.mkdtemp())
        self.addCleanup(shutil.rmtree, self.root)
        self.fixture = EvidenceFixture(self.root)

    def test_complete_set_passes(self) -> None:
        result = AUDIT.audit(self.fixture.args())
        self.assertEqual(result["source_revision"], SOURCE)
        self.assertEqual(result["image"], IMAGE)

    def test_release_secret_rejected(self) -> None:
        release = json.loads(self.fixture.paths["release"].read_text())
        release["secret"] = "must-not-be-bound"
        write_json(self.fixture.paths["release"], release)
        with self.assertRaises(AUDIT.AuditError):
            AUDIT.audit(self.fixture.args())

    def test_provider_query_rejected(self) -> None:
        provider = json.loads(self.fixture.paths["provider"].read_text())
        provider["results"][0]["query"] = "must-not-be-bound"
        write_json(self.fixture.paths["provider"], provider)
        with self.assertRaises(AUDIT.AuditError):
            AUDIT.audit(self.fixture.args())

    def test_runtime_image_mismatch(self) -> None:
        runtime = json.loads(self.fixture.paths["runtime"].read_text())
        runtime["container_runtime"]["observed_image_reference"] = ROLLBACK_IMAGE
        write_json(self.fixture.paths["runtime"], runtime)
        with self.assertRaises(AUDIT.AuditError):
            AUDIT.audit(self.fixture.args())

    def test_recovery_hash_mismatch(self) -> None:
        recovery = json.loads(self.fixture.paths["recovery"].read_text())
        recovery["artifact_bindings"]["release_evidence_sha256"] = "0" * 64
        write_json(self.fixture.paths["recovery"], recovery)
        with self.assertRaises(AUDIT.AuditError):
            AUDIT.audit(self.fixture.args())

    def test_final_hash_mismatch(self) -> None:
        final = json.loads(self.fixture.paths["final"].read_text())
        final["artifact_bindings"]["visual_evidence_sha256"] = "0" * 64
        write_json(self.fixture.paths["final"], final)
        with self.assertRaises(AUDIT.AuditError):
            AUDIT.audit(self.fixture.args())


if __name__ == "__main__":
    unittest.main()
