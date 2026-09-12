# GoreeCloud Search — TinyFish Web Intelligence Routing Acceptance Boundary

## Development acceptance

This slice can be treated as Development-source accepted only after the exact branch head passes the repository's applicable CI checks and the resulting change is reviewed and merged to `master`.

Development acceptance covers only the provider-neutral source primitive and its repository documentation. It does not constitute production-provider acceptance.

## Production acceptance remains separate

Before the router can authorize live metered TinyFish use in production, GoreeCloud must separately verify:

- exact target-runtime provider configuration;
- approved credential injection and rotation;
- durable budget and reservation state;
- provider billing/wallet reconciliation;
- multi-instance quota consistency;
- Privacy Shield operation-bound authorization;
- Wardveil handling of untrusted external content and automation output;
- live provider health and recovery behavior;
- direct-connector precedence;
- Manager status semantics and evidence freshness;
- monitoring, alerting, rollback, and recovery evidence;
- explicit production and release acceptance.

Until those gates are satisfied, this source must not be described as enabling production TinyFish budget governance or automatic metered escalation.
