# GoreeCloud Search — TinyFish Web Intelligence Routing and Budget Development Contract

## Status

Development source foundation for FR-008. This record does not establish production TinyFish routing, live budget enforcement against the TinyFish wallet, automatic provider escalation, Manager UI acceptance, Privacy Shield authorization, Wardveil protected state, Everkeep recovery evidence, Release Candidate qualification, or Stable acceptance.

## Purpose

GoreeCloud Search needs a GoreeCloud-owned control plane around external web-intelligence providers so capability choice, fallback, health, cost, and operational evidence remain under GoreeCloud authority.

The intended least-capability progression remains:

`Search → Fetch → Research → Agent → Browser`

That ordering is a policy direction, not permission to escalate automatically. The routing controller accepts one exact capability at a time. A caller must obtain separate authority before moving to a stronger capability.

## Source boundary in this slice

`native/internal/webintelligence` introduces a provider-neutral `Controller` that provides:

- exact-capability provider selection;
- free-before-metered preference within the same capability;
- configured provider priority and health-aware ordering;
- exclusion of unavailable providers;
- bounded same-capability fallback attempts;
- explicit opt-in before any metered provider is eligible;
- required estimated cost before a metered plan can be created;
- fail-closed total and per-operation budget ceilings;
- conservative reservation of the maximum planned metered attempt cost before execution;
- release of unused reservations;
- recording of actual metered spend after execution, including overrun evidence;
- privacy-preserving operational observations containing provider ID, outcome, count, and aggregate duration only;
- a bounded snapshot suitable for future GoreeCloud Manager operational presentation.

The controller is not a provider executor. It does not issue Search, Fetch, Research, Agent, or Browser network requests and does not own TinyFish credentials.

## Capability and escalation boundary

The controller never converts one capability into another. A Search request can select only Search providers; Fetch can select only Fetch providers; Research can select only Research providers; Agent can select only Agent providers; Browser can select only Browser providers.

This prevents provider failure or cost pressure from silently turning a low-authority operation into a more invasive or expensive operation. Cross-capability escalation must be separately authorized by the calling GoreeCloud workflow and must preserve the operation's purpose, Privacy Shield authority, Wardveil handling requirements, and application/runtime acceptance state.

## Free-before-metered routing

Within one capability, eligible free providers are ordered before metered providers. Provider health and configured priority are then used to order peers.

This supports GoreeCloud's current TinyFish usage direction: use Search and Fetch for ordinary discovery and retrieval when they satisfy the task, and reserve Research, Agent, or Browser for cases that genuinely require those capabilities.

Free status is not an authorization shortcut. A free provider still requires normal privacy, security, provider, and runtime acceptance before production use.

## Budget model

Metered routing fails closed unless all of the following are true:

- the caller explicitly allows metered providers;
- the caller supplies a non-zero estimated cost;
- a non-zero total budget is configured;
- a non-zero per-operation budget is configured;
- the maximum planned metered attempt cost fits within the per-operation ceiling;
- the reservation fits within remaining total budget after prior spend and outstanding reservations.

The source uses integer micro-units for accounting so policy does not depend on floating-point currency arithmetic.

A plan reserves the maximum estimated cost for all metered attempts selected into that plan. This is intentionally conservative. If an operation does not use the whole reservation, the caller must release or commit it. When actual spend exceeds the reservation, the controller records the real spend and returns an overrun error instead of hiding the policy violation.

The current source budget is an in-process accounting primitive. It is not yet a durable ledger, TinyFish wallet reconciliation service, billing authority, multi-instance quota system, or production financial control.

## Provider health and fallback

Providers have explicit `healthy`, `degraded`, and `unavailable` states. Unavailable providers are excluded from new plans. Free providers remain preferred over metered providers within the same exact capability, while health and priority order providers within those cost classes.

Health state changes are currently explicit controller inputs. This slice does not infer health automatically from live TinyFish dashboards or provider traffic. Production health policy will require bounded failure windows, recovery probes, hysteresis, evidence freshness, and target-runtime monitoring so temporary failures do not create unstable routing behavior.

Fallback is same-capability only. The maximum number of attempts is bounded by source policy. Exhausting Search providers does not authorize Fetch, Research, Agent, or Browser.

## Operational evidence and privacy

The controller's observation model deliberately excludes:

- query text;
- requested URLs;
- research prompts;
- Agent goals;
- Browser session data;
- page content;
- citations;
- cookies or local storage;
- credentials or Vault references;
- provider response bodies.

It records only the provider ID, normalized outcome, aggregate count, and aggregate duration. This gives GoreeCloud a minimal basis for health and operational visibility without turning the routing layer into a query-history or browsing-history store.

Production telemetry must continue to follow Privacy Shield data-minimization and purpose-limitation requirements. More detailed evidence, if ever required, needs explicit authority and retention rules rather than being added implicitly.

## Manager visibility direction

`Controller.Snapshot()` is the Development source boundary for future GoreeCloud Manager presentation. A future Manager integration can show bounded operational facts such as:

- configured providers and capabilities;
- provider health;
- whether a provider is metered;
- configured priority;
- total budget;
- reserved budget;
- recorded spend;
- remaining budget;
- normalized success/failure/timeout/cancellation/rejection counts and aggregate duration.

Manager must not imply production protection, authorization, or financial accuracy merely because a snapshot exists. Production Manager state must identify evidence freshness, runtime scope, configuration source, and whether the underlying runtime is actually accepted.

## Direct connector precedence

TinyFish is not a replacement for authoritative GoreeCloud connectors. When GitHub, Google Drive, Gmail, or another approved connected service can perform the required operation directly, GoreeCloud should use that service's connector instead of routing the task through TinyFish Agent or Browser.

The web-intelligence router is intended for public-web discovery and retrieval, approved research, external services without a suitable direct connector, and genuine browser interaction or rendering requirements.

## Security and failure behavior

The routing controller is intentionally independent from provider credentials and provider response content. It returns stable GoreeCloud errors for invalid requests, missing providers, missing budgets, budget exhaustion, unknown providers, and invalid accounting operations.

Budget reservation, spend, provider health, and observations are protected by controller locking so concurrent callers cannot update those in-memory values without synchronization.

This does not replace the existing provider-specific protections in Search, Fetch, Research, or Agent implementations. Provider transports must continue to enforce their own host pinning, redirect handling, response bounds, credential boundaries, public-target controls, untrusted-content handling, and other service-specific protections.

## Lifecycle boundary

This Development foundation does not:

- automatically register or enable TinyFish providers in a production runtime;
- automatically escalate between Search, Fetch, Research, Agent, and Browser;
- execute a metered TinyFish operation;
- read or modify the TinyFish wallet;
- reconcile controller accounting with provider billing;
- persist budget or telemetry state across process restarts;
- coordinate quota across multiple GoreeCloud Search instances;
- expose a public routing or budget HTTP API;
- create a GoreeCloud Manager UI;
- establish production provider-health probes;
- establish Privacy Shield or Wardveil runtime acceptance;
- establish Everkeep recovery or rollback evidence;
- authorize production deployment, Release Candidate status, or Stable status.

Those remain later acceptance and integration work under FR-008 and normal GoreeCloud release governance.
