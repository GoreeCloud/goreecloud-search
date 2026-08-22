# SPDX-License-Identifier: AGPL-3.0-or-later
"""Audit final visual and GoreeCloud Browser evidence for first-Stable review.

This defense-in-depth audit mirrors the frozen candidate's schema-version 2 visual
and Browser review contract without changing candidate #07. It verifies the review
artifacts themselves and, when supplied, requires the final manifest to preserve the
exact derived visual and Browser summaries. It never authorizes production cutover
or Stable promotion.
"""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import pathlib
import re
import sys
from typing import Any


SHA_RE = re.compile(r"^[0-9a-f]{40}$")
DIGEST_RE = re.compile(r"^[0-9a-f]{64}$")
ARTIFACT_DIGEST_RE = re.compile(r"^sha256:[0-9a-f]{64}$")
IMAGE_RE = re.compile(r"^ghcr\.io/goreecloud/goreecloud-search@sha256:[0-9a-f]{64}$")
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
    "response_body",
}


class AuditError(ValueError):
    """Raised when final review evidence is incomplete, inconsistent, or unsafe."""


def _load(path: pathlib.Path) -> dict[str, Any]:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        raise AuditError(f"Unable to read valid JSON from {path}: {exc}") from exc
    if not isinstance(value, dict):
        raise AuditError(f"{path} must contain a JSON object")
    return value


def _sha256(path: pathlib.Path) -> str:
    digest = hashlib.sha256()
    try:
        with path.open("rb") as handle:
            for chunk in iter(lambda: handle.read(1024 * 1024), b""):
                digest.update(chunk)
    except OSError as exc:
        raise AuditError(f"Unable to hash {path}: {exc}") from exc
    return digest.hexdigest()


def _mapping(value: Any, label: str) -> dict[str, Any]:
    if not isinstance(value, dict):
        raise AuditError(f"{label} must be a JSON object")
    return value


def _nonempty(value: Any, label: str) -> str:
    if not isinstance(value, str) or not value.strip():
        raise AuditError(f"{label} must be a non-empty string")
    return value.strip()


def _true(value: Any, label: str) -> None:
    if value is not True:
        raise AuditError(f"{label} must be true")


def _false(value: Any, label: str) -> None:
    if value is not False:
        raise AuditError(f"{label} must remain false")


def _sha(value: Any, label: str) -> str:
    text = _nonempty(value, label)
    if not SHA_RE.fullmatch(text):
        raise AuditError(f"{label} must be a lowercase 40-character Git SHA")
    return text


def _digest(value: Any, label: str) -> str:
    text = _nonempty(value, label)
    if not DIGEST_RE.fullmatch(text):
        raise AuditError(f"{label} must be a lowercase 64-character SHA-256 digest")
    return text


def _artifact_digest(value: Any, label: str) -> str:
    text = _nonempty(value, label)
    if not ARTIFACT_DIGEST_RE.fullmatch(text):
        raise AuditError(f"{label} must be sha256:<64-lowercase-hex>")
    return text


def _image(value: Any, label: str) -> str:
    text = _nonempty(value, label)
    if not IMAGE_RE.fullmatch(text):
        raise AuditError(f"{label} must be the GoreeCloud Search GHCR image pinned by SHA-256 digest")
    return text


def _iso_time(value: Any, label: str) -> str:
    text = _nonempty(value, label)
    try:
        parsed = dt.datetime.fromisoformat(text.replace("Z", "+00:00"))
    except ValueError as exc:
        raise AuditError(f"{label} must be ISO-8601") from exc
    if parsed.tzinfo is None:
        raise AuditError(f"{label} must include a timezone")
    return text


def _reject_sensitive_keys(value: Any, path: str) -> None:
    if isinstance(value, dict):
        for key, child in value.items():
            if str(key).strip().lower() in FORBIDDEN_KEYS:
                raise AuditError(f"Sensitive or unnecessary field is not allowed: {path}.{key}")
            _reject_sensitive_keys(child, f"{path}.{key}")
    elif isinstance(value, list):
        for index, child in enumerate(value):
            _reject_sensitive_keys(child, f"{path}[{index}]")


