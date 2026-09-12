# GoreeCloud Search — TinyFish Research Development Contract

## Status

This document records the initial GoreeCloud-owned TinyFish Research orchestration contract implemented in Development source. It is not production-provider approval, live-provider acceptance, Privacy Shield authorization, Wardveil protected-state evidence, Everkeep recovery evidence, production deployment authority, Release Candidate approval, or Stable acceptance.

## Purpose

GoreeCloud Search is the first-party Internet-research boundary for GoreeCloud AI and other approved research consumers. TinyFish Research is therefore integrated behind a GoreeCloud-owned provider-neutral contract rather than becoming a direct permanent GoreeCloud AI dependency.

The intended boundary is:

`GoreeCloud AI / approved research consumer → GoreeCloud Search → GoreeCloud research contract → TinyFish Research → normalized report and citations`

Search remains authoritative for deciding whether external research is permitted, binding the operation to an approved purpose, bounding provider input/output, normalizing the result, and determining whether later processing is allowed.

## Upstream contract used by this slice

TinyFish documents Research as the API for cited synthesis across multiple sources. The current public Research endpoint is:

`POST https://agent.tinyfish.ai/v1/automation/run-research`

REST authentication uses `X-API-Key`. The documented request body includes required `query` and optional `mode`, `stream`, and `output_language`. TinyFish documents the mode values `auto`, `standard`, `deep`, and `max`. Accepted Research requests return a `text/event-stream`; the event stream includes a terminal `final_result` event that supplies the report, citations, and termination reason.

This first GoreeCloud slice deliberately uses only `mode=standard` and `stream=true`. Higher-cost/deeper modes require a later capability-routing and budget-policy review rather than becoming available implicitly.

## GoreeCloud research contract

`native/internal/research` defines a provider-neutral `Researcher` interface plus bounded request, result, and citation types.

A request requires both:

- a non-empty research query; and
- a non-empty GoreeCloud purpose binding.

The purpose binding stays local in this adapter and is not forwarded to TinyFish. It exists so later Privacy Shield and application authorization layers can bind the external operation to an explicit purpose without making provider-specific fields part of the GoreeCloud contract.

The TinyFish implementation:

- pins requests to the official Research endpoint;
- sends the API key only in `X-API-Key`;
- uses `standard` mode only;
- requests streaming output and accepts only `text/event-stream`;
- bounds query, purpose, and output-language inputs;
- caps a provider research run to a 15-minute GoreeCloud processing window;
- caps processed event-stream bytes and individual event size;
- ignores non-terminal progress/provider events rather than exposing provider-specific event details to callers;
- requires a terminal `final_result` containing a non-empty report;
- bounds the normalized report, citation count, citation titles, citation URLs, and termination-reason text;
- permits only public HTTP(S) citation URLs and strips fragments;
- deduplicates citations by normalized URL;
- does not expose provider response bodies through HTTP-error messages;
- does not persist the query, report, citations, event stream, or API key;
- preserves caller cancellation.

Provider-specific progress events, run statistics, reusable session identifiers, internal evidence objects, raw provider errors, and arbitrary future TinyFish fields are intentionally outside the stable GoreeCloud contract in this source slice.

## Privacy and security boundary

TinyFish Research is external processing. The research query leaves GoreeCloud and TinyFish may contact multiple public sources to produce the requested synthesis.

The existence of this adapter does not create Privacy Shield authorization. A later runtime integration must independently prove that the calling operation is allowed to send the specific query to an external research provider for the declared purpose.

The existence of transport and normalization controls also does not establish a Wardveil protected state. Reports, citations, page text, and any other provider-derived material remain untrusted external content. They must not become instructions, tool authority, executable configuration, secrets, privileged actions, persistence authority, or follow-up disclosure authority merely because they were returned by a research provider.

No reusable credential belongs in source code, repository configuration, normal logs, research output, or ordinary documentation.

## Cost and capability boundary

Research is an escalation capability, not the default discovery path. GoreeCloud should continue to prefer the least capable sufficient operation:

`Search → Fetch → Research → Agent / Browser`

This source slice fixes Research to `standard` mode. Automatic mode escalation, `deep` or `max` use, per-operation budget enforcement, quotas, wallet awareness, cost telemetry, Manager controls, and provider-selection/fallback policy remain later roadmap work.

## Lifecycle

This foundation establishes an internal source primitive only. It does not yet:

- expose a public or user-facing Research HTTP route;
- automatically invoke Research from ordinary Search requests;
- connect GoreeCloud AI directly to TinyFish;
- install a production TinyFish credential;
- enable production TinyFish Research traffic;
- establish live-provider acceptance;
- establish Privacy Shield authorization or Wardveil runtime acceptance;
- establish monitoring, quota, cost, recovery, or rollback acceptance;
- authorize production deployment, Release Candidate status, or Stable status.

Those remain separate implementation and acceptance phases under the GoreeCloud Search roadmap and normal GoreeCloud release governance.
