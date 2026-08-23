# SPDX-License-Identifier: AGPL-3.0-or-later
# pylint: disable=missing-class-docstring
"""Contract for the one-shot RC #09 publication receipt."""

from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[2]
WORKFLOW = ROOT / ".github/workflows/goreecloud-rc09-publication-receipt.yml"
SUPPLY_CHAIN = ROOT / ".github/workflows/goreecloud-workflow-supply-chain.yml"


class GoreeCloudRc09ReceiptTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.workflow = WORKFLOW.read_text(encoding="utf-8")
        cls.supply_chain = SUPPLY_CHAIN.read_text(encoding="utf-8")

    def test_exact_candidate_binding(self):
        source = "ef2a7fb4e96e1f28bef53fb8bc766a1bed96b45e"
        self.assertIn(f"test \"$parent\" = '{source}'", self.workflow)
        self.assertIn(f"rc-09-{source}", self.workflow)
        self.assertIn(f"test \"$revision\" = '{source}'", self.workflow)

    def test_read_only_registry_scope(self):
        permissions = "permissions:\n  contents: read\n  packages: read\n  issues: write"
        self.assertIn(permissions, self.workflow)
        self.assertNotIn("packages: write", self.workflow)
        self.assertIn("persist-credentials: false", self.workflow)

    def test_runtime_before_receipt(self):
        pull = self.workflow.index("Pull and verify immutable RC09 identity")
        runtime = self.workflow.index("Rehearse immutable RC09 runtime")
        receipt = self.workflow.index("Post connector-visible RC09 receipt")
        self.assertLess(pull, runtime)
        self.assertLess(runtime, receipt)
        self.assertIn("Production cutover authorized: **false**", self.workflow)
        self.assertIn("Stable promotion authorized: **false**", self.workflow)

    def test_supply_chain_exception(self):
        self.assertIn("'goreecloud-rc09-publication-receipt.yml': {", self.supply_chain)
        self.assertIn("'issues': 'write'", self.supply_chain)
        self.assertIn("'packages': 'read'", self.supply_chain)


if __name__ == "__main__":
    unittest.main()
