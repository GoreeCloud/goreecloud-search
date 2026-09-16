from __future__ import annotations

from collections.abc import Iterable
from dataclasses import dataclass

from .content_policy import ContentPolicyEngine, ContentPolicyHook, ContentPolicyReport, SafeSearchMode
from .execution import ExecutionPolicy, ExecutionReport, SearchExecutor
from .models import ParsedQuery, ProviderDescriptor, QueryDisclosureBudget, SourceMode, SourcePlan
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
    content_policy: ContentPolicyReport
    results: tuple[RankedResult, ...]


class SearchCore:
    """Search orchestration core for parsing, execution, policy, and ranking."""

    def __init__(
        self,
        providers: Iterable[ProviderDescriptor] = (),
        *,
        provider_adapters: Iterable[SearchProvider] = (),
        execution_policy: ExecutionPolicy | None = None,
        content_policy_hooks: Iterable[ContentPolicyHook] = (),
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
        self._content_policy = ContentPolicyEngine(tuple(content_policy_hooks))

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
        disclosure_budget: QueryDisclosureBudget | None = None,
    ) -> tuple[ParsedQuery, SourcePlan]:
        query = self.parse(raw_query)
        return query, plan_sources(
            query,
            mode,
            self._providers,
            disclosure_budget=disclosure_budget,
        )

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
        disclosure_budget: QueryDisclosureBudget | None = None,
        safe_search: SafeSearchMode = SafeSearchMode.OFF,
    ) -> SearchResponse:
        self._content_policy.ensure_safe_search_supported(safe_search)
        query, plan = self.plan(
            raw_query,
            mode=mode,
            disclosure_budget=disclosure_budget,
        )
        executor = SearchExecutor(self._provider_adapters, policy=self._execution_policy)
        execution = await executor.execute(query, plan, limit=limit)
        normalized = self.normalize(execution.candidates)
        content_policy = self._content_policy.apply(
            normalized,
            safe_search=safe_search,
        )
        ranked = self.rank(query, content_policy.visible_results)
        return SearchResponse(
            query=query,
            plan=plan,
            execution=execution,
            content_policy=content_policy,
            results=ranked[:limit],
        )
