# SPDX-License-Identifier: AGPL-3.0-or-later
"""Audit the current GoreeCloud Search Release Candidate 09 evidence set.

This master-side audit pins every supplied acceptance artifact to the immutable RC 09
source and image while preserving Candidate 07 as historical evidence. It validates
release, target-runtime, recovery, provider, visual/device, Browser, and optional final
manifest evidence. It never authorizes production cutover or Stable promotion.
"""

from __future__ import annotations

import argparse
import pathlib
import sys
from typing import Any

import final_review_evidence_audit as review_audit
import first_stable_evidence_audit as base_audit

# pylint: disable=protected-access

RC09_REQUEST_ID = "first-stable-2026-08-23-09"
RC09_SEQUENCE = 9
RC09_REVIEWED_BASE = "61027b05ad8fef33d68078f6ec4498e1ac675112"
RC09_SOURCE = "ef2a7fb4e96e1f28bef53fb8bc766a1bed96b45e"
RC09_IMAGE = (
    "ghcr.io/goreecloud/goreecloud-search@sha256:"
    "1d6f7e9cc2b10dc4babf5a40f281b588f29aa12ba26307583ae74a1655971560"
)
RC09_OCI_VERSION = "2026.8.23-ef2a7fb"
ROLLBACK_ENVIRONMENT = "goreecloud-vps-01-production"
ROLLBACK_RECORDED_AT = "2026-08-17T18:10:00-05:00"
ROLLBACK_SOURCE = "3584da535f7ed7c3b4b8dc73cf0424fb4bdf1949"
ROLLBACK_IMAGE = (
    "ghcr.io/goreecloud/goreecloud-search@sha256:"
    "30ec99e3311fa9dcc934ac267aef123bb2541e0cd0165b360c4a7dc4fa29e3d5"
)
ROLLBACK_SHA256 = "41341d322be9b8943da969ef1aac87ea07d664b1f4ea9fb8fc69085765453524"


class AuditError(ValueError):
    """Raised when evidence does not identify the immutable RC 09 acceptance boundary."""


def _audit_baseline(baseline: dict[str, Any], digest: str) -> None:
    if digest != ROLLBACK_SHA256:
        raise AuditError("rollback baseline bytes do not match the recorded known-good artifact")
    if baseline.get("schema_version") != 1:
        raise AuditError("rollback baseline schema_version must be 1")
    if baseline.get("environment") != ROLLBACK_ENVIRONMENT:
        raise AuditError("rollback baseline environment does not match the recorded baseline")
    if baseline.get("recorded_at") != ROLLBACK_RECORDED_AT:
        raise AuditError("rollback baseline timestamp does not match the recorded baseline")
    if base_audit._sha(baseline.get("source_revision"), "rollback source") != ROLLBACK_SOURCE:
        raise AuditError("rollback baseline source does not match the recorded known-good source")
    if base_audit._image(baseline.get("image"), "rollback image") != ROLLBACK_IMAGE:
        raise AuditError("rollback baseline image does not match the recorded known-good image")
    base_audit._nonempty(baseline.get("purpose"), "rollback purpose")


