# SPDX-License-Identifier: AGPL-3.0-or-later
# pylint: disable=missing-class-docstring
"""Contract for the read-only GoreeCloud Search RC #08 registry audit."""

from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[2]
WORKFLOW = ROOT / ".github/workflows/goreecloud-rc08-registry-audit.yml"
SUPPLY_CHAIN = ROOT / ".github/workflows/goreecloud-workflow-supply-chain.yml"


class GoreeCloudRc08RegistryAuditTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.workflow = WORKFLOW.read_text(encoding="utf-8")
        cls.supply_chain = SUPPLY_CHAIN.read_text(encoding="utf-8")

    def test_exact_binding(self):
        source = "2e04d650aac7a457c594f64ba704218c168c71a8"
        self.assertIn(f"source_sha='{source}'", self.workflow)
        self.assertIn("rc-08-${source_sha}", self.workflow)
        self.assertIn('test "$revision" = "$source_sha"', self.workflow)

    def test_read_only_scope(self):
        expected = "permissions:\n  contents: read\n  packages: read"
        self.assertIn(expected, self.workflow)
        self.assertNotIn("packages: write", self.workflow)
        self.assertNotIn("issues: write", self.workflow)
        self.assertIn("persist-credentials: false", self.workflow)

    def test_runtime_audit(self):
        pull = self.workflow.index("Pull and verify immutable RC08 identity")
        runtime = self.workflow.index("Rehearse immutable RC08 runtime")
        self.assertLess(pull, runtime)
        self.assertIn("RC08 immutable image:", self.workflow)
        self.assertIn("RC08 registry audit: passed", self.workflow)

    def test_guard_exception(self):
        self.assertIn("'goreecloud-rc08-registry-audit.yml': {", self.supply_chain)
        self.assertIn("'packages': 'read'", self.supply_chain)
        self.assertIn("'goreecloud-rc08-registry-audit.yml',", self.supply_chain)


if __name__ == "__main__":
    unittest.main()
