from __future__ import annotations

from dataclasses import dataclass

from .models import ParsedQuery, SearchCategory

INDEX_CONTRACT_VERSION = "goreecloud.search-index.v1"


class IndexContractError(ValueError):
    """Raised when Search and Index contract data is invalid or incompatible."""


@dataclass(frozen=True, slots=True)
class IndexCapabilities:
    contract_version: str
    categories: frozenset[SearchCategory]
    max_page_size: int = 100
    supports_canonical_urls: bool = True
    supports_content_hashes: bool = True
    supports_last_crawled_at: bool = True

    def __post_init__(self) -> None:
        if self.contract_version != INDEX_CONTRACT_VERSION:
            raise IndexContractError(
                f"unsupported Index contract version: {self.contract_version!r}"
            )
        if not self.categories:
            raise IndexContractError("Index must advertise at least one search category")
        if self.max_page_size < 1:
            raise IndexContractError("Index max_page_size must be positive")


@dataclass(frozen=True, slots=True)
class IndexSearchRequest:
    contract_version: str
    terms: tuple[str, ...]
    phrases: tuple[str, ...]
    excluded_terms: tuple[str, ...]
    sites: tuple[str, ...]
    excluded_domains: tuple[str, ...]
    filetypes: tuple[str, ...]
    before: str | None
    after: str | None
    language: str | None
    region: str | None
    category: SearchCategory
    lens: str | None
    page_size: int
    cursor: str | None = None

    @classmethod
    def from_query(
        cls,
        query: ParsedQuery,
        *,
        page_size: int,
        cursor: str | None = None,
    ) -> "IndexSearchRequest":
        if page_size < 1:
            raise IndexContractError("page_size must be positive")
        return cls(
            contract_version=INDEX_CONTRACT_VERSION,
            terms=query.terms,
            phrases=query.phrases,
            excluded_terms=query.excluded_terms,
            sites=query.filters.sites,
            excluded_domains=query.filters.excluded_domains,
            filetypes=query.filters.filetypes,
            before=query.filters.before.isoformat() if query.filters.before else None,
            after=query.filters.after.isoformat() if query.filters.after else None,
            language=query.filters.language,
            region=query.filters.region,
            category=query.filters.category,
            lens=query.filters.lens,
            page_size=page_size,
            cursor=cursor,
        )


@dataclass(frozen=True, slots=True)
class IndexDocumentCandidate:
    document_id: str
    title: str
    url: str
    snippet: str
    rank: int | None = None
    canonical_url: str | None = None
    content_hash: str | None = None
    published_at: str | None = None
    last_crawled_at: str | None = None
    content_type: str | None = None
    language: str | None = None

    def __post_init__(self) -> None:
        if not self.document_id.strip():
            raise IndexContractError("Index document_id must not be empty")
        if not self.title.strip():
            raise IndexContractError("Index candidate title must not be empty")
        if not self.url.strip():
            raise IndexContractError("Index candidate URL must not be empty")
        if self.rank is not None and self.rank < 0:
            raise IndexContractError("Index candidate rank must be non-negative")


@dataclass(frozen=True, slots=True)
class IndexSearchResponse:
    contract_version: str
    candidates: tuple[IndexDocumentCandidate, ...]
    next_cursor: str | None = None
    degraded: bool = False
    warnings: tuple[str, ...] = ()

    def validate(self, capabilities: IndexCapabilities) -> None:
        if self.contract_version != INDEX_CONTRACT_VERSION:
            raise IndexContractError(
                f"incompatible Index response contract: {self.contract_version!r}"
            )
        if capabilities.contract_version != self.contract_version:
            raise IndexContractError("Index capability and response contract versions differ")
