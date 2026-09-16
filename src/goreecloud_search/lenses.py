from __future__ import annotations

from dataclasses import dataclass
from enum import Enum
from urllib.parse import urlsplit

from .normalization import NormalizedResult
from .ranking import RankedResult, RankingSignal


class LensError(ValueError):
    """Raised when a Lens definition or requested Lens is invalid."""


class LensRuleTarget(str, Enum):
    DOMAIN = "domain"
    FILETYPE = "filetype"
    LANGUAGE = "language"


class LensRuleAction(str, Enum):
    BOOST = "boost"
    LOWER = "lower"
    EXCLUDE = "exclude"


@dataclass(frozen=True, slots=True)
class LensRule:
    target: LensRuleTarget
    value: str
    action: LensRuleAction
    weight: float = 5.0

    def __post_init__(self) -> None:
        value = self.value.strip().casefold()
        if not value:
            raise LensError("Lens rule value must not be empty")
        object.__setattr__(self, "value", value)
        if self.action is LensRuleAction.EXCLUDE:
            object.__setattr__(self, "weight", 0.0)
        elif self.weight <= 0 or self.weight > 100:
            raise LensError("Lens boost/lower weight must be greater than 0 and at most 100")


@dataclass(frozen=True, slots=True)
class Lens:
    name: str
    rules: tuple[LensRule, ...]
    description: str | None = None

    def __post_init__(self) -> None:
        name = self.name.strip()
        if not name:
            raise LensError("Lens name must not be empty")
        if not self.rules:
            raise LensError("Lens must contain at least one rule")
        object.__setattr__(self, "name", name)


@dataclass(frozen=True, slots=True)
class LensExclusion:
    result: NormalizedResult
    rule: LensRule
    explanation: str


@dataclass(frozen=True, slots=True)
class LensApplicationReport:
    lens_name: str | None
    results: tuple[RankedResult, ...]
    exclusions: tuple[LensExclusion, ...] = ()

    @property
    def excluded_count(self) -> int:
        return len(self.exclusions)


def _lens_key(name: str) -> str:
    return "-".join(name.strip().casefold().replace("_", "-").split())


class LensRegistry:
    def __init__(self, lenses: tuple[Lens, ...] | list[Lens] = ()) -> None:
        by_name: dict[str, Lens] = {}
        for lens in lenses:
            key = _lens_key(lens.name)
            if key in by_name:
                raise LensError(f"duplicate Lens name: {lens.name!r}")
            by_name[key] = lens
        self._by_name = by_name

    @property
    def lenses(self) -> tuple[Lens, ...]:
        return tuple(self._by_name.values())

    def resolve(self, name: str | None) -> Lens | None:
        if name is None:
            return None
        lens = self._by_name.get(_lens_key(name))
        if lens is None:
            raise LensError(f"unknown Lens: {name!r}")
        return lens


def _host_matches(host: str, domain: str) -> bool:
    host = host.casefold().rstrip(".")
    domain = domain.casefold().rstrip(".")
    return host == domain or host.endswith(f".{domain}")


def _rule_matches(rule: LensRule, result: NormalizedResult) -> bool:
    parts = urlsplit(result.canonical_url)
    if rule.target is LensRuleTarget.DOMAIN:
        return _host_matches(parts.hostname or "", rule.value)
    if rule.target is LensRuleTarget.FILETYPE:
        return parts.path.casefold().endswith(f".{rule.value.lstrip('.')}")
    if rule.target is LensRuleTarget.LANGUAGE:
        return bool(result.language and result.language.casefold() == rule.value)
    raise LensError(f"unsupported Lens rule target: {rule.target!r}")


def _rule_explanation(lens: Lens, rule: LensRule) -> str:
    target = rule.target.value
    if rule.action is LensRuleAction.EXCLUDE:
        return f"Lens '{lens.name}' excluded this result because its {target} matched '{rule.value}'."
    verb = "boosted" if rule.action is LensRuleAction.BOOST else "lowered"
    return f"Lens '{lens.name}' {verb} this result because its {target} matched '{rule.value}'."


def apply_lens(
    ranked: tuple[RankedResult, ...] | list[RankedResult],
    lens: Lens | None,
) -> LensApplicationReport:
    """Apply transparent, deterministic, local-only Lens reranking rules."""
    ranked = tuple(ranked)
    if lens is None:
        return LensApplicationReport(lens_name=None, results=ranked)

    adjusted: list[RankedResult] = []
    exclusions: list[LensExclusion] = []

    for item in ranked:
        matching = tuple(rule for rule in lens.rules if _rule_matches(rule, item.result))
        excluded = next((rule for rule in matching if rule.action is LensRuleAction.EXCLUDE), None)
        if excluded is not None:
            exclusions.append(
                LensExclusion(
                    result=item.result,
                    rule=excluded,
                    explanation=_rule_explanation(lens, excluded),
                )
            )
            continue

        signals = list(item.signals)
        for rule in matching:
            score = rule.weight if rule.action is LensRuleAction.BOOST else -rule.weight
            signals.append(
                RankingSignal(
                    code=f"lens_{rule.target.value}_{rule.action.value}",
                    score=score,
                    explanation=_rule_explanation(lens, rule),
                )
            )
        adjusted.append(
            RankedResult(
                result=item.result,
                score=round(item.score + sum(s.score for s in signals[len(item.signals):]), 3),
                signals=tuple(signals),
            )
        )

    adjusted.sort(
        key=lambda item: (
            -item.score,
            -item.result.source_agreement,
            item.result.canonical_url.casefold(),
            item.result.result_id,
        )
    )
    return LensApplicationReport(
        lens_name=lens.name,
        results=tuple(adjusted),
        exclusions=tuple(exclusions),
    )
