from pathlib import Path
import unittest


REPO_ROOT = Path(__file__).resolve().parents[1]
WORKFLOW = REPO_ROOT / ".github" / "workflows" / "publish-container.yml"


class ReleasePublicationWorkflowTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.text = WORKFLOW.read_text(encoding="utf-8")

    def test_publication_is_manual_only(self):
        self.assertIn("workflow_dispatch:", self.text)
        self.assertNotIn("\npush:", self.text)
        self.assertNotIn("\npull_request:", self.text)
        self.assertNotIn("\nschedule:", self.text)

    def test_publication_fails_closed_to_main_exact_sha_and_version(self):
        self.assertIn('GITHUB_REF}" != "refs/heads/main"', self.text)
        self.assertIn('GITHUB_SHA}" != "${REQUESTED_SHA}', self.text)
        self.assertIn("REQUESTED_VERSION", self.text)
        self.assertIn("Development/RC identities only", self.text)

    def test_explicit_confirmation_and_no_latest_tag(self):
        self.assertIn(
            "PUBLISH GOREECLOUD SEARCH DEVELOPMENT CONTAINER",
            self.text,
        )
        self.assertNotIn(":latest", self.text)
        self.assertNotIn(" latest", self.text.casefold())

    def test_ghcr_permissions_are_bounded(self):
        self.assertIn("contents: read", self.text)
        self.assertIn("packages: write", self.text)
        self.assertNotIn("contents: write", self.text)
        self.assertNotIn("actions: write", self.text)

    def test_checkout_action_is_commit_pinned(self):
        self.assertIn(
            "actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683",
            self.text,
        )
        self.assertNotIn("actions/checkout@v", self.text)

    def test_build_targets_verified_vps_architecture_and_attestations(self):
        self.assertIn("--platform linux/amd64", self.text)
        self.assertIn("--provenance=mode=max", self.text)
        self.assertIn("--sbom=true", self.text)
        self.assertIn("--metadata-file", self.text)
        self.assertIn("containerimage.digest", self.text)

    def test_image_identity_is_canonical_and_immutable(self):
        self.assertIn(
            'IMAGE="ghcr.io/goreecloud/goreecloud-search"',
            self.text,
        )
        self.assertIn("immutable_ref=", self.text)
        self.assertIn("docker buildx imagetools inspect", self.text)
        self.assertIn("org.opencontainers.image.revision", self.text)
        self.assertIn("org.opencontainers.image.version", self.text)

    def test_workflow_does_not_deploy_or_use_provider_secret(self):
        lower = self.text.casefold()
        self.assertNotIn("docker compose up", lower)
        self.assertNotIn("ssh ", lower)
        self.assertNotIn("brave_search_api_key", lower)
        self.assertNotIn("search.goreecloud.com", lower)
        self.assertIn("does not deploy the vps", lower)


if __name__ == "__main__":
    unittest.main()
