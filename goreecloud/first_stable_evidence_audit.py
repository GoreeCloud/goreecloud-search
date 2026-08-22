# SPDX-License-Identifier: AGPL-3.0-or-later
"""Audit GoreeCloud Search first-Stable evidence before release governance review.

This auditor is intentionally separate from the frozen first-Stable candidate. It performs
additional defense-in-depth checks across the six companion evidence artifacts and an
optional schema-version 2 final manifest. It never authorizes production cutover or Stable
promotion.
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
LOOPBACK_URL_RE = re.compile(r"^https?://(?:127\.0\.0\.1|localhost|\[::1\])(?::[0-9]{1,5})?$")
REPOSITORY_URL = "https://github.com/GoreeCloud/goreecloud-search"
REQUIRED_PROVIDER_CATEGORIES = frozenset({"general", "images", "videos", "news", "files"})
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
    """Raised when evidence is incomplete, inconsistent, unsafe, or weakly bound."""


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


def _list(value: Any, label: str) -> list[Any]:
    if not isinstance(value, list):
        raise AuditError(f"{label} must be a JSON array")
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
            normalized = str(key).strip().lower()
            if normalized in FORBIDDEN_KEYS:
                raise AuditError(f"Sensitive or unnecessary field is not allowed: {path}.{key}")
            _reject_sensitive_keys(child, f"{path}.{key}")
    elif isinstance(value, list):
        for index, child in enumerate(value):
            _reject_sensitive_keys(child, f"{path}[{index}]")


def _schema_one(value: dict[str, Any], label: str) -> None:
    if value.get("schema_version") != 1 or value.get("product") != "GoreeCloud Search":
        raise AuditError(f"{label} must be a GoreeCloud Search schema-version 1 artifact")


def _candidate_from_release(release: dict[str, Any]) -> tuple[str, str]:
    _schema_one(release, "Release evidence")
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


def _audit_runtime(runtime: dict[str, Any], source: str, image: str) -> None:
    _schema_one(runtime, "Target-runtime evidence")
    _iso_time(runtime.get("generated_at"), "target-runtime generated_at")
    target = _mapping(runtime.get("target"), "target-runtime target")
    base_url = _nonempty(target.get("base_url"), "target-runtime target.base_url")
    if not LOOPBACK_URL_RE.fullmatch(base_url):
        raise AuditError("target-runtime target.base_url must identify loopback-only staging")
    _nonempty(target.get("container"), "target-runtime target.container")

    http_acceptance = _mapping(runtime.get("http_acceptance"), "target-runtime http_acceptance")
    for key in ("home_identity", "preferences_identity", "about_identity", "health", "privacy_headers"):
        if http_acceptance.get(key) != "passed":
            raise AuditError(f"target-runtime http_acceptance.{key} must be passed")

    container = _mapping(runtime.get("container_runtime"), "target-runtime container_runtime")
    if container.get("status") != "verified":
        raise AuditError("target-runtime container_runtime.status must be verified")
    _true(container.get("running"), "target-runtime container_runtime.running")
    if container.get("health") != "healthy":
        raise AuditError("target-runtime container_runtime.health must be healthy")
    ports = _nonempty(container.get("published_ports"), "target-runtime container_runtime.published_ports")
    if "127.0.0.1:" not in ports and "[::1]:" not in ports:
        raise AuditError("target-runtime published ports must be loopback-only")
    if container.get("identity_status") != "verified":
        raise AuditError("target-runtime container_runtime.identity_status must be verified")
    if _image(container.get("expected_image"), "target-runtime expected_image") != image:
        raise AuditError("target-runtime expected image does not match release evidence")
    if _sha(container.get("expected_source_revision"), "target-runtime expected_source_revision") != source:
        raise AuditError("target-runtime expected source does not match release evidence")
    if _image(container.get("observed_image_reference"), "target-runtime observed_image_reference") != image:
        raise AuditError("target-runtime observed image does not match release evidence")
    _nonempty(container.get("observed_image_id"), "target-runtime observed_image_id")

    oci = _mapping(container.get("oci"), "target-runtime container_runtime.oci")
    if oci.get("title") != "GoreeCloud Search":
        raise AuditError("target-runtime OCI title must be GoreeCloud Search")
    if oci.get("source") != REPOSITORY_URL:
        raise AuditError("target-runtime OCI source must identify the canonical Search repository")
    if _sha(oci.get("revision"), "target-runtime OCI revision") != source:
        raise AuditError("target-runtime OCI revision does not match release evidence")
    _nonempty(oci.get("version"), "target-runtime OCI version")
    if oci.get("licenses") != "AGPL-3.0-or-later":
        raise AuditError("target-runtime OCI licenses must be AGPL-3.0-or-later")

    scope = _mapping(runtime.get("scope"), "target-runtime scope")
    _true(scope.get("target_runtime_identity_verified"), "target-runtime identity verified")
    _false(scope.get("target_environment_configuration_rollback_tested"), "target-runtime configuration rollback")
    _false(scope.get("target_environment_data_restore_tested"), "target-runtime data restore")
    _false(scope.get("backup_restore_tested"), "target-runtime backup restore")
    _false(scope.get("production_cutover_authorized"), "target-runtime production_cutover_authorized")


def _audit_recovery(
    recovery: dict[str, Any],
    source: str,
    image: str,
    release_sha256: str,
    runtime_sha256: str,
) -> None:
    _schema_one(recovery, "Recovery evidence")
    _iso_time(recovery.get("generated_at"), "recovery generated_at")
    if recovery.get("environment") != "goreecloud-vps-01":
        raise AuditError("recovery environment must be goreecloud-vps-01")
    candidate = _mapping(recovery.get("candidate"), "recovery candidate")
    if _sha(candidate.get("source_revision"), "recovery candidate.source_revision") != source:
        raise AuditError("recovery candidate source does not match release evidence")
    if _image(candidate.get("image"), "recovery candidate.image") != image:
        raise AuditError("recovery candidate image does not match release evidence")

    rollback_identity = _mapping(recovery.get("known_good_rollback"), "recovery known_good_rollback")
    _sha(rollback_identity.get("source_revision"), "recovery rollback source_revision")
    rollback_image = _image(rollback_identity.get("image"), "recovery rollback image")
    if rollback_image == image:
        raise AuditError("recovery rollback image must differ from the candidate image")

    bindings = _mapping(recovery.get("artifact_bindings"), "recovery artifact_bindings")
    if _digest(bindings.get("release_evidence_sha256"), "recovery release_evidence_sha256") != release_sha256:
        raise AuditError("recovery release-evidence binding does not match the supplied artifact")
    if _digest(bindings.get("target_runtime_evidence_sha256"), "recovery target_runtime_evidence_sha256") != runtime_sha256:
        raise AuditError("recovery target-runtime binding does not match the supplied artifact")
    _digest(bindings.get("rollback_baseline_sha256"), "recovery rollback_baseline_sha256")

    backup = _mapping(recovery.get("backup"), "recovery backup")
    _nonempty(backup.get("system"), "recovery backup.system")
    _nonempty(backup.get("snapshot_reference"), "recovery backup.snapshot_reference")
    _iso_time(backup.get("captured_at"), "recovery backup.captured_at")
    backup_scope = _mapping(backup.get("scope"), "recovery backup.scope")
    for key in ("stack_definition", "search_settings", "protected_runtime_configuration_recovery_path", "caddy_search_route"):
        _true(backup_scope.get(key), f"recovery backup.scope.{key}")
    if backup.get("cache_policy") != "rebuildable-non-authoritative":
        raise AuditError("recovery cache policy must be rebuildable-non-authoritative")

    restore = _mapping(recovery.get("restore"), "recovery restore")
    _nonempty(restore.get("isolated_location"), "recovery restore.isolated_location")
    _false(restore.get("production_modified"), "recovery restore.production_modified")
    for key in (
        "stack_definition_restored",
        "search_settings_restored",
        "protected_runtime_configuration_recovery_verified",
        "caddy_search_route_copy_restored",
        "compose_validation_passed",
        "runtime_recreated",
        "runtime_health_passed",
        "runtime_identity_passed",
        "target_acceptance_passed",
    ):
        _true(restore.get(key), f"recovery restore.{key}")

    rollback = _mapping(recovery.get("rollback"), "recovery rollback")
    for key in (
        "known_good_image_available",
        "previous_runtime_configuration_preserved",
        "previous_caddy_route_preserved",
        "rollback_procedure_documented",
        "isolated_rollback_rehearsal_passed",
    ):
        _true(rollback.get(key), f"recovery rollback.{key}")
    mode = _nonempty(rollback.get("mode"), "recovery rollback.mode")
    if mode == "production-route-rehearsal":
        _true(rollback.get("production_route_rollback_tested"), "recovery production route rollback")
    elif mode == "equivalent-verified-evidence":
        _true(rollback.get("equivalent_verified_rollback_evidence"), "recovery equivalent rollback evidence")
    else:
        raise AuditError("recovery rollback.mode is not approved")

    monitoring = _mapping(recovery.get("monitoring"), "recovery monitoring")
    if monitoring.get("monitor_identity") != "GoreeCloud Search":
        raise AuditError("recovery monitoring.monitor_identity must be GoreeCloud Search")
    _true(monitoring.get("availability_monitor_verified"), "recovery availability monitor")
    _true(monitoring.get("alert_delivery_verified"), "recovery alert delivery")

    scope = _mapping(recovery.get("scope"), "recovery scope")
    _true(scope.get("application_level_restore_tested"), "recovery application-level restore")
    _true(scope.get("monitoring_and_alerting_verified"), "recovery monitoring and alerting")
    _true(scope.get("rollback_evidence_verified"), "recovery rollback evidence")
    _false(scope.get("production_cutover_authorized"), "recovery production_cutover_authorized")


def _audit_provider(provider: dict[str, Any], source: str, image: str) -> None:
    _schema_one(provider, "Provider evidence")
    _iso_time(provider.get("generated_at"), "provider generated_at")
    candidate = _mapping(provider.get("candidate"), "provider candidate")
    if _sha(candidate.get("source_revision"), "provider candidate.source_revision") != source:
        raise AuditError("provider candidate source does not match release evidence")
    if _image(candidate.get("image"), "provider candidate.image") != image:
        raise AuditError("provider candidate image does not match release evidence")

    runtime = _mapping(provider.get("runtime_binding"), "provider runtime_binding")
    _true(runtime.get("verified_before_and_after_requests"), "provider runtime identity before/after requests")
    base_url = _nonempty(runtime.get("base_url"), "provider runtime_binding.base_url")
    if not LOOPBACK_URL_RE.fullmatch(base_url):
        raise AuditError("provider evidence must refer to loopback-only candidate staging")
    _nonempty(runtime.get("container"), "provider runtime_binding.container")
    ports = _nonempty(runtime.get("published_port"), "provider runtime_binding.published_port")
    if "127.0.0.1:" not in ports and "[::1]:" not in ports:
        raise AuditError("provider published port must be loopback-only")
    if _image(runtime.get("observed_image_reference"), "provider observed_image_reference") != image:
        raise AuditError("provider observed image does not match release evidence")
    _nonempty(runtime.get("observed_image_id"), "provider observed_image_id")
    if _sha(runtime.get("oci_revision"), "provider oci_revision") != source:
        raise AuditError("provider OCI revision does not match release evidence")

    required = set(_list(provider.get("required_categories"), "provider required_categories"))
    if not REQUIRED_PROVIDER_CATEGORIES.issubset(required):
        missing = sorted(REQUIRED_PROVIDER_CATEGORIES - required)
        raise AuditError("provider evidence is missing required categories: " + ", ".join(missing))
    passed: set[str] = set()
    for index, item in enumerate(_list(provider.get("results"), "provider results")):
        result = _mapping(item, f"provider results[{index}]")
        category = _nonempty(result.get("category"), f"provider results[{index}].category")
        if result.get("passed") is True:
            passed.add(category)
    missing_passes = sorted(REQUIRED_PROVIDER_CATEGORIES - passed)
    if missing_passes:
        raise AuditError("provider evidence lacks passing required categories: " + ", ".join(missing_passes))

    scope = _mapping(provider.get("scope"), "provider scope")
    _true(scope.get("real_provider_requests_performed"), "provider real requests")
    _true(scope.get("runtime_identity_verified_during_provider_requests"), "provider runtime binding")
    _true(scope.get("all_required_categories_passed"), "provider required categories")
    _false(scope.get("query_text_persisted"), "provider query persistence")
    _false(scope.get("response_content_persisted"), "provider response persistence")
    _false(scope.get("production_cutover_authorized"), "provider production_cutover_authorized")


def _audit_visual(visual: dict[str, Any], source: str, image: str) -> None:
    _schema_one(visual, "Visual evidence")
    _iso_time(visual.get("generated_at"), "visual generated_at")
    candidate = _mapping(visual.get("candidate"), "visual candidate")
    if _sha(candidate.get("source_revision"), "visual candidate.source_revision") != source:
        raise AuditError("visual candidate source does not match release evidence")
    if _image(candidate.get("image"), "visual candidate.image") != image:
        raise AuditError("visual candidate image does not match release evidence")
    artifact = _mapping(visual.get("review_artifact"), "visual review_artifact")
    _nonempty(artifact.get("reference"), "visual review artifact reference")
    _artifact_digest(artifact.get("digest"), "visual review artifact digest")
    scope = _mapping(visual.get("scope"), "visual scope")
    _true(scope.get("exact_candidate_visual_artifact_verified"), "visual exact candidate artifact")
    _true(scope.get("manual_visual_acceptance_verified"), "visual manual acceptance")
    _false(scope.get("production_cutover_authorized"), "visual production_cutover_authorized")


def _audit_browser(browser: dict[str, Any], source: str, image: str) -> None:
    _schema_one(browser, "Browser evidence")
    _iso_time(browser.get("generated_at"), "browser generated_at")
    candidate = _mapping(browser.get("search_candidate"), "browser search_candidate")
    if _sha(candidate.get("source_revision"), "browser Search source_revision") != source:
        raise AuditError("Browser evidence refers to a different Search source")
    if _image(candidate.get("image"), "browser Search image") != image:
        raise AuditError("Browser evidence refers to a different Search image")
    _sha(browser.get("browser_source_revision"), "browser source_revision")
    artifact = _mapping(browser.get("runtime_artifact"), "browser runtime_artifact")
    _nonempty(artifact.get("reference"), "browser runtime artifact reference")
    _artifact_digest(artifact.get("digest"), "browser runtime artifact digest")
    scope = _mapping(browser.get("scope"), "browser scope")
    _true(scope.get("actual_browser_runtime_verified"), "browser actual runtime")
    _true(scope.get("search_candidate_runtime_verified"), "browser Search runtime")
    _false(scope.get("production_cutover_authorized"), "browser production_cutover_authorized")


def _audit_final(
    final: dict[str, Any],
    source: str,
    image: str,
    bindings: dict[str, str],
) -> None:
    if final.get("schema_version") != 2 or final.get("product") != "GoreeCloud Search":
        raise AuditError("Final evidence must be a GoreeCloud Search schema-version 2 artifact")
    _iso_time(final.get("generated_at"), "final generated_at")
    candidate = _mapping(final.get("candidate"), "final candidate")
    if _sha(candidate.get("source_revision"), "final source_revision") != source:
        raise AuditError("final source does not match release evidence")
    if _image(candidate.get("image"), "final image") != image:
        raise AuditError("final image does not match release evidence")
    artifact_bindings = _mapping(final.get("artifact_bindings"), "final artifact_bindings")
    if set(artifact_bindings) != set(bindings):
        raise AuditError("final artifact_bindings must contain exactly the six required hashes")
    for key, expected in bindings.items():
        if _digest(artifact_bindings.get(key), f"final artifact_bindings.{key}") != expected:
            raise AuditError(f"final artifact_bindings.{key} does not match the supplied artifact")
    scope = _mapping(final.get("scope"), "final scope")
    _true(scope.get("glaze_ui_1_1_final_visual_acceptance_verified"), "final visual acceptance")
    _true(scope.get("browser_runtime_integration_verified"), "final Browser integration")
    _true(scope.get("real_provider_acceptance_verified"), "final provider acceptance")
    _true(scope.get("recovery_evidence_verified"), "final recovery acceptance")
    _true(scope.get("final_candidate_acceptance_complete"), "final candidate acceptance")
    _false(scope.get("production_cutover_authorized"), "final production_cutover_authorized")


def audit(args: argparse.Namespace) -> dict[str, str]:
    paths = {
        "release": pathlib.Path(args.release_evidence),
        "runtime": pathlib.Path(args.target_runtime_evidence),
        "recovery": pathlib.Path(args.recovery_evidence),
        "provider": pathlib.Path(args.provider_evidence),
        "visual": pathlib.Path(args.visual_evidence),
        "browser": pathlib.Path(args.browser_evidence),
    }
    artifacts = {name: _load(path) for name, path in paths.items()}
    for name, value in artifacts.items():
        _reject_sensitive_keys(value, f"{name}_evidence")

    source, image = _candidate_from_release(artifacts["release"])
    release_sha256 = _sha256(paths["release"])
    runtime_sha256 = _sha256(paths["runtime"])
    _audit_runtime(artifacts["runtime"], source, image)
    _audit_recovery(artifacts["recovery"], source, image, release_sha256, runtime_sha256)
    _audit_provider(artifacts["provider"], source, image)
    _audit_visual(artifacts["visual"], source, image)
    _audit_browser(artifacts["browser"], source, image)

    bindings = {
        "release_evidence_sha256": release_sha256,
        "target_runtime_evidence_sha256": runtime_sha256,
        "recovery_evidence_sha256": _sha256(paths["recovery"]),
        "provider_evidence_sha256": _sha256(paths["provider"]),
        "visual_evidence_sha256": _sha256(paths["visual"]),
        "browser_evidence_sha256": _sha256(paths["browser"]),
    }
    if args.final_evidence:
        final = _load(pathlib.Path(args.final_evidence))
        _reject_sensitive_keys(final, "final_evidence")
        _audit_final(final, source, image, bindings)

    return {"source_revision": source, "image": image, **bindings}


def parser() -> argparse.ArgumentParser:
    root = argparse.ArgumentParser(description=__doc__)
    root.add_argument("--release-evidence", required=True)
    root.add_argument("--target-runtime-evidence", required=True)
    root.add_argument("--recovery-evidence", required=True)
    root.add_argument("--provider-evidence", required=True)
    root.add_argument("--visual-evidence", required=True)
    root.add_argument("--browser-evidence", required=True)
    root.add_argument("--final-evidence")
    return root


def main() -> int:
    args = parser().parse_args()
    try:
        result = audit(args)
    except AuditError as exc:
        print(f"First-Stable evidence audit error: {exc}", file=sys.stderr)
        return 2
    print("GoreeCloud Search first-Stable evidence safety audit passed.")
    print(f"Candidate source: {result['source_revision']}")
    print(f"Candidate image: {result['image']}")
    print("Production cutover authorized by this audit: false")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
