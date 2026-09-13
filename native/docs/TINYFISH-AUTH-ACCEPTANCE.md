# TinyFish authenticated-session acceptance evidence

**Lifecycle:** Development source evidence only. This document does not establish a production-approved TinyFish login, Privacy Shield authorization, Wardveil protected state, Everkeep recovery evidence, deployment acceptance, Release Candidate status, or Stable status.

## Purpose

GoreeCloud Search now has three separate authenticated-automation controls:

1. exact-host bindings choose the intended Browser Context Profile and only explicitly scoped Vault credential references;
2. the Browser Context Profile lifecycle can create, repair, save, or cancel authorized profile setup state;
3. this acceptance-evidence primitive records what has actually been observed for each exact-host binding without storing credential values or live browser-control endpoints.

This separation is deliberate. A configured profile ID does not prove the profile is logged in. A successful setup save does not prove a future run can reuse the session. A TinyFish run reaching `COMPLETED` does not prove the requested authenticated goal succeeded. Production claims therefore require explicit evidence rather than inference from configuration or provider lifecycle status.

## Evidence states

The Development evidence model supports normalized observations:

- `setup_saved` — an authorized setup session was successfully persisted;
- `session_reuse_verified` — a later run proved the saved session was reusable;
- `vault_repair_verified` — a stale/expired session was successfully repaired using the credential scope already bound to that exact host;
- `mfa_required` — interactive MFA or equivalent user participation prevented unattended completion;
- `site_blocked` — CAPTCHA, access denial, bot protection, or another site barrier blocked the flow;
- `expired` — the previously usable saved state is currently expired/stale;
- `revoked` — the profile or site session is known to have been revoked;
- `failed` — the authenticated flow failed without a more specific normalized category;
- `secret_disclosure_detected` — an acceptance test observed credential/session material appearing somewhere it must not appear.

Bindings with no observation are explicitly reported as `unverified` rather than being treated as healthy.

## Exact-host and profile identity rule

Every observation must match an existing exact-host authentication binding and the exact Browser Context Profile ID configured by that binding.

A result observed for `github.com` cannot be recorded against `api.github.com`, another sibling/subdomain, or a different profile. This prevents successful evidence from one authentication boundary being reused to claim acceptance for another.

## Vault repair evidence rule

`vault_repair_verified` is valid only when the exact-host binding already contains explicit Vault credential item references. A profile-only binding cannot claim Vault repair evidence.

The evidence registry never stores or accepts the underlying password, passkey, OTP, cookie, API key, or other reusable credential value.

## Temporal integrity

Each observation includes an explicit observation timestamp. A stale observation cannot overwrite a newer result for the same host.

Historical positive evidence is retained separately from the current state. For example, if session reuse was previously verified and the session later expires, the current state becomes `expired` while the historical `session_reuse_verified` evidence remains visible. This allows operators to distinguish "never worked" from "worked previously but is not currently usable."

A secret-disclosure observation is sticky. Later successful navigation does not erase evidence that a secret-handling defect was observed.

## Privacy-minimized status

The status/snapshot contract exposes only:

- exact host;
- current normalized state;
- observation time;
- reviewed `lite`/`stealth` browser profile mode;
- whether a Vault scope is configured;
- whether setup-save, session reuse, or Vault repair have ever been verified;
- whether secret disclosure has ever been detected;
- observation count.

It intentionally omits:

- Browser Context Profile IDs;
- Vault credential item IDs;
- passwords, passkeys, OTPs, cookies, API keys, or tokens;
- CDP URLs, setup-session IDs, TinyFish base URLs, screenshots, page HTML, prompts, and page content;
- account usernames/emails and other provider account identifiers.

This makes the primitive suitable as a future source for restricted Manager operational visibility without turning authentication telemetry into a credential inventory.

## Current-state semantics

The current state is evidence, not authorization. In particular:

- `session_reuse_verified` does not approve future browser actions;
- `vault_repair_verified` does not grant broad Vault access;
- `setup_saved` does not prove the profile is currently authenticated;
- `unverified` does not mean broken, only that GoreeCloud lacks acceptance evidence;
- a successful historical observation does not override a newer `expired`, `revoked`, `mfa_required`, `site_blocked`, or `failed` state.

Every real Agent/Browser operation still requires the existing GoreeCloud authorization, purpose, routing, budget, and provider controls.

## Connector precedence

GitHub, Google Drive, Gmail, and other authoritative direct connectors remain preferred wherever they provide the required operation. TinyFish authenticated-browser acceptance exists for external websites and UI-only gaps, not as a replacement for direct connectors.

## Persistence boundary

The initial registry is intentionally in-memory Development infrastructure. Durable/distributed acceptance state, retention policy, audit history, Manager UI integration, incident escalation, and Everkeep recovery evidence require separate design and review.

## Production acceptance still required

Before authenticated TinyFish automation may be accepted in production for any service/account boundary, evidence should include at minimum:

- exact-host profile binding established;
- authorized profile setup/save completed;
- later saved-session reuse verified;
- scoped Vault repair verified when Vault repair is required for that service;
- MFA/SSO/CAPTCHA/access-denial behavior understood;
- no credential/session disclosure observed;
- cost and budget behavior verified;
- revocation and rollback tested;
- Privacy Shield authorization and purpose limitation accepted in the target runtime;
- Wardveil runtime controls and evidence accepted;
- deployment and target-runtime acceptance completed.

Source implementation or CI success alone does not satisfy these gates.