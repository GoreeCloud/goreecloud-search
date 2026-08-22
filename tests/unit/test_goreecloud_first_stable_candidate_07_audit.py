# SPDX-License-Identifier: AGPL-3.0-or-later
"""Regression tests for the frozen first-Stable candidate #07 evidence contract."""

from __future__ import annotations

import copy
import importlib.util
from pathlib import Path
import sys
import unittest


MODULE_PATH = Path(__file__).parents[2] / "goreecloud" / "first_stable_candidate_07_audit.py"
MODULE_DIR = str(MODULE_PATH.parent)
SPEC = importlib.util.spec_from_file_location("first_stable_candidate_07_audit", MODULE_PATH)
assert SPEC is not None and SPEC.loader is not None
AUDIT = importlib.util.module_from_spec(SPEC)
sys.path.insert(0, MODULE_DIR)
try:
    SPEC.loader.exec_module(AUDIT)
finally:
    sys.path.remove(MODULE_DIR)

NOW = "2026-08-22T17:00:00Z"
BASELINE_DIGEST = "a" * 64


def runtime_evidence() -> dict:
    """Return the target-runtime shape emitted by frozen candidate #07 target_acceptance.sh."""
    return {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "generated_at": NOW,
        "target": {
            "base_url": "http://127.0.0.1:8888",
            "container": "goreecloud-search-candidate-07",
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
            "status": "checked",
            "running": True,
            "health": "healthy",
            "published_ports": "verified",
            "identity_status": "verified",
            "expected_image": AUDIT.FROZEN_IMAGE,
            "expected_source_revision": AUDIT.FROZEN_SOURCE,
            "observed_image_reference": AUDIT.FROZEN_IMAGE,
            "observed_image_id": "sha256:" + ("d" * 64),
            "oci": {
                "title": "GoreeCloud Search",
                "source": "https://github.com/GoreeCloud/goreecloud-search",
                "revision": AUDIT.FROZEN_SOURCE,
                "version": "2026.8.19-b355aafe7",
                "licenses": "AGPL-3.0-or-later",
            },
        },
        "scope": {
            "target_runtime_identity_verified": True,
            "target_environment_configuration_rollback_tested": False,
            "target_environment_data_restore_tested": False,
            "backup_restore_tested": False,
            "production_cutover_authorized": False,
            "statement": "Read-only exact-candidate target-runtime identity acceptance.",
        },
    }


def rollback_baseline() -> dict:
    """Return the immutable known-good rollback identity recorded with candidate #07."""
    return {
        "schema_version": 1,
        "environment": AUDIT.FROZEN_ROLLBACK_ENVIRONMENT,
        "recorded_at": "2026-08-17T18:10:00-05:00",
        "source_revision": AUDIT.FROZEN_ROLLBACK_SOURCE,
        "image": AUDIT.FROZEN_ROLLBACK_IMAGE,
        "purpose": "Known-good production image baseline for isolated rollback rehearsal.",
    }


def recovery_binding() -> dict:
    """Return the recovery fields that bind to the frozen rollback baseline."""
    return {
        "known_good_rollback": {
            "source_revision": AUDIT.FROZEN_ROLLBACK_SOURCE,
            "image": AUDIT.FROZEN_ROLLBACK_IMAGE,
        },
        "artifact_bindings": {
            "rollback_baseline_sha256": BASELINE_DIGEST,
        },
    }


