# SPDX-License-Identifier: AGPL-3.0-or-later
"""Verify immutable publication provenance for GoreeCloud Search candidate #07.

This master-side defense-in-depth layer binds the operator-supplied release evidence to the
exact release-evidence.json published by the frozen candidate workflow and requires completed
visual evidence to reference the exact immutable candidate #07 visual-review artifact. It
validates evidence only and never authorizes production cutover or Stable promotion.
"""

from __future__ import annotations

import argparse
import pathlib
import sys
from typing import Any

import first_stable_candidate_07_audit as candidate_audit


base_audit = candidate_audit.base_audit
review_audit = candidate_audit.review_audit

FROZEN_RELEASE_ARTIFACT_ID = 9382173615
FROZEN_RELEASE_ARTIFACT_NAME = (
    "goreecloud-search-candidate-b355aafe769176acebfc938b15a6f7b5b9a2db87"
)
FROZEN_RELEASE_ARTIFACT_DIGEST = (
    "sha256:1decf1341d0b876c3f3c73ed2519f6eadf952981481c5cb7934a4c2cc8ee09f0"
)
FROZEN_RELEASE_EVIDENCE_SHA256 = (
    "b0873cb4fbf244a6bcef2024add86b6579e7557a7e77f30a29f0096b2adf6752"
)
FROZEN_VISUAL_ARTIFACT_ID = 9382309578
FROZEN_VISUAL_ARTIFACT_NAME = "goreecloud-search-candidate-07-visual-evidence"
FROZEN_VISUAL_ARTIFACT_DIGEST = (
    "sha256:0b9fe7a184a6e15f01e53063ba27bec39aa477d2b08a0ca3c769e0c470451be9"
)
FROZEN_VISUAL_MANIFEST_SHA256 = (
    "e6079dcd36f5f4e8139892b98e8d804a48318f493701dd2ab4cda7b368c00979"
)


class AuditError(ValueError):
    """Raised when candidate #07 publication provenance is inconsistent or unverified."""


def _require_release_evidence_sha256(actual: str) -> None:
    """Require the exact release-evidence.json bytes retained in candidate artifact #9382173615."""
    if actual != FROZEN_RELEASE_EVIDENCE_SHA256:
        raise AuditError(
            "release evidence bytes do not match the exact file published in frozen "
            f"candidate artifact {FROZEN_RELEASE_ARTIFACT_ID}"
        )


def _audit_release_artifact(path: pathlib.Path) -> dict[str, Any]:
    """Bind release evidence to the exact independently verified candidate publication artifact."""
    actual = base_audit._sha256(path)  # pylint: disable=protected-access
    _require_release_evidence_sha256(actual)
    return {
        "artifact_id": FROZEN_RELEASE_ARTIFACT_ID,
        "artifact_name": FROZEN_RELEASE_ARTIFACT_NAME,
        "artifact_digest": FROZEN_RELEASE_ARTIFACT_DIGEST,
        "release_evidence_sha256": actual,
    }


def _require_visual_artifact(review_artifact: dict[str, Any]) -> None:
    """Require completed visual review evidence to reference the frozen workflow artifact."""
    reference = review_artifact.get("reference")
    digest = review_artifact.get("digest")
    if reference != FROZEN_VISUAL_ARTIFACT_NAME:
        raise AuditError(
            "visual evidence review_artifact.reference does not identify the frozen candidate #07 "
            "visual artifact"
        )
    if digest != FROZEN_VISUAL_ARTIFACT_DIGEST:
        raise AuditError(
            "visual evidence review_artifact.digest does not match the frozen candidate #07 "
            "visual artifact"
        )


def _audit_visual_artifact(
    visual: dict[str, Any], source: str, image: str
) -> dict[str, Any]:
    """Validate visual evidence and bind it to the exact immutable review artifact."""
    summary = review_audit._visual_summary(  # pylint: disable=protected-access
        visual, source, image
    )
    review_artifact = summary["review_artifact"]
    _require_visual_artifact(review_artifact)
    return {
        "artifact_id": FROZEN_VISUAL_ARTIFACT_ID,
        "artifact_name": FROZEN_VISUAL_ARTIFACT_NAME,
        "artifact_digest": FROZEN_VISUAL_ARTIFACT_DIGEST,
        "workflow_manifest_sha256": FROZEN_VISUAL_MANIFEST_SHA256,
    }


def audit(args: argparse.Namespace) -> dict[str, Any]:
    """Run the established candidate audit plus immutable publication-provenance checks."""
    result = candidate_audit.audit(args)
    release_path = pathlib.Path(args.release_evidence)
    release_provenance = _audit_release_artifact(release_path)

    visual = base_audit._load(  # pylint: disable=protected-access
        pathlib.Path(args.visual_evidence)
    )
    base_audit._reject_sensitive_keys(  # pylint: disable=protected-access
        visual, "visual_evidence"
    )
    visual_provenance = _audit_visual_artifact(
        visual,
        result["source_revision"],
        result["image"],
    )
    return {
        **result,
        "release_artifact_provenance": release_provenance,
        "visual_artifact_provenance": visual_provenance,
        "artifact_provenance_verified": True,
    }


def parser() -> argparse.ArgumentParser:
    """Reuse the established candidate-#07 evidence command-line contract unchanged."""
    return candidate_audit.parser()


def main() -> int:
    """Run the candidate #07 immutable artifact-provenance audit CLI."""
    args = parser().parse_args()
    try:
        result = audit(args)
    except (
        AuditError,
        candidate_audit.AuditError,
        base_audit.AuditError,
        review_audit.AuditError,
    ) as exc:
        print(f"First-Stable candidate #07 provenance audit error: {exc}", file=sys.stderr)
        return 2
    print("GoreeCloud Search first-Stable candidate #07 artifact provenance audit passed.")
    print(f"Candidate source: {result['source_revision']}")
    print(f"Candidate image: {result['image']}")
    print(f"Release artifact ID: {FROZEN_RELEASE_ARTIFACT_ID}")
    print(f"Visual artifact ID: {FROZEN_VISUAL_ARTIFACT_ID}")
    print("Production cutover authorized by this audit: false")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
