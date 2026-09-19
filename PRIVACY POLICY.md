# GoreeCloud Search — Privacy Policy

## Current development implementation

Version `0.1.0.dev6` contains local query parsing/planning, third-party disclosure budgeting, bounded provider execution, Search ↔ Index contract/pagination, normalization/deduplication, content-policy hooks, and deterministic ranking.

- No authenticated live network transport or approved external network provider ships in this candidate. The development CLI performs no network access.
- Explicitly injected network-capable adapters, if an embedding supplies them, remain bounded by the planner-approved source plan and optional disclosure budget.
- A zero disclosure budget prevents third-party execution for modes with a first-party/local path and makes an unsatisfiable External Only request fail rather than disclose.
- Content-policy hooks run on normalized results before ranking.
- Administrator domain policy is local to the Search process in this candidate.
- Moderate/Strict SafeSearch is rejected before provider execution if no configured hook truthfully declares enforcement, avoiding query disclosure under a falsely protected state.
- No search history, advertising/tracking code, behavioral profile, third-party analytics, paid placement, or click-profile ranking is implemented.

## SafeSearch boundary

Off/Moderate/Strict are intent modes, not proof of classification quality. The built-in domain hook does not classify content. Production SafeSearch requires separately verified classification inputs and applicable GoreeCloud privacy/security/platform authority.

## Future network features

Any live provider, suggestion, synchronization, analytics, AI, Browser, Index, or other network-capable feature must update this policy to match verified behavior and apply relevant GoreeCloud Identity, Privacy Shield, Wardveil Security, retention, consent, minimization, and disclosure controls before production acceptance.
