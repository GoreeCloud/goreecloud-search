from __future__ import annotations

from collections.abc import Iterable
from dataclasses import dataclass

from .execution import ExecutionPolicy, ExecutionReport, SearchExecutor
from .models import ParsedQuery, ProviderDescriptor, SourceMode, SourcePlan
from .normalization import NormalizedResult, normalize_and_deduplicate
from .planner import plan_sources
from .providers import ResultCandidate, SearchProvider
from .query_parser import parse_query
from .ranking import RankedResult, rank_results


@dataclass(frozen=True, slots=True)
class SearchResponse:
    query: ParsedQuery
    plan: SourcePlan
    execution: ExecutionReport
    results: tuple[RankedResult, ...]


class SearchCore:
    """Search orchestration core for parsing, execution, normalization, and ranking."""

    def __init__(
        self,
        providers: Iterable[ProviderDescriptor] = (),
        *,
        provider_adapters: Iterable[SearchProvider] = (),
        execution_policy: ExecutionPolicy | None = None,
    ) -> None:
        adapters = tuple(provider_adapters)
        descriptors: dict[str, ProviderDescriptor] = {
            provider.name.casefold(): provider for provider in providers
        }
        for adapter in adapters:
            descriptor = adapter.descriptor
            existing = descriptors.get(descriptor.name.casefold())
            if existing is not None and existing != descriptor:
                raise ValueError(
                    f"provider descriptor conflicts with execution adapter: {descriptor.name!r}"
                )
            descriptors[descriptor.name.casefold()] = descriptor

        self._providers = tuple(descriptors.values())
        self._provider_adapters = adapters
        self._execution_policy = execution_policy or ExecutionPolicy()

    @property
    def providers(self) -> tuple[ProviderDescriptor, ...]:
        return self._providers

    def parse(self, raw_query: str) -> ParsedQuery:
        return parse_query(raw_query)

    def plan(
        self,
        raw_query: str,
        *,
        mode: SourceMode = SourceMode.INDEX_FIRST,
    ) -> tuple[ParsedQuery, SourcePlan]:
        query = self.parse(raw_query)
        return query, plan_sources(query, mode, self._providers)

    def normalize(self, candidates: Iterable[ResultCandidate]) -> tuple[NormalizedResult, ...]:
        return normalize_and_deduplicate(tuple(candidates), self._providers)

    def rank(
        self,
        query: ParsedQuery,
        results: Iterable[NormalizedResult],
    ) -> tuple[RankedResult, ...]:
        return rank_results(query, tuple(results))

    async def search(
        self,
        raw_query: str,
        *,
        mode: SourceMode = SourceMode.INDEX_FIRST,
        limit: int = 10,
    ) -> SearchResponse:
        query, plan = self.plan(raw_query, mode=mode)
        executor = SearchExecutor(
            self._provider_adapters,
            policy=self._execution_policy,
        )
        execution = await executor.execute(query, plan, limit=limit)
        normalized = self.normalize(execution.candidates)
        ranked = self.rank(query, normalized)
        return SearchResponse(
            query=query,
            plan=plan,
            execution=execution,
            results=ranked[:limit],
        )
