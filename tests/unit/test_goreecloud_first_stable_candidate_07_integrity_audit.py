# SPDX-License-Identifier: AGPL-3.0-or-later
"""Regression tests for candidate #07 real-provider result integrity."""

from __future__ import annotations

import copy
import importlib.util
from pathlib import Path
import sys
import unittest


MODULE_PATH = Path(__file__).parents[2] / "goreecloud" / "first_stable_candidate_07_integrity_audit.py"
MODULE_DIR = str(MODULE_PATH.parent)
SPEC = importlib.util.spec_from_file_location("first_stable_candidate_07_integrity_audit", MODULE_PATH)
assert SPEC is not None and SPEC.loader is not None
AUDIT = importlib.util.module_from_spec(SPEC)
sys.path.insert(0, MODULE_DIR)
try:
    SPEC.loader.exec_module(AUDIT)
finally:
    sys.path.remove(MODULE_DIR)


def provider_evidence() -> dict:
    """Return the successful evidence shape emitted by frozen provider_acceptance.py."""
    results = []
    for category in AUDIT.FROZEN_PROVIDER_SUITE_CATEGORIES:
        results.append(
            {
                "category": category,
                "exit_code": 0,
                "http_status": 200,
                "product_identity": True,
                "result_cards": 2,
                "engine_messages": 0,
                "passed": True,
            }
        )
    return {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "generated_at": "2026-08-22T19:00:00Z",
        "candidate": {
            "source_revision": AUDIT.candidate_audit.FROZEN_SOURCE,
            "image": AUDIT.candidate_audit.FROZEN_IMAGE,
        },
        "runtime_binding": {
            "verified_before_and_after_requests": True,
            "base_url": "http://127.0.0.1:8888",
            "container": "goreecloud-search-candidate-07-provider",
            "published_port": "127.0.0.1:8888",
            "observed_image_reference": AUDIT.candidate_audit.FROZEN_IMAGE,
            "observed_image_id": "sha256:" + ("d" * 64),
            "oci_revision": AUDIT.candidate_audit.FROZEN_SOURCE,
        },
        "minimum_results": 1,
        "required_categories": sorted(AUDIT.base_audit.REQUIRED_PROVIDER_CATEGORIES),
        "results": results,
        "scope": {
            "real_provider_requests_performed": True,
            "runtime_identity_verified_during_provider_requests": True,
            "all_required_categories_passed": True,
            "full_diagnostic_suite_passed": True,
            "query_text_persisted": False,
            "response_content_persisted": False,
            "production_cutover_authorized": False,
            "statement": "Sanitized exact-candidate real-provider acceptance.",
        },
    }


class Candidate07ProviderIntegrityTests(unittest.TestCase):
    """Reject contradictory or incomplete provider result semantics."""

    def test_provider_shape_passes(self) -> None:
        AUDIT._audit_frozen_provider(  # pylint: disable=protected-access
            provider_evidence(),
            AUDIT.candidate_audit.FROZEN_SOURCE,
            AUDIT.candidate_audit.FROZEN_IMAGE,
        )

    def test_exit_code_conflict(self) -> None:
        evidence = copy.deepcopy(provider_evidence())
        evidence["results"][0]["exit_code"] = 5
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_frozen_provider(  # pylint: disable=protected-access
                evidence,
                AUDIT.candidate_audit.FROZEN_SOURCE,
                AUDIT.candidate_audit.FROZEN_IMAGE,
            )

    def test_http_status_conflict(self) -> None:
        evidence = copy.deepcopy(provider_evidence())
        evidence["results"][1]["http_status"] = 503
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_frozen_provider(  # pylint: disable=protected-access
                evidence,
                AUDIT.candidate_audit.FROZEN_SOURCE,
                AUDIT.candidate_audit.FROZEN_IMAGE,
            )

    def test_product_identity_conflict(self) -> None:
        evidence = copy.deepcopy(provider_evidence())
        evidence["results"][2]["product_identity"] = False
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_frozen_provider(  # pylint: disable=protected-access
                evidence,
                AUDIT.candidate_audit.FROZEN_SOURCE,
                AUDIT.candidate_audit.FROZEN_IMAGE,
            )

    def test_result_threshold(self) -> None:
        evidence = copy.deepcopy(provider_evidence())
        evidence["minimum_results"] = 3
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_frozen_provider(  # pylint: disable=protected-access
                evidence,
                AUDIT.candidate_audit.FROZEN_SOURCE,
                AUDIT.candidate_audit.FROZEN_IMAGE,
            )

    def test_diagnostic_suite(self) -> None:
        evidence = copy.deepcopy(provider_evidence())
        evidence["scope"]["full_diagnostic_suite_passed"] = False
        with self.assertRaises(AUDIT.base_audit.AuditError):
            AUDIT._audit_frozen_provider(  # pylint: disable=protected-access
                evidence,
                AUDIT.candidate_audit.FROZEN_SOURCE,
                AUDIT.candidate_audit.FROZEN_IMAGE,
            )

    def test_duplicate_category(self) -> None:
        evidence = copy.deepcopy(provider_evidence())
        evidence["results"][1]["category"] = "general"
        with self.assertRaises((AUDIT.AuditError, AUDIT.base_audit.AuditError)):
            AUDIT._audit_frozen_provider(  # pylint: disable=protected-access
                evidence,
                AUDIT.candidate_audit.FROZEN_SOURCE,
                AUDIT.candidate_audit.FROZEN_IMAGE,
            )

    def test_noninteger_cards(self) -> None:
        evidence = copy.deepcopy(provider_evidence())
        evidence["results"][3]["result_cards"] = True
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_frozen_provider(  # pylint: disable=protected-access
                evidence,
                AUDIT.candidate_audit.FROZEN_SOURCE,
                AUDIT.candidate_audit.FROZEN_IMAGE,
            )

    def test_required_categories(self) -> None:
        evidence = copy.deepcopy(provider_evidence())
        evidence["required_categories"].append("science")
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._audit_frozen_provider(  # pylint: disable=protected-access
                evidence,
                AUDIT.candidate_audit.FROZEN_SOURCE,
                AUDIT.candidate_audit.FROZEN_IMAGE,
            )


if __name__ == "__main__":
    unittest.main()
