from __future__ import annotations

from collections.abc import Iterable

from .models import (
    ParsedQuery,
    ProviderDescriptor,
    ProviderOrigin,
    QueryDisclosureBudget,
    SourceMode,
    SourcePlan,
    SourcePlanStep,
)


class SourcePlanningError(ValueError):
    """Raised when no provider can satisfy the requested plan."""


INDEX_ORIGINATED_DELEGATION_CONTRACT_VERSION = "goreecloud.search-index-delegation.v1"
INDEX_ORIGINATED_DELEGATION_MODE = SourceMode.EXTERNAL_ONLY.value
INDEX_ORIGINATED_INDEX_PROVIDER_REENTRY_ALLOWED = False
INDEX_ORIGINATED_FALLBACK_ALLOWED = False


def _eligible(query: ParsedQuery, providers: Iterable[ProviderDescriptor]) -> list[ProviderDescriptor]:
    requested_sources = set(query.filters.sources)
    eligible = [
        provider
        for provider in providers
        if provider.enabled
        and query.filters.category in provider.categories
        and (not requested_sources or provider.name.casefold() in requested_sources)
    ]
    return sorted(eligible, key=lambda provider: (provider.priority, provider.name.casefold()))


def _step(provider: ProviderDescriptor, stage: str, reason: str) -> SourcePlanStep:
    return SourcePlanStep(provider=provider.name, stage=stage, origin=provider.origin, reason=reason)


def _apply_disclosure_budget(
    providers: list[ProviderDescriptor],
    budget: QueryDisclosureBudget,
) -> tuple[list[ProviderDescriptor], int]:
    maximum = budget.max_third_party_providers
    if maximum is None:
        return providers, 0

    retained: list[ProviderDescriptor] = []
    disclosed = 0
    omitted = 0
    for provider in providers:
        if provider.third_party_query_disclosure:
            if disclosed >= maximum:
                omitted += 1
                continue
            disclosed += 1
        retained.append(provider)
    return retained, omitted


def plan_sources(
    query: ParsedQuery,
    mode: SourceMode,
    providers: Iterable[ProviderDescriptor],
    *,
    disclosure_budget: QueryDisclosureBudget | None = None,
) -> SourcePlan:
    """Create a deterministic provider plan without executing any provider.

    Source modes that prohibit third-party query disclosure never include external
    providers. A supplied disclosure budget is applied before the plan is returned,
    so providers over budget never receive the query.
    """

    budget = disclosure_budget or QueryDisclosureBudget()
    eligible = _eligible(query, providers)
    budgeted_eligible, omitted_by_budget = _apply_disclosure_budget(eligible, budget)
    steps: list[SourcePlanStep] = []

    if mode is SourceMode.INDEX_FIRST:
        native = [p for p in budgeted_eligible if p.origin in {ProviderOrigin.GOREECLOUD_INDEX, ProviderOrigin.GOREECLOUD_SERVICE, ProviderOrigin.LOCAL}]
        external = [p for p in budgeted_eligible if p.origin is ProviderOrigin.EXTERNAL]
        steps.extend(_step(p, "primary", "first-party or local source") for p in native)
        steps.extend(_step(p, "fallback", "external gap-filling source") for p in external)
    elif mode is SourceMode.FEDERATED:
        steps.extend(_step(p, "primary", "federated source") for p in budgeted_eligible)
    elif mode is SourceMode.GOREECLOUD_ONLY:
        steps.extend(_step(p, "primary", "GoreeCloud-only policy") for p in budgeted_eligible if p.origin is not ProviderOrigin.EXTERNAL)
    elif mode is SourceMode.EXTERNAL_ONLY:
        steps.extend(_step(p, "primary", "external-only policy") for p in budgeted_eligible if p.origin is ProviderOrigin.EXTERNAL)
    elif mode is SourceMode.OFFLINE_LOCAL:
        steps.extend(_step(p, "primary", "offline local policy") for p in budgeted_eligible if p.local_only or p.origin is ProviderOrigin.LOCAL)
    else:
        raise SourcePlanningError(f"unsupported source mode: {mode!r}")

    if not steps:
        requested = ", ".join(query.filters.sources) or "configured providers"
        budget_detail = (
            f", disclosure_budget={budget.max_third_party_providers}"
            if budget.max_third_party_providers is not None
            else ""
        )
        raise SourcePlanningError(
            f"no eligible provider for category={query.filters.category.value}, "
            f"mode={mode.value}, sources={requested}{budget_detail}"
        )

    selected_names = {step.provider.casefold() for step in steps}
    selected = [p for p in budgeted_eligible if p.name.casefold() in selected_names]
    third_party_count = sum(1 for p in selected if p.third_party_query_disclosure)
    external_disclosure = third_party_count > 0

    maximum = budget.max_third_party_providers
    if maximum is not None and third_party_count > maximum:
        raise AssertionError("privacy invariant violated: disclosure budget exceeded")
    if mode in {SourceMode.GOREECLOUD_ONLY, SourceMode.OFFLINE_LOCAL} and external_disclosure:
        raise AssertionError("privacy invariant violated: prohibited external disclosure")

    return SourcePlan(
        mode=mode,
        category=query.filters.category,
        steps=tuple(steps),
        third_party_query_disclosure=external_disclosure,
        third_party_provider_count=third_party_count,
        disclosure_budget=maximum,
        third_party_providers_omitted=omitted_by_budget,
    )


def plan_index_originated_delegation(
    query: ParsedQuery,
    providers: Iterable[ProviderDescriptor],
    *,
    disclosure_budget: QueryDisclosureBudget | None = None,
) -> SourcePlan:
    """Plan an Index-originated Internet/web delegation without recursive Index re-entry.

    Index is the caller on this path, so Search may execute only explicitly configured
    external providers. The caller cannot select INDEX_FIRST/FEDERATED behavior, Search
    cannot execute a GOREECLOUD_INDEX/GOREECLOUD_SERVICE/LOCAL provider, and no fallback
    stage is permitted.
    """

    plan = plan_sources(
        query,
        SourceMode.EXTERNAL_ONLY,
        providers,
        disclosure_budget=disclosure_budget,
    )
    if plan.mode is not SourceMode.EXTERNAL_ONLY:
        raise AssertionError("Index-originated delegation must remain external-only")
    if any(step.stage != "primary" for step in plan.steps):
        raise AssertionError("Index-originated delegation must not create fallback stages")
    if any(step.origin is not ProviderOrigin.EXTERNAL for step in plan.steps):
        raise AssertionError("Index-originated delegation must not re-enter GoreeCloud Index or another non-external provider")
    return plan