def _audit_release(release: dict[str, Any], baseline: dict[str, Any]) -> None:
    base_audit._schema_one(release, "Release evidence")
    if release.get("lifecycle") != "Release Candidate":
        raise AuditError("release lifecycle must remain Release Candidate")
    base_audit._iso_time(release.get("generated_at"), "release generated_at")

    request = base_audit._mapping(release.get("request"), "release request")
    if request.get("request_id") != RC09_REQUEST_ID:
        raise AuditError("release request_id does not identify RC 09")
    if request.get("candidate_sequence") != RC09_SEQUENCE:
        raise AuditError("release candidate_sequence does not identify RC 09")
    if request.get("reviewed_base_revision") != RC09_REVIEWED_BASE:
        raise AuditError("release reviewed base does not match the RC 09 reviewed boundary")

    candidate = base_audit._mapping(release.get("candidate"), "release candidate")
    if base_audit._sha(candidate.get("source_revision"), "release source") != RC09_SOURCE:
        raise AuditError("release evidence does not identify the immutable RC 09 source")
    if base_audit._image(candidate.get("image"), "release image") != RC09_IMAGE:
        raise AuditError("release evidence does not identify the immutable RC 09 image")
    if base_audit._sha(candidate.get("oci_revision"), "release OCI revision") != RC09_SOURCE:
        raise AuditError("release OCI revision does not match RC 09 source")
    if candidate.get("oci_version") != RC09_OCI_VERSION:
        raise AuditError("release OCI version does not match RC 09")
    base_audit._true(candidate.get("registry_digest_pull_verified"), "release registry pull")
    if candidate.get("isolated_runtime_acceptance") != "passed":
        raise AuditError("release isolated runtime acceptance must be passed")

    rollback = base_audit._mapping(release.get("rollback_baseline"), "release rollback baseline")
    for key in ("environment", "recorded_at", "source_revision", "image"):
        if rollback.get(key) != baseline.get(key):
            raise AuditError(f"release rollback_baseline.{key} does not match the supplied baseline")
    base_audit._true(rollback.get("registry_digest_pull_verified"), "rollback registry pull")
    if rollback.get("isolated_runtime_acceptance") != "passed":
        raise AuditError("rollback isolated runtime acceptance must be passed")

    scope = base_audit._mapping(release.get("rollback_scope"), "release rollback scope")
    if scope.get("image_level_rehearsal") != "passed":
        raise AuditError("release image-level rollback rehearsal must be passed")
    base_audit._false(
        scope.get("target_environment_configuration_rollback_tested"),
        "release target configuration rollback",
    )
    base_audit._false(
        scope.get("target_environment_data_restore_tested"),
        "release target data restoration",
    )
    if "production_cutover_authorized" in scope:
        base_audit._false(
            scope.get("production_cutover_authorized"),
            "legacy release production_cutover_authorized",
        )

    authorization = base_audit._mapping(release.get("authorization"), "release authorization")
    for key in (
        "production_cutover_authorized",
        "stable_release_authorized",
        "target_host_change_authorized",
    ):
        base_audit._false(authorization.get(key), f"release authorization.{key}")
    base_audit._nonempty(release.get("statement"), "release statement")


def _audit_runtime_version(runtime: dict[str, Any]) -> None:
    container = base_audit._mapping(runtime.get("container_runtime"), "target-runtime container")
    oci = base_audit._mapping(container.get("oci"), "target-runtime OCI")
    if oci.get("version") != RC09_OCI_VERSION:
        raise AuditError("target-runtime OCI version does not match RC 09")


def _audit_recovery_baseline(recovery: dict[str, Any], baseline_digest: str) -> None:
    rollback = base_audit._mapping(recovery.get("known_good_rollback"), "recovery rollback")
    if base_audit._sha(rollback.get("source_revision"), "recovery rollback source") != ROLLBACK_SOURCE:
        raise AuditError("recovery rollback source does not match the recorded baseline")
    if base_audit._image(rollback.get("image"), "recovery rollback image") != ROLLBACK_IMAGE:
        raise AuditError("recovery rollback image does not match the recorded baseline")
    bindings = base_audit._mapping(recovery.get("artifact_bindings"), "recovery bindings")
    if base_audit._digest(
        bindings.get("rollback_baseline_sha256"), "recovery rollback baseline SHA-256"
    ) != baseline_digest:
        raise AuditError("recovery rollback-baseline binding does not match the supplied baseline")


