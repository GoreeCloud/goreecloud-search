# GoreeCloud Search — TinyFish Completed-Outcome Validation

## Status

Development hardening. This document does not establish live-provider acceptance, production credentials, production deployment, Release Candidate qualification, or Stable acceptance.

## Problem

TinyFish documents that `COMPLETED` means the browser agent finished its run, not necessarily that the requested goal succeeded. A completed result can still describe a CAPTCHA, access-denied page, missing authentication, missing credentials, or another task failure.

GoreeCloud therefore must not treat TinyFish transport/run completion as sufficient evidence of successful task completion.

## Development behavior

The TinyFish Agent adapter now validates the preferred provider result (`result_json` when present, otherwise `result`) before returning a successful GoreeCloud result.

The validator intentionally uses a narrow, fail-aware contract:

- structured booleans such as `goal_succeeded: false`, `goal_success: false`, `succeeded: false`, or `success: false` map to a GoreeCloud goal-failed state;
- structured status/outcome values such as `failed`, `task_failed`, `goal_failed`, `incomplete`, `authentication_required`, or `unauthorized` map to a goal-failed state;
- explicit `blocked`, `site_blocked`, `captcha`, `access_denied`, or `bot_blocked` status/outcome values map to a site-blocked state;
- a small set of unambiguous free-form statements such as `The task cannot be completed`, `No saved credentials are available`, or an explicit CAPTCHA/access-denied barrier are rejected even when the TinyFish run status is `COMPLETED`.

The validator does **not** reject every occurrence of words such as `blocked`, `captcha`, `login`, or `authentication`. Successful diagnostic output may legitimately say that no CAPTCHA was present or that authentication succeeded. Generic keyword matching would create false failures and is prohibited.

Malformed provider output is still handled by the existing normalized-output validation path. Raw provider content is not copied into GoreeCloud error messages.

## Why this matters for Vault and navigation

This hardening directly addresses the observed failure pattern where TinyFish completed an Agent run but the returned result stated that the browser was not authenticated or that no saved credentials were available. GoreeCloud can now distinguish that outcome from a genuinely successful authenticated task.

It does not repair the live Browser Context Profile or configure TinyFish Vault by itself. Live authenticated reliability still depends on:

1. a valid service-specific Browser Context Profile containing an established signed-in session;
2. an explicitly scoped TinyFish Vault credential item for that exact account when stale-session repair is required;
3. a bounded, precise goal;
4. structured output whenever the expected result shape is known; and
5. separate authorization for stealth, proxy, retry, or direct Browser escalation.

## Provider basis

TinyFish's current documentation states that:

- `COMPLETED` means the agent finished, but the result still needs to be inspected for actual task success;
- a completed result may describe a CAPTCHA or access-denied failure;
- coding-agent integrations should check `result_json` for failure signals; and
- structured output should be used when predictable JSON result shape is important.

GoreeCloud treats those provider semantics as external-provider behavior, not as a replacement for GoreeCloud authorization, privacy, security, release, or evidence requirements.

## Lifecycle boundary

This source hardening does not:

- access or reveal any live password;
- enumerate or configure the user's TinyFish Vault;
- create, save, repair, or inspect a live Browser Context Profile;
- run a metered TinyFish Agent or Browser operation;
- authorize automatic stealth or proxy retry;
- prove Privacy Shield or Wardveil runtime acceptance; or
- establish production, Release Candidate, or Stable status.
