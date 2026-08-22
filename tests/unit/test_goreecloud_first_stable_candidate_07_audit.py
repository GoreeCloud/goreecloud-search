# SPDX-License-Identifier: AGPL-3.0-or-later
"""Regression tests for the frozen first-Stable candidate #07 evidence contract."""

from __future__ import annotations

import copy
import unittest

from goreecloud import first_stable_candidate_07_audit as audit


NOW = "2026-08-22T17:00:00Z"


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
            "expected_image": audit.FROZEN_IMAGE,
            "expected_source_revision": audit.FROZEN_SOURCE,
            "observed_image_reference": audit.FROZEN_IMAGE,
            "observed_image_id": "sha256:" + ("d" * 64),
            "oci": {
                "title": "GoreeCloud Search",
                "source": "https://github.com/GoreeCloud/goreecloud-search",
                "revision": audit.FROZEN_SOURCE,
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


class Candidate07EvidenceAuditTests(unittest.TestCase):
    """Keep the master audit compatible with the immutable candidate #07 evidence schema."""

    def test_frozen_runtime_contract_passes(self) -> None:
        audit._audit_frozen_runtime(  # pylint: disable=protected-access
            runtime_evidence(), audit.FROZEN_SOURCE, audit.FROZEN_IMAGE
        )

    def test_synthetic_status_rejected(self) -> None:
        evidence = copy.deepcopy(runtime_evidence())
        evidence["container_runtime"]["status"] = "verified"
        with self.assertRaises(audit.AuditError):
            audit._audit_frozen_runtime(evidence, audit.FROZEN_SOURCE, audit.FROZEN_IMAGE)  # pylint: disable=protected-access

    def test_literal_ports_rejected(self) -> None:
        evidence = copy.deepcopy(runtime_evidence())
        evidence["container_runtime"]["published_ports"] = "8080/tcp -> 127.0.0.1:8888"
        with self.assertRaises(audit.AuditError):
            audit._audit_frozen_runtime(evidence, audit.FROZEN_SOURCE, audit.FROZEN_IMAGE)  # pylint: disable=protected-access

    def test_non_loopback_target_rejected(self) -> None:
        evidence = copy.deepcopy(runtime_evidence())
        evidence["target"]["base_url"] = "https://search.goreecloud.com"
        with self.assertRaises(audit.AuditError):
            audit._audit_frozen_runtime(evidence, audit.FROZEN_SOURCE, audit.FROZEN_IMAGE)  # pylint: disable=protected-access

    def test_unverified_ports_rejected(self) -> None:
        evidence = copy.deepcopy(runtime_evidence())
        evidence["container_runtime"]["published_ports"] = "not_checked"
        with self.assertRaises(audit.AuditError):
            audit._audit_frozen_runtime(evidence, audit.FROZEN_SOURCE, audit.FROZEN_IMAGE)  # pylint: disable=protected-access

    def test_wrong_candidate_rejected(self) -> None:
        with self.assertRaises(audit.AuditError):
            audit._require_frozen_candidate("1" * 40, audit.FROZEN_IMAGE)  # pylint: disable=protected-access
        with self.assertRaises(audit.AuditError):
            audit._require_frozen_candidate(audit.FROZEN_SOURCE, audit.FROZEN_IMAGE[:-1] + "0")  # pylint: disable=protected-access


if __name__ == "__main__":
    unittest.main()
