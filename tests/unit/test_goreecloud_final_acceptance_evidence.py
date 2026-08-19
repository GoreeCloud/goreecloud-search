# SPDX-License-Identifier: AGPL-3.0-or-later
"""Deterministic contract for GoreeCloud Search final-candidate acceptance evidence."""

from __future__ import annotations

import json
from pathlib import Path
import subprocess
import sys
import tempfile


ROOT = Path(__file__).resolve().parents[2]
SOURCE = "a" * 40
IMAGE = "ghcr.io/goreecloud/goreecloud-search@sha256:" + "b" * 64
BROWSER_SOURCE = "c" * 40
REQUIRED = ["files", "general", "images", "news", "videos"]


def _write(path: Path, value: dict) -> None:
    path.write_text(json.dumps(value, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def _artifacts(directory: Path) -> dict[str, Path]:
    release = directory / "release.json"
    runtime = directory / "runtime.json"
    recovery = directory / "recovery.json"
    provider = directory / "provider.json"

    _write(
        release,
        {
            "schema_version": 1,
            "product": "GoreeCloud Search",
            "candidate": {
                "source_revision": SOURCE,
                "image": IMAGE,
                "isolated_runtime_acceptance": "passed",
            },
            "rollback_scope": {
                "image_level_rehearsal": "passed",
                "production_cutover_authorized": False,
            },
        },
    )
    _write(
        runtime,
        {
            "schema_version": 1,
            "product": "GoreeCloud Search",
            "container_runtime": {
                "expected_source_revision": SOURCE,
                "expected_image": IMAGE,
            },
            "scope": {
                "target_runtime_identity_verified": True,
                "production_cutover_authorized": False,
            },
        },
    )
    _write(
        recovery,
        {
            "schema_version": 1,
            "product": "GoreeCloud Search",
            "candidate": {"source_revision": SOURCE, "image": IMAGE},
            "scope": {
                "application_level_restore_tested": True,
                "monitoring_and_alerting_verified": True,
                "rollback_evidence_verified": True,
                "production_cutover_authorized": False,
            },
        },
    )
    _write(
        provider,
        {
            "schema_version": 1,
            "product": "GoreeCloud Search",
            "candidate": {"source_revision": SOURCE, "image": IMAGE},
            "runtime_binding": {
                "verified_before_and_after_requests": True,
                "base_url": "http://127.0.0.1:8888",
                "container": "goreecloud-search",
                "published_port": "8080/tcp -> 127.0.0.1:8888",
                "observed_image_reference": IMAGE,
                "observed_image_id": "sha256:" + "d" * 64,
                "oci_revision": SOURCE,
            },
            "required_categories": REQUIRED,
            "results": [
                {
                    "category": category,
                    "exit_code": 0,
                    "http_status": 200,
                    "product_identity": True,
                    "result_cards": 1,
                    "engine_messages": 0,
                    "passed": True,
                }
                for category in REQUIRED
            ],
            "scope": {
                "real_provider_requests_performed": True,
                "runtime_identity_verified_during_provider_requests": True,
                "all_required_categories_passed": True,
                "full_diagnostic_suite_passed": True,
                "query_text_persisted": False,
                "response_content_persisted": False,
                "production_cutover_authorized": False,
            },
        },
    )
    return {
        "release": release,
        "runtime": runtime,
        "recovery": recovery,
        "provider": provider,
    }


def _command(artifacts: dict[str, Path], *args: str) -> list[str]:
    return [
        sys.executable,
        str(ROOT / "goreecloud/final_acceptance_evidence.py"),
        *args,
        "--release-evidence",
        str(artifacts["release"]),
        "--target-runtime-evidence",
        str(artifacts["runtime"]),
        "--recovery-evidence",
        str(artifacts["recovery"]),
        "--provider-evidence",
        str(artifacts["provider"]),
    ]


def test_final_acceptance_template_and_validation() -> None:
    with tempfile.TemporaryDirectory() as temporary:
        directory = Path(temporary)
        artifacts = _artifacts(directory)
        final = directory / "final.json"

        subprocess.run(
            _command(artifacts, "template", "--output", str(final)),
            check=True,
            cwd=ROOT,
            capture_output=True,
            text=True,
        )
        evidence = json.loads(final.read_text(encoding="utf-8"))
        assert evidence["candidate"]["source_revision"] == SOURCE
        assert evidence["candidate"]["image"] == IMAGE
        assert evidence["scope"]["final_candidate_acceptance_complete"] is False
        assert evidence["scope"]["production_cutover_authorized"] is False

        visual = evidence["visual_acceptance"]
        for case_name in ("compact_light", "compact_dark", "expanded_light", "expanded_dark"):
            visual[case_name]["passed"] = True
            visual[case_name]["evidence_reference"] = f"acceptance/{case_name}"
        visual["physical_android_preferences_review"] = True
        visual["desktop_regression_review"] = True

        browser = evidence["browser_integration"]
        browser["browser_source_revision"] = BROWSER_SOURCE
        browser["evidence_reference"] = "browser/runtime-acceptance"
        for key in (
            "search_only_default_provider",
            "address_bar_routed_through_search",
            "new_tab_routed_through_search",
            "dedicated_search_field_routed_through_search",
            "no_external_browser_fallback",
            "search_unavailability_state_verified",
            "recovery_after_search_reachability_verified",
        ):
            browser[key] = True

        scope = evidence["scope"]
        scope["glaze_ui_1_1_final_visual_acceptance_verified"] = True
        scope["browser_runtime_integration_verified"] = True
        scope["final_candidate_acceptance_complete"] = True
        _write(final, evidence)

        completed = subprocess.run(
            _command(artifacts, "validate", "--evidence", str(final)),
            check=False,
            cwd=ROOT,
            capture_output=True,
            text=True,
        )
        assert completed.returncode == 0, completed.stderr
        assert "final-candidate evidence passed" in completed.stdout

        evidence["scope"]["production_cutover_authorized"] = True
        _write(final, evidence)
        unsafe = subprocess.run(
            _command(artifacts, "validate", "--evidence", str(final)),
            check=False,
            cwd=ROOT,
            capture_output=True,
            text=True,
        )
        assert unsafe.returncode != 0
        assert "production_cutover_authorized must remain false" in unsafe.stderr


def test_final_acceptance_rejects_provider_evidence_without_runtime_binding() -> None:
    with tempfile.TemporaryDirectory() as temporary:
        directory = Path(temporary)
        artifacts = _artifacts(directory)
        provider = json.loads(artifacts["provider"].read_text(encoding="utf-8"))
        provider.pop("runtime_binding")
        _write(artifacts["provider"], provider)

        final = directory / "final.json"
        rejected = subprocess.run(
            _command(artifacts, "template", "--output", str(final)),
            check=False,
            cwd=ROOT,
            capture_output=True,
            text=True,
        )
        assert rejected.returncode != 0
        assert "provider runtime_binding must be a JSON object" in rejected.stderr


def test_provider_runner_advertises_runtime_bound_evidence_without_network_access() -> None:
    result = subprocess.run(
        [sys.executable, str(ROOT / "goreecloud/provider_acceptance.py"), "--help"],
        check=True,
        cwd=ROOT,
        capture_output=True,
        text=True,
    )
    assert "--evidence-json" in result.stdout
    assert "--expected-source" in result.stdout
    assert "--expected-image" in result.stdout
    assert "--container" in result.stdout

    source = (ROOT / "goreecloud/provider_acceptance.py").read_text(encoding="utf-8")
    assert '"verified_before_and_after_requests": True' in source
    assert '"runtime_identity_verified_during_provider_requests": True' in source
    assert '"query_text_persisted": False' in source
    assert '"response_content_persisted": False' in source
    assert '"production_cutover_authorized": False' in source


def run_contract_checks() -> None:
    test_final_acceptance_template_and_validation()
    test_final_acceptance_rejects_provider_evidence_without_runtime_binding()
    test_provider_runner_advertises_runtime_bound_evidence_without_network_access()


if __name__ == "__main__":
    run_contract_checks()
    print("GoreeCloud Search final-candidate evidence contract passed.")
