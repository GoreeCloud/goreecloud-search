# GoreeCloud Search Benefits

## Purpose

This record explains the value GoreeCloud Search is intended to provide without treating unfinished native migration, deployment, or Stable acceptance as completed work.

## User benefits

### Greater control over web-search exposure

GoreeCloud Search is designed to keep Browser search behind one GoreeCloud-controlled discovery boundary rather than embedding unrelated commercial search engines directly into first-party clients. External providers may still observe requests from GoreeCloud infrastructure, so this improves control and reduces direct exposure rather than providing anonymity.

### No GoreeCloud advertising incentive

GoreeCloud Search does not use an advertising or sponsored-placement business model. Result treatment can therefore be governed around usefulness, source visibility, deterministic rules, privacy, and reliability instead of paid placement.

### Honest degraded behavior

The native engine tracks provider availability and timeout conditions using bounded status codes. It can represent partial provider failure without exposing raw provider errors or silently pretending that an unrelated provider is the configured Search authority.

### Safer result presentation

The native engine accepts HTTP/HTTPS destinations, removes URL fragments during normalization, rejects result URLs containing embedded user-info credentials, and prevents malformed provider identities from entering result/source evidence.

### A consistent GoreeCloud experience

The native application is being rebuilt around GoreeCloud-owned presentation, preferences, search semantics, and application boundaries rather than treating inherited SearXNG UI/application structure as permanent product architecture.

## Administrative benefits

### Native-first ownership

GoreeCloud owns the target Search service, provider contracts, product behavior, preferences, presentation, release process, migration plan, and platform integrations. The inherited SearXNG-derived runtime remains only as a transitional dependency while native parity and acceptance are completed.

### Explicit provider contracts

Native provider interfaces make source identity, category capability, execution status, and failure behavior explicit. Invalid provider identities do not become executable or visible runtime evidence, reducing ambiguity in diagnostics and policy enforcement.

### Controlled migration instead of a flag-day rewrite

The transitional runtime can continue to serve retained functionality while native equivalents are implemented, tested, documented, and accepted. This preserves continuity without allowing inherited code to remain the permanent architectural authority.

### Evidence-driven acceptance

Search distinguishes source completion, CI validation, provider acceptance, target-runtime validation, recovery/rollback, migration cutover, and Stable promotion. A green build alone is not treated as production evidence.

## Platform benefits

### Shared first-party discovery boundary

Search can provide a common discovery layer for GoreeCloud Browser and approved future first-party consumers instead of each application independently choosing provider/privacy behavior.

### Clear privacy and security authority

Privacy Shield governs data-use/minimization controls, Wardveil Security governs security acceptance where applicable, Everkeep governs recovery/continuity, GoreeCloud Identity governs account/session identity, and GoreeCloud Mesh governs cross-application coordination. Search does not create parallel authorities for those concerns.

### Replaceable external-provider layer

The native provider contract allows provider selection to evolve without making any one external service the identity of GoreeCloud Search. Production provider adapters still require explicit privacy, operational, credential, and runtime acceptance.

## Current limitations

The native application is not yet a complete Stable replacement for the transitional runtime. Production-approved provider adapters, native feature parity, complete platform-system acceptance, deployment/recovery evidence, and controlled migration/cutover remain pending.
