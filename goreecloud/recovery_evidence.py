# SPDX-License-Identifier: AGPL-3.0-or-later
"""Create and validate sanitized GoreeCloud Search recovery evidence.

This tool binds target-host restore/rollback evidence to the exact candidate image,
source revision, release evidence, target-runtime evidence, and known-good rollback
baseline. It never performs a restore and never authorizes production cutover.
Actual target-host recovery work must be completed separately before a generated
template can pass validation.
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
IMAGE_RE = re.compile(r"^ghcr\.io/goreecloud/goreecloud-search@sha256:[0-9a-f]{64}$")
FORBIDDEN_EVIDENCE_KEYS = {
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
}


class EvidenceError(ValueError):
    """Raised when recovery evidence is incomplete, inconsistent, or unsafe."""


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


def _require_mapping(value: Any, label: str) -> dict[str, Any]:
    if not isinstance(value, dict):
        raise EvidenceError(f"{label} must be a JSON object")
    return value


def _require_nonempty(value: Any, label: str) -> str:
    if not isinstance(value, str) or not value.strip():
        raise EvidenceError(f"{label} must be a non-empty string")
    return value.strip()


def _require_true(value: Any, label: str) -> None:
    if value is not True:
        raise EvidenceError(f"{label} must be true after the target acceptance step is completed")


def _require_false(value: Any, label: str) -> None:
    if value is not False:
        raise EvidenceError(f"{label} must remain false")


def _require_iso_time(value: Any, label: str) -> str:
    text = _require_nonempty(value, label)
    try:
        parsed = dt.datetime.fromisoformat(text.replace("Z", "+00:00"))
    except ValueError as exc:
        raise EvidenceError(f"{label} must be ISO-8601") from exc
    if parsed.tzinfo is None:
        raise EvidenceError(f"{label} must include a timezone")
    return text


def _require_sha(value: Any, label: str) -> str:
    text = _require_nonempty(value, label)
    if not SHA_RE.fullmatch(text):
        raise EvidenceError(f"{label} must be a lowercase 40-character Git SHA")
    return text


def _require_digest(value: Any, label: str) -> str:
    text = _require_nonempty(value, label)
    if not DIGEST_RE.fullmatch(text):
        raise EvidenceError(f"{label} must be a lowercase 64-character SHA-256 digest")
    return text


def _require_image(value: Any, label: str) -> str:
    text = _require_nonempty(value, label)
    if not IMAGE_RE.fullmatch(text):
        raise EvidenceError(
            f"{label} must be ghcr.io/goreecloud/goreecloud-search pinned by sha256 digest"
        )
    return text


def _reject_sensitive_keys(value: Any, path: str = "evidence") -> None:
    if isinstance(value, dict):
        for key, child in value.items():
            normalized = str(key).strip().lower()
            if normalized in FORBIDDEN_EVIDENCE_KEYS:
                raise EvidenceError(f"Sensitive field name is not allowed in recovery evidence: {path}.{key}")
            _reject_sensitive_keys(child, f"{path}.{key}")
    elif isinstance(value, list):
        for index, child in enumerate(value):
            _reject_sensitive_keys(child, f"{path}[{index}]")


def _binding(
    release_path: pathlib.Path,
    runtime_path: pathlib.Path,
    baseline_path: pathlib.Path,
) -> dict[str, str]:
    release = _load(release_path)
    runtime = _load(runtime_path)
    baseline = _load(baseline_path)

    if release.get("schema_version") != 1 or release.get("product") != "GoreeCloud Search":
        raise EvidenceError("Release evidence is not a GoreeCloud Search schema-version 1 artifact")
    candidate = _require_mapping(release.get("candidate"), "release candidate")
    source = _require_sha(candidate.get("source_revision"), "release candidate source_revision")
    image = _require_image(candidate.get("image"), "release candidate image")
    if candidate.get("isolated_runtime_acceptance") != "passed":
        raise EvidenceError("Release candidate isolated runtime acceptance must have passed")
    rollback_scope = _require_mapping(release.get("rollback_scope"), "release rollback_scope")
    if rollback_scope.get("image_level_rehearsal") != "passed":
        raise EvidenceError("Release image-level rollback rehearsal must have passed")
    _require_false(rollback_scope.get("production_cutover_authorized"), "release production_cutover_authorized")

    if runtime.get("schema_version") != 1 or runtime.get("product") != "GoreeCloud Search":
        raise EvidenceError("Target runtime evidence is not a GoreeCloud Search schema-version 1 artifact")
    runtime_container = _require_mapping(runtime.get("container_runtime"), "target container_runtime")
    runtime_scope = _require_mapping(runtime.get("scope"), "target runtime scope")
    _require_true(runtime_scope.get("target_runtime_identity_verified"), "target runtime identity")
    _require_false(runtime_scope.get("production_cutover_authorized"), "target production_cutover_authorized")
    runtime_source = _require_sha(
        runtime_container.get("expected_source_revision"), "target expected_source_revision"
    )
    runtime_image = _require_image(runtime_container.get("expected_image"), "target expected_image")
    if runtime_source != source or runtime_image != image:
        raise EvidenceError("Release and target-runtime evidence refer to different candidates")

    if baseline.get("schema_version") != 1:
        raise EvidenceError("Rollback baseline schema_version must be 1")
    baseline_source = _require_sha(baseline.get("source_revision"), "rollback baseline source_revision")
    baseline_image = _require_image(baseline.get("image"), "rollback baseline image")
    if baseline_image == image:
        raise EvidenceError("Candidate image must differ from the recorded rollback baseline")

    return {
        "candidate_source": source,
        "candidate_image": image,
        "rollback_source": baseline_source,
        "rollback_image": baseline_image,
        "release_sha256": _sha256(release_path),
        "runtime_sha256": _sha256(runtime_path),
        "baseline_sha256": _sha256(baseline_path),
    }


def build_template(args: argparse.Namespace) -> dict[str, Any]:
    release_path = pathlib.Path(args.release_evidence)
    runtime_path = pathlib.Path(args.target_runtime_evidence)
    baseline_path = pathlib.Path(args.rollback_baseline)
    binding = _binding(release_path, runtime_path, baseline_path)

    evidence = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "environment": "goreecloud-vps-01",
        "generated_at": dt.datetime.now(dt.timezone.utc).isoformat().replace("+00:00", "Z"),
        "candidate": {
            "source_revision": binding["candidate_source"],
            "image": binding["candidate_image"],
        },
        "known_good_rollback": {
            "source_revision": binding["rollback_source"],
            "image": binding["rollback_image"],
        },
        "artifact_bindings": {
            "release_evidence_sha256": binding["release_sha256"],
            "target_runtime_evidence_sha256": binding["runtime_sha256"],
            "rollback_baseline_sha256": binding["baseline_sha256"],
        },
        "backup": {
            "system": "",
            "snapshot_reference": "",
            "captured_at": "",
            "scope": {
                "stack_definition": False,
                "search_settings": False,
                "protected_runtime_configuration_recovery_path": False,
                "caddy_search_route": False,
            },
            "cache_policy": "rebuildable-non-authoritative",
        },
        "restore": {
            "isolated_location": "",
            "production_modified": False,
            "stack_definition_restored": False,
            "search_settings_restored": False,
            "protected_runtime_configuration_recovery_verified": False,
            "caddy_search_route_copy_restored": False,
            "compose_validation_passed": False,
            "runtime_recreated": False,
            "runtime_health_passed": False,
            "runtime_identity_passed": False,
            "target_acceptance_passed": False,
        },
        "rollback": {
            "mode": "",
            "known_good_image_available": False,
            "previous_runtime_configuration_preserved": False,
            "previous_caddy_route_preserved": False,
            "rollback_procedure_documented": False,
            "isolated_rollback_rehearsal_passed": False,
            "production_route_rollback_tested": False,
            "equivalent_verified_rollback_evidence": False,
        },
        "monitoring": {
            "monitor_identity": "",
            "availability_monitor_verified": False,
            "alert_delivery_verified": False,
        },
        "scope": {
            "application_level_restore_tested": False,
            "monitoring_and_alerting_verified": False,
            "rollback_evidence_verified": False,
            "production_cutover_authorized": False,
            "statement": (
                "This artifact is incomplete until an actual isolated restore, monitoring check, "
                "and rollback-evidence review are performed. It must never contain reusable secrets "
                "or independently authorize production cutover."
            ),
        },
    }
    _write(pathlib.Path(args.output), evidence)
    return evidence


def validate_evidence(args: argparse.Namespace) -> dict[str, Any]:
    evidence_path = pathlib.Path(args.evidence)
    release_path = pathlib.Path(args.release_evidence)
    runtime_path = pathlib.Path(args.target_runtime_evidence)
    baseline_path = pathlib.Path(args.rollback_baseline)
    evidence = _load(evidence_path)
    _reject_sensitive_keys(evidence)
    binding = _binding(release_path, runtime_path, baseline_path)

    if evidence.get("schema_version") != 1 or evidence.get("product") != "GoreeCloud Search":
        raise EvidenceError("Recovery evidence must be a GoreeCloud Search schema-version 1 artifact")
    if evidence.get("environment") != "goreecloud-vps-01":
        raise EvidenceError("Recovery evidence environment must be goreecloud-vps-01")
    _require_iso_time(evidence.get("generated_at"), "recovery generated_at")

    candidate = _require_mapping(evidence.get("candidate"), "recovery candidate")
    if _require_sha(candidate.get("source_revision"), "recovery candidate source_revision") != binding[
        "candidate_source"
    ]:
        raise EvidenceError("Recovery candidate source does not match release evidence")
    if _require_image(candidate.get("image"), "recovery candidate image") != binding["candidate_image"]:
        raise EvidenceError("Recovery candidate image does not match release evidence")

    rollback_identity = _require_mapping(evidence.get("known_good_rollback"), "known_good_rollback")
    if _require_sha(rollback_identity.get("source_revision"), "rollback source_revision") != binding[
        "rollback_source"
    ]:
        raise EvidenceError("Recovery rollback source does not match the source-controlled baseline")
    if _require_image(rollback_identity.get("image"), "rollback image") != binding["rollback_image"]:
        raise EvidenceError("Recovery rollback image does not match the source-controlled baseline")

    artifacts = _require_mapping(evidence.get("artifact_bindings"), "artifact_bindings")
    expected_artifacts = {
        "release_evidence_sha256": binding["release_sha256"],
        "target_runtime_evidence_sha256": binding["runtime_sha256"],
        "rollback_baseline_sha256": binding["baseline_sha256"],
    }
    for key, expected in expected_artifacts.items():
        actual = _require_digest(artifacts.get(key), f"artifact_bindings.{key}")
        if actual != expected:
            raise EvidenceError(f"artifact_bindings.{key} does not match the supplied artifact")

    backup = _require_mapping(evidence.get("backup"), "backup")
    _require_nonempty(backup.get("system"), "backup.system")
    _require_nonempty(backup.get("snapshot_reference"), "backup.snapshot_reference")
    _require_iso_time(backup.get("captured_at"), "backup.captured_at")
    backup_scope = _require_mapping(backup.get("scope"), "backup.scope")
    for key in (
        "stack_definition",
        "search_settings",
        "protected_runtime_configuration_recovery_path",
        "caddy_search_route",
    ):
        _require_true(backup_scope.get(key), f"backup.scope.{key}")
    if backup.get("cache_policy") != "rebuildable-non-authoritative":
        raise EvidenceError("Search cache must be classified as rebuildable-non-authoritative")

    restore = _require_mapping(evidence.get("restore"), "restore")
    _require_nonempty(restore.get("isolated_location"), "restore.isolated_location")
    _require_false(restore.get("production_modified"), "restore.production_modified")
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
        _require_true(restore.get(key), f"restore.{key}")

    rollback = _require_mapping(evidence.get("rollback"), "rollback")
    for key in (
        "known_good_image_available",
        "previous_runtime_configuration_preserved",
        "previous_caddy_route_preserved",
        "rollback_procedure_documented",
        "isolated_rollback_rehearsal_passed",
    ):
        _require_true(rollback.get(key), f"rollback.{key}")
    mode = _require_nonempty(rollback.get("mode"), "rollback.mode")
    if mode == "production-route-rehearsal":
        _require_true(rollback.get("production_route_rollback_tested"), "rollback.production_route_rollback_tested")
    elif mode == "equivalent-verified-evidence":
        _require_true(
            rollback.get("equivalent_verified_rollback_evidence"),
            "rollback.equivalent_verified_rollback_evidence",
        )
    else:
        raise EvidenceError(
            "rollback.mode must be production-route-rehearsal or equivalent-verified-evidence"
        )

    monitoring = _require_mapping(evidence.get("monitoring"), "monitoring")
    if monitoring.get("monitor_identity") != "GoreeCloud Search":
        raise EvidenceError("monitoring.monitor_identity must be GoreeCloud Search")
    _require_true(monitoring.get("availability_monitor_verified"), "monitoring.availability_monitor_verified")
    _require_true(monitoring.get("alert_delivery_verified"), "monitoring.alert_delivery_verified")

    scope = _require_mapping(evidence.get("scope"), "scope")
    _require_true(scope.get("application_level_restore_tested"), "scope.application_level_restore_tested")
    _require_true(scope.get("monitoring_and_alerting_verified"), "scope.monitoring_and_alerting_verified")
    _require_true(scope.get("rollback_evidence_verified"), "scope.rollback_evidence_verified")
    _require_false(scope.get("production_cutover_authorized"), "scope.production_cutover_authorized")
    _require_nonempty(scope.get("statement"), "scope.statement")

    return evidence


def parser() -> argparse.ArgumentParser:
    root = argparse.ArgumentParser(description=__doc__)
    commands = root.add_subparsers(dest="command", required=True)

    template = commands.add_parser("template", help="Create an incomplete candidate-bound recovery template")
    template.add_argument("--release-evidence", required=True)
    template.add_argument("--target-runtime-evidence", required=True)
    template.add_argument("--rollback-baseline", required=True)
    template.add_argument("--output", required=True)

    validate = commands.add_parser("validate", help="Validate completed target recovery evidence")
    validate.add_argument("--evidence", required=True)
    validate.add_argument("--release-evidence", required=True)
    validate.add_argument("--target-runtime-evidence", required=True)
    validate.add_argument("--rollback-baseline", required=True)
    return root


def main() -> int:
    args = parser().parse_args()
    try:
        if args.command == "template":
            evidence = build_template(args)
            print(f"Incomplete recovery template written: {args.output}")
            print(f"Candidate: {evidence['candidate']['image']}")
            print("Production cutover authorized: false")
            return 0
        if args.command == "validate":
            validate_evidence(args)
            print("GoreeCloud Search target recovery evidence passed validation.")
            print("Production cutover authorized by this artifact: false")
            return 0
    except EvidenceError as exc:
        print(f"Recovery evidence error: {exc}", file=sys.stderr)
        return 2
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
