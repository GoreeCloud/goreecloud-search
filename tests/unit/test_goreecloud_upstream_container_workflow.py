# SPDX-License-Identifier: AGPL-3.0-or-later
# pylint: disable=missing-class-docstring,invalid-name
"""Regression contract for inherited upstream container publication."""

from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[2]
UPSTREAM_WORKFLOW = ROOT / ".github/workflows/container.yml"
GOREECLOUD_WORKFLOW = ROOT / ".github/workflows/goreecloud-container-build.yml"
CONTAINER_HELPER = ROOT / "utils/lib_sxng_container.sh"


class GoreeCloudUpstreamContainerWorkflowTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.upstream = UPSTREAM_WORKFLOW.read_text(encoding="utf-8")
        cls.goreecloud = GOREECLOUD_WORKFLOW.read_text(encoding="utf-8")
        cls.helper = CONTAINER_HELPER.read_text(encoding="utf-8")
        cls.build = cls.upstream.split("\n  build:\n", 1)[1].split("\n  test:\n", 1)[0]
        cls.release = cls.upstream.split("\n  release:\n", 1)[1]

    def test_inherited_registry_build_is_upstream_owner_gated(self):
        self.assertIn("github.repository_owner == 'searxng'", self.build)
        self.assertIn("github.event_name == 'workflow_dispatch'", self.build)
        self.assertIn("github.event.workflow_run.conclusion == 'success'", self.build)
        self.assertIn("packages: write", self.build)
        self.assertIn("docker/login-action@dbcb813823bdd20940b903addbd779551569679f", self.build)
        owner_check = self.build.index("github.repository_owner == 'searxng'")
        dispatch_check = self.build.index("github.event_name == 'workflow_dispatch'")
        self.assertLess(owner_check, dispatch_check)

    def test_manual_dispatch_no_longer_bypasses_owner_gate(self):
        self.assertNotIn(
            "github.event_name == 'workflow_dispatch'\n" "      || (github.repository_owner == 'searxng'",
            self.upstream,
        )

    def test_inherited_helper_pushes_cache_when_actions_mode_is_enabled(self):
        self.assertIn('if [ "$GITHUB_ACTIONS" = "true" ]; then', self.helper)
        self.assertIn(
            '"$container_engine" push "ghcr.io/$CONTAINER_IMAGE_ORGANIZATION/cache:$CONTAINER_IMAGE_NAME-$arch$variant"',
            self.helper,
        )

    def test_goreecloud_acceptance_build_remains_local_only(self):
        self.assertIn("permissions:\n  contents: read", self.goreecloud)
        self.assertIn("GITHUB_ACTIONS=false ./manage container.build podman", self.goreecloud)
        self.assertNotIn("packages: write", self.goreecloud)
        self.assertNotIn("docker/login-action@", self.goreecloud)

    def test_release_remains_upstream_master_only(self):
        self.assertIn("github.repository_owner == 'searxng'", self.release)
        self.assertIn("github.ref_name == 'master'", self.release)
        self.assertIn("run: make container.push", self.release)


if __name__ == "__main__":
    unittest.main()