class Candidate07EvidenceAuditTests(unittest.TestCase):
    """Keep the master audit compatible with immutable candidate #07 evidence."""

    def test_frozen_runtime_passes(self) -> None:
        AUDIT._audit_frozen_runtime(  # pylint: disable=protected-access
            runtime_evidence(), AUDIT.FROZEN_SOURCE, AUDIT.FROZEN_IMAGE
        )

    def test_frozen_rollback_passes(self) -> None:
        result = AUDIT._audit_frozen_rollback_baseline(  # pylint: disable=protected-access
            rollback_baseline()
        )
        self.assertEqual(
            result,
            (AUDIT.FROZEN_ROLLBACK_SOURCE, AUDIT.FROZEN_ROLLBACK_IMAGE),
        )

    def test_recovery_baseline_binding_passes(self) -> None:
        AUDIT._audit_recovery_baseline_binding(  # pylint: disable=protected-access
            recovery_binding(),
            AUDIT.FROZEN_ROLLBACK_SOURCE,
            AUDIT.FROZEN_ROLLBACK_IMAGE,
            BASELINE_DIGEST,
        )

    def test_synthetic_status_rejected(self) -> None:
        evidence = copy.deepcopy(runtime_evidence())
        evidence["container_runtime"]["status"] = "verified"
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_frozen_runtime(  # pylint: disable=protected-access
                evidence, AUDIT.FROZEN_SOURCE, AUDIT.FROZEN_IMAGE
            )

    def test_literal_ports_rejected(self) -> None:
        evidence = copy.deepcopy(runtime_evidence())
        evidence["container_runtime"]["published_ports"] = "8080/tcp -> 127.0.0.1:8888"
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_frozen_runtime(  # pylint: disable=protected-access
                evidence, AUDIT.FROZEN_SOURCE, AUDIT.FROZEN_IMAGE
            )

    def test_nonloopback_rejected(self) -> None:
        evidence = copy.deepcopy(runtime_evidence())
        evidence["target"]["base_url"] = "https://search.goreecloud.com"
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_frozen_runtime(  # pylint: disable=protected-access
                evidence, AUDIT.FROZEN_SOURCE, AUDIT.FROZEN_IMAGE
            )

    def test_unverified_ports_rejected(self) -> None:
        evidence = copy.deepcopy(runtime_evidence())
        evidence["container_runtime"]["published_ports"] = "not_checked"
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_frozen_runtime(  # pylint: disable=protected-access
                evidence, AUDIT.FROZEN_SOURCE, AUDIT.FROZEN_IMAGE
            )

    def test_wrong_candidate_rejected(self) -> None:
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._require_frozen_candidate(  # pylint: disable=protected-access
                "1" * 40, AUDIT.FROZEN_IMAGE
            )
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._require_frozen_candidate(  # pylint: disable=protected-access
                AUDIT.FROZEN_SOURCE, AUDIT.FROZEN_IMAGE[:-1] + "0"
            )

    def test_wrong_rollback_image_rejected(self) -> None:
        baseline = copy.deepcopy(rollback_baseline())
        baseline["image"] = AUDIT.FROZEN_IMAGE
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_frozen_rollback_baseline(  # pylint: disable=protected-access
                baseline
            )

    def test_wrong_rollback_environment_rejected(self) -> None:
        baseline = copy.deepcopy(rollback_baseline())
        baseline["environment"] = "staging"
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_frozen_rollback_baseline(  # pylint: disable=protected-access
                baseline
            )

    def test_wrong_rollback_digest_rejected(self) -> None:
        recovery = copy.deepcopy(recovery_binding())
        recovery["artifact_bindings"]["rollback_baseline_sha256"] = "b" * 64
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_recovery_baseline_binding(  # pylint: disable=protected-access
                recovery,
                AUDIT.FROZEN_ROLLBACK_SOURCE,
                AUDIT.FROZEN_ROLLBACK_IMAGE,
                BASELINE_DIGEST,
            )

    def test_wrong_recovery_rollback_source_rejected(self) -> None:
        recovery = copy.deepcopy(recovery_binding())
        recovery["known_good_rollback"]["source_revision"] = "1" * 40
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_recovery_baseline_binding(  # pylint: disable=protected-access
                recovery,
                AUDIT.FROZEN_ROLLBACK_SOURCE,
                AUDIT.FROZEN_ROLLBACK_IMAGE,
                BASELINE_DIGEST,
            )


if __name__ == "__main__":
    unittest.main()
