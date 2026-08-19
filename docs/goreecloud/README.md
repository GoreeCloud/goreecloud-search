# GoreeCloud Search Documentation

GoreeCloud-owned documentation for the maintained SearXNG fork.

- `UPSTREAM.md` — upstream baseline and sync policy.
- `API.md` — governed machine-interface boundary.
- `DEPLOYMENT.md` — deployment model and production boundary.
- `READINESS.md` — source-versus-production readiness contract.
- `BROWSER-INTEGRATION.md` — GoreeCloud Browser search-provider and failure-boundary contract.
- `STABLE-CUTOVER.md` — first-Stable evidence gates, immutable release identity, runtime binding, and cutover/retirement controls.
- `TARGET-ACCEPTANCE.md` — controlled `goreecloud-vps-01` staging, validation, cutover, recovery, and rollback procedure.
- `RECOVERY-ACCEPTANCE.md` — candidate-bound application-level backup, isolated restore, monitoring, and rollback-evidence procedure.
- `PRODUCTION-CHECKLIST.md` — concise operational gate checklist for the live promotion sequence.

Production authorization requires the complete applicable evidence set. A successful source build, candidate image, target health check, restore artifact, or any other individual gate is insufficient by itself.
