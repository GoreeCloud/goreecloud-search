# TinyFish Browser Context Profile lifecycle

**Lifecycle:** Development source foundation only. This document does not establish a live authenticated profile, production provider approval, Privacy Shield authorization, Wardveil protection, Everkeep recovery evidence, deployment acceptance, Release Candidate status, or Stable status.

## Purpose

GoreeCloud Search now has exact-host authentication bindings, but a binding can only reference state that actually exists and is usable in TinyFish. Recent authenticated-navigation failures showed that a Browser Context Profile may be absent, attached but unauthenticated, or unable to repair itself because no usable Vault credential is available.

This slice adds a controlled Browser Context Profile lifecycle primitive so an approved operator can create or repair saved browser state without broadening GoreeCloud's credential boundary.

## Official TinyFish lifecycle

Current TinyFish documentation defines the persistent-profile setup flow as:

1. create a named profile with `POST /v1/profiles`;
2. start an interactive setup browser with `POST /v1/profiles/{profileId}/setup-session`;
3. connect an approved Playwright/Puppeteer/CDP controller to the returned short-lived `cdp_url` and complete login;
4. persist the authenticated state with `POST /v1/profiles/{profileId}/save` and the returned `session_id`, or explicitly discard it with `POST /v1/profiles/{profileId}/setup-session/cancel`;
5. use the saved profile on later Agent runs with `use_profile: true` and a specific `profile_id` when required.

TinyFish documents Browser Context Profiles as saved cookies, local storage, and session storage. They are distinct from the `lite`/`stealth` Browser Profile runtime mode.

## GoreeCloud contract

The source contract under `native/internal/webautomation` intentionally exposes only four lifecycle operations:

- create profile;
- start setup session;
- save setup session;
- cancel setup session.

Raw cookie upload, arbitrary profile mutation, profile deletion, bulk credential operations, and generic provider administration are deliberately outside this first stable contract.

Every lifecycle operation requires a local `ProfileAuthorizer`. No allow-all authorizer ships with the package. Authorization receives only normalized local operation context: operation, profile ID when applicable, target URL when applicable, and purpose.

The local purpose binding is not forwarded to TinyFish.

## Sensitive setup-session state

A setup session returns a short-lived `session_id`, `cdp_url`, `base_url`, timeout, and expiry. These values can control or address a live setup browser and therefore must be treated as sensitive ephemeral runtime state.

They must not be placed in:

- normal logs;
- analytics;
- persistent operational-status responses;
- repository configuration;
- feature-roadmap records;
- long-lived Manager history;
- provider-visible goals unrelated to setup.

An approved setup UI or local operator workflow may hold them only for the duration needed to complete or cancel setup.

## Transport controls

The TinyFish profile transport:

- uses only `https://agent.tinyfish.ai/v1/profiles` and its documented child endpoints;
- authenticates through deployment-supplied `X-API-Key` only;
- reuses the existing pinned TinyFish Agent HTTP transport, which disables ambient proxy use, rejects redirects, requires TLS 1.2+, and guards DNS/dial behavior for the official Agent host;
- bounds provider response size;
- validates provider profile/session references;
- requires a `wss://` CDP endpoint and `https://` base endpoint from setup responses;
- bounds setup duration to at most 15 minutes in GoreeCloud even if the provider later permits a longer value;
- normalizes saved-domain summaries and counters;
- maps provider 400/401/404/409 classes to stable GoreeCloud errors without returning provider response bodies.

## Relationship to Vault

This lifecycle primitive does not read a password manager or inject credential values. Its role is to establish or repair saved browser state.

The preferred recurring-authentication order remains:

1. exact-host binding selects the intended Browser Context Profile;
2. the saved authenticated state is reused;
3. only the explicitly scoped Vault credential references attached to that binding may be used for stale-session repair during an Agent run;
4. if reauthentication cannot be completed, the run fails rather than broadening credential access;
5. an authorized operator may later use this lifecycle flow to repair the saved profile deliberately.

## Connector precedence

Direct authoritative GoreeCloud connectors remain preferred. GitHub work should use the GitHub connector, Google Drive work should use the Google Drive connector, and Gmail work should use the Gmail connector whenever those connectors support the requested operation. A Browser Context Profile is not justification to route ordinary connector-supported work through browser automation.

## Production acceptance still required

Before this lifecycle may be enabled in production, GoreeCloud still needs evidence for:

- a reviewed operator/setup surface;
- explicit Privacy Shield authorization and purpose binding for create/setup/save/cancel;
- Wardveil handling of live browser-control endpoints and untrusted page content;
- no logging or persistence of `cdp_url`, session IDs, cookies, passwords, OTPs, or other reusable secrets;
- exact service/account binding and revocation behavior;
- MFA, SSO, CAPTCHA, access-denial, and bot-block behavior;
- cost/budget treatment for setup browsers and later Agent/Browser runs;
- session expiry and abandoned-setup cleanup;
- rollback and profile revocation;
- target-runtime configuration and deployment acceptance;
- live provider evidence that the saved profile can be reused and, when authorized, repaired with the intended scoped Vault credential.

Source implementation and CI success are necessary Development evidence only; they do not prove production authentication works.