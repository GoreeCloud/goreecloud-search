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
VISUAL_ARTIFACT_DIGEST = "sha256:" + "e" * 64
BROWSER_ARTIFACT_DIGEST = "sha256:" + "f" * 64


def _write(path: Path, value: dict) -> None:
    path.write_text(json.dumps(value, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def _review(reference: str) -> dict:
    return {"passed": True, "evidence_reference": reference}


def _artifacts(directory: Path) -> dict[str, Path]:
    release = directory / "release.json"
    runtime = directory / "runtime.json"
    recovery = directory / "recovery.json"
    provider = directory / "provider.json"
    visual = directory / "visual.json"
    browser = directory / "browser.json"

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
    _write(
        visual,
        {
            "schema_version": 1,
            "product": "GoreeCloud Search",
            "generated_at": "2026-08-19T20:00:00Z",
            "candidate": {"source_revision": SOURCE, "image": IMAGE},
            "glaze_ui_version": "1.1.0",
            "review_artifact": {
                "reference": "github-actions/candidate-visual-evidence",
                "digest": VISUAL_ARTIFACT_DIGEST,
            },
            "reviews": {
                "compact_light": _review("visual/compact-light"),
                "compact_dark": _review("visual/compact-dark"),
                "expanded_light": _review("visual/expanded-light"),
                "expanded_dark": _review("visual/expanded-dark"),
            },
            "physical_android_preferences_review": _review("device/android-preferences"),
            "desktop_regression_review": _review("device/desktop-regression"),
            "persisted_theme_preference_review": _review("device/theme-persistence"),
            "scope": {
                "exact_candidate_visual_artifact_verified": True,
                "manual_visual_acceptance_verified": True,
                "production_cutover_authorized": False,
                "statement": "Manual visual/device review completed against the exact immutable candidate.",
            },
        },
    )
    _write(
        browser,
        {
            "schema_version": 1,
            "product": "GoreeCloud Search",
            "generated_at": "2026-08-19T20:05:00Z",
            "search_candidate": {"source_revision": SOURCE, "image": IMAGE},
            "browser_source_revision": BROWSER_SOURCE,
            "runtime_artifact": {
                "reference": "github-actions/browser-runtime-acceptance",
                "digest": BROWSER_ARTIFACT_DIGEST,
            },
            "behaviors": {
                "search_only_default_provider": True,
                "address_bar_routed_through_search": True,
                "new_tab_routed_through_search": True,
                "dedicated_search_field_routed_through_search": True,
                "no_external_browser_fallback": True,
                "search_unavailability_state_verified": True,
                "recovery_after_search_reachability_verified": True,
            },
            "scope": {
                "actual_browser_runtime_verified": True,
                "search_candidate_runtime_verified": True,
                "production_cutover_authorized": False,
                "statement": "Actual GoreeCloud Browser runtime acceptance completed against this Search candidate.",
            },
        },
    )
    return {
        "release": release,
        "runtime": runtime,
        "recovery": recovery,
        "provider": provider,
        "visual": visual,
        "browser": browser,
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
        "--visual-evidence",
        str(artifacts["visual"]),
        "--browser-evidence",
        str(artifacts["browser"]),
    ]


def _assemble(directory: Path, artifacts: dict[str, Path]) -> Path:
    final = directory / "final.json"
    assembled = subprocess.run(
        _command(artifacts, "assemble", "--output", str(final)),
        check=False,
        cwd=ROOT,
        capture_output=True,
        text=True,
    )
    assert assembled.returncode == 0, assembled.stderr
    return final


def test_final_acceptance_assembles_and_validates_six_bound_artifacts() -> None:
    with tempfile.TemporaryDirectory() as temporary:
        directory = Path(temporary)
        artifacts = _artifacts(directory)
        final = _assemble(directory, artifacts)

        evidence = json.loads(final.read_text(encoding="utf-8"))
        assert evidence["schema_version"] == 2
        assert evidence["candidate"]["source_revision"] == SOURCE
        assert evidence["candidate"]["image"] == IMAGE
        assert evidence["scope"]["final_candidate_acceptance_complete"] is True
        assert evidence["scope"]["production_cutover_authorized"] is False
        assert set(evidence["artifact_bindings"]) == {
            "release_evidence_sha256",
            "target_runtime_evidence_sha256",
            "recovery_evidence_sha256",
            "provider_evidence_sha256",
            "visual_evidence_sha256",
            "browser_evidence_sha256",
        }
        assert evidence["visual_acceptance"]["review_artifact"]["digest"] == VISUAL_ARTIFACT_DIGEST
        assert evidence["browser_integration"]["runtime_artifact"]["digest"] == BROWSER_ARTIFACT_DIGEST

        completed = subprocess.run(
            _command(artifacts, "validate", "--evidence", str(final)),
            check=False,
            cwd=ROOT,
            capture_output=True,
            text=True,
        )
        assert completed.returncode == 0, completed.stderr
        assert "final-candidate evidence passed" in completed.stdout


def test_final_acceptance_rejects_mutated_bound_visual_or_browser_evidence() -> None:
    with tempfile.TemporaryDirectory() as temporary:
        directory = Path(temporary)
        artifacts = _artifacts(directory)
        final = _assemble(directory, artifacts)

        visual = json.loads(artifacts["visual"].read_text(encoding="utf-8"))
        visual["reviews"]["compact_light"]["evidence_reference"] = "visual/changed-after-assembly"
        _write(artifacts["visual"], visual)
        rejected_visual = subprocess.run(
            _command(artifacts, "validate", "--evidence", str(final)),
            check=False,
            cwd=ROOT,
            capture_output=True,
            text=True,
        )
        assert rejected_visual.returncode != 0
        assert "visual_evidence_sha256 does not match" in rejected_visual.stderr

        artifacts = _artifacts(directory)
        final = _assemble(directory, artifacts)
        browser = json.loads(artifacts["browser"].read_text(encoding="utf-8"))
        browser["runtime_artifact"]["reference"] = "browser/changed-after-assembly"
        _write(artifacts["browser"], browser)
        rejected_browser = subprocess.run(
            _command(artifacts, "validate", "--evidence", str(final)),
            check=False,
            cwd=ROOT,
            capture_output=True,
            text=True,
        )
        assert rejected_browser.returncode != 0
        assert "browser_evidence_sha256 does not match" in rejected_browser.stderr


def test_final_acceptance_rejects_incomplete_or_wrong_candidate_visual_evidence() -> None:
    with tempfile.TemporaryDirectory() as temporary:
        directory = Path(temporary)
        artifacts = _artifacts(directory)
        visual = json.loads(artifacts["visual"].read_text(encoding="utf-8"))
        visual["physical_android_preferences_review"]["passed"] = False
        _write(artifacts["visual"], visual)

        rejected = subprocess.run(
            _command(artifacts, "assemble", "--output", str(directory / "final.json")),
            check=False,
            cwd=ROOT,
            capture_output=True,
            text=True,
        )
        assert rejected.returncode != 0
        assert "physical_android_preferences_review.passed must be true" in rejected.stderr

        artifacts = _artifacts(directory)
        visual = json.loads(artifacts["visual"].read_text(encoding="utf-8"))
        visual["candidate"]["source_revision"] = "9" * 40
        _write(artifacts["visual"], visual)
        wrong_candidate = subprocess.run(
            _command(artifacts, "assemble", "--output", str(directory / "wrong.json")),
            check=False,
            cwd=ROOT,
            capture_output=True,
            text=True,
        )
        assert wrong_candidate.returncode != 0
        assert "Visual evidence refers to a different source revision" in wrong_candidate.stderr


def test_final_acceptance_rejects_incomplete_browser_runtime_evidence() -> None:
    with tempfile.TemporaryDirectory() as temporary:
        directory = Path(temporary)
        artifacts = _artifacts(directory)
        browser = json.loads(artifacts["browser"].read_text(encoding="utf-8"))
        browser["behaviors"]["no_external_browser_fallback"] = False
        _write(artifacts["browser"], browser)

        rejected = subprocess.run(
            _command(artifacts, "assemble", "--output", str(directory / "final.json")),
            check=False,
            cwd=ROOT,
            capture_output=True,
            text=True,
        )
        assert rejected.returncode != 0
        assert "no_external_browser_fallback must be true" in rejected.stderr


def test_final_acceptance_rejects_provider_evidence_without_runtime_binding() -> None:
    with tempfile.TemporaryDirectory() as temporary:
        directory = Path(temporary)
        artifacts = _artifacts(directory)
        provider = json.loads(artifacts["provider"].read_text(encoding="utf-8"))
        provider.pop("runtime_binding")
        _write(artifacts["provider"], provider)

        rejected = subprocess.run(
            _command(artifacts, "assemble", "--output", str(directory / "final.json")),
            check=False,
            cwd=ROOT,
            capture_output=True,
            text=True,
        )
        assert rejected.returncode != 0
        assert "provider runtime_binding must be a JSON object" in rejected.stderr


def test_final_acceptance_rejects_old_unbound_schema_and_cutover_authorization() -> None:
    with tempfile.TemporaryDirectory() as temporary:
        directory = Path(temporary)
        artifacts = _artifacts(directory)
        final = _assemble(directory, artifacts)
        evidence = json.loads(final.read_text(encoding="utf-8"))

        evidence["schema_version"] = 1
        _write(final, evidence)
        old_schema = subprocess.run(
            _command(artifacts, "validate", "--evidence", str(final)),
            check=False,
            cwd=ROOT,
            capture_output=True,
            text=True,
        )
        assert old_schema.returncode != 0
        assert "schema-version 2" in old_schema.stderr

        final = _assemble(directory, artifacts)
        evidence = json.loads(final.read_text(encoding="utf-8"))
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
    test_final_acceptance_assembles_and_validates_six_bound_artifacts()
    test_final_acceptance_rejects_mutated_bound_visual_or_browser_evidence()
    test_final_acceptance_rejects_incomplete_or_wrong_candidate_visual_evidence()
    test_final_acceptance_rejects_incomplete_browser_runtime_evidence()
    test_final_acceptance_rejects_provider_evidence_without_runtime_binding()
    test_final_acceptance_rejects_old_unbound_schema_and_cutover_authorization()
    test_provider_runner_advertises_runtime_bound_evidence_without_network_access()


if __name__ == "__main__":
    run_contract_checks()
    print("GoreeCloud Search final-candidate evidence contract passed.")
