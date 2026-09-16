from __future__ import annotations

from dataclasses import dataclass
from enum import Enum
from typing import Protocol
from urllib.parse import urlsplit

from .normalization import NormalizedResult


class ContentPolicyError(RuntimeError):
    """Raised when requested content policy cannot be enforced safely."""


class SafeSearchMode(str, Enum):
    OFF = "off"
    MODERATE = "moderate"
    STRICT = "strict"


class ContentPolicyAction(str, Enum):
    ALLOW = "allow"
    WARN = "warn"
    BLOCK = "block"


@dataclass(frozen=True, slots=True)
class ContentPolicyDecision:
    hook: str
    action: ContentPolicyAction
    reason: str


@dataclass(frozen=True, slots=True)
class ContentPolicyOutcome:
    result: NormalizedResult
    action: ContentPolicyAction
    decisions: tuple[ContentPolicyDecision, ...]


@dataclass(frozen=True, slots=True)
class ContentPolicyReport:
    safe_search_mode: SafeSearchMode
    safe_search_enforced: bool
    outcomes: tuple[ContentPolicyOutcome, ...]

    @property
    def visible_results(self) -> tuple[NormalizedResult, ...]:
        return tuple(
            outcome.result
            for outcome in self.outcomes
            if outcome.action is not ContentPolicyAction.BLOCK
        )

    @property
    def blocked_count(self) -> int:
        return sum(1 for outcome in self.outcomes if outcome.action is ContentPolicyAction.BLOCK)

    @property
    def warned_count(self) -> int:
        return sum(1 for outcome in self.outcomes if outcome.action is ContentPolicyAction.WARN)


class ContentPolicyHook(Protocol):
    """Deterministic policy hook over an already-normalized result."""

    @property
    def name(self) -> str:
        ...

    @property
    def supports_safe_search(self) -> bool:
        ...

    def evaluate(
        self,
        result: NormalizedResult,
        *,
        safe_search: SafeSearchMode,
    ) -> ContentPolicyDecision | None:
        ...


def _normalize_domain(value: str) -> str:
    domain = value.strip().casefold().rstrip(".")
    if not domain or "/" in domain or "://" in domain:
        raise ValueError(f"invalid policy domain: {value!r}")
    return domain


def _host_matches(host: str, domain: str) -> bool:
    normalized_host = host.casefold().rstrip(".")
    return normalized_host == domain or normalized_host.endswith(f".{domain}")


class DomainPolicyHook:
    """Local administrator allow/block domain policy; not a content classifier."""

    def __init__(
        self,
        *,
        blocked_domains: tuple[str, ...] = (),
        allowed_domains: tuple[str, ...] = (),
        name: str = "domain-policy",
    ) -> None:
        if not name.strip():
            raise ValueError("content-policy hook name must not be empty")
        self._name = name
        self._blocked = tuple(dict.fromkeys(_normalize_domain(v) for v in blocked_domains))
        self._allowed = tuple(dict.fromkeys(_normalize_domain(v) for v in allowed_domains))

    @property
    def name(self) -> str:
        return self._name

    @property
    def supports_safe_search(self) -> bool:
        return False

    def evaluate(
        self,
        result: NormalizedResult,
        *,
        safe_search: SafeSearchMode,
    ) -> ContentPolicyDecision | None:
        del safe_search
        host = (urlsplit(result.canonical_url).hostname or "").casefold().rstrip(".")
        if any(_host_matches(host, domain) for domain in self._blocked):
            return ContentPolicyDecision(
                hook=self.name,
                action=ContentPolicyAction.BLOCK,
                reason="Result domain is explicitly blocked by administrator policy.",
            )
        if self._allowed and not any(_host_matches(host, domain) for domain in self._allowed):
            return ContentPolicyDecision(
                hook=self.name,
                action=ContentPolicyAction.BLOCK,
                reason="Result domain is outside the administrator allowlist.",
            )
        return None


class ContentPolicyEngine:
    """Apply configured local policy hooks before user-facing ranking."""

    def __init__(self, hooks: tuple[ContentPolicyHook, ...] | list[ContentPolicyHook] = ()) -> None:
        names: set[str] = set()
        ordered: list[ContentPolicyHook] = []
        for hook in hooks:
            key = hook.name.casefold()
            if not key:
                raise ValueError("content-policy hook name must not be empty")
            if key in names:
                raise ValueError(f"duplicate content-policy hook name: {hook.name!r}")
            names.add(key)
            ordered.append(hook)
        self._hooks = tuple(ordered)

    @property
    def hooks(self) -> tuple[ContentPolicyHook, ...]:
        return self._hooks

    def can_enforce_safe_search(self, safe_search: SafeSearchMode) -> bool:
        return (
            safe_search is SafeSearchMode.OFF
            or any(hook.supports_safe_search for hook in self._hooks)
        )

    def ensure_safe_search_supported(self, safe_search: SafeSearchMode) -> None:
        if not self.can_enforce_safe_search(safe_search):
            raise ContentPolicyError(
                f"SafeSearch mode {safe_search.value!r} was requested but no configured hook can enforce it"
            )

    def apply(
        self,
        results: tuple[NormalizedResult, ...] | list[NormalizedResult],
        *,
        safe_search: SafeSearchMode = SafeSearchMode.OFF,
    ) -> ContentPolicyReport:
        self.ensure_safe_search_supported(safe_search)
        outcomes: list[ContentPolicyOutcome] = []

        for result in results:
            decisions: list[ContentPolicyDecision] = []
            for hook in self._hooks:
                try:
                    decision = hook.evaluate(result, safe_search=safe_search)
                except Exception as exc:
                    raise ContentPolicyError(
                        f"content-policy hook {hook.name!r} failed"
                    ) from exc
                if decision is not None:
                    if decision.hook.casefold() != hook.name.casefold():
                        raise ContentPolicyError(
                            f"content-policy hook {hook.name!r} returned mismatched provenance"
                        )
                    decisions.append(decision)

            if any(d.action is ContentPolicyAction.BLOCK for d in decisions):
                action = ContentPolicyAction.BLOCK
            elif any(d.action is ContentPolicyAction.WARN for d in decisions):
                action = ContentPolicyAction.WARN
            else:
                action = ContentPolicyAction.ALLOW

            outcomes.append(
                ContentPolicyOutcome(
                    result=result,
                    action=action,
                    decisions=tuple(decisions),
                )
            )

        return ContentPolicyReport(
            safe_search_mode=safe_search,
            safe_search_enforced=True,
            outcomes=tuple(outcomes),
        )
