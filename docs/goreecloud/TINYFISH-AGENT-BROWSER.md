# GoreeCloud Search — TinyFish Agent and Browser Escalation Development Contract

## Status

Development source foundation. This record does not establish live-provider acceptance, production TinyFish credentials, Privacy Shield authorization, Wardveil protected state, Everkeep recovery evidence, production deployment authority, Release Candidate qualification, or Stable acceptance.

## Purpose

TinyFish Agent and Browser are escalation capabilities behind GoreeCloud-owned boundaries. They must not become the default route for discovery or retrieval, and they must not create a direct permanent GoreeCloud AI dependency on TinyFish.

The intended least-capability order remains:

`Search → Fetch → Research → Agent → Browser`

Agent is appropriate only when an approved web task requires interaction such as navigation, clicking, or form work. Browser is a lower-level capability for direct CDP/automation control and therefore requires a separate executor and stronger authorization rather than silently inheriting Agent authority.

## Source boundary in this slice

`native/internal/webautomation` defines:

- a provider-neutral `Request`, `Result`, `Authorizer`, and `Executor` contract;
- explicit `agent` and `browser` capabilities;
- a fail-closed `Escalator` that has no built-in allow-all policy;
- bounded target URL, goal, purpose, profile, credential-reference, step, and duration validation;
- `lite` as the normalized default browser runtime profile;
- an explicit `stealth` runtime profile value that policy can reject unless separately justified;
- explicit Browser Context Profile reuse through `use_profile` plus a required profile ID;
- optional TinyFish Vault credential references only when Vault use is explicitly enabled;
- a TinyFish Agent implementation using the cancellable asynchronous Agent lifecycle.

The TinyFish Agent implementation is pinned to the official Agent host and uses:

- `POST /v1/automation/run-async` to create a run;
- `GET /v1/runs/{id}` to poll bounded status;
- `POST /v1/runs/{id}/cancel` when the GoreeCloud caller cancels or the local processing window expires;
- `X-API-Key` authentication only;
- `browser_profile` plus optional Browser Context Profile and Vault references;
- optional bounded `agent_config.max_steps` and required bounded `agent_config.max_duration_seconds`.

TinyFish documents `agent_config.max_steps` as beta-limited. GoreeCloud therefore does not treat that field as a complete cost-control mechanism. Live automated spending remains blocked on later FR-008 budget, quota, telemetry, and Manager controls.

## Authorization boundary

The package deliberately cannot authorize itself. Every operation must pass through an injected GoreeCloud `Authorizer` before an executor is selected.

A future production authorizer must independently evaluate the calling identity and application, declared purpose, requested target, external-processing permission, Privacy Shield authorization, Wardveil handling requirements, session-state scope, Vault use, credential references, requested runtime profile, capability level, cost/budget state, and any additional service-specific policy.

Missing, malformed, stale, contradictory, unavailable, or insufficient authority must fail closed. The existence of an API key, Browser Context Profile, Vault credential item, or TinyFish account balance is not authorization.

## Browser Context Profiles and Vault

Browser Context Profiles persist cookies, local storage, and session storage across Agent runs. They are distinct from TinyFish Browser Profiles (`lite` and `stealth`), which select runtime behavior.

GoreeCloud should use narrowly scoped Browser Context Profiles for specific approved external services or purposes instead of one shared global profile. A profile ID must be explicit whenever session reuse is requested.

Vault credential references are provider-side references, not authorization tokens. GoreeCloud source, ordinary logs, evidence, prompts, documentation, and caller-visible errors must not contain reusable credentials. Credential item IDs are accepted only when Vault use is explicitly requested and are bounded and deduplicated before provider submission.

## Security and privacy controls

The source foundation:

- accepts only public HTTP(S) target URLs without embedded credentials;
- rejects localhost, local-domain, and literal non-public IP targets;
- strips URL fragments before provider submission;
- requires a local GoreeCloud purpose binding and does not forward that purpose to TinyFish;
- pins transport to `agent.tinyfish.ai`, disables ambient proxy use, rejects redirects, and requires TLS 1.2 or later;
- bounds provider response and normalized output sizes;
- validates provider run identities before using them in status or cancellation paths;
- normalizes terminal states and does not expose raw provider error payloads;
- preserves caller cancellation and attempts provider cancellation for cancellable async runs.

Provider-derived page content and automation output remain untrusted external material. They do not become instructions, tool authority, configuration, code, persistence authority, credential authority, or follow-up disclosure authority merely because TinyFish returned them.

## Direct connector precedence

TinyFish Agent or Browser must not replace an available authoritative GoreeCloud connector for the same system. GitHub work remains routed through the GitHub plugin, Google Drive work through the Google Drive plugin, and Gmail work through the Gmail plugin when those connected capabilities can perform the task. TinyFish is reserved for approved public-web work, external services without an appropriate direct connector, and actual browser-rendering or UI-interaction requirements.

## Browser capability state

This slice defines the provider-neutral Browser capability boundary but intentionally does not create a TinyFish Browser API session transport yet. Browser API sessions expose low-level remote browser control and can carry broader interaction authority than a bounded Agent goal. The Browser executor therefore remains unavailable unless a separately reviewed executor is injected.

A later Browser implementation must independently validate the current TinyFish Browser API request/response contract, remote-session lifecycle, CDP URL handling, Browser Context Profile setup/save lifecycle, timeout and cleanup behavior, profile isolation, Vault pairing, metered Browser duration, evidence minimization, and target-runtime acceptance.

## Lifecycle boundary

This Development foundation does not:

- expose a public or user-facing web-automation HTTP route;
- automatically escalate Search, Fetch, or Research calls into Agent or Browser;
- configure or disclose a production TinyFish API key;
- create, modify, or reuse any live Browser Context Profile;
- read or inject any live TinyFish Vault credential;
- execute a live metered Agent or Browser run;
- establish automatic budget or wallet control;
- establish Privacy Shield authorization or Wardveil runtime acceptance;
- establish monitoring, recovery, rollback, or target-runtime acceptance;
- authorize production deployment, Release Candidate status, or Stable status.

Those remain separate implementation and acceptance phases under FR-007, FR-008, the GoreeCloud Search project specification, and normal GoreeCloud release governance.
