# SPDX-License-Identifier: AGPL-3.0-or-later
"""Run the complete master-side integrity audit for frozen first-Stable candidate #07.

This wrapper composes the established candidate-#07 evidence audit, immutable publication-
provenance checks, and candidate-specific consistency checks for the authentic real-provider
artifact. It validates supplied evidence only and never authorizes production cutover or Stable
promotion.
"""

from __future__ import annotations

import argparse
import pathlib
import sys
from typing import Any

import first_stable_candidate_07_provenance_audit as provenance_audit


candidate_audit = provenance_audit.candidate_audit
base_audit = candidate_audit.base_audit
FROZEN_PROVIDER_SUITE_CATEGORIES = (
    "general",
    "images",
    "news",
    "videos",
    "files",
    "it",
    "science",
)


class AuditError(ValueError):
    """Raised when candidate #07 provider evidence is internally inconsistent."""


def _integer(value: Any, label: str, *, minimum: int = 0) -> int:
    if isinstance(value, bool) or not isinstance(value, int) or value < minimum:
        raise AuditError(f"{label} must be an integer greater than or equal to {minimum}")
    return value


def _boolean(value: Any, label: str) -> bool:
    if not isinstance(value, bool):
        raise AuditError(f"{label} must be a boolean")
    return value


def _audit_frozen_provider(provider: dict[str, Any], source: str, image: str) -> None:
    """Verify the complete result semantics emitted by frozen provider_acceptance.py."""
    base_audit._audit_provider(provider, source, image)  # pylint: disable=protected-access

    minimum_results = _integer(provider.get("minimum_results"), "provider minimum_results", minimum=1)
    required = base_audit._list(  # pylint: disable=protected-access
        provider.get("required_categories"), "provider required_categories"
    )
    expected_required = sorted(base_audit.REQUIRED_PROVIDER_CATEGORIES)
    if required != expected_required:
        raise AuditError("provider required_categories must exactly match the frozen release set")

    results = base_audit._list(provider.get("results"), "provider results")  # pylint: disable=protected-access
    if len(results) != len(FROZEN_PROVIDER_SUITE_CATEGORIES):
        raise AuditError("provider results must contain exactly the frozen representative suite")

    observed_categories: list[str] = []
    for index, item in enumerate(results):
        result = base_audit._mapping(  # pylint: disable=protected-access
            item, f"provider results[{index}]"
        )
        category = base_audit._nonempty(  # pylint: disable=protected-access
            result.get("category"), f"provider results[{index}].category"
        )
        observed_categories.append(category)

        exit_code = _integer(result.get("exit_code"), f"provider results[{index}].exit_code")
        http_status = _integer(
            result.get("http_status"), f"provider results[{index}].http_status", minimum=100
        )
        product_identity = _boolean(
            result.get("product_identity"), f"provider results[{index}].product_identity"
        )
        result_cards = _integer(
            result.get("result_cards"), f"provider results[{index}].result_cards"
        )
        _integer(result.get("engine_messages"), f"provider results[{index}].engine_messages")
        passed = _boolean(result.get("passed"), f"provider results[{index}].passed")

        if not passed:
            raise AuditError(f"provider category {category} must pass in successful candidate evidence")
        if exit_code != 0:
            raise AuditError(f"provider category {category} passed flag conflicts with exit_code")
        if http_status != 200:
            raise AuditError(f"provider category {category} passed flag conflicts with HTTP status")
        if not product_identity:
            raise AuditError(f"provider category {category} passed without GoreeCloud Search identity")
        if result_cards < minimum_results:
            raise AuditError(f"provider category {category} passed below minimum_results")

    if tuple(observed_categories) != FROZEN_PROVIDER_SUITE_CATEGORIES:
        raise AuditError("provider result categories or order do not match the frozen representative suite")

    scope = base_audit._mapping(provider.get("scope"), "provider scope")  # pylint: disable=protected-access
    base_audit._true(  # pylint: disable=protected-access
        scope.get("full_diagnostic_suite_passed"), "provider full diagnostic suite"
    )
    base_audit._nonempty(scope.get("statement"), "provider scope.statement")  # pylint: disable=protected-access


def audit(args: argparse.Namespace) -> dict[str, Any]:
    """Run candidate, artifact-provenance, and deep provider-result integrity checks."""
    result = provenance_audit.audit(args)
    provider = base_audit._load(pathlib.Path(args.provider_evidence))  # pylint: disable=protected-access
    base_audit._reject_sensitive_keys(provider, "provider_evidence")  # pylint: disable=protected-access
    _audit_frozen_provider(provider, result["source_revision"], result["image"])
    return {**result, "provider_result_integrity_verified": True}


def parser() -> argparse.ArgumentParser:
    """Reuse the established candidate-#07 command-line contract unchanged."""
    return candidate_audit.parser()


def main() -> int:
    """Run the complete candidate #07 integrity audit CLI."""
    args = parser().parse_args()
    try:
        result = audit(args)
    except (
        AuditError,
        provenance_audit.AuditError,
        candidate_audit.AuditError,
        base_audit.AuditError,
        candidate_audit.review_audit.AuditError,
    ) as exc:
        print(f"First-Stable candidate #07 integrity audit error: {exc}", file=sys.stderr)
        return 2
    print("GoreeCloud Search first-Stable candidate #07 integrity audit passed.")
    print(f"Candidate source: {result['source_revision']}")
    print(f"Candidate image: {result['image']}")
    print("Artifact provenance verified: true")
    print("Provider result integrity verified: true")
    print("Production cutover authorized by this audit: false")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
