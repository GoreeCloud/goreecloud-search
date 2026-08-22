# SPDX-License-Identifier: AGPL-3.0-or-later
# pylint: disable=missing-class-docstring,invalid-name
"""Contracts for the manual GoreeCloud real-provider acceptance workflow."""

from pathlib import Path

from tests import SearxTestCase


WORKFLOW = Path(".github/workflows/goreecloud-provider-acceptance.yml")


class GoreeCloudProviderWorkflowTest(SearxTestCase):
    @classmethod
    def setUpClass(cls):
        super().setUpClass()
        cls.workflow = WORKFLOW.read_text(encoding="utf-8")

    def test_requires_exact_candidate_identity(self):
        self.assertIn("candidate_source:", self.workflow)
        self.assertIn("candidate_image:", self.workflow)
        self.assertIn("^[0-9a-f]{40}$", self.workflow)
        self.assertIn(
            "^ghcr\\.io/goreecloud/goreecloud-search@sha256:[0-9a-f]{64}$",
            self.workflow,
        )
        self.assertIn("ref: ${{ inputs.candidate_source }}", self.workflow)
        self.assertIn('test "$observed_source" = "$GC_CANDIDATE_SOURCE"', self.workflow)

    def test_stages_only_the_immutable_candidate_on_loopback(self):
        self.assertIn('docker pull "$GC_CANDIDATE_IMAGE"', self.workflow)
        self.assertIn("--publish 127.0.0.1:8888:8080", self.workflow)
        self.assertIn('"$GC_CANDIDATE_IMAGE"', self.workflow)
        self.assertIn("goreecloud-search-provider-evidence", self.workflow)
        self.assertNotIn("python -m searx.webapp", self.workflow)

    def test_suite_generates_candidate_bound_sanitized_evidence(self):
        for required_argument in (
            "--expected-source",
            "--expected-image",
            "--container",
            "--evidence-json",
        ):
            self.assertIn(required_argument, self.workflow)

        self.assertIn("provider-evidence.json", self.workflow)
        self.assertIn("verified_before_and_after_requests", self.workflow)
        self.assertIn("all_required_categories_passed", self.workflow)
        self.assertIn("query_text_persisted", self.workflow)
        self.assertIn("response_content_persisted", self.workflow)
        self.assertIn("production_cutover_authorized", self.workflow)
        self.assertIn("actions/upload-artifact@v4", self.workflow)
        self.assertIn("sha256sum --check SHA256SUMS", self.workflow)

    def test_untrusted_diagnostic_inputs_are_passed_through_environment(self):
        self.assertIn("GC_QUERY: ${{ inputs.query }}", self.workflow)
        self.assertIn("GC_CATEGORY: ${{ inputs.category }}", self.workflow)
        self.assertIn(' --query "$GC_QUERY"', self.workflow.replace("\\\n", " "))
        self.assertIn(' --category "$GC_CATEGORY"', self.workflow.replace("\\\n", " "))
