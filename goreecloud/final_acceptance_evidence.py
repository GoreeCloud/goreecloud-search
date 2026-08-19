# SPDX-License-Identifier: AGPL-3.0-or-later
"""Assemble and validate GoreeCloud Search first-Stable final-candidate evidence.

The final manifest cryptographically binds release, target-runtime, recovery,
real-provider, manual visual-review, and actual GoreeCloud Browser runtime
evidence for one exact immutable Search candidate. Validation proves that the
reviewed evidence set is internally consistent and unchanged. It never sets or
permits production_cutover_authorized=true; the explicit release decision
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
ARTIFACT_DIGEST_RE = re.compile(r"^sha256:[0-9a-f]{64}$")
IMAGE_RE = re.compile(r"^ghcr\.io/goreecloud/goreecloud-search@sha256:[0-9a-f]{64}$")
LOOPBACK_URL_RE = re.compile(r"^https?://(?:127\.0\.0\.1|localhost|\[::1\])(?::[0-9]{1,5})?$")
REQUIRED_PROVIDER_CATEGORIES = frozenset({"general", "images", "videos", "news", "files"})
REQUIRED_VISUAL_CASES = (
    "compact_light",
    "compact_dark",
    "expanded_light",
    "expanded_dark",
)
REQUIRED_BROWSER_BEHAVIORS = (
    "search_only_default_provider",
    "address_bar_routed_through_search",
    "new_tab_routed_through_search",
    "dedicated_search_field_routed_through_search",
    "no_external_browser_fallback",
    "search_unavailability_state_verified",
    "recovery_after_search_reachability_verified",
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


def _artifact_digest(value: Any, label: str) -> str:
    text = _nonempty(value, label)
    if not ARTIFACT_DIGEST_RE.fullmatch(text):
        raise EvidenceError(f"{label} must be an immutable sha256:<64-lowercase-hex> artifact digest")
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

    runtime_binding = _mapping(provider.get("runtime_binding"), "provider runtime_binding")
    _true(
        runtime_binding.get("verified_before_and_after_requests"),
        "provider runtime_binding.verified_before_and_after_requests",
    )
    base_url = _nonempty(runtime_binding.get("base_url"), "provider runtime_binding.base_url")
    if not LOOPBACK_URL_RE.fullmatch(base_url):
        raise EvidenceError("Provider evidence base_url must identify the loopback-staged candidate")
    _nonempty(runtime_binding.get("container"), "provider runtime_binding.container")
    published_port = _nonempty(runtime_binding.get("published_port"), "provider runtime_binding.published_port")
    if "127.0.0.1:" not in published_port and "[::1]:" not in published_port:
        raise EvidenceError("Provider evidence published_port must be loopback-only")
    if _image(
        runtime_binding.get("observed_image_reference"),
        "provider runtime_binding.observed_image_reference",
    ) != image:
        raise EvidenceError("Provider runtime binding refers to a different image")
    _nonempty(runtime_binding.get("observed_image_id"), "provider runtime_binding.observed_image_id")
    if _sha(runtime_binding.get("oci_revision"), "provider runtime_binding.oci_revision") != source:
        raise EvidenceError("Provider runtime binding refers to a different source revision")

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
    _true(
        scope.get("runtime_identity_verified_during_provider_requests"),
        "provider runtime_identity_verified_during_provider_requests",
    )
    _true(scope.get("all_required_categories_passed"), "provider all_required_categories_passed")
    _false(scope.get("query_text_persisted"), "provider query_text_persisted")
    _false(scope.get("response_content_persisted"), "provider response_content_persisted")
    _false(scope.get("production_cutover_authorized"), "provider production_cutover_authorized")


def _passed_review(value: Any, label: str) -> dict[str, Any]:
    review = _mapping(value, label)
    _true(review.get("passed"), f"{label}.passed")
    _nonempty(review.get("evidence_reference"), f"{label}.evidence_reference")
    return {
        "passed": True,
        "evidence_reference": review["evidence_reference"].strip(),
    }


def _verify_visual(visual: dict[str, Any], source: str, image: str) -> dict[str, Any]:
    _reject_sensitive_keys(visual, "visual_evidence")
    if visual.get("schema_version") != 1 or visual.get("product") != "GoreeCloud Search":
        raise EvidenceError("Visual evidence must be a GoreeCloud Search schema-version 1 artifact")
    _iso_time(visual.get("generated_at"), "visual generated_at")
    candidate = _mapping(visual.get("candidate"), "visual candidate")
    if _sha(candidate.get("source_revision"), "visual source_revision") != source:
        raise EvidenceError("Visual evidence refers to a different source revision")
    if _image(candidate.get("image"), "visual image") != image:
        raise EvidenceError("Visual evidence refers to a different image")
    if visual.get("glaze_ui_version") != "1.1.0":
        raise EvidenceError("Visual evidence glaze_ui_version must be 1.1.0")

    review_artifact = _mapping(visual.get("review_artifact"), "visual review_artifact")
    artifact_summary = {
        "reference": _nonempty(review_artifact.get("reference"), "visual review_artifact.reference"),
        "digest": _artifact_digest(review_artifact.get("digest"), "visual review_artifact.digest"),
    }

    reviews = _mapping(visual.get("reviews"), "visual reviews")
    review_summary = {
        case_name: _passed_review(reviews.get(case_name), f"visual reviews.{case_name}")
        for case_name in REQUIRED_VISUAL_CASES
    }
    device_summary = {
        "physical_android_preferences_review": _passed_review(
            visual.get("physical_android_preferences_review"),
            "visual physical_android_preferences_review",
        ),
        "desktop_regression_review": _passed_review(
            visual.get("desktop_regression_review"),
            "visual desktop_regression_review",
        ),
        "persisted_theme_preference_review": _passed_review(
            visual.get("persisted_theme_preference_review"),
            "visual persisted_theme_preference_review",
        ),
    }

    scope = _mapping(visual.get("scope"), "visual scope")
    _true(scope.get("exact_candidate_visual_artifact_verified"), "visual exact_candidate_visual_artifact_verified")
    _true(scope.get("manual_visual_acceptance_verified"), "visual manual_visual_acceptance_verified")
    _false(scope.get("production_cutover_authorized"), "visual production_cutover_authorized")
    _nonempty(scope.get("statement"), "visual scope.statement")

    return {
        "glaze_ui_version": "1.1.0",
        "review_artifact": artifact_summary,
        "reviews": review_summary,
        **device_summary,
    }


def _verify_browser(browser: dict[str, Any], source: str, image: str) -> dict[str, Any]:
    _reject_sensitive_keys(browser, "browser_evidence")
    if browser.get("schema_version") != 1 or browser.get("product") != "GoreeCloud Search":
        raise EvidenceError("Browser evidence must be a GoreeCloud Search schema-version 1 artifact")
    _iso_time(browser.get("generated_at"), "browser generated_at")
    candidate = _mapping(browser.get("search_candidate"), "browser search_candidate")
    if _sha(candidate.get("source_revision"), "browser Search source_revision") != source:
        raise EvidenceError("Browser evidence refers to a different Search source revision")
    if _image(candidate.get("image"), "browser Search image") != image:
        raise EvidenceError("Browser evidence refers to a different Search image")

    browser_source = _sha(browser.get("browser_source_revision"), "browser source_revision")
    runtime_artifact = _mapping(browser.get("runtime_artifact"), "browser runtime_artifact")
    runtime_summary = {
        "reference": _nonempty(runtime_artifact.get("reference"), "browser runtime_artifact.reference"),
        "digest": _artifact_digest(runtime_artifact.get("digest"), "browser runtime_artifact.digest"),
    }
    behaviors = _mapping(browser.get("behaviors"), "browser behaviors")
    behavior_summary: dict[str, bool] = {}
    for key in REQUIRED_BROWSER_BEHAVIORS:
        _true(behaviors.get(key), f"browser behaviors.{key}")
        behavior_summary[key] = True

    scope = _mapping(browser.get("scope"), "browser scope")
    _true(scope.get("actual_browser_runtime_verified"), "browser actual_browser_runtime_verified")
    _true(scope.get("search_candidate_runtime_verified"), "browser search_candidate_runtime_verified")
    _false(scope.get("production_cutover_authorized"), "browser production_cutover_authorized")
    _nonempty(scope.get("statement"), "browser scope.statement")

    return {
        "browser_source_revision": browser_source,
        "runtime_artifact": runtime_summary,
        "behaviors": behavior_summary,
    }


def _bindings(
    args: argparse.Namespace,
) -> tuple[str, str, dict[str, str], dict[str, Any], dict[str, Any]]:
    release_path = pathlib.Path(args.release_evidence)
    runtime_path = pathlib.Path(args.target_runtime_evidence)
    recovery_path = pathlib.Path(args.recovery_evidence)
    provider_path = pathlib.Path(args.provider_evidence)
    visual_path = pathlib.Path(args.visual_evidence)
    browser_path = pathlib.Path(args.browser_evidence)

    release = _load(release_path)
    runtime = _load(runtime_path)
    recovery = _load(recovery_path)
    provider = _load(provider_path)
    visual = _load(visual_path)
    browser = _load(browser_path)

    source, image = _candidate_from_release(release)
    _verify_runtime(runtime, source, image)
    _verify_recovery(recovery, source, image)
    _verify_provider(provider, source, image)
    visual_summary = _verify_visual(visual, source, image)
    browser_summary = _verify_browser(browser, source, image)

    return source, image, {
        "release_evidence_sha256": _sha256(release_path),
        "target_runtime_evidence_sha256": _sha256(runtime_path),
        "recovery_evidence_sha256": _sha256(recovery_path),
        "provider_evidence_sha256": _sha256(provider_path),
        "visual_evidence_sha256": _sha256(visual_path),
        "browser_evidence_sha256": _sha256(browser_path),
    }, visual_summary, browser_summary


def assemble_manifest(args: argparse.Namespace) -> dict[str, Any]:
    source, image, bindings, visual_summary, browser_summary = _bindings(args)
    evidence = {
        "schema_version": 2,
        "product": "GoreeCloud Search",
        "generated_at": dt.datetime.now(dt.timezone.utc).isoformat().replace("+00:00", "Z"),
        "candidate": {
            "source_revision": source,
            "image": image,
        },
        "artifact_bindings": bindings,
        "visual_acceptance": visual_summary,
        "browser_integration": browser_summary,
        "scope": {
            "glaze_ui_1_1_final_visual_acceptance_verified": True,
            "browser_runtime_integration_verified": True,
            "real_provider_acceptance_verified": True,
            "recovery_evidence_verified": True,
            "final_candidate_acceptance_complete": True,
            "production_cutover_authorized": False,
            "statement": (
                "This manifest binds the complete first-Stable final-candidate evidence set for one exact "
                "Search candidate. Passing validation confirms evidence completeness and integrity only; it "
                "does not independently authorize production cutover or Stable promotion."
            ),
        },
    }
    _write(pathlib.Path(args.output), evidence)
    return evidence


def validate_evidence(args: argparse.Namespace) -> dict[str, Any]:
    evidence_path = pathlib.Path(args.evidence)
    evidence = _load(evidence_path)
    _reject_sensitive_keys(evidence)
    source, image, expected_bindings, visual_summary, browser_summary = _bindings(args)

    if evidence.get("schema_version") != 2 or evidence.get("product") != "GoreeCloud Search":
        raise EvidenceError("Final evidence must be a GoreeCloud Search schema-version 2 artifact")
    _iso_time(evidence.get("generated_at"), "final generated_at")

    candidate = _mapping(evidence.get("candidate"), "final candidate")
    if _sha(candidate.get("source_revision"), "final source_revision") != source:
        raise EvidenceError("Final evidence source revision does not match the bound candidate")
    if _image(candidate.get("image"), "final image") != image:
        raise EvidenceError("Final evidence image does not match the bound candidate")

    bindings = _mapping(evidence.get("artifact_bindings"), "artifact_bindings")
    if set(bindings) != set(expected_bindings):
        raise EvidenceError("artifact_bindings must contain exactly the six required evidence hashes")
    for key, expected in expected_bindings.items():
        actual = _digest(bindings.get(key), f"artifact_bindings.{key}")
        if actual != expected:
            raise EvidenceError(f"artifact_bindings.{key} does not match the supplied artifact")

    visual = _mapping(evidence.get("visual_acceptance"), "visual_acceptance")
    if visual != visual_summary:
        raise EvidenceError("Final visual_acceptance does not exactly match the bound visual evidence")

    browser = _mapping(evidence.get("browser_integration"), "browser_integration")
    if browser != browser_summary:
        raise EvidenceError("Final browser_integration does not exactly match the bound Browser evidence")

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
    parser.add_argument("--visual-evidence", required=True)
    parser.add_argument("--browser-evidence", required=True)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    subparsers = parser.add_subparsers(dest="command", required=True)

    assemble = subparsers.add_parser(
        "assemble",
        help="Assemble a completed candidate-bound final manifest from six validated evidence artifacts.",
    )
    _common_artifact_args(assemble)
    assemble.add_argument("--output", required=True)

    validate = subparsers.add_parser("validate", help="Validate a completed final-candidate evidence manifest.")
    _common_artifact_args(validate)
    validate.add_argument("--evidence", required=True)

    args = parser.parse_args()
    try:
        if args.command == "assemble":
            assemble_manifest(args)
            print(f"Wrote complete candidate-bound final evidence manifest: {args.output}")
        else:
            validate_evidence(args)
            print("GoreeCloud Search final-candidate evidence passed.")
    except EvidenceError as exc:
        parser.error(str(exc))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
