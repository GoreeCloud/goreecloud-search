# GoreeCloud Search — Web Intelligence Execution Coordinator Development Contract

## Status

Development source integration under FR-008. This contract does not establish production TinyFish execution, production provider acceptance, production budget authority, Privacy Shield runtime acceptance, Wardveil protected state, Everkeep recovery evidence, Release Candidate qualification, or Stable acceptance.

## Purpose

The web-intelligence routing controller already decides which providers are eligible for one exact capability and reserves metered budget before execution. This Development slice adds a provider-neutral execution coordinator between that plan and provider-specific executors.

The coordinator exists so fallback, metered accounting, cancellation, and minimal operational evidence remain under GoreeCloud authority rather than being reimplemented differently by each external provider integration.

## Capability boundary

The coordinator receives one already-authorized `webintelligence.Request` and calls `Controller.Plan` exactly once for that capability.

It does not convert capabilities. In particular, it never performs any of the following implicitly:

- Search → Fetch;
- Fetch → Research;
- Research → Agent;
- Agent → Browser.

Cross-capability escalation remains a separate caller authorization decision that must preserve purpose, Privacy Shield authority, Wardveil handling requirements, and runtime acceptance state.

## Executor registration

Provider-specific transports implement the `Executor` interface and are registered by provider ID.

Registration fails closed when:

- the controller is unavailable;
- an executor has no provider ID;
- duplicate executor IDs are supplied;
- an executor references a provider not configured in the controller; or
- an executor's declared capability differs from the configured provider capability.

Not every configured provider must have an executor in this Development primitive. If a planned provider has no registered executor, the attempt is recorded as rejected and same-capability fallback may continue. If no selected provider has an executor, execution fails closed.

## Execution and fallback

The coordinator executes only the ordered providers returned by `Controller.Plan`.

Same-capability provider failure may move to the next selected provider. Provider error details are not returned to callers by this layer; exhausted provider execution returns a stable GoreeCloud error.

Cancellation and deadline semantics are preserved as `context.Canceled` and `context.DeadlineExceeded`. The coordinator does not retry across cancellation or timeout as though those conditions were ordinary provider failures.

## Budget accounting

Metered reservations are created by the existing controller before provider execution.

The coordinator:

- accumulates actual metered cost reported by every attempted metered executor, including failed attempts;
- requires each attempted metered executor to explicitly state that its returned cost is backed by provider/runtime billing evidence;
- treats an evidence-backed zero as zero cost, rather than assuming that an omitted or unknown cost is free;
- commits actual accumulated cost when metered work was attempted and all attempted metered costs are observed;
- releases an unused metered reservation when execution succeeds or terminates before any metered provider is attempted;
- retains the reservation and fails closed if metered execution has occurred but actual cost evidence is unavailable, preventing an unverified zero from being recorded as spend;
- rejects non-zero cost reported by a provider configured as free;
- surfaces controller budget-overrun errors instead of hiding them; and
- never reads or modifies a TinyFish wallet directly.

Provider rates and wallet balances remain mutable external state and are not hard-coded in this coordinator. The current retained-reservation behavior is deliberately conservative Development behavior; a future durable billing reconciliation mechanism must settle or release unresolved reservations from authoritative provider evidence before production use.

## Privacy and security

The coordinator treats provider input and output as opaque values. It does not persist, normalize, inspect, or copy them into controller observations.

Its operational evidence remains limited to the existing controller model:

- provider ID;
- normalized outcome;
- aggregate count; and
- aggregate duration.

It does not add query text, URLs, research prompts, Agent goals, Browser session state, page content, citations, cookies, credentials, Vault references, Browser Context Profile IDs, or provider response bodies to web-intelligence observations.

Provider executors remain responsible for their own transport security, credential boundaries, host pinning, redirect policy, response limits, public-target rules, untrusted-content handling, service-specific validation, and authoritative billing/cost evidence where they represent a provider as metered.

## Direct connector precedence

This coordinator does not change GoreeCloud connector precedence. GitHub, Google Drive, Gmail, and other approved authoritative connected services must continue to use their direct connectors when those connectors can perform the requested operation.

TinyFish Agent or Browser must not be selected merely to bypass an available authoritative connector.

## Current Development limits

This source slice does not:

- wire Search, Fetch, Research, Agent, or Browser transports automatically into a live execution registry;
- expose a public execution HTTP endpoint;
- execute a TinyFish request as part of source validation;
- deploy or read production TinyFish credentials;
- read or mutate the TinyFish wallet;
- create or use a live Browser Context Profile;
- create or use a live Vault credential;
- reconcile unresolved retained reservations against provider billing;
- infer provider health from live traffic;
- persist budget state across restarts;
- coordinate budget across multiple Search instances;
- establish Manager UI acceptance;
- establish Privacy Shield or Wardveil runtime acceptance;
- establish Everkeep recovery/rollback evidence; or
- authorize production deployment, Release Candidate status, or Stable status.

Those remain later integration and acceptance work under normal GoreeCloud governance.
