# SPDX-License-Identifier: AGPL-3.0-or-later
"""Report first-Stable evidence readiness for frozen GoreeCloud Search candidate #07.

The report validates only supplied evidence. It never creates missing evidence, changes
production, authorizes production cutover, or authorizes Stable promotion.
"""

from __future__ import annotations

import argparse
from dataclasses import dataclass, field
import json
import pathlib
import sys
from typing import Any

import first_stable_candidate_07_integrity_audit as integrity_audit


# The readiness layer intentionally composes the established audit modules rather than
# duplicating their schema logic.
# pylint: disable=protected-access
candidate_audit = integrity_audit.candidate_audit
base_audit = integrity_audit.base_audit
review_audit = candidate_audit.review_audit
AUDIT_ERRORS = (
    integrity_audit.AuditError,
    candidate_audit.AuditError,
    base_audit.AuditError,
    review_audit.AuditError,
)
REQUIRED_INPUTS = (
    ("release", "release_evidence", "release_evidence"),
    ("target_runtime", "target_runtime_evidence", "target_runtime_evidence"),
    ("recovery", "recovery_evidence", "recovery_evidence"),
    ("rollback_baseline", "rollback_baseline", "rollback_baseline"),
    ("provider", "provider_evidence", "provider_evidence"),
    ("visual", "visual_evidence", "visual_evidence"),
    ("browser", "browser_evidence", "browser_evidence"),
)


@dataclass
class ReadinessContext:
    """Mutable validation context used while classifying supplied evidence."""

    artifacts: dict[str, dict[str, Any]] = field(default_factory=dict)
    values: dict[str, dict[str, Any] | None] = field(default_factory=dict)
    paths: dict[str, pathlib.Path | None] = field(default_factory=dict)
    source: str | None = None
    image: str | None = None
    rollback_source: str | None = None
    rollback_image: str | None = None
    rollback_sha256: str | None = None


def _entry(
    status: str,
    path: pathlib.Path | None = None,
    detail: str | None = None,
) -> dict[str, Any]:
    value: dict[str, Any] = {"status": status}
    if path is not None:
        value["path"] = str(path)
    if detail:
        value["detail"] = detail
    return value


def _load_artifact(
    raw_path: str | None, label: str
) -> tuple[dict[str, Any], dict[str, Any] | None, pathlib.Path | None]:
    if not raw_path:
        return _entry("missing", detail="input was not supplied"), None, None
    path = pathlib.Path(raw_path)
    if not path.is_file():
        return _entry("missing", path, "file does not exist"), None, path
    try:
        value = base_audit._load(path)
        base_audit._reject_sensitive_keys(value, label)
    except AUDIT_ERRORS as exc:
        return _entry("invalid", path, str(exc)), None, path
    return _entry("present", path), value, path


def _set_valid(entry: dict[str, Any]) -> None:
    entry["status"] = "valid"
    entry.pop("detail", None)


def _set_invalid(entry: dict[str, Any], exc: Exception) -> None:
    entry["status"] = "invalid"
    entry["detail"] = str(exc)


def _set_pending(entry: dict[str, Any], detail: str) -> None:
    if entry["status"] == "present":
        entry["status"] = "pending_dependency"
        entry["detail"] = detail


def _all_valid(artifacts: dict[str, dict[str, Any]]) -> bool:
    return all(artifacts[name]["status"] == "valid" for name, _, _ in REQUIRED_INPUTS)


def _audit_namespace(args: argparse.Namespace, *, include_final: bool) -> argparse.Namespace:
    return argparse.Namespace(
        release_evidence=args.release_evidence,
        target_runtime_evidence=args.target_runtime_evidence,
        recovery_evidence=args.recovery_evidence,
        rollback_baseline=args.rollback_baseline,
        provider_evidence=args.provider_evidence,
        visual_evidence=args.visual_evidence,
        browser_evidence=args.browser_evidence,
        final_evidence=args.final_evidence if include_final else None,
    )


def _load_inputs(args: argparse.Namespace) -> tuple[ReadinessContext, pathlib.Path | None]:
    context = ReadinessContext()
    for name, attribute, sensitive_label in REQUIRED_INPUTS:
        entry, value, path = _load_artifact(getattr(args, attribute), sensitive_label)
        context.artifacts[name] = entry
        context.values[name] = value
        context.paths[name] = path

    final_entry, _, final_path = _load_artifact(args.final_evidence, "final_evidence")
    if not args.final_evidence:
        final_entry = _entry(
            "not_supplied",
            detail="final manifest is optional until assembly",
        )
        final_path = None
    context.artifacts["final_manifest"] = final_entry
    return context, final_path