def audit(args: argparse.Namespace) -> dict[str, Any]:
    """Audit one complete RC 09 acceptance set without mutating supplied evidence."""
    paths = {
        "release": pathlib.Path(args.release_evidence),
        "runtime": pathlib.Path(args.target_runtime_evidence),
        "recovery": pathlib.Path(args.recovery_evidence),
        "provider": pathlib.Path(args.provider_evidence),
        "visual": pathlib.Path(args.visual_evidence),
        "browser": pathlib.Path(args.browser_evidence),
    }
    baseline_path = pathlib.Path(args.rollback_baseline)
    artifacts = {name: base_audit._load(path) for name, path in paths.items()}
    baseline = base_audit._load(baseline_path)
    for name, value in artifacts.items():
        base_audit._reject_sensitive_keys(value, f"{name}_evidence")
    base_audit._reject_sensitive_keys(baseline, "rollback_baseline")

    baseline_digest = base_audit._sha256(baseline_path)
    _audit_baseline(baseline, baseline_digest)
    _audit_release(artifacts["release"], baseline)

    release_digest = base_audit._sha256(paths["release"])
    runtime_digest = base_audit._sha256(paths["runtime"])
    base_audit._audit_runtime(artifacts["runtime"], RC09_SOURCE, RC09_IMAGE)
    _audit_runtime_version(artifacts["runtime"])
    base_audit._audit_recovery(
        artifacts["recovery"],
        RC09_SOURCE,
        RC09_IMAGE,
        release_digest,
        runtime_digest,
    )
    _audit_recovery_baseline(artifacts["recovery"], baseline_digest)
    base_audit._audit_provider(artifacts["provider"], RC09_SOURCE, RC09_IMAGE)
    base_audit._audit_visual(artifacts["visual"], RC09_SOURCE, RC09_IMAGE)
    base_audit._audit_browser(artifacts["browser"], RC09_SOURCE, RC09_IMAGE)

    bindings = {
        "release_evidence_sha256": release_digest,
        "target_runtime_evidence_sha256": runtime_digest,
        "recovery_evidence_sha256": base_audit._sha256(paths["recovery"]),
        "provider_evidence_sha256": base_audit._sha256(paths["provider"]),
        "visual_evidence_sha256": base_audit._sha256(paths["visual"]),
        "browser_evidence_sha256": base_audit._sha256(paths["browser"]),
    }
    if args.final_evidence:
        final = base_audit._load(pathlib.Path(args.final_evidence))
        base_audit._reject_sensitive_keys(final, "final_evidence")
        base_audit._audit_final(final, RC09_SOURCE, RC09_IMAGE, bindings)

    visual_summary = review_audit._visual_summary(
        artifacts["visual"], RC09_SOURCE, RC09_IMAGE
    )
    browser_summary = review_audit._browser_summary(
        artifacts["browser"], RC09_SOURCE, RC09_IMAGE
    )
    if args.final_evidence:
        review_audit._audit_final(
            final,
            RC09_SOURCE,
            RC09_IMAGE,
            paths["visual"],
            paths["browser"],
            visual_summary,
            browser_summary,
        )

    return {
        "source_revision": RC09_SOURCE,
        "image": RC09_IMAGE,
        "rollback_baseline_sha256": baseline_digest,
        **bindings,
        "visual_acceptance": visual_summary,
        "browser_integration": browser_summary,
        "production_cutover_authorized": False,
        "stable_promotion_authorized": False,
    }


def parser() -> argparse.ArgumentParser:
    """Build the RC 09 evidence-audit command line."""
    root = argparse.ArgumentParser(description=__doc__)
    root.add_argument("--release-evidence", required=True)
    root.add_argument("--target-runtime-evidence", required=True)
    root.add_argument("--recovery-evidence", required=True)
    root.add_argument("--rollback-baseline", required=True)
    root.add_argument("--provider-evidence", required=True)
    root.add_argument("--visual-evidence", required=True)
    root.add_argument("--browser-evidence", required=True)
    root.add_argument("--final-evidence")
    return root


def main() -> int:
    """Run the RC 09 evidence audit CLI."""
    try:
        result = audit(parser().parse_args())
    except (AuditError, base_audit.AuditError, review_audit.AuditError) as exc:
        print(f"RC 09 evidence audit error: {exc}", file=sys.stderr)
        return 2
    print("GoreeCloud Search RC 09 evidence audit passed.")
    print(f"Candidate source: {result['source_revision']}")
    print(f"Candidate image: {result['image']}")
    print("Production cutover authorized by this audit: false")
    print("Stable promotion authorized by this audit: false")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
