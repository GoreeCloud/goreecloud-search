# SPDX-License-Identifier: AGPL-3.0-or-later
"""Create and validate GoreeCloud Search first-Stable final-candidate evidence.

The final artifact binds already-validated release, target-runtime, recovery, and
real-provider evidence to the manual Glaze UI and GoreeCloud Browser runtime
acceptance that cannot be proven by source CI alone. Validation proves the
reviewed evidence set is internally consistent for one exact candidate. It never
sets or permits production_cutover_authorized=true; the explicit release decision
remains a separate human governance step.
"""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import pathlib
import re
from typing import Any


SHA_RE = re.compile(r"^[0-9a-f]{40}$")
DIGEST_RE = re.compile(r"^[0-9a-f]{64}$")
IMAGE_RE = re.compile(r"^ghcr\.io/goreecloud/goreecloud-search@sha256:[0-9a-f]{64}$")
REQUIRED_PROVIDER_CATEGORIES = frozenset({"general", "images", "videos", "news", "files"})
REQUIRED_VISUAL_CASES = (
    "compact_light",
    "compact_dark",
    "expanded_light",
    "expanded_dark",
)
FORBIDDEN_KEYS = {
    "password",
    "passwords",
    "token",
    "tokens",
    "cookie",
    "cookies",
    "secret",
    "secrets",
    "credential",
    "credentials",
    "environment_values",
    "env",
    "query",
    "queries",
    "response_content",
}


class EvidenceError(ValueError):
    """Raised when final-candidate evidence is incomplete, inconsistent, or unsafe."""


def _load(path: pathlib.Path) -> dict[str, Any]:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        raise EvidenceError(f"Unable to read valid JSON from {path}: {exc}") from exc
    if not isinstance(value, dict):
        raise EvidenceError(f"{path} must contain a JSON object")
    return value


