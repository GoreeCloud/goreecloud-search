from __future__ import annotations

from dataclasses import dataclass
from hashlib import sha256
from ipaddress import AddressValueError, IPv4Address, IPv6Address
from unicodedata import category as unicode_category
from urllib.parse import parse_qsl, urlencode, urlsplit, urlunsplit

from .models import ProviderDescriptor, ProviderOrigin
from .providers import ResultCandidate

_MAX_RESULT_TITLE_CHARS = 512
_MAX_RESULT_SNIPPET_CHARS = 4096
_BIDI_CONTROLS = frozenset(
    {
        "\u061c",
        "\u200e",
        "\u200f",
        "\u202a",
        "\u202b",
        "\u202c",
        "\u202d",
        "\u202e",
        "\u2066",
        "\u2067",
        "\u2068",
        "\u2069",
    }
)

_TRACKING_QUERY_KEYS = frozenset(
    {
        "dclid",
        "fbclid",
        "gclid",
        "igshid",
        "mc_cid",
        "mc_eid",
        "msclkid",
        "vero_conv",
        "vero_id",
    }
)


class ResultNormalizationError(ValueError):
    """Raised when a candidate cannot be normalized safely."""


@dataclass(frozen=True, slots=True)
class ResultProvenance:
    provider: str
    origin: ProviderOrigin
    provider_rank: int | None
    source_id: str | None
    contract_version: str | None
    indexed_by_goreecloud: bool
    content_hash: str | None
    last_crawled_at: str | None


@dataclass(frozen=True, slots=True)
class NormalizedResult:
    result_id: str
    title: str
    url: str
    canonical_url: str
    snippet: str
    published_at: str | None
    content_type: str | None
    language: str | None
    source_agreement: int
    provenance: tuple[ResultProvenance, ...]


def _sanitize_result_text(
    value: str,
    *,
    field: str,
    max_chars: int,
    allow_empty: bool,
) -> str:
    if not isinstance(value, str):
        raise ResultNormalizationError(f"{field} must be text")

    sanitized = "".join(
        " "
        if unicode_category(character) == "Cc" or character in _BIDI_CONTROLS
        else character
        for character in value
    )
    normalized = " ".join(sanitized.split())

    if not normalized and not allow_empty:
        raise ResultNormalizationError(f"{field} must contain visible text")
    if len(normalized) > max_chars:
        normalized = normalized[: max_chars - 1].rstrip() + "…"
    return normalized


def _safe_web_url_parts(url: str):
    if not isinstance(url, str):
        raise ResultNormalizationError("result URL must be text")
    raw = url.strip()
    if not raw or raw != url:
        raise ResultNormalizationError("result URL must not be empty or padded with whitespace")
    if any(
        character.isspace()
        or unicode_category(character) in {"Cc", "Cf"}
        or character == "\\"
        for character in raw
    ):
        raise ResultNormalizationError(
            "result URL contains unsupported whitespace, control, format, or backslash characters"
        )

    try:
        parts = urlsplit(raw)
        host = parts.hostname
    except ValueError as exc:
        raise ResultNormalizationError(f"invalid result URL: {url!r}") from exc

    if parts.scheme.casefold() not in {"http", "https"} or not host:
        raise ResultNormalizationError(f"unsupported or invalid result URL: {url!r}")
    if parts.username is not None or parts.password is not None:
        raise ResultNormalizationError("credential-bearing result URLs are not accepted")
    return raw, parts, host


def _looks_like_numeric_ipv4(host: str) -> bool:
    labels = host.split(".")
    if not 1 <= len(labels) <= 4:
        return False

    def numeric_label(label: str) -> bool:
        if label.isdecimal():
            return True
        lowered = label.casefold()
        return (
            lowered.startswith("0x")
            and len(lowered) > 2
            and all(character in "0123456789abcdef" for character in lowered[2:])
        )

    return all(numeric_label(label) for label in labels)


