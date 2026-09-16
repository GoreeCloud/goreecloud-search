from __future__ import annotations

from collections.abc import Sequence
from typing import Protocol

from .index_contract import (
    INDEX_CONTRACT_VERSION,
    IndexCapabilities,
    IndexContractError,
    IndexSearchRequest,
    IndexSearchResponse,
)
from .models import ParsedQuery, ProviderDescriptor, ProviderOrigin
from .providers import ResultCandidate


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

    async def search(self, query: ParsedQuery, *, limit: int) -> Sequence[ResultCandidate]:
        if query.filters.category not in self._capabilities.categories:
            raise IndexContractError(
                f"Index does not support category: {query.filters.category.value}"
            )
        if limit < 1:
            raise IndexContractError("limit must be positive")

        page_size = min(limit, self._capabilities.max_page_size)
        response = await self._transport.search(
            IndexSearchRequest.from_query(query, page_size=page_size)
        )
        response.validate(self._capabilities)

        return tuple(
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
            for item in response.candidates[:limit]
        )
