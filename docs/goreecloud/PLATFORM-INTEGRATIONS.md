# GoreeCloud Search Platform Integrations

## Purpose

GoreeCloud Search integrates GoreeCloud Privacy Shield and Wardveil Security as separate first-party platform systems. The integrations in this repository are not decorative branding. Each identity is tied to a bounded Search-specific contract, canonical source provenance, machine-verifiable consumer assets, and an explicit authority boundary.

## Privacy Shield

GoreeCloud Privacy Shield is the privacy authority for the Search privacy capabilities declared by this repository.

The Search adapter covers:

- Search-data minimization: usage metrics are disabled by default, query text is excluded from page titles, and GoreeCloud Search does not implement query-history retention.
- Request privacy: image proxying is enabled in the controlled runtime baseline, referrer leakage is suppressed with `Referrer-Policy: no-referrer`, and search pages are marked `noindex, nofollow`.
- User control: native Search privacy preferences remain available, and provider/engine choices remain visible and configurable.

The adapter does not claim that GoreeCloud Search makes external providers anonymous. External providers remain separate services with their own retention, geographic, ranking, authentication, and abuse-prevention policies.

Canonical authority: `GoreeCloud/goreecloud-privacy-shield` at revision `1edb768336591e0656226f555c57d32537de9274`.

## Wardveil Security

Wardveil Security is the security authority and security-integration framework for the Search security posture represented by this repository.

The current Search integration records source and application-runtime hardening controls including:

- loopback-only application bind baseline;
- runtime secret separation;
- security response-header baseline;
- HTML-only default output boundary; and
- the requirement that production application ports are not exposed directly merely for convenience.

The Search adapter intentionally reports the overall Wardveil runtime status as `unknown` and `protected_by_wardveil=false`. Source integration alone is not enough to assert an overall **Protected by Wardveil** state. Production network, reverse-proxy, host, monitoring, vulnerability, recovery, and target-environment evidence remain authoritative in their owning systems and acceptance records.

Canonical authority: `GoreeCloud/goreecloud-wardveil-security` at revision `d044e04f35fc09d623dc2ee55810a0e1453b6c01`.

## Non-Overlap Boundary

Privacy Shield and Wardveil Security remain separate authorities.

Privacy Shield owns privacy concerns such as data minimization, tracking resistance, privacy-safe status, privacy controls, and privacy-facing exceptions.

Wardveil Security owns security concerns such as integrity, secret protection, access and authorization protection, exposure control, vulnerability posture, security events, and compromise-related response.

Search may present the two systems together for user convenience, but it must not collapse privacy state into security state or security state into privacy state.

## Canonical Assets

The Search consumer copies are:

- `searx/static/themes/simple/img/privacy-shield.svg`
- `searx/static/themes/simple/img/wardveil-security.svg`

Their source repository revisions and SHA-256 digests are recorded in `goreecloud/platform-integrations.json`. They must remain byte-identical to the approved canonical source copied at the recorded revision unless the integration contract is deliberately advanced and revalidated.

## UI Presentation

The Search UI may present Privacy Shield and Wardveil Security only with language that matches the integration boundary.

Privacy Shield may be presented as active for the declared Search privacy capabilities when the controlled baseline remains intact.

Wardveil may be presented as integrated, but the UI must not present an overall **Protected by Wardveil** claim while `protected_by_wardveil` is false or required runtime evidence is unknown, stale, unavailable, incomplete, or unverified.

## Validation

`tests/unit/test_goreecloud_platform_integrations.py` validates:

- separate platform authorities;
- exact canonical revision pins;
- consumer-asset SHA-256 integrity;
- the Privacy Shield adapter boundary;
- the fail-closed Wardveil protection claim; and
- the required Search runtime-baseline evidence markers.

The GoreeCloud foundation workflow runs this validation on the exact source under test.
