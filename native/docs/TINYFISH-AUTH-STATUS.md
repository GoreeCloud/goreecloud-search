# TinyFish authenticated web-automation operational status

**Lifecycle:** Development control-plane visibility only. This document does not establish production TinyFish authentication, Privacy Shield authorization, Wardveil protected state, Everkeep recovery evidence, deployment acceptance, Release Candidate status, or Stable status.

## Purpose

GoreeCloud Search now has source primitives for exact-host authentication bindings, Browser Context Profile lifecycle control, managed authenticated execution, and privacy-minimized acceptance evidence. Operators still need a safe way to determine whether authenticated automation is configured and whether GoreeCloud has actually observed usable session state without exposing Browser Context Profile IDs or Vault credential references.

This slice adds bounded Development runtime loading and a read-only status surface for that purpose.

## Deployment-controlled binding file

Set `GOREECLOUD_SEARCH_WEB_AUTOMATION_AUTH_BINDINGS_FILE` to a deployment-controlled JSON document accepted by the existing exact-host authentication-binding parser.

The file may contain only:

- schema version;
- exact public DNS host;
- Browser Context Profile reference;
- optional explicitly scoped Vault credential item references;
- reviewed `lite` or `stealth` browser mode.

The file must not contain passwords, passkeys, OTPs, cookies, API keys, session tokens, page content, or CDP endpoints.

The loader is bounded to the existing authentication-binding size limit and fails closed when the configured file is absent from disk, oversized, malformed, uses an unsupported schema, contains unsafe hosts, or otherwise violates the binding contract. An unset environment variable is not an error and means authenticated web automation is explicitly unconfigured.

## Status endpoint

The native Development service exposes:

`GET /api/v1/web-intelligence/authentication/status`

The response reports:

- whether the authenticated automation control is configured;
- Development control-plane management scope;
- explicit `credentials_exposed: false`;
- explicit `profile_ids_exposed: false`;
- explicit `credential_item_ids_exposed: false`;
- explicit `production_approved: false`;
- when configured, the privacy-minimized authenticated-session acceptance snapshot.

The snapshot may contain only the already reviewed acceptance fields: exact host, normalized evidence state/time, browser mode, whether a Vault scope exists, historical setup/reuse/repair booleans, disclosure flag, and observation count.

## Unverified is intentional

Loading a Browser Context Profile reference does not prove that the provider-side profile contains a valid login session. A newly loaded binding therefore appears as `unverified` until trusted execution/setup evidence is recorded.

This avoids the earlier failure mode where configuration was mistaken for successful authentication.

## No identifier disclosure

The status endpoint deliberately omits:

- Browser Context Profile IDs;
- Vault credential item IDs;
- passwords, passkeys, OTPs, cookies, API keys, or tokens;
- setup-session IDs, CDP URLs, TinyFish base URLs, prompts, screenshots, HTML, and page content;
- provider account identifiers.

The runtime may use profile and credential references internally to select authorized provider state, but operational visibility must not become a credential inventory.

## Connector precedence

GitHub, Google Drive, Gmail, and other authoritative direct connectors remain preferred whenever they provide the required operation. TinyFish authenticated browser automation remains for external websites and UI-only gaps.

## Production acceptance still required

This status endpoint is not production acceptance. Per-service live acceptance still requires evidence for authorized profile setup/save, later saved-session reuse, scoped Vault repair where needed, MFA/SSO/CAPTCHA/access-denial behavior, `lite` versus explicitly authorized `stealth`, cost/budget behavior, revocation/rollback, secret non-disclosure, Privacy Shield authorization, Wardveil controls, deployment acceptance, and target-runtime verification.
