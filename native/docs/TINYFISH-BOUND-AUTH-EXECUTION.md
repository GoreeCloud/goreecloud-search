# TinyFish managed authenticated execution

**Lifecycle:** Development source foundation only. This document does not establish production TinyFish authentication, Privacy Shield authorization, Wardveil protected state, Everkeep recovery evidence, deployment acceptance, Release Candidate status, or Stable status.

## Purpose

Recurring authenticated TinyFish automation must not depend on each caller remembering which Browser Context Profile or Vault credential reference belongs to a website.

GoreeCloud Search therefore applies the managed exact-host authentication binding before invoking the underlying web-automation executor. The target host determines the Browser Context Profile and any explicitly scoped Vault credential references. Caller-supplied profile or Vault overrides remain rejected by the binding registry.

This directly addresses previously observed failure modes where a TinyFish run had no Browser Context Profile attached, used unusable authentication state, or had no usable credential available for reauthentication.

## Execution flow

For an authenticated request:

1. validate the target URL through the existing GoreeCloud web-automation request boundary;
2. resolve the exact target hostname against the managed authentication-binding registry;
3. inject the configured Browser Context Profile and only the credential item references already approved for that exact host;
4. invoke the existing provider-neutral executor;
5. translate only trusted normalized execution outcomes into authenticated-session acceptance evidence.

The wrapper does not broaden the authorized capability. Agent remains Agent and Browser remains Browser. It also does not authorize a different host, sibling hostname, or subdomain.

## Evidence semantics

A successful authenticated execution records `session_reuse_verified` for the bound profile because the requested operation succeeded while the managed saved-session state was attached.

The wrapper deliberately **does not** record `vault_repair_verified` merely because Vault references were available to TinyFish. Credential availability is not proof that TinyFish actually used a credential to repair a stale session. Vault-repair evidence requires a separate trusted signal that explicitly proves repair occurred.

Normalized site barriers such as CAPTCHA/access denial record `site_blocked`. A normalized requested-goal failure records `failed`.

Caller cancellation, deadline expiration, provider transport failure, and other infrastructure failures are not treated as authentication-state evidence. Doing so would incorrectly mark a Browser Context Profile as unhealthy when the failure may be unrelated to authentication.

## Fail-closed behavior

The execution wrapper requires:

- a non-empty managed exact-host authentication-binding registry;
- an authenticated-session acceptance registry;
- an underlying executor.

Unbound websites fail before provider execution. Caller-supplied profile/Vault references fail before provider execution. If a provider operation succeeds but the resulting acceptance evidence cannot be recorded, the wrapper fails closed instead of returning an untracked authenticated success.

## Secret-handling boundary

This layer handles references and normalized outcomes only. It does not accept, store, log, or expose passwords, passkeys, OTPs, cookies, API keys, session tokens, CDP URLs, or page content.

The exact-host binding continues to carry only Browser Context Profile IDs and optional credential item references. Privacy-minimized acceptance snapshots continue to omit those identifiers and all reusable secrets.

## Connector precedence

GitHub, Google Drive, Gmail, and other authoritative direct connectors remain preferred wherever they provide the required operation. TinyFish authenticated execution is intended for external websites and browser-only gaps, not as a replacement for first-party connectors.

## Navigation reliability direction

This wrapper improves deterministic session selection, but it cannot make an empty, expired, revoked, or incorrectly configured provider-side profile valid by itself. Reliable production use still requires live acceptance for each service/account boundary, including:

- authorized profile setup and save;
- later saved-session reuse;
- scoped Vault repair where needed;
- MFA/SSO/CAPTCHA/access-denial behavior;
- `lite` versus explicitly authorized `stealth` behavior;
- cost and budget evidence;
- revocation and rollback;
- secret non-disclosure;
- Privacy Shield and Wardveil target-runtime acceptance.

## Development boundary

This source slice does not enumerate or modify live TinyFish profiles or Vault items, run a metered TinyFish task, enable production credentials, authorize automatic stealth/proxy retries, persist acceptance state, expose a public execution route, or qualify GoreeCloud Search as Release Candidate or Stable.
