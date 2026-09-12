# GoreeCloud Search — TinyFish Operational Status Development Contract

## Status

Development source integration under FR-008. This record does not establish production TinyFish routing, live wallet reconciliation, durable accounting, automatic health inference, Manager UI acceptance, Privacy Shield authorization, Wardveil protected state, Everkeep recovery evidence, Release Candidate qualification, or Stable acceptance.

## Purpose

GoreeCloud Search now has a bounded way to load the provider-neutral web-intelligence control-plane configuration and expose a privacy-minimized operational snapshot for future GoreeCloud Manager consumption.

The status surface is informational Development evidence only. It does not execute Search, Fetch, Research, Agent, or Browser requests and does not authorize escalation.

## Configuration boundary

The optional configuration file is selected by:

`GOREECLOUD_SEARCH_WEB_INTELLIGENCE_CONFIG_FILE`

Schema version 1 contains only:

- a total integer micro-unit budget ceiling;
- a per-operation integer micro-unit budget ceiling;
- provider ID;
- exact capability (`search`, `fetch`, `research`, `agent`, or `browser`);
- whether the provider is metered;
- configured priority;
- optional initial health (`healthy`, `degraded`, or `unavailable`).

The loader deliberately accepts no API keys, cookies, Browser Context Profile identifiers, Vault credential identifiers, account identifiers, query text, URLs, research prompts, Agent goals, Browser session state, page content, or provider response bodies.

Provider credentials remain owned by the provider-specific runtime configuration and secret boundary. This control-plane configuration must never become a second credential store.

Configuration parsing is strict and fail-closed: unknown JSON fields, unsupported schema versions, trailing data, duplicate or invalid provider identities, invalid capabilities, invalid health values, invalid budgets, and oversized files are rejected.

If the environment variable is absent, the web-intelligence control plane remains explicitly unconfigured rather than silently inventing default authority.

## Development status endpoint

The native Search development service exposes:

`GET /api/v1/web-intelligence/status`

The response identifies whether the control plane is configured, states that credentials are not exposed, states that production is not approved, and—when configured—returns `Controller.Snapshot()`.

The snapshot is limited to:

- provider ID;
- provider capability;
- metered flag;
- configured priority;
- current health state;
- total budget;
- per-operation budget;
- currently reserved budget;
- recorded spend;
- remaining budget;
- normalized operational outcome aggregates and aggregate duration.

It does not include user search history, browsing history, provider response content, URLs, prompts, credentials, cookies, citations, or session identifiers.

## Manager boundary

This endpoint is intended to become an input to a future GoreeCloud Manager operational view. Manager must preserve the distinction between configuration evidence and runtime acceptance.

A configured provider is not necessarily reachable, healthy in production, privacy-authorized, security-accepted, financially reconciled, or approved for a specific application/runtime. The status endpoint therefore continues to report `production_approved: false` in this Development slice.

Future Manager work should add evidence freshness, runtime identity, source-of-configuration information, wallet/billing reconciliation state, durable quota state, monitoring evidence, and accepted deployment scope before presenting stronger operational claims.

## Runtime and spending boundary

This slice does not wire the controller into request execution. It cannot cause a TinyFish operation or spend wallet funds.

In particular, it does not:

- automatically select TinyFish for a user request;
- automatically escalate Search → Fetch → Research → Agent → Browser;
- execute TinyFish Research, Agent, or Browser;
- read or modify the TinyFish wallet;
- infer cost from hard-coded TinyFish pricing;
- persist reservations or spend across restarts;
- coordinate budgets across multiple Search instances;
- infer provider health from traffic or dashboards;
- expose a public mutation endpoint for budgets or health;
- expose provider credentials.

## Direct connector precedence

GitHub, Google Drive, Gmail, and other authoritative GoreeCloud connectors retain precedence whenever they can perform the requested operation directly. The TinyFish web-intelligence control plane is for approved public-web intelligence, research, external sites without a suitable direct connector, and genuine browser-interaction requirements.

## Next acceptance work

Later FR-008 work still requires durable and multi-instance budget control, current wallet/rate evidence, live-provider health policy with bounded recovery behavior, secure Manager integration, runtime wiring, Privacy Shield authorization, Wardveil untrusted-content handling, monitoring, recovery/rollback evidence, target-runtime validation, and explicit production acceptance.
