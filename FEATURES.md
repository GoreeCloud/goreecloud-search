# GoreeCloud Search — Current Features

**Lifecycle:** Development  
**Version:** `0.1.0.dev1`

This file records behavior in the current development candidate and does not claim release, deployment, production acceptance, or Stable qualification.

## Implemented in the current development candidate

- Native Python package under `src/goreecloud_search`.
- Typed search categories and source modes.
- Initial query-operator parser.
- ISO date-range validation.
- Domain and file-extension normalization.
- Deterministic provider eligibility and source planning.
- Index-first, Federated, GoreeCloud-only, External-only, and Offline/local planning.
- Explicit third-party query-disclosure signal.
- Fail-closed external-disclosure invariant for restrictive modes.
- Replaceable provider protocol and normalized result-candidate contract.
- Development CLI that performs no network access.
- Unit tests and CI for Python 3.11 and 3.12.
- Explicit internal application version.

## Not implemented

No provider performs live search in this candidate. No web UI, Browser integration, AI synthesis, persistent history, or production deployment exists yet.