def _canonicalize_web_host(host: str) -> str:
    # Zone identifiers are machine-local routing details and are not portable
    # result identities. Percent-encoded host ambiguity is rejected with them.
    if "%" in host:
        raise ResultNormalizationError("result URL host contains unsupported percent/zone syntax")

    if ":" in host:
        try:
            return str(IPv6Address(host)).casefold()
        except AddressValueError as exc:
            raise ResultNormalizationError("result URL contains invalid IPv6 host syntax") from exc

    source_host = host.rstrip(".")
    if not source_host:
        raise ResultNormalizationError("result URL host must not be empty")

    try:
        ascii_host = source_host.encode("idna").decode("ascii").casefold()
    except UnicodeError as exc:
        raise ResultNormalizationError("result URL host cannot be represented as a valid IDN") from exc

    if len(ascii_host) > 253:
        raise ResultNormalizationError("result URL host is too long")

    labels = ascii_host.split(".")
    if any(
        not label
        or len(label) > 63
        or label.startswith("-")
        or label.endswith("-")
        or any(not (character.isascii() and (character.isalnum() or character == "-")) for character in label)
        for label in labels
    ):
        raise ResultNormalizationError("result URL host contains invalid DNS label syntax")

    # WHATWG/browser URL stacks may reinterpret shortened, integer, octal-like,
    # or hexadecimal numeric forms as IPv4. Accept only canonical four-octet
    # decimal IPv4 so Search cannot display one host identity and open another.
    if _looks_like_numeric_ipv4(ascii_host):
        try:
            canonical_ipv4 = str(IPv4Address(ascii_host))
        except AddressValueError as exc:
            raise ResultNormalizationError("result URL contains ambiguous numeric host syntax") from exc
        if ascii_host != canonical_ipv4:
            raise ResultNormalizationError("result URL contains non-canonical IPv4 syntax")
        return canonical_ipv4

    return ascii_host


def canonicalize_url(url: str) -> str:
    """Return a conservative canonical URL suitable for result identity.

    The function removes fragments and known tracking parameters, normalizes the
    host and default ports, and otherwise preserves path/query semantics. It does
    not collapse HTTP into HTTPS or invent a canonical URL for the publisher.
    """

    raw, parts, parsed_host = _safe_web_url_parts(url)

    scheme = parts.scheme.casefold()
    host = _canonicalize_web_host(parsed_host)
    try:
        port = parts.port
    except ValueError as exc:
        raise ResultNormalizationError(f"invalid result URL port: {url!r}") from exc

    if port == 0:
        raise ResultNormalizationError("result URL port 0 is not accepted")

    default_port = (scheme == "http" and port == 80) or (scheme == "https" and port == 443)
    host_for_netloc = f"[{host}]" if ":" in host else host
    netloc = host_for_netloc if port is None or default_port else f"{host_for_netloc}:{port}"

    filtered_query: list[tuple[str, str]] = []
    for key, value in parse_qsl(parts.query, keep_blank_values=True):
        normalized_key = key.casefold()
        if normalized_key.startswith("utm_") or normalized_key in _TRACKING_QUERY_KEYS:
            continue
        filtered_query.append((key, value))

    path = parts.path or "/"
    query = urlencode(filtered_query, doseq=True)
    return urlunsplit((scheme, netloc, path, query, ""))


def _result_id(canonical_url: str, content_hash: str | None) -> str:
    identity = f"hash:{content_hash}" if content_hash else f"url:{canonical_url}"
    return sha256(identity.encode("utf-8")).hexdigest()[:24]


