from __future__ import annotations

from collections.abc import Iterable

from .models import ParsedQuery, ProviderDescriptor, SourceMode, SourcePlan
from .planner import plan_sources
from .query_parser import parse_query


class SearchCore:
    """Side-effect-free Search orchestration core for parsing and source planning."""

    def __init__(self, providers: Iterable[ProviderDescriptor] = ()) -> None:
        self._providers = tuple(providers)

    @property
    def providers(self) -> tuple[ProviderDescriptor, ...]:
        return self._providers

    def parse(self, raw_query: str) -> ParsedQuery:
        return parse_query(raw_query)

    def plan(self, raw_query: str, *, mode: SourceMode = SourceMode.INDEX_FIRST) -> tuple[ParsedQuery, SourcePlan]:
        query = self.parse(raw_query)
        return query, plan_sources(query, mode, self._providers)
