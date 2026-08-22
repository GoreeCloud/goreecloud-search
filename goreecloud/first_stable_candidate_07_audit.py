# SPDX-License-Identifier: AGPL-3.0-or-later
"""Audit the frozen GoreeCloud Search first-Stable candidate #07 evidence set.

This is the authoritative master-side defense-in-depth entrypoint for candidate #07. It
preserves the frozen candidate evidence schema, reuses the broader six-artifact safety audit,
and also runs the deep visual/Browser review audit. It validates evidence only and never
authorizes production cutover or Stable promotion.
"""

from __future__ import annotations

import argparse
import pathlib
import sys
from typing import Any

import final_review_evidence_audit as review_audit
import first_stable_evidence_audit as base_audit


FROZEN_SOURCE = "b355aafe769176acebfc938b15a6f7b5b9a2db87"
FROZEN_IMAGE = (
    "ghcr.io/goreecloud/goreecloud-search@sha256:"
    "3ce3a509675ee396cba33b77ca429aaeea1b1d995f42c62778f9d24de40a09d8"
)


class AuditError(ValueError):
    """Raised when candidate #07 evidence is incomplete, inconsistent, or incompatible."""


def _require_frozen_candidate(source: str, image: str) -> None:
    if source != FROZEN_SOURCE:
        raise AuditError("release evidence does not identify frozen first-Stable candidate #07 source")
    if image != FROZEN_IMAGE:
        raise AuditError("release evidence does not identify frozen first-Stable candidate #07 image")


def _audit_frozen_runtime(runtime: dict[str, Any], source: str, image: str) -> None:
    """Validate the exact target-runtime schema emitted by frozen candidate #07."""
    base_audit._schema_one(runtime, "Target-runtime evidence")  # pylint: disable=protected-access
    base_audit._iso_time(  # pylint: disable=protected-access
        runtime.get("generated_at"), "target-runtime generated_at"
    )
    target = base_audit._mapping(runtime.get("target"), "target-runtime target")  # pylint: disable=protected-access
    base_url = base_audit._nonempty(  # pylint: disable=protected-access
        target.get("base_url"), "target-runtime target.base_url"
    )
    if not base_audit.LOOPBACK_URL_RE.fullmatch(base_url):
        raise AuditError("target-runtime target.base_url must identify loopback-only staging")
    base_audit._nonempty(target.get("container"), "target-runtime target.container")  # pylint: disable=protected-access

    http_acceptance = base_audit._mapping(  # pylint: disable=protected-access
        runtime.get("http_acceptance"), "target-runtime http_acceptance"
    )
    for key in ("home_identity", "preferences_identity", "about_identity", "health", "privacy_headers"):
        if http_acceptance.get(key) != "passed":
            raise AuditError(f"target-runtime http_acceptance.{key} must be passed")

    if runtime.get("providers") not in {"passed", "skipped"}:
        raise AuditError("target-runtime providers must be passed or skipped")

    container = base_audit._mapping(  # pylint: disable=protected-access
        runtime.get("container_runtime"), "target-runtime container_runtime"
    )
    if container.get("status") != "checked":
        raise AuditError("target-runtime container_runtime.status must be checked")
    base_audit._true(  # pylint: disable=protected-access
        container.get("running"), "target-runtime container_runtime.running"
    )
    if container.get("health") != "healthy":
        raise AuditError("target-runtime container_runtime.health must be healthy")
    if container.get("published_ports") != "verified":
        raise AuditError("target-runtime container_runtime.published_ports must be verified")
    if container.get("identity_status") != "verified":
        raise AuditError("target-runtime container_runtime.identity_status must be verified")
    if base_audit._image(  # pylint: disable=protected-access
        container.get("expected_image"), "target-runtime expected_image"
    ) != image:
        raise AuditError("target-runtime expected image does not match release evidence")
    if base_audit._sha(  # pylint: disable=protected-access
        container.get("expected_source_revision"), "target-runtime expected_source_revision"
    ) != source:
        raise AuditError("target-runtime expected source does not match release evidence")
    if base_audit._image(  # pylint: disable=protected-access
        container.get("observed_image_reference"), "target-runtime observed_image_reference"
    ) != image:
        raise AuditError("target-runtime observed image does not match release evidence")
    base_audit._nonempty(  # pylint: disable=protected-access
        container.get("observed_image_id"), "target-runtime observed_image_id"
    )

    oci = base_audit._mapping(  # pylint: disable=protected-access
        container.get("oci"), "target-runtime container_runtime.oci"
    )
    if oci.get("title") != "GoreeCloud Search":
        raise AuditError("target-runtime OCI title must be GoreeCloud Search")
    if oci.get("source") != base_audit.REPOSITORY_URL:
        raise AuditError("target-runtime OCI source must identify the canonical Search repository")
    if base_audit._sha(oci.get("revision"), "target-runtime OCI revision") != source:  # pylint: disable=protected-access
        raise AuditError("target-runtime OCI revision does not match release evidence")
    base_audit._nonempty(oci.get("version"), "target-runtime OCI version")  # pylint: disable=protected-access
    if oci.get("licenses") != "AGPL-3.0-or-later":
        raise AuditError("target-runtime OCI licenses must be AGPL-3.0-or-later")

    scope = base_audit._mapping(runtime.get("scope"), "target-runtime scope")  # pylint: disable=protected-access
    base_audit._true(scope.get("target_runtime_identity_verified"), "target-runtime identity verified")  # pylint: disable=protected-access
    base_audit._false(  # pylint: disable=protected-access
        scope.get("target_environment_configuration_rollback_tested"), "target-runtime configuration rollback"
    )
    base_audit._false(  # pylint: disable=protected-access
        scope.get("target_environment_data_restore_tested"), "target-runtime data restore"
    )
    base_audit._false(scope.get("backup_restore_tested"), "target-runtime backup restore")  # pylint: disable=protected-access
    base_audit._false(  # pylint: disable=protected-access
        scope.get("production_cutover_authorized"), "target-runtime production_cutover_authorized"
    )
    base_audit._nonempty(scope.get("statement"), "target-runtime scope.statement")  # pylint: disable=protected-access


