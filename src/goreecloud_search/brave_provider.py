from __future__ import annotations

import asyncio
from dataclasses import dataclass
import json
from json import JSONDecodeError
from typing import Any, Mapping, Protocol
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode
from urllib.request import Request, urlopen

from .models import ParsedQuery, ProviderDescriptor, ProviderOrigin, SearchCategory
from .providers import ProviderSearchBatch, ResultCandidate


BRAVE_WEB_SEARCH_URL = "https://api.search.brave.com/res/v1/web/search"
BRAVE_MAX_RESULTS = 20
BRAVE_MAX_QUERY_CHARS = 600
BRAVE_MAX_QUERY_WORDS = 75
_MAX_RESPONSE_BYTES = 2 * 1024 * 1024
_SAFESEARCH_VALUES = frozenset({"off", "moderate", "strict"})


class BraveSearchProviderError(RuntimeError):
    """Raised when the Brave Search adapter cannot safely complete a request."""


@dataclass(frozen=True, slots=True)
class BraveWebSearchRequest:
    query: str
    count: int
    country: str | None
    search_lang: str | None
    safesearch: str
    freshness: str | None = None


class BraveSearchTransport(Protocol):
    async def search(
        self,
        request: BraveWebSearchRequest,
        *,
        api_key: str,
    ) -> Mapping[str, Any]:
        ...


def _quote(value: str) -> str:
    escaped = value.replace("\\", "\\\\").replace('"', '\\"')
    return f'"{escaped}"'


def _provider_query_text(query: ParsedQuery) -> str:
    parts: list[str] = []
    parts.extend(query.terms)
    parts.extend(_quote(value) for value in query.phrases)
    parts.extend(
        f"-{_quote(value) if any(character.isspace() for character in value) else value}"
        for value in query.excluded_terms
    )
    parts.extend(f"site:{domain}" for domain in query.filters.sites)
    parts.extend(f"-site:{domain}" for domain in query.filters.excluded_domains)
    parts.extend(f"filetype:{extension}" for extension in query.filters.filetypes)
    rendered = " ".join(part for part in parts if part).strip()
    if not rendered:
        raise BraveSearchProviderError(
            "provider query is empty after removing Search-local operators"
        )
    if (
        len(rendered) > BRAVE_MAX_QUERY_CHARS
        or len(rendered.split()) > BRAVE_MAX_QUERY_WORDS
    ):
        raise BraveSearchProviderError("provider query exceeds Brave Search API bounds")
    return rendered


def _freshness(query: ParsedQuery) -> str | None:
    after = query.filters.after
    before = query.filters.before
    if after is None and before is None:
        return None
    if after is None or before is None:
        raise BraveSearchProviderError(
            "Brave provider currently requires both after: and before: "
            "for a bounded date range"
        )
    return f"{after.isoformat()}to{before.isoformat()}"


def _country(query: ParsedQuery, configured: str | None) -> str | None:
    value = query.filters.region or configured
    if value is None:
        return None
    normalized = value.strip().upper()
    if len(normalized) != 2 or not normalized.isalpha():
        raise BraveSearchProviderError(
            "Brave provider country must be a two-letter code"
        )
    return normalized


def _language(query: ParsedQuery, configured: str | None) -> str | None:
    value = query.filters.language or configured
    if value is None:
        return None
    normalized = value.strip().lower()
    if not (2 <= len(normalized) <= 16) or not all(
        character.isalpha() or character == "-" for character in normalized
    ):
        raise BraveSearchProviderError("Brave provider language is invalid")
    return normalized


