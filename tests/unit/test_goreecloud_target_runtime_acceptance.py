# SPDX-License-Identifier: AGPL-3.0-or-later
"""Regression tests for the read-only GoreeCloud target-runtime evidence harness."""

from __future__ import annotations

import importlib.util
import json
from pathlib import Path
import sys
import unittest


MODULE_PATH = Path(__file__).parents[2] / "goreecloud" / "target_runtime_acceptance.py"
MODULE_DIR = str(MODULE_PATH.parent)
SPEC = importlib.util.spec_from_file_location("target_runtime_acceptance", MODULE_PATH)
assert SPEC is not None and SPEC.loader is not None
ACCEPTANCE = importlib.util.module_from_spec(SPEC)
sys.path.insert(0, MODULE_DIR)
try:
    SPEC.loader.exec_module(ACCEPTANCE)
finally:
    sys.path.remove(MODULE_DIR)

IMAGE = "ghcr.io/goreecloud/goreecloud-search@sha256:" + ("1" * 64)
SOURCE = "2" * 40
VERSION = "2026.8.23-2a2a2a2"
BASE_URL = "http://127.0.0.1:8888"


def runtime() -> dict:
    return {
        "status": "verified",
        "running": True,
        "health": "healthy",
        "published_ports": "8080/tcp -> 127.0.0.1:8888",
        "identity_status": "verified",
        "expected_image": IMAGE,
        "expected_source_revision": SOURCE,
        "observed_image_reference": IMAGE,
        "observed_image_id": "sha256:" + ("3" * 64),
        "oci": {
            "title": "GoreeCloud Search",
            "source": ACCEPTANCE.REPOSITORY_URL,
            "revision": SOURCE,
            "version": VERSION,
            "licenses": "AGPL-3.0-or-later",
        },
    }


def http_acceptance() -> dict:
    return {
        "home_identity": "passed",
        "preferences_identity": "passed",
        "about_identity": "passed",
        "health": "passed",
        "privacy_headers": "passed",
    }


class TargetRuntimeTests(unittest.TestCase):
    """Keep target-runtime evidence loopback-only, immutable, sanitized, and non-authorizing."""

    def test_inputs(self) -> None:
        ACCEPTANCE.validate_inputs(BASE_URL, IMAGE, SOURCE, VERSION)

    def test_external_url(self) -> None:
        with self.assertRaises(ACCEPTANCE.AcceptanceError):
            ACCEPTANCE.validate_inputs("https://search.goreecloud.com", IMAGE, SOURCE, VERSION)

    def test_mutable_image(self) -> None:
        with self.assertRaises(ACCEPTANCE.AcceptanceError):
            ACCEPTANCE.validate_inputs(BASE_URL, "ghcr.io/goreecloud/goreecloud-search:latest", SOURCE, VERSION)

    def test_external_port(self) -> None:
        with self.assertRaises(ACCEPTANCE.AcceptanceError):
            ACCEPTANCE.validate_ports("8080/tcp -> 0.0.0.0:8888")

    def test_loopback_ports(self) -> None:
        value = "8080/tcp -> 127.0.0.1:8888\n8443/tcp -> [::1]:8889"
        self.assertEqual(ACCEPTANCE.validate_ports(value), value)

    def test_evidence(self) -> None:
        evidence = ACCEPTANCE.build_evidence(
            BASE_URL,
            "goreecloud-search-rc09",
            http_acceptance(),
            runtime(),
        )
        self.assertEqual(evidence["schema_version"], 1)
        self.assertEqual(evidence["product"], "GoreeCloud Search")
        self.assertEqual(evidence["providers"], "skipped")
        self.assertEqual(evidence["container_runtime"]["status"], "verified")
        self.assertIn("127.0.0.1:", evidence["container_runtime"]["published_ports"])
        self.assertFalse(evidence["scope"]["production_cutover_authorized"])

    def test_sanitized(self) -> None:
        evidence = ACCEPTANCE.build_evidence(
            BASE_URL,
            "goreecloud-search-rc09",
            http_acceptance(),
            runtime(),
        )
        serialized = json.dumps(evidence, sort_keys=True)
        forbidden = (
            '"password"',
            '"token"',
            '"secret"',
            '"credential"',
            '"query"',
            '"response_content"',
            '"environment_values"',
        )
        for key in forbidden:
            self.assertNotIn(key, serialized)


if __name__ == "__main__":
    unittest.main()
