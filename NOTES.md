# GoreeCloud Search — Repository Notes

## Current development baseline

The repository's authoritative `main` line was initialized on September 16, 2026. Current implementation work remains on stacked development candidates and is not accepted on `main`.

## Design decisions

- Python 3.11+ keeps the initial core dependency-light.
- Provider execution is policy-following rather than provider-selecting: only source-plan-approved providers can run.
- Query-disclosure budgets narrow the plan before execution and never expand provider eligibility.
- Provider exceptions/timeouts become explicit execution state; caller cancellation propagates.
- Content policy is a distinct stage after normalization and before ranking.
- Policy hooks are attributable; exceptions or spoofed decision provenance fail closed.
- Moderate/Strict SafeSearch must be enforceable before any provider call or Search rejects the request.
- The built-in domain hook is administrator policy only, not a semantic/safety classifier.
- The Brave Web Search adapter is an opt-in Development path; its presence is not production provider acceptance.
- The local HTTP API remains fixed to IPv4 loopback and is a Development boundary only.
- Authoritative GoreeCloud Docker inventory establishes that the operational Search/SearXNG lineage is hosted on `goreecloud-vps-01`, not on the owner laptop. The verified August 17 deployment uses container name `searxng-core`, Compose project/path `searxng` at `/srv/docker/stacks/searxng/docker-compose.yml`, Caddy backend `searxng-core:8080`, and private hostname `search.goreecloud.com`; historical `searxng-*` names were retained intentionally for compatibility.
- The deployed VPS image `ghcr.io/goreecloud/goreecloud-search:3584da535f7ed7c3b4b8dc73cf0424fb4bdf1949` at recorded digest `sha256:30ec99e3311fa9dcc934ac267aef123bb2541e0cd0165b360c4a7dc4fa29e3d5` is outside the current rewritten repository lineage and is therefore treated as the current rollback baseline, not as source evidence for this native candidate.
- This branch adds a native VPS Docker Development candidate that preserves the existing Caddy backend identity, avoids host-port publication, and joins only the `proxy` network because the native service does not use Valkey. The legacy `searxng-valkey` service and `searxng-internal` network remain live-state compatibility/rollback concerns until separately verified for retirement. The candidate has not been built into an accepted GHCR artifact or deployed on the VPS.
- Earlier Draft PRs #13 and #14 were closed after correcting the mistaken laptop-daemon replacement assumption. The laptop is a Browser/client acceptance target, not the Search server runtime.
- `deploy/vps/read-only-preflight.sh` is a pre-cutover evidence collector only. It performs no Docker/Compose/Caddy/file mutation, does not read provider secret values or container environment, does not emit raw Compose/Caddy contents or logs, performs no provider query, and probes only the Search HTTPS homepage. Its output does not authorize deployment; the actual VPS run and evidence review remain open.
- This version is Development only.

## Open decisions

- Final production service/runtime framework.
- Public source licensing/rights model.
- Production acceptance criteria and provider set beyond the Development Brave adapter.
- Production safety/content classification sources and Wardveil integration.
- Exact current Stable Platform-System contract versions at integration time.
- Verified native VPS deployment artifact, exact live Compose reconciliation, Caddy/private-DNS acceptance, monitoring, rollback, and retirement of any obsolete SearXNG-derived supporting state.

## Stacked implementation candidates

The current stack continues through the Lens, snippet, and external-provider candidates: `feature/native-search-core-foundation` → `feature/index-contract-normalization` → `feature/deterministic-ranking-explanations` → `feature/provider-execution-pipeline` → `feature/query-disclosure-budget` → `feature/content-policy-hooks` → licensing/Lens candidates → `feature/snippet-generation` → `feature/brave-web-provider` → `feature/local-http-api`. Each child must be retargeted/reconciled and revalidated when a parent changes or merges.
