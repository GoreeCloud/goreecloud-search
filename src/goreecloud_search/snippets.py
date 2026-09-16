from __future__ import annotations

import re
from dataclasses import dataclass

from .models import ParsedQuery

DEFAULT_SNIPPET_MAX_CHARS = 280
MAX_SNIPPET_SOURCE_CHARS = 32_768
_MIN_SNIPPET_MAX_CHARS = 80
_WHITESPACE = re.compile(r"\s+")
_SENTENCE_BREAK = re.compile(r"(?<=[.!?])\s+")


class SnippetGenerationError(ValueError):
    """Raised when snippet-generation input violates the bounded contract."""


@dataclass(frozen=True, slots=True)
class GeneratedSnippet:
    text: str
    matched_query: bool
    source_truncated: bool
    start_offset: int
    end_offset: int


def _normalize_source(text: str) -> tuple[str, bool]:
    if not isinstance(text, str):
        raise SnippetGenerationError("snippet source must be text")

    bounded = text[:MAX_SNIPPET_SOURCE_CHARS]
    source_truncated = len(text) > len(bounded)
    normalized = _WHITESPACE.sub(" ", bounded).strip()
    return normalized, source_truncated


def _needles(query: ParsedQuery) -> tuple[str, ...]:
    values: list[str] = []
    for value in (*query.phrases, *query.terms):
        cleaned = _WHITESPACE.sub(" ", value).strip()
        if cleaned and cleaned.casefold() not in {item.casefold() for item in values}:
            values.append(cleaned)
    return tuple(values)


def _best_match(text: str, needles: tuple[str, ...]) -> tuple[int, int] | None:
    folded = text.casefold()
    matches: list[tuple[int, int, int]] = []
    for needle in needles:
        start = folded.find(needle.casefold())
        if start >= 0:
            matches.append((start, -len(needle), start + len(needle)))
    if not matches:
        return None
    start, _, end = min(matches)
    return start, end


def _clip(text: str, *, focus_start: int, focus_end: int, max_chars: int) -> tuple[str, int, int]:
    if len(text) <= max_chars:
        return text, 0, len(text)

    focus_mid = (focus_start + focus_end) // 2
    start = max(0, focus_mid - max_chars // 2)
    end = min(len(text), start + max_chars)
    start = max(0, end - max_chars)

    # Avoid starting or ending in the middle of a word when a nearby boundary exists.
    if start > 0:
        boundary = text.find(" ", start, min(start + 24, end))
        if boundary >= 0:
            start = boundary + 1
    if end < len(text):
        boundary = text.rfind(" ", max(start, end - 24), end)
        if boundary > start:
            end = boundary

    body = text[start:end].strip()
    prefix = "…" if start > 0 else ""
    suffix = "…" if end < len(text) else ""
    return f"{prefix}{body}{suffix}", start, end


def generate_snippet(
    query: ParsedQuery,
    source_text: str,
    *,
    max_chars: int = DEFAULT_SNIPPET_MAX_CHARS,
) -> GeneratedSnippet:
    """Generate a deterministic, local-only snippet from already-authorized plain text.

    The generator performs no network or filesystem I/O and receives no authority to
    fetch content. Callers must supply text they are already authorized to process.
    Source text is bounded before processing and is never retained by this result.
    """

    if max_chars < _MIN_SNIPPET_MAX_CHARS:
        raise SnippetGenerationError(
            f"max_chars must be at least {_MIN_SNIPPET_MAX_CHARS}"
        )
    if max_chars > DEFAULT_SNIPPET_MAX_CHARS:
        raise SnippetGenerationError(
            f"max_chars must not exceed {DEFAULT_SNIPPET_MAX_CHARS}"
        )

    text, source_truncated = _normalize_source(source_text)
    if not text:
        return GeneratedSnippet(
            text="",
            matched_query=False,
            source_truncated=source_truncated,
            start_offset=0,
            end_offset=0,
        )

    needles = _needles(query)
    match = _best_match(text, needles)
    matched_query = match is not None

    if match is None:
        # Prefer a complete leading sentence when it fits; otherwise use the same
        # deterministic bounded clipping path as query-focused snippets.
        first_sentence = _SENTENCE_BREAK.split(text, maxsplit=1)[0]
        if first_sentence and len(first_sentence) <= max_chars:
            end = len(first_sentence)
            suffix = "…" if end < len(text) else ""
            return GeneratedSnippet(
                text=f"{first_sentence}{suffix}",
                matched_query=False,
                source_truncated=source_truncated,
                start_offset=0,
                end_offset=end,
            )
        focus_start = 0
        focus_end = min(len(text), max_chars)
    else:
        focus_start, focus_end = match

    snippet, start, end = _clip(
        text,
        focus_start=focus_start,
        focus_end=focus_end,
        max_chars=max_chars,
    )
    return GeneratedSnippet(
        text=snippet,
        matched_query=matched_query,
        source_truncated=source_truncated,
        start_offset=start,
        end_offset=end,
    )
