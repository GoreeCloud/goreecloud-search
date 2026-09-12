# GoreeCloud Search — TinyFish Agent and Browser Escalation Development Contract

## Status

Development source foundation. This record does not establish live-provider acceptance, production TinyFish credentials, Privacy Shield authorization, Wardveil protected state, Everkeep recovery evidence, production deployment authority, Release Candidate qualification, or Stable acceptance.

## Purpose

TinyFish Agent and Browser are escalation capabilities behind GoreeCloud-owned boundaries. They must not become the default route for discovery or retrieval, and they must not create a direct permanent GoreeCloud AI dependency on TinyFish.

The intended least-capability order remains:

`Search → Fetch → Research → Agent → Browser`

Agent is appropriate only when an approved web task requires interaction such as navigation, clicking, or form work. Browser is a lower-level capability for direct CDP/automation control and therefore requires a separate executor and stronger authorization rather than silently inheriting Agent authority.

## Source boundary

`native/internal/webautomation` defines:

- a provider-neutral `Request`, `Result`, `Authorizer`, and `Executor` contract;
- explicit `agent` and `browser` capabilities;
- a fail-closed `Escalator` that has no built-in allow-all policy;
- bounded target URL, goal, purpose, profile, credential-reference, structured-output, step, and duration validation;
- `lite` as the normalized default browser runtime profile;
- an explicit `stealth` runtime profile value that policy can reject unless separately justified;
- explicit Browser Context Profile reuse through `use_profile` plus a required profile ID;
- TinyFish Vault use only with explicitly scoped credential item IDs;
- provider-supported structured output through a bounded object `output_schema`;
- a TinyFish Agent implementation using the cancellable asynchronous Agent lifecycle.

The TinyFish Agent implementation is pinned to the official Agent host and uses:

- `POST /v1/automation/run-async` to create a run;
- `GET /v1/runs/{id}` to poll bounded status;
- `POST /v1/runs/{id}/cancel` when the GoreeCloud caller cancels or the local processing window expires;
- `X-API-Key` authentication only;
- `browser_profile` plus optional Browser Context Profile and explicitly scoped Vault references;
- optional bounded `output_schema` for result structure;
- required bounded `agent_config.max_duration_seconds`;
- `agent_config.max_steps` only when the runtime explicitly enables TinyFish's reviewed beta contract.

TinyFish currently documents `agent_config.max_steps` as beta-limited. GoreeCloud therefore fails closed when a caller requests a max-step bound but the configured TinyFish runtime has not explicitly enabled that beta capability. GoreeCloud will not silently discard a requested execution limit.

## Authenticated automation reliability policy

TinyFish's current guidance favors persistent browser state over repeated cold logins. GoreeCloud therefore treats authenticated automation as a session-reuse problem first and a password-injection problem second.

For recurring authenticated workflows, the preferred sequence is:

1. create a dedicated Browser Context Profile for the specific approved service and purpose;
2. establish and save a valid signed-in session in that profile;
3. run later Agent operations with that explicit profile ID;
4. pair the profile with TinyFish Vault only for stale-session repair when reauthentication becomes necessary; and
5. scope the Vault run to the exact credential item IDs required for that service/account.

GoreeCloud does not authorize `use_vault: true` with an omitted credential list. TinyFish permits that form and may expose all enabled Vault items to the run, but TinyFish's own guidance states that explicit `credential_item_ids` are more reliable when multiple accounts or credentials exist. GoreeCloud additionally requires explicit scoping for data-minimization and authority reasons.

One broad shared Browser Context Profile spanning unrelated services is prohibited. Profiles should be separated by service, account, and operational purpose where practical so cookies, local storage, session storage, workspace selection, and site-specific state do not bleed across unrelated automation.

## Navigation reliability policy

TinyFish's prompting guidance indicates that explicit goals, structured output, and edge-case handling materially improve success. The GoreeCloud TinyFish Agent adapter therefore augments approved goals with bounded provider-specific reliability instructions that:

- treat the saved Browser Context Profile as the primary authentication state when profile reuse is enabled;
- use a scoped Vault credential only when authentication or reauthentication is actually necessary;
- tell the agent not to reveal or return credential values;
- wait for dynamic content after navigation;
- use visible site navigation, scrolling, and pagination when required; and
- stop and report CAPTCHA, access-denied, bot-block, or similar barriers instead of looping through repeated login or navigation attempts.

Callers should use a bounded `output_schema` when the expected result has a known structure. This gives the provider a more precise completion contract and reduces ambiguous free-form completion output.

A TinyFish run marked `COMPLETED` is not automatically treated as a successful GoreeCloud operation when the provider returns a documented run-level error. The adapter now parses only the provider's machine-readable error code and maps it to stable GoreeCloud errors without exposing provider messages or page content. Current normalized classes include:

- `SITE_BLOCKED` → site-blocked state;
- `TASK_FAILED` → goal-failed state;
- `MAX_STEPS_EXCEEDED` and `TIMEOUT` → execution-limit state;
- billing/credit rejection codes → billing-rejected state; and
- provider cancellation → caller-visible cancellation semantics.

