# SPDX-License-Identifier: AGPL-3.0-or-later
"""Regression contract for retired historical SearXNG RC workflows."""

import json
from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[2]
WORKFLOW_ROOT = ROOT / ".github/workflows"
SUPPLY_CHAIN = WORKFLOW_ROOT / "goreecloud-workflow-supply-chain.yml"
REQUEST = ROOT / "goreecloud/release_candidate_request.json"

RETIRED_WORKFLOWS = (
    "goreecloud-rc-publication.yml",
    "goreecloud-rc08-publication-receipt.yml",
    "goreecloud-rc09-publication-receipt.yml",
)


class GoreeCloudHistoricalRcWorkflowRetirementTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.supply_chain = SUPPLY_CHAIN.read_text(encoding="utf-8")
        cls.request = json.loads(REQUEST.read_text(encoding="utf-8"))

    def test_historical_rc_workflows_are_not_executable_on_native_line(self):
        for name in RETIRED_WORKFLOWS:
            self.assertFalse(
                (WORKFLOW_ROOT / name).exists(),
                f"historical RC workflow must remain retired on native line: {name}",
            )

    def test_supply_chain_guard_blocks_reintroduction(self):
        self.assertIn("retired_workflows = {", self.supply_chain)
        for name in RETIRED_WORKFLOWS:
            self.assertIn(f"'{name}'", self.supply_chain)
        self.assertIn("must remain retired on the", self.supply_chain)
        self.assertNotIn("'packages': 'write'", self.supply_chain)
        self.assertNotIn("'issues': 'write'", self.supply_chain)

    def test_historical_release_request_record_is_preserved(self):
        self.assertEqual(self.request["product"], "GoreeCloud Search")
        self.assertEqual(self.request["request_id"], "first-stable-2026-08-23-09")
        self.assertEqual(self.request["candidate_sequence"], 9)
        self.assertEqual(self.request["lifecycle"], "Release Candidate")
        self.assertFalse(self.request["production_cutover_authorized"])
        self.assertFalse(self.request["stable_release_authorized"])


if __name__ == "__main__":
    unittest.main()
