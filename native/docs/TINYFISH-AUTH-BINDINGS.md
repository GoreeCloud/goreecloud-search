# TinyFish authenticated-session bindings

**Lifecycle:** Development source control only. This document does not establish production acceptance, a usable live TinyFish login, Privacy Shield authorization, Wardveil protection, Everkeep recovery evidence, or Release Candidate/Stable status.

## Purpose

Recurring authenticated web automation must not guess which saved Browser Context Profile or TinyFish Vault item belongs to a website. GoreeCloud Search therefore provides an exact-host authentication-binding primitive under `native/internal/webautomation`.

A binding associates one exact public DNS host with:

- one Browser Context Profile ID;
- zero or more explicitly scoped Vault credential item IDs used only for stale-session repair or reauthentication;
- the reviewed TinyFish browser profile (`lite` by default, `stealth` only when explicitly configured).

The binding stores references only. Passwords, passkeys, one-time codes, API keys, cookies, session contents, and other reusable secret values must not be placed in this document, committed configuration, logs, status output, or provider-visible goals.

## Exact-host rule

Host matching is deliberately exact and fail closed. A binding for `example.com` does not authorize `login.example.com`, `admin.example.com`, another subdomain, an IP literal, localhost, or a different registrable domain.

This prevents a saved login session or Vault credential from being silently paired with a different website because of broad suffix matching or navigation ambiguity.

## Managed-binding rule

When a managed binding is applied, the caller may not also supply ad-hoc `profile_id`, `credential_item_ids`, `use_profile`, or `use_vault` values. Mixed authority is rejected rather than choosing one source implicitly.

The resulting provider-neutral request always enables the saved Browser Context Profile. Vault is enabled only when the matched binding contains explicit credential item IDs. This preserves the preferred authentication order:

1. reuse the saved authenticated browser session;
2. use only the matched scoped Vault credential when reauthentication is actually required;
3. stop and return a normalized failure when authentication cannot be repaired;
4. never broaden Vault access to all enabled credentials by omission.

## Configuration document

`ParseAuthBindings` accepts a strict bounded JSON document with schema version `1`:

```json
{
  "schema_version": 1,
  "bindings": [
    {
      "host": "portal.example.com",
      "profile_id": "prof_example",
      "credential_item_ids": ["credential_example"],
      "browser_profile": "lite"
    }
  ]
}
```

The parser rejects unknown fields, unsupported schema versions, duplicate canonical hosts, unsafe/ambiguous hosts, malformed IDs, excessive credential references, oversized documents, and trailing JSON.

The example identifiers above are placeholders only. Real profile or credential-reference identifiers are deployment-controlled sensitive configuration and must not be committed to the repository.

## Relationship to observed TinyFish failures

Recent GoreeCloud TinyFish runs demonstrated three distinct failure classes that must not be conflated:

- no Browser Context Profile was attached;
- a Browser Context Profile was attached but did not provide a usable authenticated session;
- the browser reported that no saved credential was available for reauthentication.

A service-scoped binding addresses the first and third classes by making selection deterministic and preventing an unrelated credential/profile from being used. It cannot make an expired or never-authenticated Browser Context Profile valid by itself. The profile must still be established in TinyFish and the referenced Vault item must exist and be enabled on the provider side.

## Completed-run semantics

A TinyFish provider run reaching `COMPLETED` is not sufficient proof that the requested authenticated task succeeded. The existing TinyFish Agent adapter validates the preferred provider result and converts explicit CAPTCHA/access-denied/site-blocked, authentication failure, missing-credential, and incomplete-task outcomes into GoreeCloud failures before returning success.

Managed bindings complement that check; they do not replace it.

## Connector precedence

Direct authoritative GoreeCloud connectors retain precedence. GitHub operations should use the GitHub connector, Google Drive operations should use the Google Drive connector, and Gmail operations should use the Gmail connector whenever the required capability exists. TinyFish authenticated browser automation is for external sites or UI-only gaps that do not have an appropriate direct connector.

## Production acceptance still required

Before authenticated TinyFish automation can be accepted for production, GoreeCloud still needs live-provider evidence for each approved service/account binding, including:

- successful Browser Context Profile creation and reuse;
- scoped Vault repair without credential disclosure;
- correct exact-host isolation;
- explicit handling of sign-out, expired sessions, MFA, CAPTCHA, access denial, and bot blocks;
- cost/budget evidence;
- rollback and profile revocation behavior;
- Privacy Shield authorization and purpose binding;
- Wardveil runtime protection/evidence;
- target-runtime and deployment acceptance.

Source implementation or CI success alone does not satisfy those gates.
