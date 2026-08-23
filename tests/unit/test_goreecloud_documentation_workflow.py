# SPDX-License-Identifier: AGPL-3.0-or-later
# pylint: disable=missing-class-docstring,invalid-name
"""Contracts for GoreeCloud's inherited documentation workflow boundary."""

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
WORKFLOW = ROOT / ".github/workflows/documentation.yml"


class GoreeCloudDocumentationWorkflowTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.workflow = WORKFLOW.read_text(encoding="utf-8")
        cls.docs = cls.workflow.split("\n  docs:\n", 1)[1].split("\n  release:\n", 1)[0]
        cls.release = cls.workflow.split("\n  release:\n", 1)[1]

    def test_fork_build_is_read_only(self):
        self.assertIn("permissions:\n  contents: read", self.workflow)
        self.assertIn("name: Build", self.docs)
        self.assertIn("run: make V=1 docs.html", self.docs)
        self.assertNotIn("contents: write", self.docs)
        self.assertNotIn("github-pages-deploy-action", self.docs)

    def test_publication_is_upstream_owner_gated(self):
        self.assertIn("github.repository_owner == 'searxng'", self.release)
        self.assertIn("github.ref_name == 'master'", self.release)
        self.assertIn("contents: write", self.release)
        self.assertIn("github-pages-deploy-action@fa24774553152dd7873cd16ebd8d959b010c5445", self.release)
        self.assertNotIn("github.repository_owner == 'searxng' || github.event_name == 'workflow_dispatch'", self.workflow)

    def test_manual_dispatch_does_not_bypass_owner_gate(self):
        self.assertIn("workflow_dispatch:", self.workflow)
        self.assertIn("github.event_name == 'workflow_dispatch'", self.release)
        owner_check = self.release.index("github.repository_owner == 'searxng'")
        dispatch_check = self.release.index("github.event_name == 'workflow_dispatch'")
        self.assertLess(owner_check, dispatch_check)

    def test_checkout_credentials_are_not_persisted(self):
        self.assertEqual(self.workflow.count("persist-credentials: \"false\""), 2)
        self.assertEqual(
            self.workflow.count("actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1"),
            2,
        )


if __name__ == "__main__":
    unittest.main()