This distinction is important because TinyFish documents `TASK_FAILED` for navigation, content, or authentication failure and `SITE_BLOCKED` for anti-bot, CAPTCHA, or IP blocking even when the API request itself returned HTTP 200.

## Stealth, anti-bot, and retry behavior

`lite` remains the default because it is the least specialized runtime. TinyFish recommends `stealth` and, where authorized, an appropriate proxy for sites with confirmed anti-bot behavior.

GoreeCloud does **not** automatically retry every failed Agent task in stealth mode. An Agent run may already have performed state-changing actions before the final failure signal, so an automatic retry can duplicate form submissions, purchases, changes, messages, administrative operations, or other effects.

A caller or higher-level GoreeCloud policy may authorize a stealth retry only after evaluating the operation's idempotency, prior-run evidence, cost, site policy, Privacy Shield authority, Wardveil constraints, and target-specific risk. Read-only navigation and extraction are better candidates for bounded retry than state-changing workflows.

When a site remains difficult for Agent navigation but the operation is approved and requires deterministic interaction, the next architectural option is a separately authorized Browser executor using direct browser/CDP control rather than repeatedly increasing Agent freedom.

## Authorization boundary

The package deliberately cannot authorize itself. Every operation must pass through an injected GoreeCloud `Authorizer` before an executor is selected.

A future production authorizer must independently evaluate the calling identity and application, declared purpose, requested target, external-processing permission, Privacy Shield authorization, Wardveil handling requirements, session-state scope, Vault use, credential references, requested runtime profile, capability level, cost/budget state, retry safety, and any additional service-specific policy.

Missing, malformed, stale, contradictory, unavailable, or insufficient authority must fail closed. The existence of an API key, Browser Context Profile, Vault credential item, TinyFish account balance, or prior successful login is not authorization.

## Browser Context Profiles and Vault

Browser Context Profiles persist cookies, local storage, and session storage across Agent runs. They are distinct from TinyFish Browser Profiles (`lite` and `stealth`), which select runtime behavior.

GoreeCloud should use narrowly scoped Browser Context Profiles for specific approved external services or purposes instead of one shared global profile. A profile ID must be explicit whenever session reuse is requested.

Vault credential references are provider-side references, not authorization tokens. GoreeCloud source, ordinary logs, evidence, prompts, documentation, and caller-visible errors must not contain reusable credentials. Credential item IDs are accepted only when Vault use is explicitly requested, must be explicit whenever Vault use is enabled, and are bounded and deduplicated before provider submission.

## Security and privacy controls

The source foundation:

- accepts only public HTTP(S) target URLs without embedded credentials;
- rejects localhost, local-domain, and literal non-public IP targets;
- strips URL fragments before provider submission;
- requires a local GoreeCloud purpose binding and does not forward that purpose to TinyFish;
- pins transport to `agent.tinyfish.ai`, disables ambient proxy use, rejects redirects, and requires TLS 1.2 or later;
- bounds provider response, structured-output schema, and normalized output sizes;
- validates provider run identities before using them in status or cancellation paths;
- recognizes both provider `result` and current `result_json` response fields;
- normalizes documented run-level failures and does not expose raw provider error payloads;
- preserves caller cancellation and attempts provider cancellation for cancellable async runs; and
- refuses unscoped Vault access and unsupported requested execution controls.

Provider-derived page content and automation output remain untrusted external material. They do not become instructions, tool authority, configuration, code, persistence authority, credential authority, or follow-up disclosure authority merely because TinyFish returned them.

## Direct connector precedence

TinyFish Agent or Browser must not replace an available authoritative GoreeCloud connector for the same system. GitHub work remains routed through the GitHub plugin, Google Drive work through the Google Drive plugin, and Gmail work through the Gmail plugin when those connected capabilities can perform the task. TinyFish is reserved for approved public-web work, external services without an appropriate direct connector, and actual browser-rendering or UI-interaction requirements.

## Browser capability state

This source line defines the provider-neutral Browser capability boundary but intentionally does not create a TinyFish Browser API session transport yet. Browser API sessions expose low-level remote browser control and can carry broader interaction authority than a bounded Agent goal. The Browser executor therefore remains unavailable unless a separately reviewed executor is injected.

A later Browser implementation must independently validate the current TinyFish Browser API request/response contract, remote-session lifecycle, CDP URL handling, Browser Context Profile setup/save lifecycle, timeout and cleanup behavior, profile isolation, Vault pairing, metered Browser duration, evidence minimization, and target-runtime acceptance.

## Lifecycle boundary

This Development foundation does not:

- expose a public or user-facing web-automation HTTP route;
- automatically escalate Search, Fetch, or Research calls into Agent or Browser;
- configure or disclose a production TinyFish API key;
- create, modify, or reuse any live Browser Context Profile;
- read or inject any live TinyFish Vault credential;
- execute a live metered Agent or Browser run;
- establish automatic stealth retry or proxy use;
- establish automatic budget or wallet control;
- establish Privacy Shield authorization or Wardveil runtime acceptance;
- establish monitoring, recovery, rollback, or target-runtime acceptance;
- authorize production deployment, Release Candidate status, or Stable status.

Those remain separate implementation and acceptance phases under FR-007, FR-008, the GoreeCloud Search project specification, and normal GoreeCloud release governance.
