# SPDX-License-Identifier: AGPL-3.0-or-later
# pylint: disable=missing-class-docstring
"""Contract for the one-shot RC #08 publication receipt."""

from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[2]
WORKFLOW = ROOT / ".github/workflows/goreecloud-rc08-publication-receipt.yml"
SUPPLY_CHAIN = ROOT / ".github/workflows/goreecloud-workflow-supply-chain.yml"


class GoreeCloudRc08ReceiptTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.workflow = WORKFLOW.read_text(encoding="utf-8")
        cls.supply_chain = SUPPLY_CHAIN.read_text(encoding="utf-8")

    def test_exact_candidate_binding(self):
        source = "2e04d650aac7a457c594f64ba704218c168c71a8"
        self.assertIn(f"test \"$parent\" = '{source}'", self.workflow)
        self.assertIn(f"rc-08-{source}", self.workflow)
        self.assertIn(f"test \"$revision\" = '{source}'", self.workflow)

    def test_read_only_registry_scope(self):
        permissions = "permissions:\n  contents: read\n  packages: read\n  issues: write"
        self.assertIn(permissions, self.workflow)
        self.assertNotIn("packages: write", self.workflow)
        self.assertIn("persist-credentials: false", self.workflow)

    def test_runtime_before_receipt(self):
        pull = self.workflow.index("Pull and verify immutable RC08 identity")
        runtime = self.workflow.index("Rehearse immutable RC08 runtime")
        receipt = self.workflow.index("Post connector-visible RC08 receipt")
        self.assertLess(pull, runtime)
        self.assertLess(runtime, receipt)
        self.assertIn("Production cutover authorized: **false**", self.workflow)
        self.assertIn("Stable promotion authorized: **false**", self.workflow)

    def test_supply_chain_exception(self):
        self.assertIn("'goreecloud-rc08-publication-receipt.yml': {", self.supply_chain)
        self.assertIn("'issues': 'write'", self.supply_chain)
        self.assertIn("'packages': 'read'", self.supply_chain)


if __name__ == "__main__":
    unittest.main()
