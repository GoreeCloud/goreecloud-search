from __future__ import annotations

from typing import Protocol

from .index_contract import (
    INDEX_CONTRACT_VERSION,
    IndexCapabilities,
    IndexContractError,
    IndexSearchRequest,
    IndexSearchResponse,
)
from .models import ParsedQuery, ProviderDescriptor, ProviderOrigin
from .providers import ProviderSearchBatch, ResultCandidate


class IndexTransport(Protocol):
    """Transport boundary implemented later by the authenticated Index client."""

    async def search(self, request: IndexSearchRequest) -> IndexSearchResponse:
        ...


class GoreeCloudIndexProvider:
    """First-party Search provider adapter for the versioned GoreeCloud Index contract.

    The adapter contains no network implementation. A transport must be injected so
    authentication, Privacy Shield, Wardveil, Mesh, timeout, and deployment concerns
    can be implemented and verified at the proper boundary later.
    """

    def __init__(
        self,
        transport: IndexTransport,
        capabilities: IndexCapabilities,
        *,
        name: str = "goreecloud-index",
        priority: int = 10,
    ) -> None:
        self._transport = transport
        self._capabilities = capabilities
        self._descriptor = ProviderDescriptor(
            name=name,
            origin=ProviderOrigin.GOREECLOUD_INDEX,
            categories=capabilities.categories,
            priority=priority,
            third_party_query_disclosure=False,
        )

    @property
    def descriptor(self) -> ProviderDescriptor:
        return self._descriptor

    @property
    def capabilities(self) -> IndexCapabilities:
        return self._capabilities

    async def search(self, query: ParsedQuery, *, limit: int) -> ProviderSearchBatch:
        if query.filters.category not in self._capabilities.categories:
            raise IndexContractError(
                f"Index does not support category: {query.filters.category.value}"
            )
        if limit < 1:
            raise IndexContractError("limit must be positive")

        candidates: list[ResultCandidate] = []
        warnings: list[str] = []
        degraded = False
        cursor: str | None = None
        seen_cursors: set[str] = set()

        while len(candidates) < limit:
            page_size = min(limit - len(candidates), self._capabilities.max_page_size)
            response = await self._transport.search(
                IndexSearchRequest.from_query(query, page_size=page_size, cursor=cursor)
            )
            response.validate(self._capabilities)
            degraded = degraded or response.degraded
            warnings.extend(response.warnings)

            candidates.extend(
                ResultCandidate(
                    title=item.title,
                    url=item.url,
                    snippet=item.snippet,
                    provider=self._descriptor.name,
                    provider_rank=item.rank,
                    published_at=item.published_at,
                    content_type=item.content_type,
                    canonical_url=item.canonical_url,
                    source_id=item.document_id,
                    content_hash=item.content_hash,
                    language=item.language,
                    last_crawled_at=item.last_crawled_at,
                    provider_contract_version=INDEX_CONTRACT_VERSION,
                )
                for item in response.candidates
            )

            next_cursor = response.next_cursor
            if not next_cursor or len(candidates) >= limit:
                break
            if next_cursor in seen_cursors or next_cursor == cursor:
                raise IndexContractError("Index pagination cursor repeated")
            seen_cursors.add(next_cursor)
            if not response.candidates:
                degraded = True
                warnings.append("Index returned an empty page with a continuation cursor")
                break
            cursor = next_cursor

        unique_warnings = tuple(dict.fromkeys(warnings))
        return ProviderSearchBatch(
            candidates=tuple(candidates[:limit]),
            degraded=degraded,
            warnings=unique_warnings,
        )
