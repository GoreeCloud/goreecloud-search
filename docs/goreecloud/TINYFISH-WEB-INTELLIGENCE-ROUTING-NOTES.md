# GoreeCloud Search — TinyFish Web Intelligence Routing Implementation Notes

This Development slice intentionally stops at a provider-neutral policy primitive. It does not wire the controller into the Search HTTP runtime or any live TinyFish provider.

The next implementation phase after this foundation is accepted should add an adapter layer that maps the already-existing Search, Fetch, Research, and Agent providers into the router without creating a second provider implementation. The adapter must preserve each provider's existing validation and transport controls, and must not make Browser available until the separately reviewed Browser executor exists.

Production work must also decide where durable budget state belongs. The current in-memory accounting is suitable for source-level policy testing but is not sufficient for multiple replicas, process restarts, billing reconciliation, or administrative quota changes. A durable implementation should use a GoreeCloud-owned state service or other approved persistent store with atomic reservations and explicit recovery semantics rather than persisting TinyFish wallet credentials or provider secrets in the router.

Provider-health automation should likewise remain separate from content telemetry. Health evaluation can be driven by bounded normalized success/failure/timeout signals and explicit probes, while user queries, URLs, research prompts, page content, and automation goals remain outside the health record unless separately authorized.
