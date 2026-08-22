# SPDX-License-Identifier: AGPL-3.0-or-later
"""Regression tests for frozen candidate #07 immutable artifact provenance."""

from __future__ import annotations

import importlib.util
from pathlib import Path
import sys
import tempfile
import unittest


MODULE_PATH = Path(__file__).parents[2] / "goreecloud" / "first_stable_candidate_07_provenance_audit.py"
MODULE_DIR = str(MODULE_PATH.parent)
SPEC = importlib.util.spec_from_file_location("first_stable_candidate_07_provenance_audit", MODULE_PATH)
assert SPEC is not None and SPEC.loader is not None
AUDIT = importlib.util.module_from_spec(SPEC)
sys.path.insert(0, MODULE_DIR)
try:
    SPEC.loader.exec_module(AUDIT)
finally:
    sys.path.remove(MODULE_DIR)

RELEASE_EVIDENCE = """{
  "candidate": {
    "image": "ghcr.io/goreecloud/goreecloud-search@sha256:3ce3a509675ee396cba33b77ca429aaeea1b1d995f42c62778f9d24de40a09d8",
    "isolated_runtime_acceptance": "passed",
    "oci_revision": "b355aafe769176acebfc938b15a6f7b5b9a2db87",
    "oci_version": "2026.8.19-b355aafe7",
    "registry_digest_pull_verified": true,
    "source_revision": "b355aafe769176acebfc938b15a6f7b5b9a2db87"
  },
  "generated_at": "2026-08-19T20:28:33.282138Z",
  "product": "GoreeCloud Search",
  "rollback_baseline": {
    "environment": "goreecloud-vps-01-production",
    "image": "ghcr.io/goreecloud/goreecloud-search@sha256:30ec99e3311fa9dcc934ac267aef123bb2541e0cd0165b360c4a7dc4fa29e3d5",
    "isolated_runtime_acceptance": "passed",
    "recorded_at": "2026-08-17T18:10:00-05:00",
    "registry_digest_pull_verified": true,
    "source_revision": "3584da535f7ed7c3b4b8dc73cf0424fb4bdf1949"
  },
  "rollback_scope": {
    "image_level_rehearsal": "passed",
    "production_cutover_authorized": false,
    "statement": "This artifact proves immutable image identity, registry retrieval, and isolated candidate-to-known-good-image rollback execution only. Target-host configuration, persistent data, backup restoration, private routing, monitoring, and live rollback remain separate acceptance requirements.",
    "target_environment_configuration_rollback_tested": false,
    "target_environment_data_restore_tested": false
  },
  "schema_version": 1
}
"""


class Candidate07ProvenanceAuditTests(unittest.TestCase):
    """Keep candidate publication provenance exact and fail closed."""

    def test_release_bytes_pass(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "release-evidence.json"
            path.write_text(RELEASE_EVIDENCE, encoding="utf-8")
            result = AUDIT._audit_release_artifact(path)  # pylint: disable=protected-access
        self.assertEqual(result["artifact_id"], AUDIT.FROZEN_RELEASE_ARTIFACT_ID)
        self.assertEqual(
            result["release_evidence_sha256"],
            AUDIT.FROZEN_RELEASE_EVIDENCE_SHA256,
        )

    def test_release_mutation_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "release-evidence.json"
            path.write_text(RELEASE_EVIDENCE + "\n", encoding="utf-8")
            with self.assertRaises(AUDIT.AuditError):
                AUDIT._audit_release_artifact(path)  # pylint: disable=protected-access

    def test_visual_artifact_passes(self) -> None:
        AUDIT._require_visual_artifact(  # pylint: disable=protected-access
            {
                "reference": AUDIT.FROZEN_VISUAL_ARTIFACT_NAME,
                "digest": AUDIT.FROZEN_VISUAL_ARTIFACT_DIGEST,
            }
        )

    def test_visual_reference_rejected(self) -> None:
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._require_visual_artifact(  # pylint: disable=protected-access
                {
                    "reference": "different-artifact",
                    "digest": AUDIT.FROZEN_VISUAL_ARTIFACT_DIGEST,
                }
            )

    def test_visual_digest_rejected(self) -> None:
        with self.assertRaises(AUDIT.AuditError):
            AUDIT._require_visual_artifact(  # pylint: disable=protected-access
                {
                    "reference": AUDIT.FROZEN_VISUAL_ARTIFACT_NAME,
                    "digest": "sha256:" + ("0" * 64),
                }
            )

    def test_published_ids_are_frozen(self) -> None:
        self.assertEqual(AUDIT.FROZEN_RELEASE_ARTIFACT_ID, 9382173615)
        self.assertEqual(AUDIT.FROZEN_VISUAL_ARTIFACT_ID, 9382309578)


if __name__ == "__main__":
    unittest.main()
