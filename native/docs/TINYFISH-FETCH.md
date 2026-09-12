# GoreeCloud Search — TinyFish Fetch Development Contract

## Status

This document records the initial GoreeCloud-owned TinyFish Fetch retrieval contract implemented in Development source. It is not production-provider approval, live-provider acceptance, Privacy Shield authorization, Wardveil protected-state evidence, Everkeep recovery evidence, production deployment authority, Release Candidate approval, or Stable acceptance.

## Purpose

GoreeCloud Search uses TinyFish Fetch only as a bounded external retrieval capability after discovery. GoreeCloud retains ownership of when retrieval is permitted, which URLs may be sent, how returned content is bounded and normalized, how failures are represented, and whether retrieved material is allowed to flow into later research or agent stages.

The initial source contract is intentionally narrower than TinyFish's complete Fetch API.

## Upstream contract used by this slice

TinyFish documents the public Fetch endpoint as `https://api.fetch.tinyfish.ai`, authenticated with `X-API-Key`. The API accepts a POST JSON request containing `urls`, supports up to 10 URLs in a request, and reports per-URL failures without requiring the whole batch to fail. The GoreeCloud adapter requests Markdown content explicitly and consumes only the documented page URL, final URL, title, extracted text, and per-URL error surfaces required by this first slice.

TinyFish supports additional Fetch controls and metadata. They are deliberately not exposed through the initial GoreeCloud contract and require separate review before use.

## GoreeCloud retrieval contract

`native/internal/retrieval` defines a provider-neutral `Fetcher` interface and normalized `Response`, `Document`, and `Failure` types. Callers do not receive raw TinyFish error text or arbitrary provider-specific fields.

The TinyFish implementation:

- permits at most 10 requested URLs per call;
- accepts only HTTP or HTTPS targets;
- rejects embedded URL credentials;
- rejects localhost, `.localhost`, `.local`, single-label hosts, and literal private/loopback/link-local/multicast addresses;
- strips URL fragments before transmission or return;
- sends the API key only in `X-API-Key`;
- pins transport to the official `api.fetch.tinyfish.ai` service;
- rejects redirects from the TinyFish API endpoint;
- ignores ambient HTTP proxy configuration;
- requires the TinyFish API host to resolve only to public addresses at connection time;
- limits the TinyFish response body to 8 MiB;
- bounds normalized titles and extracted content before returning them to callers;
- converts per-URL provider failures to the stable GoreeCloud code `fetch_failed` instead of exposing raw provider diagnostics;
- preserves request cancellation through the caller context.

## Privacy and security boundary

A retrieval request sends the approved target URL to TinyFish and causes TinyFish infrastructure to fetch that target. GoreeCloud must therefore treat Fetch as external data egress and remote processing.

The existence of these source controls does not create Privacy Shield authorization and must not be presented as a Wardveil protected state. Any production use still requires explicit purpose/egress authorization, provider acceptance, target-runtime evidence, observability and cost controls, recovery/rollback evidence where applicable, and the appropriate GoreeCloud Integral Platform System evidence.

Sensitive signed URLs, private-network URLs, credential-bearing URLs, local application endpoints, and other protected targets must not be routed through this external Fetch path merely because the target parser accepts a public HTTP(S) URL.

## Lifecycle

This source foundation establishes only a bounded retrieval primitive. It does not yet:

- expose a user-facing retrieval endpoint;
- automatically fetch Search results;
- connect retrieval to TinyFish Research, Agent, or Browser;
- establish cost/quota policy;
- establish Manager administration controls;
- establish production credentials or live traffic;
- establish production-provider approval;
- establish production deployment, cutover, Release Candidate, or Stable status.

Those remain later acceptance and implementation phases under the GoreeCloud Search roadmap.
