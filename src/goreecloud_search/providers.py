from __future__ import annotations

from dataclasses import dataclass
from typing import Protocol

from .models import ParsedQuery, ProviderDescriptor


@dataclass(frozen=True, slots=True)
class ResultCandidate:
    title: str
    url: str
    snippet: str
    provider: str
    provider_rank: int | None = None
    published_at: str | None = None
    content_type: str | None = None
    canonical_url: str | None = None
    source_id: str | None = None
    content_hash: str | None = None
    language: str | None = None
    last_crawled_at: str | None = None
    provider_contract_version: str | None = None


@dataclass(frozen=True, slots=True)
class ProviderSearchBatch:
    """One provider's bounded response to a Search execution request."""

    candidates: tuple[ResultCandidate, ...]
    degraded: bool = False
    warnings: tuple[str, ...] = ()


class SearchProvider(Protocol):
    """Replaceable provider boundary for Index and federated adapters."""

    @property
    def descriptor(self) -> ProviderDescriptor:
        ...

    async def search(self, query: ParsedQuery, *, limit: int) -> ProviderSearchBatch:
        ...