def _release_candidate(release: dict[str, Any]) -> tuple[str, str]:
    if release.get("schema_version") != 1 or release.get("product") != "GoreeCloud Search":
        raise AuditError("Release evidence must be a GoreeCloud Search schema-version 1 artifact")
    candidate = _mapping(release.get("candidate"), "release candidate")
    source = _sha(candidate.get("source_revision"), "release candidate.source_revision")
    image = _image(candidate.get("image"), "release candidate.image")
    if candidate.get("isolated_runtime_acceptance") != "passed":
        raise AuditError("release candidate.isolated_runtime_acceptance must be passed")
    rollback = _mapping(release.get("rollback_scope"), "release rollback_scope")
    if rollback.get("image_level_rehearsal") != "passed":
        raise AuditError("release rollback_scope.image_level_rehearsal must be passed")
    _false(rollback.get("production_cutover_authorized"), "release production_cutover_authorized")
    return source, image


def _passed_review(value: Any, label: str) -> dict[str, Any]:
    review = _mapping(value, label)
    _true(review.get("passed"), f"{label}.passed")
    reference = _nonempty(review.get("evidence_reference"), f"{label}.evidence_reference")
    return {"passed": True, "evidence_reference": reference}


def _visual_summary(visual: dict[str, Any], source: str, image: str) -> dict[str, Any]:
    _reject_sensitive_keys(visual, "visual_evidence")
    if visual.get("schema_version") != 1 or visual.get("product") != "GoreeCloud Search":
        raise AuditError("Visual evidence must be a GoreeCloud Search schema-version 1 artifact")
    _iso_time(visual.get("generated_at"), "visual generated_at")
    candidate = _mapping(visual.get("candidate"), "visual candidate")
    if _sha(candidate.get("source_revision"), "visual source_revision") != source:
        raise AuditError("Visual evidence refers to a different Search source revision")
    if _image(candidate.get("image"), "visual image") != image:
        raise AuditError("Visual evidence refers to a different Search image")
    if visual.get("glaze_ui_version") != "1.1.0":
        raise AuditError("Visual evidence glaze_ui_version must be 1.1.0")

    review_artifact = _mapping(visual.get("review_artifact"), "visual review_artifact")
    artifact_summary = {
        "reference": _nonempty(review_artifact.get("reference"), "visual review artifact.reference"),
        "digest": _artifact_digest(review_artifact.get("digest"), "visual review artifact.digest"),
    }
    reviews = _mapping(visual.get("reviews"), "visual reviews")
    review_summary = {
        name: _passed_review(reviews.get(name), f"visual reviews.{name}")
        for name in REQUIRED_VISUAL_CASES
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
    _true(scope.get("exact_candidate_visual_artifact_verified"), "visual exact candidate artifact")
    _true(scope.get("manual_visual_acceptance_verified"), "visual manual acceptance")
    _false(scope.get("production_cutover_authorized"), "visual production_cutover_authorized")
    _nonempty(scope.get("statement"), "visual scope.statement")
    return {
        "glaze_ui_version": "1.1.0",
        "review_artifact": artifact_summary,
        "reviews": review_summary,
        **device_summary,
    }


def _browser_summary(browser: dict[str, Any], source: str, image: str) -> dict[str, Any]:
    _reject_sensitive_keys(browser, "browser_evidence")
    if browser.get("schema_version") != 1 or browser.get("product") != "GoreeCloud Search":
        raise AuditError("Browser evidence must be a GoreeCloud Search schema-version 1 artifact")
    _iso_time(browser.get("generated_at"), "browser generated_at")
    candidate = _mapping(browser.get("search_candidate"), "browser search_candidate")
    if _sha(candidate.get("source_revision"), "browser Search source_revision") != source:
        raise AuditError("Browser evidence refers to a different Search source revision")
    if _image(candidate.get("image"), "browser Search image") != image:
        raise AuditError("Browser evidence refers to a different Search image")

    browser_source = _sha(browser.get("browser_source_revision"), "browser source_revision")
    runtime_artifact = _mapping(browser.get("runtime_artifact"), "browser runtime_artifact")
    runtime_summary = {
        "reference": _nonempty(runtime_artifact.get("reference"), "browser runtime artifact.reference"),
        "digest": _artifact_digest(runtime_artifact.get("digest"), "browser runtime artifact.digest"),
    }
    behaviors = _mapping(browser.get("behaviors"), "browser behaviors")
    behavior_summary: dict[str, bool] = {}
    for key in REQUIRED_BROWSER_BEHAVIORS:
        _true(behaviors.get(key), f"browser behaviors.{key}")
        behavior_summary[key] = True

    scope = _mapping(browser.get("scope"), "browser scope")
    _true(scope.get("actual_browser_runtime_verified"), "browser actual runtime")
    _true(scope.get("search_candidate_runtime_verified"), "browser Search runtime")
    _false(scope.get("production_cutover_authorized"), "browser production_cutover_authorized")
    _nonempty(scope.get("statement"), "browser scope.statement")
    return {
        "browser_source_revision": browser_source,
        "runtime_artifact": runtime_summary,
        "behaviors": behavior_summary,
    }


def _audit_final(
    final: dict[str, Any],
    source: str,
    image: str,
    visual_path: pathlib.Path,
    browser_path: pathlib.Path,
    visual_summary: dict[str, Any],
    browser_summary: dict[str, Any],
) -> None:
    _reject_sensitive_keys(final, "final_evidence")
    if final.get("schema_version") != 2 or final.get("product") != "GoreeCloud Search":
        raise AuditError("Final evidence must be a GoreeCloud Search schema-version 2 artifact")
    _iso_time(final.get("generated_at"), "final generated_at")
    candidate = _mapping(final.get("candidate"), "final candidate")
    if _sha(candidate.get("source_revision"), "final source_revision") != source:
        raise AuditError("Final evidence refers to a different Search source revision")
    if _image(candidate.get("image"), "final image") != image:
        raise AuditError("Final evidence refers to a different Search image")

    bindings = _mapping(final.get("artifact_bindings"), "final artifact_bindings")
    expected_visual = _sha256(visual_path)
    expected_browser = _sha256(browser_path)
    if _digest(bindings.get("visual_evidence_sha256"), "final visual_evidence_sha256") != expected_visual:
        raise AuditError("Final visual evidence hash does not match the supplied visual artifact")
    if _digest(bindings.get("browser_evidence_sha256"), "final browser_evidence_sha256") != expected_browser:
        raise AuditError("Final Browser evidence hash does not match the supplied Browser artifact")

    visual = _mapping(final.get("visual_acceptance"), "final visual_acceptance")
    if visual != visual_summary:
        raise AuditError("Final visual_acceptance does not exactly match the supplied visual evidence")
    browser = _mapping(final.get("browser_integration"), "final browser_integration")
    if browser != browser_summary:
        raise AuditError("Final browser_integration does not exactly match the supplied Browser evidence")

    scope = _mapping(final.get("scope"), "final scope")
    _true(scope.get("glaze_ui_1_1_final_visual_acceptance_verified"), "final visual acceptance")
    _true(scope.get("browser_runtime_integration_verified"), "final Browser integration")
    _true(scope.get("real_provider_acceptance_verified"), "final provider acceptance")
    _true(scope.get("recovery_evidence_verified"), "final recovery acceptance")
    _true(scope.get("final_candidate_acceptance_complete"), "final candidate acceptance")
    _false(scope.get("production_cutover_authorized"), "final production_cutover_authorized")
    _nonempty(scope.get("statement"), "final scope.statement")


def audit(args: argparse.Namespace) -> dict[str, Any]:
    """Audit final visual and Browser evidence and return the derived summaries."""
    release_path = pathlib.Path(args.release_evidence)
    visual_path = pathlib.Path(args.visual_evidence)
    browser_path = pathlib.Path(args.browser_evidence)
    release = _load(release_path)
    visual = _load(visual_path)
    browser = _load(browser_path)
    _reject_sensitive_keys(release, "release_evidence")
    source, image = _release_candidate(release)
    visual_summary = _visual_summary(visual, source, image)
    browser_summary = _browser_summary(browser, source, image)

    if args.final_evidence:
        final = _load(pathlib.Path(args.final_evidence))
        _audit_final(
            final,
            source,
            image,
            visual_path,
            browser_path,
            visual_summary,
            browser_summary,
        )
    return {
        "source_revision": source,
        "image": image,
        "visual_acceptance": visual_summary,
        "browser_integration": browser_summary,
        "production_cutover_authorized": False,
    }


def parser() -> argparse.ArgumentParser:
    """Build the command-line parser."""
    root = argparse.ArgumentParser(description=__doc__)
    root.add_argument("--release-evidence", required=True)
    root.add_argument("--visual-evidence", required=True)
    root.add_argument("--browser-evidence", required=True)
    root.add_argument("--final-evidence")
    return root


def main() -> int:
    """Run the audit CLI."""
    args = parser().parse_args()
    try:
        result = audit(args)
    except AuditError as exc:
        print(f"Final review evidence audit error: {exc}", file=sys.stderr)
        return 2
    print("GoreeCloud Search final visual/Browser review evidence passed defense-in-depth audit.")
    print(f"Candidate source: {result['source_revision']}")
    print(f"Candidate image: {result['image']}")
    print("Production cutover authorized by this audit: false")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
