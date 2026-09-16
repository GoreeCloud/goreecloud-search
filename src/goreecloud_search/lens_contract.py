from __future__ import annotations

import json
from collections.abc import Sequence

from .lenses import Lens, LensError, LensRule, LensRuleAction, LensRuleTarget

LENS_FORMAT_VERSION = "goreecloud.search-lens.v1"
MAX_LENS_DOCUMENT_BYTES = 64 * 1024
MAX_LENS_RULES = 256
MAX_LENS_NAME_LENGTH = 128
MAX_LENS_DESCRIPTION_LENGTH = 2048
MAX_LENS_RULE_VALUE_LENGTH = 512


class LensFormatError(LensError):
    """Raised when a portable Lens document is invalid or incompatible."""


def _object_without_duplicate_keys(pairs: Sequence[tuple[str, object]]) -> dict[str, object]:
    output: dict[str, object] = {}
    for key, value in pairs:
        if key in output:
            raise LensFormatError(f"duplicate JSON key: {key!r}")
        output[key] = value
    return output


def _reject_nonfinite(value: str) -> None:
    raise LensFormatError(f"non-finite JSON number is not allowed: {value}")


def _ensure_exact_keys(value: dict[str, object], *, allowed: set[str], context: str) -> None:
    unknown = set(value) - allowed
    if unknown:
        names = ", ".join(sorted(unknown))
        raise LensFormatError(f"unknown {context} field(s): {names}")


def _text_field(
    value: object,
    *,
    name: str,
    max_length: int,
    allow_none: bool = False,
) -> str | None:
    if value is None and allow_none:
        return None
    if not isinstance(value, str):
        raise LensFormatError(f"{name} must be a string")
    normalized = value.strip()
    if not normalized:
        raise LensFormatError(f"{name} must not be empty")
    if len(normalized) > max_length:
        raise LensFormatError(f"{name} exceeds maximum length {max_length}")
    return normalized


def lens_to_document(lens: Lens) -> dict[str, object]:
    """Return the versioned, JSON-compatible portable representation of a Lens."""
    return {
        "format_version": LENS_FORMAT_VERSION,
        "name": lens.name,
        "description": lens.description,
        "rules": [
            {
                "target": rule.target.value,
                "value": rule.value,
                "action": rule.action.value,
                "weight": float(rule.weight),
            }
            for rule in lens.rules
        ],
    }


def export_lens(lens: Lens) -> str:
    """Serialize a Lens deterministically for local export/share workflows."""
    if len(lens.name) > MAX_LENS_NAME_LENGTH:
        raise LensFormatError(f"Lens name exceeds maximum length {MAX_LENS_NAME_LENGTH}")
    if lens.description is not None and len(lens.description) > MAX_LENS_DESCRIPTION_LENGTH:
        raise LensFormatError(
            f"Lens description exceeds maximum length {MAX_LENS_DESCRIPTION_LENGTH}"
        )
    if len(lens.rules) > MAX_LENS_RULES:
        raise LensFormatError(f"Lens contains more than {MAX_LENS_RULES} rules")
    for rule in lens.rules:
        if len(rule.value) > MAX_LENS_RULE_VALUE_LENGTH:
            raise LensFormatError(
                f"Lens rule value exceeds maximum length {MAX_LENS_RULE_VALUE_LENGTH}"
            )

    payload = json.dumps(
        lens_to_document(lens),
        ensure_ascii=False,
        indent=2,
        sort_keys=True,
        allow_nan=False,
    ) + "\n"
    if len(payload.encode("utf-8")) > MAX_LENS_DOCUMENT_BYTES:
        raise LensFormatError(
            f"Lens document exceeds maximum size {MAX_LENS_DOCUMENT_BYTES} bytes"
        )
    return payload


def import_lens(payload: str) -> Lens:
    """Parse a portable Lens document using a strict, fail-closed v1 schema."""
    if not isinstance(payload, str):
        raise LensFormatError("Lens document must be text")
    if len(payload.encode("utf-8")) > MAX_LENS_DOCUMENT_BYTES:
        raise LensFormatError(
            f"Lens document exceeds maximum size {MAX_LENS_DOCUMENT_BYTES} bytes"
        )

    try:
        document = json.loads(
            payload,
            object_pairs_hook=_object_without_duplicate_keys,
            parse_constant=_reject_nonfinite,
        )
    except LensFormatError:
        raise
    except (json.JSONDecodeError, TypeError, ValueError) as exc:
        raise LensFormatError(f"invalid Lens JSON: {exc}") from exc

    if not isinstance(document, dict):
        raise LensFormatError("Lens document root must be an object")
    _ensure_exact_keys(
        document,
        allowed={"format_version", "name", "description", "rules"},
        context="Lens document",
    )

    format_version = document.get("format_version")
    if format_version != LENS_FORMAT_VERSION:
        raise LensFormatError(f"unsupported Lens format version: {format_version!r}")

    name = _text_field(
        document.get("name"),
        name="Lens name",
        max_length=MAX_LENS_NAME_LENGTH,
    )
    description = _text_field(
        document.get("description"),
        name="Lens description",
        max_length=MAX_LENS_DESCRIPTION_LENGTH,
        allow_none=True,
    )
    raw_rules = document.get("rules")
    if not isinstance(raw_rules, list):
        raise LensFormatError("Lens rules must be an array")
    if not raw_rules:
        raise LensFormatError("Lens must contain at least one rule")
    if len(raw_rules) > MAX_LENS_RULES:
        raise LensFormatError(f"Lens contains more than {MAX_LENS_RULES} rules")

    rules: list[LensRule] = []
    for index, raw_rule in enumerate(raw_rules):
        if not isinstance(raw_rule, dict):
            raise LensFormatError(f"Lens rule {index} must be an object")
        _ensure_exact_keys(
            raw_rule,
            allowed={"target", "value", "action", "weight"},
            context=f"Lens rule {index}",
        )
        target_text = _text_field(
            raw_rule.get("target"),
            name=f"Lens rule {index} target",
            max_length=32,
        )
        value = _text_field(
            raw_rule.get("value"),
            name=f"Lens rule {index} value",
            max_length=MAX_LENS_RULE_VALUE_LENGTH,
        )
        action_text = _text_field(
            raw_rule.get("action"),
            name=f"Lens rule {index} action",
            max_length=32,
        )
        weight = raw_rule.get("weight")
        if isinstance(weight, bool) or not isinstance(weight, (int, float)):
            raise LensFormatError(f"Lens rule {index} weight must be numeric")

        try:
            target = LensRuleTarget(target_text)
        except ValueError as exc:
            raise LensFormatError(
                f"unsupported Lens rule {index} target: {target_text!r}"
            ) from exc
        try:
            action = LensRuleAction(action_text)
        except ValueError as exc:
            raise LensFormatError(
                f"unsupported Lens rule {index} action: {action_text!r}"
            ) from exc

        try:
            rules.append(LensRule(target, value, action, float(weight)))
        except LensError as exc:
            raise LensFormatError(f"invalid Lens rule {index}: {exc}") from exc

    try:
        return Lens(name=name, description=description, rules=tuple(rules))
    except LensError as exc:
        raise LensFormatError(f"invalid Lens document: {exc}") from exc
