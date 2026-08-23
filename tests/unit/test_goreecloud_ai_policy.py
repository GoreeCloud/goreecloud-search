# SPDX-License-Identifier: AGPL-3.0-or-later
# pylint: disable=missing-class-docstring,invalid-name
"""Regression contract for GoreeCloud AI contribution governance."""

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
POLICY = ROOT / "AI_POLICY.rst"
CONTRIBUTING = ROOT / "CONTRIBUTING.rst"
WORKFLOW = ROOT / ".github/workflows/goreecloud-ai-policy.yml"
LEGACY_WORKFLOW = ROOT / ".github/workflows/ai-policy.yml"
LEGACY_SCRIPT = ROOT / ".github/scripts/ai_policy.cjs"
SUPPLY_CHAIN = ROOT / ".github/workflows/goreecloud-workflow-supply-chain.yml"


class GoreeCloudAIPolicyContractTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.policy = POLICY.read_text(encoding="utf-8")
        cls.contributing = CONTRIBUTING.read_text(encoding="utf-8")
        cls.workflow = WORKFLOW.read_text(encoding="utf-8")
        cls.supply_chain = SUPPLY_CHAIN.read_text(encoding="utf-8")

    def test_policy_allows_agent_authorship(self):
        self.assertIn("AI-assisted and agent-authored contributions are permitted", self.policy)
        self.assertIn("AI agents may be the primary implementation author", self.policy)
        self.assertIn("Quality, evidence,", self.policy)
        self.assertIn("Upstream SearXNG Boundary", self.policy)
        self.assertNotIn("AI should never be the main author", self.policy)
        self.assertNotIn("Issues and PR descriptions must be fully human-written", self.policy)
        self.assertNotIn("invalid:slop", self.policy)

    def test_contributing_uses_goreecloud_ai_policy(self):
        self.assertIn("``AI_POLICY.rst``", self.contributing)
        self.assertIn(
            "AI-assisted and agent-authored contributions are permitted",
            self.contributing,
        )
        self.assertIn("does not impose a human-written-majority requirement", self.contributing)
        self.assertIn(
            "upstream SearXNG project controls its own contribution and AI policies",
            self.contributing,
        )

    def test_destructive_upstream_enforcement_is_removed(self):
        self.assertFalse(LEGACY_WORKFLOW.exists())
        self.assertFalse(LEGACY_SCRIPT.exists())
        self.assertNotIn("pull_request_target", self.workflow)
        self.assertNotIn("issues: write", self.workflow)
        self.assertNotIn("pull-requests: write", self.workflow)
        self.assertNotIn("invalid:slop", self.workflow)
        self.assertNotIn("github.rest.issues.update", self.workflow)
        self.assertIn("permissions:\n  contents: read", self.workflow)

    def test_policy_workflow_is_under_supply_chain_governance(self):
        self.assertIn("'goreecloud-ai-policy.yml'", self.supply_chain)
        self.assertIn(
            "actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1",
            self.workflow,
        )
        self.assertIn("persist-credentials: false", self.workflow)
        self.assertIn(
            "actions/setup-python@5fda3b95a4ea91299a34e894583c3862153e4b97",
            self.workflow,
        )


if __name__ == "__main__":
    unittest.main()