class UrllibBraveSearchTransport:
    """Dependency-free HTTPS transport for the fixed Brave Search endpoint."""

    def __init__(self, *, timeout_seconds: float = 2.5) -> None:
        if timeout_seconds <= 0:
            raise ValueError("timeout_seconds must be positive")
        self._timeout_seconds = timeout_seconds

    async def search(
        self,
        request: BraveWebSearchRequest,
        *,
        api_key: str,
    ) -> Mapping[str, Any]:
        return await asyncio.to_thread(self._search_sync, request, api_key)

    def _search_sync(
        self,
        request: BraveWebSearchRequest,
        api_key: str,
    ) -> Mapping[str, Any]:
        params: dict[str, str] = {
            "q": request.query,
            "count": str(request.count),
            "safesearch": request.safesearch,
        }
        if request.country is not None:
            params["country"] = request.country
        if request.search_lang is not None:
            params["search_lang"] = request.search_lang
        if request.freshness is not None:
            params["freshness"] = request.freshness

        url = f"{BRAVE_WEB_SEARCH_URL}?{urlencode(params)}"
        http_request = Request(
            url,
            headers={
                "Accept": "application/json",
                "Accept-Encoding": "identity",
                "User-Agent": "GoreeCloud-Search/0.1",
                "X-Subscription-Token": api_key,
            },
            method="GET",
        )
        try:
            with urlopen(http_request, timeout=self._timeout_seconds) as response:
                body = response.read(_MAX_RESPONSE_BYTES + 1)
        except HTTPError as exc:
            if exc.code in {401, 403}:
                raise BraveSearchProviderError(
                    "Brave Search API authentication failed"
                ) from exc
            if exc.code == 429:
                raise BraveSearchProviderError(
                    "Brave Search API rate limit reached"
                ) from exc
            raise BraveSearchProviderError(
                f"Brave Search API returned HTTP status {exc.code}"
            ) from exc
        except (URLError, TimeoutError, OSError) as exc:
            raise BraveSearchProviderError(
                "Brave Search API transport failed"
            ) from exc

        if len(body) > _MAX_RESPONSE_BYTES:
            raise BraveSearchProviderError(
                "Brave Search API response exceeded size limit"
            )
        try:
            payload = json.loads(body)
        except (JSONDecodeError, UnicodeDecodeError) as exc:
            raise BraveSearchProviderError(
                "Brave Search API returned malformed JSON"
            ) from exc
        if not isinstance(payload, dict):
            raise BraveSearchProviderError(
                "Brave Search API returned an invalid response object"
            )
        return payload


class BraveWebSearchProvider:
    """Opt-in external provider for Brave Web Search API results."""

    def __init__(
        self,
        *,
        api_key: str,
        transport: BraveSearchTransport | None = None,
        country: str | None = None,
        search_lang: str | None = None,
        safesearch: str = "moderate",
        priority: int = 100,
    ) -> None:
        secret = api_key.strip()
        if not secret:
            raise ValueError("api_key must not be empty")
        safe = safesearch.strip().lower()
        if safe not in _SAFESEARCH_VALUES:
            raise ValueError("safesearch must be off, moderate, or strict")
        self._api_key = secret
        self._transport = transport or UrllibBraveSearchTransport()
        self._country = country
        self._search_lang = search_lang
        self._safesearch = safe
        self._descriptor = ProviderDescriptor(
            name="brave",
            origin=ProviderOrigin.EXTERNAL,
            categories=frozenset({SearchCategory.GENERAL}),
            priority=priority,
            third_party_query_disclosure=True,
        )

    @property
    def descriptor(self) -> ProviderDescriptor:
        return self._descriptor

    async def search(
        self,
        query: ParsedQuery,
        *,
        limit: int,
    ) -> ProviderSearchBatch:
        if limit < 1:
            raise ValueError("limit must be positive")
        request = BraveWebSearchRequest(
            query=_provider_query_text(query),
            count=min(limit, BRAVE_MAX_RESULTS),
            country=_country(query, self._country),
            search_lang=_language(query, self._search_lang),
            safesearch=self._safesearch,
            freshness=_freshness(query),
        )
        payload = await self._transport.search(request, api_key=self._api_key)
        web = payload.get("web")
        if web is None:
            return ProviderSearchBatch(
                candidates=(),
                degraded=True,
                warnings=(
                    "Brave response did not contain a web-results section.",
                ),
            )
        if not isinstance(web, Mapping):
            raise BraveSearchProviderError(
                "Brave web-results section is invalid"
            )
        raw_results = web.get("results", ())
        if not isinstance(raw_results, list):
            raise BraveSearchProviderError("Brave results list is invalid")

        candidates: list[ResultCandidate] = []
        malformed = 0
        for rank, raw in enumerate(
            raw_results[: request.count],
            start=1,
        ):
            if not isinstance(raw, Mapping):
                malformed += 1
                continue
            title = raw.get("title")
            url = raw.get("url")
            description = raw.get("description", "")
            if (
                not isinstance(title, str)
                or not title.strip()
                or not isinstance(url, str)
                or not url.strip()
            ):
                malformed += 1
                continue
            if not isinstance(description, str):
                description = ""
            candidates.append(
                ResultCandidate(
                    title=title.strip(),
                    url=url.strip(),
                    snippet=description.strip(),
                    provider=self.descriptor.name,
                    provider_rank=rank,
                    content_type="text/html",
                )
            )

        warnings = (
            (f"Brave response omitted {malformed} malformed result(s).",)
            if malformed
            else ()
        )
        return ProviderSearchBatch(
            candidates=tuple(candidates),
            degraded=bool(malformed),
            warnings=warnings,
        )
