# SPDX-License-Identifier: AGPL-3.0-or-later
# pylint: disable=missing-class-docstring
"""Regression contract for GoreeCloud Search Release Candidate publication."""

import json
from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[2]
REQUEST = ROOT / "goreecloud/release_candidate_request.json"
WORKFLOW = ROOT / ".github/workflows/goreecloud-rc-publication.yml"
SUPPLY_CHAIN = ROOT / ".github/workflows/goreecloud-workflow-supply-chain.yml"
EVIDENCE_HELPER = ROOT / "goreecloud/release_evidence.py"


class GoreeCloudRcPublicationTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.request = json.loads(REQUEST.read_text(encoding="utf-8"))
        cls.workflow = WORKFLOW.read_text(encoding="utf-8")
        cls.supply_chain = SUPPLY_CHAIN.read_text(encoding="utf-8")
        cls.helper = EVIDENCE_HELPER.read_text(encoding="utf-8")

    def test_rc09_request(self):
        self.assertEqual(self.request["product"], "GoreeCloud Search")
        self.assertEqual(self.request["request_id"], "first-stable-2026-08-23-09")
        self.assertEqual(self.request["candidate_sequence"], 9)
        self.assertEqual(self.request["lifecycle"], "Release Candidate")
        self.assertEqual(
            self.request["reviewed_base_revision"],
            "61027b05ad8fef33d68078f6ec4498e1ac675112",
        )
        self.assertFalse(self.request["production_cutover_authorized"])
        self.assertFalse(self.request["stable_release_authorized"])
        self.assertFalse(self.request["target_host_change_authorized"])

    def test_master_parent_binding(self):
        self.assertIn("branches:\n      - master", self.workflow)
        self.assertNotIn("workflow_dispatch:", self.workflow)
        self.assertNotIn("pull_request:", self.workflow)
        self.assertIn("github.repository_owner == 'GoreeCloud'", self.workflow)
        self.assertIn("github.ref_name == 'master'", self.workflow)
        self.assertIn('git rev-parse "${GITHUB_SHA}^"', self.workflow)
        self.assertIn('test "$parent" = "$reviewed_base"', self.workflow)
        self.assertIn("Unapproved RC boundary file", self.workflow)

    def test_rehearsal_before_publish(self):
        local_rehearsal = self.workflow.index("Rehearse local candidate and known-good rollback")
        registry_login = self.workflow.index("Authenticate to GHCR")
        publication = self.workflow.index("Publish immutable Release Candidate")
        digest_rehearsal = self.workflow.index("Verify published digest and rehearse immutable candidate")
        self.assertLess(local_rehearsal, registry_login)
        self.assertLess(registry_login, publication)
        self.assertLess(publication, digest_rehearsal)
        self.assertIn("GITHUB_ACTIONS=false ./manage container.build podman", self.workflow)
        self.assertIn("ghcr.io/goreecloud/goreecloud-search:rc-", self.workflow)
        self.assertIn("candidate-image.txt", self.workflow)

    def test_guarded_registry_write(self):
        self.assertIn("permissions:\n  contents: read\n  packages: write", self.workflow)
        self.assertIn("'goreecloud-rc-publication.yml': {", self.supply_chain)
        self.assertIn("'packages': 'write'", self.supply_chain)
        self.assertIn("'goreecloud-rc-publication.yml',", self.supply_chain)
        self.assertIn("persist-credentials: false", self.workflow)

    def test_evidence_is_rc_only(self):
        self.assertIn('"lifecycle": "Release Candidate"', self.helper)
        self.assertIn('"production_cutover_authorized": False', self.helper)
        self.assertIn('"stable_release_authorized": False', self.helper)
        self.assertIn("target_environment_data_restore_tested", self.helper)
        self.assertIn("actual compiled GoreeCloud Browser acceptance", self.helper)


if __name__ == "__main__":
    unittest.main()
