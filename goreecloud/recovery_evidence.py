# SPDX-License-Identifier: AGPL-3.0-or-later
"""Create and validate sanitized GoreeCloud Search RC 09 recovery evidence.

This master-side tool binds recovery, rollback, monitoring, and alert-delivery
evidence to the immutable RC 09 candidate, target-runtime evidence, and the
source-controlled rollback baseline. It never performs recovery work and never
authorizes target-host changes, production cutover, or Stable promotion.
"""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import pathlib
import sys
from typing import Any

import first_stable_evidence_audit as base_audit
import release_candidate_09_audit as rc09_audit

# pylint: disable=protected-access

RC09_SOURCE = rc09_audit.RC09_SOURCE
RC09_IMAGE = rc09_audit.RC09_IMAGE
ROLLBACK_SOURCE = rc09_audit.ROLLBACK_SOURCE
ROLLBACK_IMAGE = rc09_audit.ROLLBACK_IMAGE
ROLLBACK_SHA256 = rc09_audit.ROLLBACK_SHA256

FORBIDDEN_EVIDENCE_KEYS = frozenset(
    {
        *base_audit.FORBIDDEN_KEYS,
        "api_key",
        "api_keys",
    }
)


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


def _reject_sensitive_keys(value: Any, path: str = "evidence") -> None:
    if isinstance(value, dict):
        for key, child in value.items():
            normalized = str(key).strip().lower()
            if normalized in FORBIDDEN_EVIDENCE_KEYS:
                raise EvidenceError(
                    f"Sensitive or unnecessary field is not allowed: {path}.{key}"
                )
            _reject_sensitive_keys(child, f"{path}.{key}")
    elif isinstance(value, list):
        for index, child in enumerate(value):
            _reject_sensitive_keys(child, f"{path}[{index}]")


def _audit_companions(
    release_path: pathlib.Path,
    runtime_path: pathlib.Path,
    baseline_path: pathlib.Path,
) -> dict[str, str]:
    release = _load(release_path)
    runtime = _load(runtime_path)
    baseline = _load(baseline_path)

    for label, value in (
        ("release_evidence", release),
        ("target_runtime_evidence", runtime),
        ("rollback_baseline", baseline),
    ):
        _reject_sensitive_keys(value, label)

    release_digest = _sha256(release_path)
    runtime_digest = _sha256(runtime_path)
    baseline_digest = _sha256(baseline_path)

    try:
        rc09_audit._audit_baseline(baseline, baseline_digest)
        rc09_audit._audit_release(release, baseline)
        base_audit._audit_runtime(runtime, RC09_SOURCE, RC09_IMAGE)
        rc09_audit._audit_runtime_version(runtime)
    except (base_audit.AuditError, rc09_audit.AuditError) as exc:
        raise EvidenceError(str(exc)) from exc

    return {
        "candidate_source": RC09_SOURCE,
        "candidate_image": RC09_IMAGE,
        "rollback_source": ROLLBACK_SOURCE,
        "rollback_image": ROLLBACK_IMAGE,
        "release_sha256": release_digest,
        "runtime_sha256": runtime_digest,
        "baseline_sha256": baseline_digest,
    }


def build_template(args: argparse.Namespace) -> dict[str, Any]:
    """Write an intentionally incomplete RC 09 recovery-evidence template."""
    binding = _audit_companions(
        pathlib.Path(args.release_evidence),
        pathlib.Path(args.target_runtime_evidence),
        pathlib.Path(args.rollback_baseline),
    )
    evidence = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "environment": "goreecloud-vps-01",
        "generated_at": dt.datetime.now(dt.timezone.utc).isoformat().replace(
            "+00:00", "Z"
        ),
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
                "This artifact is incomplete until actual isolated restore, rollback "
                "review, availability-monitor verification, and approved alert-delivery "
                "verification are completed. It does not authorize target-host changes, "
                "production cutover, or Stable promotion."
            ),
        },
    }
    _reject_sensitive_keys(evidence)
    _write(pathlib.Path(args.output), evidence)
    return evidence


def validate_evidence(args: argparse.Namespace) -> dict[str, Any]:
    """Validate completed evidence against the exact RC 09 audit contract."""
    evidence_path = pathlib.Path(args.evidence)
    release_path = pathlib.Path(args.release_evidence)
    runtime_path = pathlib.Path(args.target_runtime_evidence)
    baseline_path = pathlib.Path(args.rollback_baseline)

    evidence = _load(evidence_path)
    _reject_sensitive_keys(evidence)
    binding = _audit_companions(release_path, runtime_path, baseline_path)

    try:
        base_audit._audit_recovery(
            evidence,
            RC09_SOURCE,
            RC09_IMAGE,
            binding["release_sha256"],
            binding["runtime_sha256"],
        )
        rc09_audit._audit_recovery_baseline(
            evidence,
            binding["baseline_sha256"],
        )
        scope = base_audit._mapping(evidence.get("scope"), "recovery scope")
        base_audit._nonempty(scope.get("statement"), "recovery scope.statement")
    except (base_audit.AuditError, rc09_audit.AuditError) as exc:
        raise EvidenceError(str(exc)) from exc

    rollback = base_audit._mapping(evidence.get("rollback"), "recovery rollback")
    if rollback.get("mode") == "production-route-rehearsal":
        if rollback.get("equivalent_verified_rollback_evidence") is not False:
            raise EvidenceError(
                "production-route-rehearsal evidence must keep the equivalent-evidence "
                "flag false"
            )
    elif rollback.get("mode") == "equivalent-verified-evidence":
        if rollback.get("production_route_rollback_tested") is not False:
            raise EvidenceError(
                "equivalent-verified-evidence must keep the production-route flag false"
            )

    return evidence


def parser() -> argparse.ArgumentParser:
    """Build the recovery-evidence command line."""
    root = argparse.ArgumentParser(description=__doc__)
    commands = root.add_subparsers(dest="command", required=True)

    template = commands.add_parser(
        "template",
        help="Create an incomplete RC 09 recovery-evidence template",
    )
    template.add_argument("--release-evidence", required=True)
    template.add_argument("--target-runtime-evidence", required=True)
    template.add_argument("--rollback-baseline", required=True)
    template.add_argument("--output", required=True)

    validate = commands.add_parser(
        "validate",
        help="Validate completed RC 09 recovery evidence",
    )
    validate.add_argument("--evidence", required=True)
    validate.add_argument("--release-evidence", required=True)
    validate.add_argument("--target-runtime-evidence", required=True)
    validate.add_argument("--rollback-baseline", required=True)
    return root


def main() -> int:
    """Run the recovery-evidence CLI."""
    args = parser().parse_args()
    try:
        if args.command == "template":
            evidence = build_template(args)
            print(f"Incomplete RC 09 recovery template written: {args.output}")
            print(f"Candidate source: {evidence['candidate']['source_revision']}")
            print(f"Candidate image: {evidence['candidate']['image']}")
        elif args.command == "validate":
            validate_evidence(args)
            print("GoreeCloud Search RC 09 recovery evidence passed validation.")
        else:
            return 2
    except EvidenceError as exc:
        print(f"RC 09 recovery evidence error: {exc}", file=sys.stderr)
        return 2

    print("Target-host changes authorized by this artifact: false")
    print("Production cutover authorized by this artifact: false")
    print("Stable promotion authorized by this artifact: false")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
