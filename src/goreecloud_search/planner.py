from __future__ import annotations

from collections.abc import Iterable

from .models import ParsedQuery, ProviderDescriptor, ProviderOrigin, SourceMode, SourcePlan, SourcePlanStep


class SourcePlanningError(ValueError):
    """Raised when no provider can satisfy the requested plan."""


def _eligible(query: ParsedQuery, providers: Iterable[ProviderDescriptor]) -> list[ProviderDescriptor]:
    requested_sources = set(query.filters.sources)
    eligible = [
        provider for provider in providers
        if provider.enabled
        and query.filters.category in provider.categories
        and (not requested_sources or provider.name.casefold() in requested_sources)
    ]
    return sorted(eligible, key=lambda provider: (provider.priority, provider.name.casefold()))


def _step(provider: ProviderDescriptor, stage: str, reason: str) -> SourcePlanStep:
    return SourcePlanStep(provider=provider.name, stage=stage, origin=provider.origin, reason=reason)


def plan_sources(query: ParsedQuery, mode: SourceMode, providers: Iterable[ProviderDescriptor]) -> SourcePlan:
    """Create a deterministic provider plan without executing any provider."""
    eligible = _eligible(query, providers)
    steps: list[SourcePlanStep] = []

    if mode is SourceMode.INDEX_FIRST:
        native = [p for p in eligible if p.origin in {ProviderOrigin.GOREECLOUD_INDEX, ProviderOrigin.GOREECLOUD_SERVICE, ProviderOrigin.LOCAL}]
        external = [p for p in eligible if p.origin is ProviderOrigin.EXTERNAL]
        steps.extend(_step(p, "primary", "first-party or local source") for p in native)
        steps.extend(_step(p, "fallback", "external gap-filling source") for p in external)
    elif mode is SourceMode.FEDERATED:
        steps.extend(_step(p, "primary", "federated source") for p in eligible)
    elif mode is SourceMode.GOREECLOUD_ONLY:
        steps.extend(_step(p, "primary", "GoreeCloud-only policy") for p in eligible if p.origin is not ProviderOrigin.EXTERNAL)
    elif mode is SourceMode.EXTERNAL_ONLY:
        steps.extend(_step(p, "primary", "external-only policy") for p in eligible if p.origin is ProviderOrigin.EXTERNAL)
    elif mode is SourceMode.OFFLINE_LOCAL:
        steps.extend(_step(p, "primary", "offline local policy") for p in eligible if p.local_only or p.origin is ProviderOrigin.LOCAL)
    else:
        raise SourcePlanningError(f"unsupported source mode: {mode!r}")

    if not steps:
        requested = ", ".join(query.filters.sources) or "configured providers"
        raise SourcePlanningError(
            f"no eligible provider for category={query.filters.category.value}, mode={mode.value}, sources={requested}"
        )

    selected_names = {step.provider for step in steps}
    external_disclosure = any(
        p.third_party_query_disclosure for p in eligible if p.name in selected_names
    )
    if mode in {SourceMode.GOREECLOUD_ONLY, SourceMode.OFFLINE_LOCAL} and external_disclosure:
        raise AssertionError("privacy invariant violated: prohibited external disclosure")

    return SourcePlan(
        mode=mode,
        category=query.filters.category,
        steps=tuple(steps),
        third_party_query_disclosure=external_disclosure,
    )