def _write(path: pathlib.Path, value: dict[str, Any]) -> None:
    path.write_text(json.dumps(value, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def _sha256(path: pathlib.Path) -> str:
    digest = hashlib.sha256()
    try:
        with path.open("rb") as handle:
            for chunk in iter(lambda: handle.read(1024 * 1024), b""):
                digest.update(chunk)
    except OSError as exc:
        raise EvidenceError(f"Unable to hash {path}: {exc}") from exc
    return digest.hexdigest()


def _mapping(value: Any, label: str) -> dict[str, Any]:
    if not isinstance(value, dict):
        raise EvidenceError(f"{label} must be a JSON object")
    return value


def _list(value: Any, label: str) -> list[Any]:
    if not isinstance(value, list):
        raise EvidenceError(f"{label} must be a JSON array")
    return value


def _nonempty(value: Any, label: str) -> str:
    if not isinstance(value, str) or not value.strip():
        raise EvidenceError(f"{label} must be a non-empty string")
    return value.strip()


def _true(value: Any, label: str) -> None:
    if value is not True:
        raise EvidenceError(f"{label} must be true after the acceptance step is completed")


def _false(value: Any, label: str) -> None:
    if value is not False:
        raise EvidenceError(f"{label} must remain false")


def _sha(value: Any, label: str) -> str:
    text = _nonempty(value, label)
    if not SHA_RE.fullmatch(text):
        raise EvidenceError(f"{label} must be a lowercase 40-character Git SHA")
    return text


def _digest(value: Any, label: str) -> str:
    text = _nonempty(value, label)
    if not DIGEST_RE.fullmatch(text):
        raise EvidenceError(f"{label} must be a lowercase 64-character SHA-256 digest")
    return text


def _image(value: Any, label: str) -> str:
    text = _nonempty(value, label)
    if not IMAGE_RE.fullmatch(text):
        raise EvidenceError(
            f"{label} must be ghcr.io/goreecloud/goreecloud-search pinned by sha256 digest"
        )
    return text


def _iso_time(value: Any, label: str) -> str:
    text = _nonempty(value, label)
    try:
        parsed = dt.datetime.fromisoformat(text.replace("Z", "+00:00"))
    except ValueError as exc:
        raise EvidenceError(f"{label} must be ISO-8601") from exc
    if parsed.tzinfo is None:
        raise EvidenceError(f"{label} must include a timezone")
    return text


def _reject_sensitive_keys(value: Any, path: str = "evidence") -> None:
    if isinstance(value, dict):
        for key, child in value.items():
            normalized = str(key).strip().lower()
            if normalized in FORBIDDEN_KEYS:
                raise EvidenceError(f"Sensitive or unnecessary field is not allowed: {path}.{key}")
            _reject_sensitive_keys(child, f"{path}.{key}")
    elif isinstance(value, list):
        for index, child in enumerate(value):
            _reject_sensitive_keys(child, f"{path}[{index}]")


def _candidate_from_release(release: dict[str, Any]) -> tuple[str, str]:
    if release.get("schema_version") != 1 or release.get("product") != "GoreeCloud Search":
        raise EvidenceError("Release evidence must be a GoreeCloud Search schema-version 1 artifact")
    candidate = _mapping(release.get("candidate"), "release candidate")
    source = _sha(candidate.get("source_revision"), "release candidate source_revision")
    image = _image(candidate.get("image"), "release candidate image")
    if candidate.get("isolated_runtime_acceptance") != "passed":
        raise EvidenceError("Release candidate isolated runtime acceptance must have passed")
    rollback_scope = _mapping(release.get("rollback_scope"), "release rollback_scope")
    if rollback_scope.get("image_level_rehearsal") != "passed":
        raise EvidenceError("Release image-level rollback rehearsal must have passed")
    _false(rollback_scope.get("production_cutover_authorized"), "release production_cutover_authorized")
    return source, image


def _verify_runtime(runtime: dict[str, Any], source: str, image: str) -> None:
    if runtime.get("schema_version") != 1 or runtime.get("product") != "GoreeCloud Search":
        raise EvidenceError("Target-runtime evidence must be a GoreeCloud Search schema-version 1 artifact")
    container = _mapping(runtime.get("container_runtime"), "target container_runtime")
    if _sha(container.get("expected_source_revision"), "target expected_source_revision") != source:
        raise EvidenceError("Target-runtime evidence refers to a different source revision")
    if _image(container.get("expected_image"), "target expected_image") != image:
        raise EvidenceError("Target-runtime evidence refers to a different image")
    scope = _mapping(runtime.get("scope"), "target runtime scope")
    _true(scope.get("target_runtime_identity_verified"), "target runtime identity")
    _false(scope.get("production_cutover_authorized"), "target production_cutover_authorized")


def _verify_recovery(recovery: dict[str, Any], source: str, image: str) -> None:
    if recovery.get("schema_version") != 1 or recovery.get("product") != "GoreeCloud Search":
        raise EvidenceError("Recovery evidence must be a GoreeCloud Search schema-version 1 artifact")
    candidate = _mapping(recovery.get("candidate"), "recovery candidate")
    if _sha(candidate.get("source_revision"), "recovery source_revision") != source:
        raise EvidenceError("Recovery evidence refers to a different source revision")
    if _image(candidate.get("image"), "recovery image") != image:
        raise EvidenceError("Recovery evidence refers to a different image")
    scope = _mapping(recovery.get("scope"), "recovery scope")
    _true(scope.get("application_level_restore_tested"), "recovery application_level_restore_tested")
    _true(scope.get("monitoring_and_alerting_verified"), "recovery monitoring_and_alerting_verified")
    _true(scope.get("rollback_evidence_verified"), "recovery rollback_evidence_verified")
    _false(scope.get("production_cutover_authorized"), "recovery production_cutover_authorized")


def _verify_provider(provider: dict[str, Any], source: str, image: str) -> None:
    if provider.get("schema_version") != 1 or provider.get("product") != "GoreeCloud Search":
        raise EvidenceError("Provider evidence must be a GoreeCloud Search schema-version 1 artifact")
    candidate = _mapping(provider.get("candidate"), "provider candidate")
    if _sha(candidate.get("source_revision"), "provider source_revision") != source:
        raise EvidenceError("Provider evidence refers to a different source revision")
    if _image(candidate.get("image"), "provider image") != image:
        raise EvidenceError("Provider evidence refers to a different image")

    required_categories = set(_list(provider.get("required_categories"), "provider required_categories"))
    if not REQUIRED_PROVIDER_CATEGORIES.issubset(required_categories):
        missing = sorted(REQUIRED_PROVIDER_CATEGORIES - required_categories)
        raise EvidenceError("Provider evidence is missing required categories: " + ", ".join(missing))

    results = _list(provider.get("results"), "provider results")
    passed_categories: set[str] = set()
    for index, item in enumerate(results):
        result = _mapping(item, f"provider results[{index}]")
        category = _nonempty(result.get("category"), f"provider results[{index}].category")
        if result.get("passed") is True:
            passed_categories.add(category)
    missing_passes = sorted(REQUIRED_PROVIDER_CATEGORIES - passed_categories)
    if missing_passes:
        raise EvidenceError(
            "Provider evidence does not contain passing final-candidate results for: "
            + ", ".join(missing_passes)
        )

    scope = _mapping(provider.get("scope"), "provider scope")
    _true(scope.get("real_provider_requests_performed"), "provider real_provider_requests_performed")
    _true(scope.get("all_required_categories_passed"), "provider all_required_categories_passed")
    _false(scope.get("query_text_persisted"), "provider query_text_persisted")
    _false(scope.get("response_content_persisted"), "provider response_content_persisted")
    _false(scope.get("production_cutover_authorized"), "provider production_cutover_authorized")


def _bindings(args: argparse.Namespace) -> tuple[str, str, dict[str, str]]:
    release_path = pathlib.Path(args.release_evidence)
    runtime_path = pathlib.Path(args.target_runtime_evidence)
    recovery_path = pathlib.Path(args.recovery_evidence)
    provider_path = pathlib.Path(args.provider_evidence)

    release = _load(release_path)
    runtime = _load(runtime_path)
    recovery = _load(recovery_path)
    provider = _load(provider_path)

    source, image = _candidate_from_release(release)
    _verify_runtime(runtime, source, image)
    _verify_recovery(recovery, source, image)
    _verify_provider(provider, source, image)

    return source, image, {
        "release_evidence_sha256": _sha256(release_path),
        "target_runtime_evidence_sha256": _sha256(runtime_path),
        "recovery_evidence_sha256": _sha256(recovery_path),
        "provider_evidence_sha256": _sha256(provider_path),
    }


def build_template(args: argparse.Namespace) -> dict[str, Any]:
    source, image, bindings = _bindings(args)
    evidence = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "generated_at": dt.datetime.now(dt.timezone.utc).isoformat().replace("+00:00", "Z"),
        "candidate": {
            "source_revision": source,
            "image": image,
        },
        "artifact_bindings": bindings,
        "visual_acceptance": {
            "glaze_ui_version": "1.1.0",
            "compact_light": {"passed": False, "evidence_reference": ""},
            "compact_dark": {"passed": False, "evidence_reference": ""},
            "expanded_light": {"passed": False, "evidence_reference": ""},
            "expanded_dark": {"passed": False, "evidence_reference": ""},
            "physical_android_preferences_review": False,
            "desktop_regression_review": False,
        },
        "browser_integration": {
            "browser_source_revision": "",
            "evidence_reference": "",
            "search_only_default_provider": False,
            "address_bar_routed_through_search": False,
            "new_tab_routed_through_search": False,
            "dedicated_search_field_routed_through_search": False,
            "no_external_browser_fallback": False,
            "search_unavailability_state_verified": False,
            "recovery_after_search_reachability_verified": False,
        },
        "scope": {
            "glaze_ui_1_1_final_visual_acceptance_verified": False,
            "browser_runtime_integration_verified": False,
            "real_provider_acceptance_verified": True,
            "recovery_evidence_verified": True,
            "final_candidate_acceptance_complete": False,
            "production_cutover_authorized": False,
            "statement": (
                "This artifact binds the first-Stable final-candidate evidence set. It remains incomplete "
                "until the manual Glaze UI and GoreeCloud Browser runtime review fields are completed and "
                "validated. Passing validation does not independently authorize production cutover."
            ),
        },
    }
    _write(pathlib.Path(args.output), evidence)
    return evidence


def validate_evidence(args: argparse.Namespace) -> dict[str, Any]:
    evidence_path = pathlib.Path(args.evidence)
    evidence = _load(evidence_path)
    _reject_sensitive_keys(evidence)
    source, image, expected_bindings = _bindings(args)

    if evidence.get("schema_version") != 1 or evidence.get("product") != "GoreeCloud Search":
        raise EvidenceError("Final evidence must be a GoreeCloud Search schema-version 1 artifact")
    _iso_time(evidence.get("generated_at"), "final generated_at")

    candidate = _mapping(evidence.get("candidate"), "final candidate")
    if _sha(candidate.get("source_revision"), "final source_revision") != source:
        raise EvidenceError("Final evidence source revision does not match the bound candidate")
    if _image(candidate.get("image"), "final image") != image:
        raise EvidenceError("Final evidence image does not match the bound candidate")

    bindings = _mapping(evidence.get("artifact_bindings"), "artifact_bindings")
    for key, expected in expected_bindings.items():
        actual = _digest(bindings.get(key), f"artifact_bindings.{key}")
        if actual != expected:
            raise EvidenceError(f"artifact_bindings.{key} does not match the supplied artifact")

    visual = _mapping(evidence.get("visual_acceptance"), "visual_acceptance")
    if visual.get("glaze_ui_version") != "1.1.0":
        raise EvidenceError("visual_acceptance.glaze_ui_version must be 1.1.0")
    for case_name in REQUIRED_VISUAL_CASES:
        case = _mapping(visual.get(case_name), f"visual_acceptance.{case_name}")
        _true(case.get("passed"), f"visual_acceptance.{case_name}.passed")
        _nonempty(case.get("evidence_reference"), f"visual_acceptance.{case_name}.evidence_reference")
    _true(
        visual.get("physical_android_preferences_review"),
        "visual_acceptance.physical_android_preferences_review",
    )
    _true(visual.get("desktop_regression_review"), "visual_acceptance.desktop_regression_review")

    browser = _mapping(evidence.get("browser_integration"), "browser_integration")
    _sha(browser.get("browser_source_revision"), "browser_integration.browser_source_revision")
    _nonempty(browser.get("evidence_reference"), "browser_integration.evidence_reference")
    for key in (
        "search_only_default_provider",
        "address_bar_routed_through_search",
        "new_tab_routed_through_search",
        "dedicated_search_field_routed_through_search",
        "no_external_browser_fallback",
        "search_unavailability_state_verified",
        "recovery_after_search_reachability_verified",
    ):
        _true(browser.get(key), f"browser_integration.{key}")

    scope = _mapping(evidence.get("scope"), "scope")
    _true(
        scope.get("glaze_ui_1_1_final_visual_acceptance_verified"),
        "scope.glaze_ui_1_1_final_visual_acceptance_verified",
    )
    _true(scope.get("browser_runtime_integration_verified"), "scope.browser_runtime_integration_verified")
    _true(scope.get("real_provider_acceptance_verified"), "scope.real_provider_acceptance_verified")
    _true(scope.get("recovery_evidence_verified"), "scope.recovery_evidence_verified")
    _true(scope.get("final_candidate_acceptance_complete"), "scope.final_candidate_acceptance_complete")
    _false(scope.get("production_cutover_authorized"), "scope.production_cutover_authorized")
    _nonempty(scope.get("statement"), "scope.statement")

    return evidence


def _common_artifact_args(parser: argparse.ArgumentParser) -> None:
    parser.add_argument("--release-evidence", required=True)
    parser.add_argument("--target-runtime-evidence", required=True)
    parser.add_argument("--recovery-evidence", required=True)
    parser.add_argument("--provider-evidence", required=True)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    subparsers = parser.add_subparsers(dest="command", required=True)

    template = subparsers.add_parser("template", help="Create an incomplete candidate-bound final evidence template.")
    _common_artifact_args(template)
    template.add_argument("--output", required=True)

    validate = subparsers.add_parser("validate", help="Validate a completed final-candidate evidence artifact.")
    _common_artifact_args(validate)
    validate.add_argument("--evidence", required=True)

    args = parser.parse_args()
    try:
        if args.command == "template":
            build_template(args)
            print(f"Wrote incomplete final-candidate evidence template: {args.output}")
        else:
            validate_evidence(args)
            print("GoreeCloud Search final-candidate evidence passed.")
    except EvidenceError as exc:
        parser.error(str(exc))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