def normalize_and_deduplicate(
    candidates: tuple[ResultCandidate, ...] | list[ResultCandidate],
    providers: tuple[ProviderDescriptor, ...] | list[ProviderDescriptor],
) -> tuple[NormalizedResult, ...]:
    """Normalize provider candidates and merge duplicates without reranking them."""

    descriptors = {provider.name.casefold(): provider for provider in providers}
    groups: list[dict[str, object]] = []
    url_groups: dict[str, int] = {}
    hash_groups: dict[str, int] = {}

    for candidate in candidates:
        descriptor = descriptors.get(candidate.provider.casefold())
        if descriptor is None:
            raise ResultNormalizationError(
                f"candidate references undeclared provider: {candidate.provider!r}"
            )

        # Validate the actual result/open URL even when a provider supplies a
        # separate canonical identity. A safe canonical alias must not launder
        # credential-bearing or control-bearing navigation targets.
        result_url_canonical = canonicalize_url(candidate.url)
        canonical = (
            canonicalize_url(candidate.canonical_url)
            if candidate.canonical_url
            else result_url_canonical
        )
        safe_title = _sanitize_result_text(
            candidate.title,
            field="result title",
            max_chars=_MAX_RESULT_TITLE_CHARS,
            allow_empty=False,
        )
        safe_snippet = _sanitize_result_text(
            candidate.snippet,
            field="result snippet",
            max_chars=_MAX_RESULT_SNIPPET_CHARS,
            allow_empty=True,
        )
        content_hash = candidate.content_hash.casefold() if candidate.content_hash else None

        indexes = {
            index
            for index in (
                url_groups.get(canonical),
                hash_groups.get(content_hash) if content_hash else None,
            )
            if index is not None
        }

        provenance = ResultProvenance(
            provider=descriptor.name,
            origin=descriptor.origin,
            provider_rank=candidate.provider_rank,
            source_id=candidate.source_id,
            contract_version=candidate.provider_contract_version,
            indexed_by_goreecloud=descriptor.origin is ProviderOrigin.GOREECLOUD_INDEX,
            content_hash=content_hash,
            last_crawled_at=candidate.last_crawled_at,
        )

        if not indexes:
            index = len(groups)
            groups.append(
                {
                    "candidate": candidate,
                    "title": safe_title,
                    "snippet": safe_snippet,
                    "canonical": canonical,
                    "hash": content_hash,
                    "provenance": [provenance],
                    "active": True,
                }
            )
        else:
            index = min(indexes)
            group = groups[index]
            assert group["active"] is True
            group_provenance = group["provenance"]
            assert isinstance(group_provenance, list)
            group_provenance.append(provenance)
            if group["hash"] is None and content_hash:
                group["hash"] = content_hash

            for other_index in sorted(indexes - {index}, reverse=True):
                other = groups[other_index]
                if other["active"] is not True:
                    continue
                other_provenance = other["provenance"]
                assert isinstance(other_provenance, list)
                group_provenance.extend(other_provenance)
                other["active"] = False
                for key, mapped in list(url_groups.items()):
                    if mapped == other_index:
                        url_groups[key] = index
                for key, mapped in list(hash_groups.items()):
                    if mapped == other_index:
                        hash_groups[key] = index

        url_groups[canonical] = index
        if content_hash:
            hash_groups[content_hash] = index

    normalized: list[NormalizedResult] = []
    for group in groups:
        if group["active"] is not True:
            continue
        candidate = group["candidate"]
        provenance = group["provenance"]
        title = group["title"]
        snippet = group["snippet"]
        canonical = group["canonical"]
        content_hash = group["hash"]
        assert isinstance(candidate, ResultCandidate)
        assert isinstance(provenance, list)
        assert isinstance(title, str)
        assert isinstance(snippet, str)
        assert isinstance(canonical, str)
        assert content_hash is None or isinstance(content_hash, str)

        unique_providers = {item.provider.casefold() for item in provenance}
        normalized.append(
            NormalizedResult(
                result_id=_result_id(canonical, content_hash),
                title=title,
                url=candidate.url,
                canonical_url=canonical,
                snippet=snippet,
                published_at=candidate.published_at,
                content_type=candidate.content_type,
                language=candidate.language,
                source_agreement=len(unique_providers),
                provenance=tuple(provenance),
            )
        )

    return tuple(normalized)
