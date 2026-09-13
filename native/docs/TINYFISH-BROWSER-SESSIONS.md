# TinyFish direct Browser session lifecycle

**Lifecycle:** Development source foundation only. This document does not establish production TinyFish Browser execution, Privacy Shield authorization, Wardveil protected state, Everkeep recovery evidence, deployment acceptance, Release Candidate status, or Stable status.

## Purpose

TinyFish Browser is the lowest-level escalation in GoreeCloud Search's external web-intelligence path. It creates a remote browser session and returns short-lived browser-control endpoints for a caller-owned CDP controller.

GoreeCloud must treat this as a narrowly bounded browser-session lifecycle, not as a substitute for Search, Fetch, Research, Agent, GoreeCloud Browser, or authoritative first-party connectors.

This Development foundation adds a provider-neutral start/terminate lifecycle, an exact TinyFish Browser REST transport, and a scoped cleanup lease that prevents normal caller cancellation from silently leaking a confirmed remote browser session.

## Verified provider contract

The implementation is based on the official TinyFish Browser API documentation reviewed on September 12, 2026.

The documented REST endpoints are:

- `POST https://api.browser.tinyfish.ai` — create a browser session;
- `DELETE https://api.browser.tinyfish.ai/{session_id}` — terminate a browser session.

All requests require `X-API-Key`.

The TinyFish SDKs use `https://agent.tinyfish.ai/v1/browser` by default, while the official REST reference documents `https://api.browser.tinyfish.ai`. TinyFish documents both hosts as supported and equivalent. GoreeCloud deliberately pins the documented REST host.

### Exact create request

The documented create-session request body contains only:

- `url` — initial URL;
- `timeout_seconds` — inactivity timeout.

GoreeCloud forwards only those two fields. It does not send local purpose, Browser Context Profile identifiers, Vault references, proxy configuration, or other fields through this transport.

TinyFish allows `url` to be omitted, which starts `about:blank`. GoreeCloud is stricter: the local lifecycle requires an explicit approved public HTTP(S) target URL so authorization is bound to a concrete operation before a provider session can be created.

TinyFish documents `timeout_seconds` from 5 to 86400 seconds. GoreeCloud currently narrows that provider surface to 5 seconds through 10 minutes, with a five-minute default, so Development callers cannot accidentally create unusually long-lived direct browser sessions.

### Exact create response

A successful create returns HTTP `201 Created` with:

- `session_id`;
- `cdp_url`;
- `base_url`.

The provider documents that `201` can be returned before startup navigation completes. Therefore session creation proves only that the remote browser session was allocated. It does not prove that the target page loaded, authentication succeeded, or the intended browser goal completed.

### Termination

A successful terminate returns HTTP `204 No Content`. TinyFish documents termination as idempotent for an already-ended session.

GoreeCloud keeps a runtime-local registry of sessions it actually created and refuses to issue a destructive termination request for an arbitrary caller-supplied provider session ID. The session remains locally active when provider termination is not confirmed, allowing a deliberate retry rather than falsely recording cleanup success.

## Sensitive browser-control state

The following values are sensitive ephemeral runtime state:

- Browser session ID;
- CDP WebSocket URL;
- browser base URL.

They must not be written to normal logs, analytics, public status endpoints, long-lived configuration, feature roadmaps, or ordinary audit summaries.

The transport validates `cdp_url` as `wss://` and `base_url` as `https://` before returning them to an authorized in-process caller.

## Fail-closed authority boundary

A `BrowserSessionAuthorizer` is mandatory. No allow-all authorizer is provided.

Before a create operation, GoreeCloud:

1. validates and normalizes the target URL;
2. validates the local purpose;
3. applies the GoreeCloud inactivity-timeout bound;
4. authorizes the exact start operation;
5. only then calls TinyFish.

Before termination, GoreeCloud:

1. validates the session ID and purpose;
2. verifies the session belongs to the current manager's active-session registry;
3. resolves the original target URL associated with that session;
4. authorizes the exact termination operation;
5. only then calls TinyFish.

This keeps Privacy Shield, Wardveil, caller/session authority, target-site policy, cost policy, and future runtime acceptance logic at the GoreeCloud boundary rather than delegating authority to the provider.

## Scoped cleanup lease

A confirmed direct Browser session is a remote resource that can remain active and potentially billable after the caller's task stops. Relying only on each caller to remember a final terminate call is not sufficient.

`BrowserSessionLeaseRunner` scopes one confirmed direct Browser session to one unit of in-process work and guarantees a termination attempt when that work returns. The lease validates its cleanup purpose before session creation, so GoreeCloud never creates a session that it already knows it cannot clean up through the local contract.

Cleanup deliberately runs with a fresh, short-lived internal context rather than inheriting the task context. This matters when the caller cancels, times out, or otherwise abandons the browser task: a canceled task context must not automatically cancel the cleanup request before it can reach TinyFish.

The default cleanup window is 15 seconds and is bounded to at most 30 seconds. The cleanup context exists only for termination; it does not authorize continued browsing or extend the caller's task authority.

If browser work fails and cleanup also fails, GoreeCloud preserves both errors. If browser work succeeds but cleanup is not confirmed, the overall lease reports cleanup failure instead of representing the operation as fully clean. The underlying manager retains the session in its active registry when termination is unconfirmed so an authorized retry remains possible.

The lease does not retry uncertain creation. If session creation is not confirmed, GoreeCloud cannot safely know which provider session to control and does not invent a cleanup identifier.

## Network and provider boundary

The TinyFish Browser transport:

- pins `api.browser.tinyfish.ai`;
- disables ambient HTTP proxy use;
- rejects redirects;
- requires TLS 1.2 or newer;
- resolves and dials only the expected provider hostname;
- rejects non-public resolved addresses;
- caps response size;
- validates response content type;
- uses an HTTP timeout above TinyFish's documented minimum of 60 seconds for session creation.

## Uncertain provider outcomes

Direct browser creation and termination are side-effecting operations. A transport interruption does not prove that the provider failed to create or terminate a session.

GoreeCloud therefore distinguishes:

- `browser session creation is not confirmed`;
- `browser session termination is not confirmed`.

Retry policy must treat these states as uncertain outcomes rather than blindly issuing another create or falsely declaring cleanup complete. Provider HTTP responses such as retry-required, service-busy, and timeout are mapped conservatively according to the reviewed Browser API contract.

## Isolated direct sessions

TinyFish documents direct Browser API sessions as isolated: they do not persist cookies or storage by themselves.

Reusable authenticated state belongs in Browser Context Profiles and the existing GoreeCloud exact-host binding/profile-lifecycle path. Direct Browser sessions must not be represented as a replacement for that authenticated-session architecture.

## What this foundation does not implement

This Development source foundation does not yet provide:

- a CDP automation/controller implementation;
- a public Browser execution route;
- automatic Search → Fetch → Research → Agent → Browser escalation;
- automatic retries after uncertain create outcomes;
- persistent direct-browser session state;
- production API-key installation;
- live TinyFish Browser acceptance;
- production cost/spend acceptance;
- Privacy Shield or Wardveil production runtime acceptance;
- a claim that Browser Context Profiles or Vault authentication are healthy.

Creating and safely leasing a low-level browser session is not the same as completing a browser task. A separately reviewed CDP controller must be added before GoreeCloud can claim direct Browser capability for goal execution.

## Connector precedence

GitHub, Google Drive, Gmail, and other authoritative direct connectors remain preferred whenever they provide the required operation. TinyFish direct Browser sessions are reserved for external browser-only gaps after less-powerful capabilities are insufficient and policy permits escalation.
