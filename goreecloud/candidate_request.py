# SPDX-License-Identifier: AGPL-3.0-or-later
"""Validate explicit GoreeCloud Search candidate-publication requests.

The request marker authorizes immutable candidate publication and isolated image-level
rehearsal only. It never authorizes production cutover, Stable promotion, target-host
changes, compatibility-name retirement, or rollback-material deletion.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import pathlib
import re
from typing import Any

SHA_RE = re.compile(r"^[0-9a-f]{40}$")
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
}


class CandidateRequestError(ValueError):
    """Raised when a candidate request is unsafe or internally inconsistent."""


def _load(path: pathlib.Path) -> dict[str, Any]:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        raise CandidateRequestError(f"Unable to read valid JSON from {path}: {exc}") from exc
    if not isinstance(value, dict):
        raise CandidateRequestError("Candidate request must contain a JSON object")
    return value


def _require_nonempty(value: Any, label: str) -> str:
    if not isinstance(value, str) or not value.strip():
        raise CandidateRequestError(f"{label} must be a non-empty string")
    return value.strip()


def _require_false(value: Any, label: str) -> None:
    if value is not False:
        raise CandidateRequestError(f"{label} must remain false")


def _require_sha(value: Any, label: str) -> str:
    text = _require_nonempty(value, label)
    if not SHA_RE.fullmatch(text):
        raise CandidateRequestError(f"{label} must be a lowercase 40-character Git SHA")
    return text


def _require_iso_time(value: Any, label: str) -> str:
    text = _require_nonempty(value, label)
    try:
        parsed = dt.datetime.fromisoformat(text.replace("Z", "+00:00"))
    except ValueError as exc:
        raise CandidateRequestError(f"{label} must be ISO-8601") from exc
    if parsed.tzinfo is None:
        raise CandidateRequestError(f"{label} must include a timezone")
    return text


def _reject_sensitive_keys(value: Any, path: str = "candidate_request") -> None:
    if isinstance(value, dict):
        for key, child in value.items():
            normalized = str(key).strip().lower()
            if normalized in FORBIDDEN_KEYS:
                raise CandidateRequestError(f"Sensitive field is not allowed: {path}.{key}")
            _reject_sensitive_keys(child, f"{path}.{key}")
    elif isinstance(value, list):
        for index, child in enumerate(value):
            _reject_sensitive_keys(child, f"{path}[{index}]")


def validate_request(
    path: pathlib.Path,
    *,
    expected_reviewed_base: str | None = None,
    candidate_source: str | None = None,
    candidate_parent: str | None = None,
) -> dict[str, Any]:
    request = _load(path)
    _reject_sensitive_keys(request)

    expected_keys = {
        "schema_version",
        "product",
        "request_id",
        "requested_at",
        "request",
        "reviewed_base_revision",
        "production_cutover_authorized",
        "stable_release_authorized",
        "target_host_change_authorized",
        "statement",
    }
    actual_keys = set(request)
    if actual_keys != expected_keys:
        missing = sorted(expected_keys - actual_keys)
        extra = sorted(actual_keys - expected_keys)
        detail: list[str] = []
        if missing:
            detail.append("missing=" + ",".join(missing))
        if extra:
            detail.append("extra=" + ",".join(extra))
        raise CandidateRequestError("Candidate request fields are not exact: " + "; ".join(detail))

    if request.get("schema_version") != 1:
        raise CandidateRequestError("schema_version must be 1")
    if request.get("product") != "GoreeCloud Search":
        raise CandidateRequestError("product must be GoreeCloud Search")
    _require_nonempty(request.get("request_id"), "request_id")
    _require_iso_time(request.get("requested_at"), "requested_at")
    if request.get("request") != "publish-and-rehearse-final-candidate":
        raise CandidateRequestError("request must be publish-and-rehearse-final-candidate")

    reviewed_base = _require_sha(request.get("reviewed_base_revision"), "reviewed_base_revision")
    _require_false(request.get("production_cutover_authorized"), "production_cutover_authorized")
    _require_false(request.get("stable_release_authorized"), "stable_release_authorized")
    _require_false(request.get("target_host_change_authorized"), "target_host_change_authorized")
    _require_nonempty(request.get("statement"), "statement")

    if expected_reviewed_base is not None:
        expected = _require_sha(expected_reviewed_base, "expected_reviewed_base")
        if reviewed_base != expected:
            raise CandidateRequestError(
                "reviewed_base_revision does not match the reviewed pull-request base revision"
            )

    if candidate_source is not None:
        source = _require_sha(candidate_source, "candidate_source")
        if source == reviewed_base:
            raise CandidateRequestError("Candidate source must include the explicit request commit")

    if candidate_parent is not None:
        parent = _require_sha(candidate_parent, "candidate_parent")
        if parent != reviewed_base:
            raise CandidateRequestError(
                "Candidate commit parent does not match reviewed_base_revision; refusing publication"
            )

    return request


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("path", type=pathlib.Path)
    parser.add_argument("--expected-reviewed-base")
    parser.add_argument("--candidate-source")
    parser.add_argument("--candidate-parent")
    args = parser.parse_args()
    validate_request(
        args.path,
        expected_reviewed_base=args.expected_reviewed_base,
        candidate_source=args.candidate_source,
        candidate_parent=args.candidate_parent,
    )
    print("GoreeCloud Search candidate request is valid and non-authorizing.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
