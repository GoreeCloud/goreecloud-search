# TinyFish Agent Reliability and Health Policy

**Status:** Development source foundation  
**Lifecycle:** Pre-Stable / production acceptance pending  
**Scope:** GoreeCloud Search web-intelligence control plane

## Purpose

TinyFish Agent is a metered escalation capability. GoreeCloud Search must not treat provider availability, billing availability, or a configured API key as proof that Agent is healthy enough to receive production traffic.

This document defines a provider-neutral reliability gate over GoreeCloud's existing privacy-minimized operational observations. The gate is intended to support least-capability routing, fail-closed degradation, and future Manager visibility without exposing user queries, URLs, goals, page content, credentials, citations, or provider response bodies.

## Current external evidence

A TinyFish dashboard screenshot supplied to the project on September 13, 2026 showed a seven-day Agent snapshot of 23 runs, 10 passed, 8 failed, 5 cancelled, a displayed 43.5% success rate, and a displayed 639.5-second average duration. The screenshot also showed a fastest run of 105.0 seconds and slowest run of 1223.0 seconds.

This is useful operational evidence, but it is not a durable GoreeCloud telemetry source and must not be hard-coded as current truth. It demonstrates why GoreeCloud needs an explicit health gate before production acceptance rather than assuming that configured Agent access is healthy.

## Development primitive

`ProviderHealthPolicy` and `ProviderHealthEvaluation` provide a provider-neutral mechanism for evaluating existing Controller observations.

A policy can define:

- minimum decision samples before health may change;
- a success-rate threshold below which a provider becomes degraded;
- a lower success-rate threshold below which a provider becomes unavailable;
- an optional maximum average duration threshold that can degrade a provider even when success rate is otherwise acceptable;
- whether cancelled operations count as failures;
- whether locally/provider-rejected operations count as failures.

Cancelled and rejected operations are excluded by default. Cancellation can reflect caller behavior, and rejection can reflect local policy or budget enforcement rather than provider reliability. Deployments that intentionally want those outcomes to affect provider health must opt in explicitly.

## Fail-closed behavior

`EvaluateHealth` is read-only. It returns a point-in-time recommendation and never mutates routing state.

`ApplyHealthPolicy` changes a provider's Controller health only after the configured minimum sample threshold is met. If the result is `unavailable`, the existing Controller planner excludes that provider from new plans. A degraded provider remains eligible but ranks behind a healthy provider in the same cost class.

No implicit cross-capability escalation is introduced. A degraded or unavailable Agent provider does not automatically authorize Browser. Search, Fetch, Research, Agent, and Browser remain separately authorized capabilities.

## Privacy and security boundary

The health evaluator consumes only the Controller's normalized outcome counts and aggregate duration. It has no fields for:

- query text;
- target URLs;
- research prompts or goals;
- page content;
- Browser Context Profile identifiers;
- Vault credential item identifiers or values;
- cookies, tokens, API keys, or session endpoints;
- provider response bodies.

Health evidence does not establish Privacy Shield authorization, Wardveil protected state, production approval, or release acceptance.

## Important current limitation

The existing Controller observation store is cumulative for the lifetime of one process. This foundation therefore must not be described as a rolling seven-day SLO, a durable health ledger, or a multi-instance health consensus.

Before production acceptance, GoreeCloud should add a bounded recent-window or durable privacy-minimized telemetry source, define deployment-specific thresholds, and verify behavior across multiple runtimes. Provider dashboard statistics may be used as external corroborating evidence, not as the sole source of truth.

## Recommended production acceptance gates

Before Agent can be accepted for production routing, GoreeCloud should require all of the following:

1. A documented health policy for the Agent provider with a meaningful minimum sample size.
2. A recent-window reliability measure that distinguishes success, failure, timeout, cancellation, and rejection.
3. Latency limits appropriate to the user-facing operation and cost class.
4. Budget reservation and actual-cost reconciliation for metered attempts.
5. Automatic routing exclusion when provider health becomes unavailable.
6. Recovery/hysteresis rules so a single success does not immediately restore a failing provider.
7. Manager visibility that explains why a provider is healthy, degraded, or unavailable without exposing user content or secrets.
8. Production Privacy Shield and Wardveil acceptance for the exact runtime and policy path.

## Non-goals of this slice

This Development slice does not:

- run TinyFish Agent;
- spend wallet funds;
- install production API credentials;
- change TinyFish account settings;
- create or repair Browser Context Profiles or Vault items;
- implement durable or distributed telemetry;
- claim that the dashboard snapshot is still current;
- qualify GoreeCloud Search as Release Candidate or Stable.
