## Production acceptance change

Describe the target-environment or release-readiness change and the evidence it adds.

### Required review

- [ ] No secret, private key, token, password, or reusable credential is included.
- [ ] No unnecessary public backend port is introduced.
- [ ] Docker, Caddy, DNS, NetBird, firewall, monitoring, backup, and recovery boundaries remain explicit.
- [ ] The exact source/image provenance is recorded when applicable.
- [ ] Glaze UI and privacy-by-default behavior are preserved.
- [ ] Rollback remains possible.
- [ ] Production is not declared accepted without target-environment evidence.
