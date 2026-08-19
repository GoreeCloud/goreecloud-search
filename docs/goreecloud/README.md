# GoreeCloud Search Documentation

GoreeCloud-owned documentation for the maintained SearXNG fork.

- `UPSTREAM.md` — upstream baseline and sync policy.
- `API.md` — governed machine-interface boundary.
- `DEPLOYMENT.md` — deployment model and production boundary.
- `READINESS.md` — source-versus-production readiness contract.
- `GLAZE-UI-CONFORMANCE.md` — exact Glaze UI 1.1 source binding, semantic mapping, automated evidence, and Stable visual-acceptance boundary.
- `BROWSER-INTEGRATION.md` — GoreeCloud Browser search-provider and failure-boundary contract.
- `CANDIDATE-PUBLICATION.md` — explicit stabilization candidate-request marker, reviewed-base binding, immutable GHCR publication, and non-authorizing rehearsal control.
- `STABLE-CUTOVER.md` — first-Stable evidence gates, immutable release identity, runtime binding, and cutover/retirement controls.
- `TARGET-ACCEPTANCE.md` — controlled `goreecloud-vps-01` staging, validation, cutover, recovery, and rollback procedure.
- `RECOVERY-ACCEPTANCE.md` — candidate-bound application-level backup, isolated restore, monitoring, and rollback-evidence procedure.
- `FINAL-ACCEPTANCE.md` — candidate-bound real-provider, Glaze UI 1.1, Browser runtime, and final evidence-manifest procedure.
- `PRODUCTION-CHECKLIST.md` — concise operational gate checklist for the live promotion sequence.

Production authorization requires the complete applicable evidence set. A successful source build, candidate image, target health check, restore artifact, final-candidate evidence manifest, or any other individual gate is insufficient by itself.