def audit(args: argparse.Namespace) -> dict[str, Any]:
    """Run the full candidate #07 evidence audit without modifying any supplied artifact."""
    paths = {
        "release": pathlib.Path(args.release_evidence),
        "runtime": pathlib.Path(args.target_runtime_evidence),
        "recovery": pathlib.Path(args.recovery_evidence),
        "provider": pathlib.Path(args.provider_evidence),
        "visual": pathlib.Path(args.visual_evidence),
        "browser": pathlib.Path(args.browser_evidence),
    }
    artifacts = {
        name: base_audit._load(path)  # pylint: disable=protected-access
        for name, path in paths.items()
    }
    for name, value in artifacts.items():
        base_audit._reject_sensitive_keys(value, f"{name}_evidence")  # pylint: disable=protected-access

    source, image = base_audit._candidate_from_release(artifacts["release"])  # pylint: disable=protected-access
    _require_frozen_candidate(source, image)
    release_sha256 = base_audit._sha256(paths["release"])  # pylint: disable=protected-access
    runtime_sha256 = base_audit._sha256(paths["runtime"])  # pylint: disable=protected-access
    _audit_frozen_runtime(artifacts["runtime"], source, image)
    base_audit._audit_recovery(  # pylint: disable=protected-access
        artifacts["recovery"], source, image, release_sha256, runtime_sha256
    )
    base_audit._audit_provider(artifacts["provider"], source, image)  # pylint: disable=protected-access
    base_audit._audit_visual(artifacts["visual"], source, image)  # pylint: disable=protected-access
    base_audit._audit_browser(artifacts["browser"], source, image)  # pylint: disable=protected-access

    bindings = {
        "release_evidence_sha256": release_sha256,
        "target_runtime_evidence_sha256": runtime_sha256,
        "recovery_evidence_sha256": base_audit._sha256(paths["recovery"]),  # pylint: disable=protected-access
        "provider_evidence_sha256": base_audit._sha256(paths["provider"]),  # pylint: disable=protected-access
        "visual_evidence_sha256": base_audit._sha256(paths["visual"]),  # pylint: disable=protected-access
        "browser_evidence_sha256": base_audit._sha256(paths["browser"]),  # pylint: disable=protected-access
    }
    if args.final_evidence:
        final = base_audit._load(pathlib.Path(args.final_evidence))  # pylint: disable=protected-access
        base_audit._reject_sensitive_keys(final, "final_evidence")  # pylint: disable=protected-access
        base_audit._audit_final(final, source, image, bindings)  # pylint: disable=protected-access

    review_result = review_audit.audit(
        argparse.Namespace(
            release_evidence=args.release_evidence,
            visual_evidence=args.visual_evidence,
            browser_evidence=args.browser_evidence,
            final_evidence=args.final_evidence,
        )
    )
    if review_result["source_revision"] != source or review_result["image"] != image:
        raise AuditError("deep visual/Browser audit resolved a different Search candidate")

    return {
        "source_revision": source,
        "image": image,
        **bindings,
        "visual_acceptance": review_result["visual_acceptance"],
        "browser_integration": review_result["browser_integration"],
        "production_cutover_authorized": False,
    }


def parser() -> argparse.ArgumentParser:
    """Build the candidate #07 audit command-line parser."""
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
    """Run the candidate #07 audit CLI."""
    args = parser().parse_args()
    try:
        result = audit(args)
    except (AuditError, base_audit.AuditError, review_audit.AuditError) as exc:
        print(f"First-Stable candidate #07 evidence audit error: {exc}", file=sys.stderr)
        return 2
    print("GoreeCloud Search first-Stable candidate #07 evidence audit passed.")
    print(f"Candidate source: {result['source_revision']}")
    print(f"Candidate image: {result['image']}")
    print("Production cutover authorized by this audit: false")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