def _validate_baseline(context: ReadinessContext) -> None:
    entry = context.artifacts["rollback_baseline"]
    baseline = context.values["rollback_baseline"]
    path = context.paths["rollback_baseline"]
    if entry["status"] != "present" or baseline is None or path is None:
        return
    try:
        context.rollback_sha256 = base_audit._sha256(path)
        context.rollback_source, context.rollback_image = (
            candidate_audit._audit_frozen_rollback_baseline(
                baseline,
                context.rollback_sha256,
            )
        )
        _set_valid(entry)
    except AUDIT_ERRORS as exc:
        _set_invalid(entry, exc)


def _validate_release(context: ReadinessContext) -> None:
    entry = context.artifacts["release"]
    release = context.values["release"]
    baseline = context.values["rollback_baseline"]
    if entry["status"] != "present" or release is None:
        return
    try:
        context.source, context.image = base_audit._candidate_from_release(release)
        candidate_audit._require_frozen_candidate(context.source, context.image)
        if context.artifacts["rollback_baseline"]["status"] != "valid" or baseline is None:
            _set_pending(entry, "valid rollback baseline is required")
            return
        candidate_audit._audit_frozen_release(
            release,
            context.source,
            context.image,
            baseline,
        )
        _set_valid(entry)
    except AUDIT_ERRORS as exc:
        _set_invalid(entry, exc)


def _release_ready(context: ReadinessContext) -> bool:
    return (
        context.artifacts["release"]["status"] == "valid"
        and context.source is not None
        and context.image is not None
    )


def _validate_runtime(context: ReadinessContext) -> None:
    entry = context.artifacts["target_runtime"]
    runtime = context.values["target_runtime"]
    if entry["status"] != "present" or runtime is None:
        return
    if not _release_ready(context):
        _set_pending(entry, "valid release evidence is required")
        return
    try:
        candidate_audit._audit_frozen_runtime(runtime, context.source, context.image)
        _set_valid(entry)
    except AUDIT_ERRORS as exc:
        _set_invalid(entry, exc)


def _validate_provider(context: ReadinessContext) -> None:
    entry = context.artifacts["provider"]
    provider = context.values["provider"]
    if entry["status"] != "present" or provider is None:
        return
    if not _release_ready(context):
        _set_pending(entry, "valid release evidence is required")
        return
    try:
        integrity_audit._audit_frozen_provider(provider, context.source, context.image)
        _set_valid(entry)
    except AUDIT_ERRORS as exc:
        _set_invalid(entry, exc)


def _validate_visual(context: ReadinessContext) -> None:
    entry = context.artifacts["visual"]
    visual = context.values["visual"]
    if entry["status"] != "present" or visual is None:
        return
    if not _release_ready(context):
        _set_pending(entry, "valid release evidence is required")
        return
    try:
        review_audit._visual_summary(visual, context.source, context.image)
        _set_valid(entry)
    except AUDIT_ERRORS as exc:
        _set_invalid(entry, exc)


def _validate_browser(context: ReadinessContext) -> None:
    entry = context.artifacts["browser"]
    browser = context.values["browser"]
    if entry["status"] != "present" or browser is None:
        return
    if not _release_ready(context):
        _set_pending(entry, "valid release evidence is required")
        return
    try:
        review_audit._browser_summary(browser, context.source, context.image)
        _set_valid(entry)
    except AUDIT_ERRORS as exc:
        _set_invalid(entry, exc)


def _validate_recovery(context: ReadinessContext) -> None:
    entry = context.artifacts["recovery"]
    recovery = context.values["recovery"]
    if entry["status"] != "present" or recovery is None:
        return
    dependencies_ready = (
        _release_ready(context)
        and context.artifacts["target_runtime"]["status"] == "valid"
        and context.artifacts["rollback_baseline"]["status"] == "valid"
        and context.rollback_source is not None
        and context.rollback_image is not None
        and context.rollback_sha256 is not None
        and context.paths["release"] is not None
        and context.paths["target_runtime"] is not None
    )
    if not dependencies_ready:
        _set_pending(
            entry,
            "valid release, target-runtime, and rollback-baseline evidence is required",
        )
        return
    try:
        release_path = context.paths["release"]
        runtime_path = context.paths["target_runtime"]
        assert release_path is not None and runtime_path is not None
        release_sha256 = base_audit._sha256(release_path)
        runtime_sha256 = base_audit._sha256(runtime_path)
        base_audit._audit_recovery(
            recovery,
            context.source,
            context.image,
            release_sha256,
            runtime_sha256,
        )
        candidate_audit._audit_recovery_baseline_binding(
            recovery,
            context.rollback_source,
            context.rollback_image,
            context.rollback_sha256,
        )
        _set_valid(entry)
    except AUDIT_ERRORS as exc:
        _set_invalid(entry, exc)


def _validate_required(context: ReadinessContext) -> None:
    _validate_baseline(context)
    _validate_release(context)
    _validate_runtime(context)
    _validate_provider(context)
    _validate_visual(context)
    _validate_browser(context)
    _validate_recovery(context)


