from __future__ import annotations

from dataclasses import dataclass, field
from datetime import date
from enum import Enum


class SearchCategory(str, Enum):
    GENERAL = "general"
    IMAGES = "images"
    VIDEOS = "videos"
    NEWS = "news"
    FILES = "files"
    DISCUSSIONS = "discussions"
    FORUMS = "forums"
    SOCIAL = "social"
    MAPS = "maps"
    SHOPPING = "shopping"
    ACADEMIC = "academic"
    DOCUMENTATION = "documentation"
    SOURCE_CODE = "source_code"
    BOOKS = "books"
    PODCASTS = "podcasts"
    MUSIC = "music"
    PEOPLE = "people"
    SOFTWARE = "software"


class SourceMode(str, Enum):
    INDEX_FIRST = "index_first"
    FEDERATED = "federated"
    GOREECLOUD_ONLY = "goreecloud_only"
    EXTERNAL_ONLY = "external_only"
    OFFLINE_LOCAL = "offline_local"


class ProviderOrigin(str, Enum):
    GOREECLOUD_INDEX = "goreecloud_index"
    GOREECLOUD_SERVICE = "goreecloud_service"
    LOCAL = "local"
    EXTERNAL = "external"


@dataclass(frozen=True, slots=True)
class QueryFilters:
    sites: tuple[str, ...] = ()
    excluded_domains: tuple[str, ...] = ()
    filetypes: tuple[str, ...] = ()
    before: date | None = None
    after: date | None = None
    language: str | None = None
    region: str | None = None
    sources: tuple[str, ...] = ()
    category: SearchCategory = SearchCategory.GENERAL
    lens: str | None = None


@dataclass(frozen=True, slots=True)
class ParsedQuery:
    raw: str
    terms: tuple[str, ...] = ()
    phrases: tuple[str, ...] = ()
    excluded_terms: tuple[str, ...] = ()
    filters: QueryFilters = field(default_factory=QueryFilters)


@dataclass(frozen=True, slots=True)
class ProviderDescriptor:
    name: str
    origin: ProviderOrigin
    categories: frozenset[SearchCategory]
    enabled: bool = True
    local_only: bool = False
    priority: int = 100
    third_party_query_disclosure: bool = False

    def __post_init__(self) -> None:
        if not self.name.strip():
            raise ValueError("provider name must not be empty")
        if not self.categories:
            raise ValueError("provider must support at least one category")
        if self.priority < 0:
            raise ValueError("provider priority must be non-negative")
        if self.origin is ProviderOrigin.EXTERNAL and not self.third_party_query_disclosure:
            object.__setattr__(self, "third_party_query_disclosure", True)


@dataclass(frozen=True, slots=True)
class SourcePlanStep:
    provider: str
    stage: str
    origin: ProviderOrigin
    reason: str


@dataclass(frozen=True, slots=True)
class SourcePlan:
    mode: SourceMode
    category: SearchCategory
    steps: tuple[SourcePlanStep, ...]
    third_party_query_disclosure: bool
