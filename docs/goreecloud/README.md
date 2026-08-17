# GoreeCloud Search Documentation

GoreeCloud-owned documentation for the maintained SearXNG fork.

- `UPSTREAM.md` — upstream baseline and sync policy.
- `API.md` — governed machine-interface boundary.
- `DEPLOYMENT.md` — deployment model and production boundary.
- `READINESS.md` — source-versus-production readiness contract.
- `TARGET-ACCEPTANCE.md` — controlled `goreecloud-vps-01` staging, validation, cutover, recovery, and rollback procedure.
- `PRODUCTION-CHECKLIST.md` — concise operational gate checklist for the live promotion sequence.

Production authorization requires the applicable gates in `READINESS.md` and `TARGET-ACCEPTANCE.md`; a successful source build alone is insufficient.
