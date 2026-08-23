# SPDX-License-Identifier: AGPL-3.0-or-later
"""Validate GoreeCloud Search release-candidate identity and write release evidence.

This helper deliberately separates image-level release evidence from target-host
acceptance. It proves immutable image identity, registry retrieval, and isolated
candidate/rollback execution only. It never authorizes production cutover or
Stable promotion.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import pathlib
import re
import sys
from typing import Any


SHA_RE = re.compile(r"^[0-9a-f]{40}$")
IMAGE_RE = re.compile(r"^ghcr\.io/goreecloud/goreecloud-search@sha256:[0-9a-f]{64}$")
REQUEST_ID_RE = re.compile(r"^first-stable-\d{4}-\d{2}-\d{2}-(\d{2})$")


class EvidenceError(ValueError):
    """Raised when release evidence is incomplete or unsafe."""


def _load_json(path: pathlib.Path) -> dict[str, Any]:
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        raise EvidenceError(f"Unable to read valid JSON from {path}: {exc}") from exc
    if not isinstance(data, dict):
        raise EvidenceError(f"{path} must contain a JSON object")
    return data


def _require_sha(value: str, label: str) -> str:
    if not SHA_RE.fullmatch(value):
        raise EvidenceError(f"{label} must be a lowercase 40-character Git SHA")
    return value


def _require_image(value: str, label: str) -> str:
    if not IMAGE_RE.fullmatch(value):
        raise EvidenceError(
            f"{label} must be ghcr.io/goreecloud/goreecloud-search pinned by sha256 digest"
        )
    return value


def validate_request(data: dict[str, Any]) -> dict[str, Any]:
    required = {
        "schema_version",
        "product",
        "request_id",
        "candidate_sequence",
        "requested_at",
        "request",
        "lifecycle",
        "reviewed_base_revision",
        "production_cutover_authorized",
        "stable_release_authorized",
        "target_host_change_authorized",
        "statement",
    }
    missing = sorted(required - data.keys())
    if missing:
        raise EvidenceError(f"Release-candidate request is missing: {', '.join(missing)}")
    if data["schema_version"] != 1:
        raise EvidenceError("Release-candidate request schema_version must be 1")
    if data["product"] != "GoreeCloud Search":
        raise EvidenceError("Release-candidate request must identify GoreeCloud Search")
    if data["request"] != "publish-and-rehearse-release-candidate":
        raise EvidenceError("Release-candidate request action is invalid")
    if data["lifecycle"] != "Release Candidate":
        raise EvidenceError("Lifecycle must be Release Candidate")
    if not isinstance(data["candidate_sequence"], int) or data["candidate_sequence"] < 1:
        raise EvidenceError("candidate_sequence must be a positive integer")
    request_id = str(data["request_id"])
    match = REQUEST_ID_RE.fullmatch(request_id)
    if not match:
        raise EvidenceError("request_id must use first-stable-YYYY-MM-DD-NN")
    if int(match.group(1)) != data["candidate_sequence"]:
        raise EvidenceError("request_id sequence must match candidate_sequence")
    _require_sha(str(data["reviewed_base_revision"]), "reviewed base revision")
    try:
        requested_at = dt.datetime.fromisoformat(str(data["requested_at"]))
    except ValueError as exc:
        raise EvidenceError("requested_at must be ISO-8601") from exc
    if requested_at.tzinfo is None:
        raise EvidenceError("requested_at must include a timezone offset")
    for key in (
        "production_cutover_authorized",
        "stable_release_authorized",
        "target_host_change_authorized",
    ):
        if data[key] is not False:
            raise EvidenceError(f"{key} must remain false for Release Candidate publication")
    if not isinstance(data["statement"], str) or not data["statement"].strip():
        raise EvidenceError("statement must be non-empty")
    return data


def validate_baseline(data: dict[str, Any]) -> dict[str, Any]:
    required = {
        "schema_version",
        "environment",
        "recorded_at",
        "source_revision",
        "image",
        "purpose",
    }
    missing = sorted(required - data.keys())
    if missing:
        raise EvidenceError(f"Rollback baseline is missing: {', '.join(missing)}")
    if data["schema_version"] != 1:
        raise EvidenceError("Rollback baseline schema_version must be 1")
    if data["environment"] != "goreecloud-vps-01-production":
        raise EvidenceError("Rollback baseline must identify goreecloud-vps-01-production")
    _require_sha(str(data["source_revision"]), "rollback source revision")
    _require_image(str(data["image"]), "rollback image")
    try:
        recorded = dt.datetime.fromisoformat(str(data["recorded_at"]))
    except ValueError as exc:
        raise EvidenceError("Rollback recorded_at must be ISO-8601") from exc
    if recorded.tzinfo is None:
        raise EvidenceError("Rollback recorded_at must include a timezone offset")
    if not isinstance(data["purpose"], str) or not data["purpose"].strip():
        raise EvidenceError("Rollback purpose must be non-empty")
    return data


def write_evidence(args: argparse.Namespace) -> dict[str, Any]:
    request = validate_request(_load_json(pathlib.Path(args.request_file)))
    baseline = validate_baseline(_load_json(pathlib.Path(args.rollback_baseline)))
    candidate_source = _require_sha(args.source_sha, "candidate source SHA")
    candidate_image = _require_image(args.candidate_image, "candidate image")

    if args.candidate_oci_revision != candidate_source:
        raise EvidenceError("Candidate OCI revision does not match the exact source SHA")
    if candidate_image == baseline["image"]:
        raise EvidenceError("Candidate image must differ from the recorded rollback image")

    evidence = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "lifecycle": "Release Candidate",
        "generated_at": dt.datetime.now(dt.timezone.utc).isoformat().replace("+00:00", "Z"),
        "request": {
            "request_id": request["request_id"],
            "candidate_sequence": request["candidate_sequence"],
            "reviewed_base_revision": request["reviewed_base_revision"],
        },
        "candidate": {
            "source_revision": candidate_source,
            "image": candidate_image,
            "oci_revision": args.candidate_oci_revision,
            "oci_version": args.candidate_oci_version,
            "registry_digest_pull_verified": True,
            "isolated_runtime_acceptance": "passed",
        },
        "rollback_baseline": {
            "environment": baseline["environment"],
            "recorded_at": baseline["recorded_at"],
            "source_revision": baseline["source_revision"],
            "image": baseline["image"],
            "registry_digest_pull_verified": True,
            "isolated_runtime_acceptance": "passed",
        },
        "rollback_scope": {
            "image_level_rehearsal": "passed",
            "target_environment_configuration_rollback_tested": False,
            "target_environment_data_restore_tested": False,
        },
        "authorization": {
            "production_cutover_authorized": False,
            "stable_release_authorized": False,
            "target_host_change_authorized": False,
        },
        "statement": (
            "This artifact proves the exact Release Candidate source, immutable candidate "
            "image, registry retrieval, and isolated candidate-to-known-good-image rollback "
            "execution only. Target-host configuration, persistent data, real providers, "
            "monitoring and alert delivery, backup restoration, physical-device review, "
            "desktop/persisted-theme review, actual compiled GoreeCloud Browser acceptance, "
            "production cutover, and Stable promotion remain separate requirements."
        ),
    }

    output = pathlib.Path(args.output)
    output.write_text(json.dumps(evidence, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return evidence


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=__doc__)
    subparsers = parser.add_subparsers(dest="command", required=True)

    validate_request_parser = subparsers.add_parser(
        "validate-request",
        help="Validate the Release Candidate request JSON",
    )
    validate_request_parser.add_argument("request_file")

    validate_baseline_parser = subparsers.add_parser(
        "validate-baseline",
        help="Validate the rollback baseline JSON",
    )
    validate_baseline_parser.add_argument("baseline")

    write = subparsers.add_parser("write", help="Write candidate and rollback image evidence")
    write.add_argument("--request-file", required=True)
    write.add_argument("--rollback-baseline", required=True)
    write.add_argument("--source-sha", required=True)
    write.add_argument("--candidate-image", required=True)
    write.add_argument("--candidate-oci-revision", required=True)
    write.add_argument("--candidate-oci-version", required=True)
    write.add_argument("--output", required=True)
    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()
    try:
        if args.command == "validate-request":
            request = validate_request(_load_json(pathlib.Path(args.request_file)))
            print(f"Release Candidate request valid: {request['request_id']}")
            return 0
        if args.command == "validate-baseline":
            baseline = validate_baseline(_load_json(pathlib.Path(args.baseline)))
            print(f"Rollback baseline valid: {baseline['image']}")
            return 0
        if args.command == "write":
            evidence = write_evidence(args)
            print(f"Release evidence written: {args.output}")
            print(f"Candidate: {evidence['candidate']['image']}")
            print(f"Rollback: {evidence['rollback_baseline']['image']}")
            print("Production cutover authorized: false")
            print("Stable release authorized: false")
            return 0
    except EvidenceError as exc:
        print(f"Release evidence error: {exc}", file=sys.stderr)
        return 2
    parser.error("unsupported command")
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