def _cross_binding(
    context: ReadinessContext,
    args: argparse.Namespace,
) -> dict[str, Any]:
    if not _all_valid(context.artifacts):
        return _entry("not_run", detail="all seven required inputs must validate first")
    try:
        integrity_audit.audit(_audit_namespace(args, include_final=False))
    except AUDIT_ERRORS as exc:
        return _entry("invalid", detail=str(exc))
    return _entry(
        "valid",
        detail="six companion artifacts and rollback provenance agree",
    )


def _validate_final(
    context: ReadinessContext,
    args: argparse.Namespace,
    cross_binding: dict[str, Any],
) -> None:
    entry = context.artifacts["final_manifest"]
    if not args.final_evidence or entry["status"] != "present":
        return
    if cross_binding["status"] != "valid":
        _set_pending(entry, "valid companion cross-binding is required")
        return
    try:
        integrity_audit.audit(_audit_namespace(args, include_final=True))
        _set_valid(entry)
    except AUDIT_ERRORS as exc:
        _set_invalid(entry, exc)


def _readiness_status(
    artifacts: dict[str, dict[str, Any]],
    cross_binding: dict[str, Any],
) -> str:
    if not _all_valid(artifacts) or cross_binding["status"] != "valid":
        return "blocked"
    final_status = artifacts["final_manifest"]["status"]
    if final_status == "not_supplied":
        return "ready_for_final_manifest"
    if final_status == "valid":
        return "ready_for_governance_review"
    return "blocked"


def _operator_action(status: str) -> str:
    if status == "ready_for_final_manifest":
        return (
            "Assemble the schema-version 2 final manifest from the six validated companion "
            "artifacts, then rerun with --final-evidence."
        )
    if status == "ready_for_governance_review":
        return (
            "Submit the validated evidence set for the separate release-governance decision. "
            "This report does not authorize production cutover or Stable promotion."
        )
    return (
        "Complete or correct every required artifact marked missing, invalid, or "
        "pending_dependency, then rerun the readiness report."
    )


def build_report(args: argparse.Namespace) -> dict[str, Any]:
    """Validate supplied evidence and return a non-authorizing readiness report."""
    context, final_path = _load_inputs(args)
    _validate_required(context)
    cross_binding = _cross_binding(context, args)
    _validate_final(context, args, cross_binding)
    status = _readiness_status(context.artifacts, cross_binding)

    report = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "candidate": {
            "source_revision": context.source or candidate_audit.FROZEN_SOURCE,
            "image": context.image or candidate_audit.FROZEN_IMAGE,
        },
        "status": status,
        "artifacts": context.artifacts,
        "cross_binding": cross_binding,
        "final_contract": {
            "companion_artifact_count": 6,
            "rollback_baseline_is_supporting_provenance": True,
        },
        "production_cutover_authorized": False,
        "stable_promotion_authorized": False,
        "operator_action": _operator_action(status),
    }
    if final_path is not None:
        report["final_manifest_path"] = str(final_path)
    return report


def parser() -> argparse.ArgumentParser:
    """Build the readiness-report command-line parser."""
    root = argparse.ArgumentParser(description=__doc__)
    root.add_argument("--release-evidence")
    root.add_argument("--target-runtime-evidence")
    root.add_argument("--recovery-evidence")
    root.add_argument("--rollback-baseline")
    root.add_argument("--provider-evidence")
    root.add_argument("--visual-evidence")
    root.add_argument("--browser-evidence")
    root.add_argument("--final-evidence")
    root.add_argument("--json", action="store_true", dest="json_output")
    return root


def _print_human(report: dict[str, Any]) -> None:
    print(f"GoreeCloud Search first-Stable readiness: {report['status']}")
    print(f"Candidate source: {report['candidate']['source_revision']}")
    print(f"Candidate image: {report['candidate']['image']}")
    for name, entry in report["artifacts"].items():
        detail = f" — {entry['detail']}" if entry.get("detail") else ""
        print(f"{name}: {entry['status']}{detail}")
    print(f"cross_binding: {report['cross_binding']['status']}")
    print("Production cutover authorized by this report: false")
    print("Stable promotion authorized by this report: false")
    print(f"Next action: {report['operator_action']}")


def exit_code(status: str) -> int:
    """Map readiness state to a fail-closed process exit code."""
    if status == "ready_for_governance_review":
        return 0
    if status == "ready_for_final_manifest":
        return 3
    return 2


def main() -> int:
    """Generate a fail-closed readiness report."""
    args = parser().parse_args()
    report = build_report(args)
    if args.json_output:
        print(json.dumps(report, indent=2, sort_keys=True))
    else:
        _print_human(report)
    return exit_code(report["status"])


if __name__ == "__main__":
    raise SystemExit(main())
