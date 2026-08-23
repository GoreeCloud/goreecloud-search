# SPDX-License-Identifier: AGPL-3.0-or-later
"""Regression tests for the GoreeCloud Search RC 09 recovery-evidence tooling."""

from __future__ import annotations

import argparse
import importlib.util
import json
from pathlib import Path
import sys
import tempfile
import unittest


MODULE_PATH = Path(__file__).parents[2] / "goreecloud" / "recovery_evidence.py"
MODULE_DIR = str(MODULE_PATH.parent)
SPEC = importlib.util.spec_from_file_location("recovery_evidence", MODULE_PATH)
assert SPEC is not None and SPEC.loader is not None
RECOVERY = importlib.util.module_from_spec(SPEC)
sys.path.insert(0, MODULE_DIR)
try:
    SPEC.loader.exec_module(RECOVERY)
finally:
    sys.path.remove(MODULE_DIR)

NOW = "2026-08-23T21:30:00Z"
BASELINE = MODULE_PATH.parent / "release_baseline.json"


def write_json(path: Path, value: dict) -> None:
    path.write_text(json.dumps(value, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def release_evidence() -> dict:
    return {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "lifecycle": "Release Candidate",
        "generated_at": NOW,
        "request": {
            "request_id": RECOVERY.rc09_audit.RC09_REQUEST_ID,
            "candidate_sequence": RECOVERY.rc09_audit.RC09_SEQUENCE,
            "reviewed_base_revision": RECOVERY.rc09_audit.RC09_REVIEWED_BASE,
        },
        "candidate": {
            "source_revision": RECOVERY.RC09_SOURCE,
            "image": RECOVERY.RC09_IMAGE,
            "oci_revision": RECOVERY.RC09_SOURCE,
            "oci_version": RECOVERY.rc09_audit.RC09_OCI_VERSION,
            "registry_digest_pull_verified": True,
            "isolated_runtime_acceptance": "passed",
        },
        "rollback_baseline": {
            "environment": RECOVERY.rc09_audit.ROLLBACK_ENVIRONMENT,
            "recorded_at": RECOVERY.rc09_audit.ROLLBACK_RECORDED_AT,
            "source_revision": RECOVERY.ROLLBACK_SOURCE,
            "image": RECOVERY.ROLLBACK_IMAGE,
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


def runtime_evidence() -> dict:
    return {
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
            "expected_image": RECOVERY.RC09_IMAGE,
            "expected_source_revision": RECOVERY.RC09_SOURCE,
            "observed_image_reference": RECOVERY.RC09_IMAGE,
            "observed_image_id": "sha256:" + ("d" * 64),
            "oci": {
                "title": "GoreeCloud Search",
                "source": "https://github.com/GoreeCloud/goreecloud-search",
                "revision": RECOVERY.RC09_SOURCE,
                "version": RECOVERY.rc09_audit.RC09_OCI_VERSION,
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


def companion_args(root: Path) -> argparse.Namespace:
    release = root / "release.json"
    runtime = root / "runtime.json"
    output = root / "recovery.json"
    write_json(release, release_evidence())
    write_json(runtime, runtime_evidence())
    return argparse.Namespace(
        release_evidence=str(release),
        target_runtime_evidence=str(runtime),
        rollback_baseline=str(BASELINE),
        output=str(output),
        evidence=str(output),
    )


def complete(evidence: dict) -> dict:
    evidence["backup"].update(
        {
            "system": "operator-verified-backup",
            "snapshot_reference": "snapshot-rc09",
            "captured_at": NOW,
        }
    )
    for key in evidence["backup"]["scope"]:
        evidence["backup"]["scope"][key] = True
    evidence["restore"].update(
        {
            "isolated_location": "/srv/goreecloud-search/recovery/rc09",
            "stack_definition_restored": True,
            "search_settings_restored": True,
            "protected_runtime_configuration_recovery_verified": True,
            "caddy_search_route_copy_restored": True,
            "compose_validation_passed": True,
            "runtime_recreated": True,
            "runtime_health_passed": True,
            "runtime_identity_passed": True,
            "target_acceptance_passed": True,
        }
    )
    evidence["rollback"].update(
        {
            "mode": "equivalent-verified-evidence",
            "known_good_image_available": True,
            "previous_runtime_configuration_preserved": True,
            "previous_caddy_route_preserved": True,
            "rollback_procedure_documented": True,
            "isolated_rollback_rehearsal_passed": True,
            "equivalent_verified_rollback_evidence": True,
        }
    )
    evidence["monitoring"].update(
        {
            "monitor_identity": "GoreeCloud Search",
            "availability_monitor_verified": True,
            "alert_delivery_verified": True,
        }
    )
    evidence["scope"].update(
        {
            "application_level_restore_tested": True,
            "monitoring_and_alerting_verified": True,
            "rollback_evidence_verified": True,
        }
    )
    return evidence


class RecoveryEvidenceTests(unittest.TestCase):
    """Keep recovery evidence exact, sanitized, explicit, and non-authorizing."""

    def test_template_is_incomplete_and_non_authorizing(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = companion_args(Path(tempdir))
            evidence = RECOVERY.build_template(args)
            self.assertEqual(evidence["candidate"]["source_revision"], RECOVERY.RC09_SOURCE)
            self.assertEqual(evidence["candidate"]["image"], RECOVERY.RC09_IMAGE)
            self.assertFalse(evidence["restore"]["production_modified"])
            self.assertFalse(evidence["scope"]["production_cutover_authorized"])
            with self.assertRaises(RECOVERY.EvidenceError):
                RECOVERY.validate_evidence(args)

    def test_completed_evidence(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = companion_args(Path(tempdir))
            evidence = complete(RECOVERY.build_template(args))
            write_json(Path(args.evidence), evidence)
            validated = RECOVERY.validate_evidence(args)
            self.assertTrue(validated["monitoring"]["alert_delivery_verified"])
            self.assertFalse(validated["scope"]["production_cutover_authorized"])

    def test_wrong_candidate_source_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = companion_args(Path(tempdir))
            release = Path(args.release_evidence)
            value = json.loads(release.read_text(encoding="utf-8"))
            value["candidate"]["source_revision"] = "1" * 40
            write_json(release, value)
            with self.assertRaises(RECOVERY.EvidenceError):
                RECOVERY.build_template(args)

    def test_release_authorization_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = companion_args(Path(tempdir))
            release = Path(args.release_evidence)
            value = json.loads(release.read_text(encoding="utf-8"))
            value["authorization"]["production_cutover_authorized"] = True
            write_json(release, value)
            with self.assertRaises(RECOVERY.EvidenceError):
                RECOVERY.build_template(args)

    def test_production_modified_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = companion_args(Path(tempdir))
            evidence = complete(RECOVERY.build_template(args))
            evidence["restore"]["production_modified"] = True
            write_json(Path(args.evidence), evidence)
            with self.assertRaises(RECOVERY.EvidenceError):
                RECOVERY.validate_evidence(args)

    def test_production_cutover_authorization_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = companion_args(Path(tempdir))
            evidence = complete(RECOVERY.build_template(args))
            evidence["scope"]["production_cutover_authorized"] = True
            write_json(Path(args.evidence), evidence)
            with self.assertRaises(RECOVERY.EvidenceError):
                RECOVERY.validate_evidence(args)

    def test_binding_mismatch_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = companion_args(Path(tempdir))
            evidence = complete(RECOVERY.build_template(args))
            evidence["artifact_bindings"]["release_evidence_sha256"] = "0" * 64
            write_json(Path(args.evidence), evidence)
            with self.assertRaises(RECOVERY.EvidenceError):
                RECOVERY.validate_evidence(args)

    def test_sensitive_fields_rejected(self) -> None:
        for key in ("query", "response_content", "api_key", "credential"):
            with self.subTest(key=key), tempfile.TemporaryDirectory() as tempdir:
                args = companion_args(Path(tempdir))
                evidence = complete(RECOVERY.build_template(args))
                evidence["monitoring"][key] = "must-not-be-recorded"
                write_json(Path(args.evidence), evidence)
                with self.assertRaises(RECOVERY.EvidenceError):
                    RECOVERY.validate_evidence(args)

    def test_rollback_modes_are_unambiguous(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = companion_args(Path(tempdir))
            evidence = complete(RECOVERY.build_template(args))
            evidence["rollback"]["production_route_rollback_tested"] = True
            write_json(Path(args.evidence), evidence)
            with self.assertRaises(RECOVERY.EvidenceError):
                RECOVERY.validate_evidence(args)

    def test_template_contains_no_sensitive_fields(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            args = companion_args(Path(tempdir))
            serialized = json.dumps(RECOVERY.build_template(args), sort_keys=True)
            for key in (
                '"password"',
                '"token"',
                '"secret"',
                '"credential"',
                '"api_key"',
                '"query"',
                '"response_content"',
                '"environment_values"',
            ):
                self.assertNotIn(key, serialized)


if __name__ == "__main__":
    unittest.main()
