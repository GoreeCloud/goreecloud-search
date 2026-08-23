# SPDX-License-Identifier: AGPL-3.0-or-later
"""GoreeCloud Search provider-acceptance contract tests."""

import json
import tempfile
from pathlib import Path
from unittest import TestCase, main

from goreecloud import provider_acceptance


class ProviderContractTestCase(TestCase):
    """Protect the first-Stable provider and evidence boundary."""

    def test_required_categories(self):
        self.assertEqual(
            provider_acceptance.RELEASE_REQUIRED_CATEGORIES,
            frozenset({"general", "images", "videos", "news", "files"}),
        )
        suite = {case.category for case in provider_acceptance.REPRESENTATIVE_SUITE}
        self.assertTrue(provider_acceptance.RELEASE_REQUIRED_CATEGORIES <= suite)
        provider_acceptance.validate_representative_suite()

    def test_sanitized_evidence(self):
        source = "a" * 40
        image = "ghcr.io/goreecloud/goreecloud-search@sha256:" + "b" * 64
        runtime = provider_acceptance.RuntimeIdentity(
            container="goreecloud-search",
            base_url="http://127.0.0.1:8888",
            published_port="127.0.0.1:8888",
            image_reference=image,
            image_id="sha256:" + "c" * 64,
            oci_title="GoreeCloud Search",
            oci_source="https://github.com/GoreeCloud/goreecloud-search",
            oci_revision=source,
            oci_version="test",
            oci_licenses="AGPL-3.0-or-later",
        )
        results = [
            provider_acceptance.AcceptanceResult(
                category=category,
                exit_code=0,
                http_status=200,
                product_identity=True,
                result_cards=1,
                engine_messages=0,
                passed=True,
            )
            for category in sorted(provider_acceptance.RELEASE_REQUIRED_CATEGORIES)
        ]

        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "provider-evidence.json"
            provider_acceptance.write_evidence(
                str(path),
                source,
                image,
                1,
                0,
                results,
                runtime,
            )
            evidence = json.loads(path.read_text(encoding="utf-8"))

        self.assertEqual(
            set(evidence["required_categories"]),
            provider_acceptance.RELEASE_REQUIRED_CATEGORIES,
        )
        self.assertTrue(evidence["runtime_binding"]["verified_before_and_after_requests"])
        self.assertTrue(evidence["scope"]["all_required_categories_passed"])
        self.assertFalse(evidence["scope"]["query_text_persisted"])
        self.assertFalse(evidence["scope"]["response_content_persisted"])
        self.assertFalse(evidence["scope"]["production_cutover_authorized"])
        self.assertNotIn("query", evidence)
        self.assertNotIn("response_content", evidence)

    def test_workflow_cli_binding(self):
        workflow = Path(".github/workflows/goreecloud-provider-acceptance.yml").read_text(
            encoding="utf-8"
        )
        for argument in (
            "--expected-source",
            "--expected-image",
            "--container",
            "--evidence-json",
        ):
            self.assertIn(argument, workflow)
        self.assertIn("runtime_identity_verified_during_provider_requests", workflow)
        self.assertIn("all_required_categories_passed", workflow)


if __name__ == "__main__":
    main()
