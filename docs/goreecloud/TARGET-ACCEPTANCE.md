# GoreeCloud Search Target-Environment Acceptance

## Status

Production-acceptance procedure. This document does not authorize cutover merely because a container starts.

The current target is `goreecloud-vps-01`. The current known-good GoreeCloud Search production image and preserved pre-Stable compatibility configuration remain rollback material until the final candidate completes every mandatory gate below.

## Governing deployment model

The permanent GoreeCloud Search stack should use the standard GoreeCloud Docker structure:

```text
/srv/docker/stacks/goreecloud-search/
├── docker-compose.yml
├── .env
├── .env.example
├── config/
│   └── settings.yml
├── scripts/
├── README.md
└── archive/

/srv/docker/appdata/goreecloud-search/
└── cache/
```

Only the directories actually required by the runtime should be created. Secrets remain outside source control and ordinary documentation.

The production backend must not publish an unnecessary host port. Caddy should reach `goreecloud-search:8080` through the approved external `proxy` Docker network. `goreecloud/compose.production.yml.example` records this production topology. The existing loopback-only `goreecloud/compose.yml.example` remains the preferred isolated staging topology.

The Search cache is rebuildable, non-authoritative runtime state for recovery acceptance. The deployment definition, reviewed Search settings, protected runtime-configuration recovery path, Search-specific Caddy route/backend material, and immutable image identities are the application-level recovery scope that must remain reproducible.

## Phase 1 — Record rollback state before change

Before modifying the current runtime, inspect the live host and record enough evidence to recreate or restore the verified pre-change state. Do not guess the current container name from historical documentation.

Example inspection after identifying the actual production container:

```bash
current_container='<verified-current-search-container>'
sudo docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}'
sudo docker inspect "$current_container" > /tmp/goreecloud-search-pre-cutover-inspect.json
sudo docker image inspect "$(sudo docker inspect -f '{{.Config.Image}}' "$current_container")" \
  > /tmp/goreecloud-search-pre-cutover-image.json
```

Inspect and record the active Compose file, settings/configuration path, protected environment source, Docker networks, mounts, immutable or resolved image reference, health state, and current Caddy backend. Do not print secret values into logs or documentation.

Back up the active Caddyfile before any material route change. The currently documented active path on `goreecloud-vps-01` is `/srv/docker/caddy/Caddyfile`; inspect the live host first and use the active path actually present.

The source-controlled known-good image identity is recorded in `goreecloud/release_baseline.json`. Target-host inspection must confirm the actual pre-change state before relying on any documented baseline.

## Phase 2 — Select an immutable GoreeCloud Search candidate

Production acceptance must identify:

- exact Git commit SHA;
- exact OCI image reference;
- immutable image digest;
- build/CI and release-evidence artifacts for that source revision;
- the GoreeCloud runtime settings derived from `goreecloud/settings.yml.example`.

Do not use an unreviewed mutable `latest` tag as production provenance.

The candidate-image workflow must produce `release-evidence.json` and bind the candidate source revision to the immutable GHCR digest before target acceptance begins.

## Phase 3 — Stage without replacing production

Create a temporary isolated acceptance deployment or use an equivalent existing controlled staging path. The temporary deployment may use a loopback-only host mapping so it cannot become an unintended public service.

Example staging preparation from a reviewed checkout:

```bash
sudo install -d -m 0750 /srv/docker/stacks/goreecloud-search-acceptance/config
sudo install -d -m 0750 /srv/docker/appdata/goreecloud-search-acceptance/cache
```

Copy reviewed deployment files deliberately; do not copy credentials from the repository. Populate the protected runtime `.env` separately and use restrictive permissions.

Validate Compose before starting:

```bash
cd /srv/docker/stacks/goreecloud-search-acceptance
sudo docker compose config --quiet
sudo docker compose up -d
sudo docker compose ps
```

The staged service must remain separate from the current production Search runtime and route. Before first-Stable acceptance, confirm the staging container is the expected `goreecloud-search` instance and that its direct host publication is loopback-only, normally `127.0.0.1:8888`.

## Phase 4 — Run candidate-bound target-environment acceptance

From the exact reviewed source checkout, run the read-only acceptance harness with the exact candidate digest and source revision. Invoke the script explicitly through Bash so the procedure does not depend on checkout executable-mode metadata:

```bash
bash goreecloud/target_acceptance.sh \
  --base-url http://127.0.0.1:8888 \
  --container goreecloud-search \
  --expected-image 'ghcr.io/goreecloud/goreecloud-search@sha256:<candidate-digest>' \
  --expected-source '<40-character-release-source-sha>' \
  --evidence-json target-runtime-evidence.json
```

The harness validates application identity, Preferences/About surfaces, the health endpoint, privacy/security headers, Docker running/health state, loopback-only direct published ports, exact running-container image identity, OCI source/revision/license metadata, and the representative provider suite.

After target-runtime identity succeeds, create the separately retained first-Stable provider artifact against that **same staged container**:

```bash
python goreecloud/provider_acceptance.py \
  --base-url http://127.0.0.1:8888 \
  --container goreecloud-search \
  --suite \
  --expected-image 'ghcr.io/goreecloud/goreecloud-search@sha256:<candidate-digest>' \
  --expected-source '<40-character-release-source-sha>' \
  --evidence-json provider-evidence.json
```

Candidate-bound provider evidence fails closed unless the loopback URL and named container resolve to the exact immutable candidate. It verifies the running/healthy container, configured image reference, image ID, and OCI revision before the provider requests, repeats the identity check after the suite, and rejects the artifact if the runtime changed during acceptance.

Do not use `https://search.goreecloud.com` to create pre-cutover provider evidence while that hostname still routes to the previous known-good production runtime. The production hostname is reserved for post-cutover rechecks after a separately authorized release decision.

Provider acceptance covers seven representative categories:

- `general`;
- `images`;
- `news`;
- `videos`;
- `files`;
- `it`;
- `science`.

The first-Stable required set is General, Images, Videos, News, and Files. IT and Science remain additional diagnostics.

Provider failures must be classified. A rate limit, captcha, provider block, engine initialization failure, or provider-specific empty result is not automatically an application failure, but it must be recorded and evaluated before acceptance.

The generated target-runtime and provider artifacts are sanitized. They prove which staged candidate runtime was tested, but they do not prove backup restoration, persistent-data recovery, route rollback, or cutover authorization.

## Phase 5 — Validate production network topology before cutover

The approved private path is:

```text
Approved NetBird client
        |
        v
AdGuard Home private DNS
        |
        v
100.71.27.119:443
        |
        v
Caddy
        |
        v
proxy Docker network
        |
        v
GoreeCloud Search :8080
```

Required checks include:

```bash
dig +short search.goreecloud.com
curl -I https://search.goreecloud.com
openssl s_client \
  -connect 100.71.27.119:443 \
  -servername search.goreecloud.com \
  </dev/null 2>/dev/null |
openssl x509 -noout -issuer -subject -ext subjectAltName -dates
sudo docker network inspect proxy
sudo ss -tulpen
```

An approved NetBird client must resolve `search.goreecloud.com` to the documented private endpoint and succeed. An unapproved source must not reach the private backend. The GoreeCloud Search container must not add a direct public listener.

## Phase 6 — Prepare the permanent production stack

Use `/srv/docker/stacks/goreecloud-search/` as the stable stack name and `/srv/docker/appdata/goreecloud-search/` for required application-controlled runtime state. Use `goreecloud/compose.production.yml.example` as the reviewed production starting point.

The active production Compose file must identify the immutable candidate image and join only networks required for its role. Do not attach GoreeCloud Search to unrelated Docker networks.

Before Caddy cutover:

1. validate the GoreeCloud Search container is healthy;
2. verify Caddy can reach the backend on the `proxy` network;
3. back up the active Caddyfile;
4. edit only the verified Search route/backend;
5. validate the complete Caddyfile before reload or recreation;
6. preserve the previous verified Search route/configuration as rollback evidence.

Caddy validation is mandatory:

```bash
sudo docker exec caddy \
  caddy validate --config /etc/caddy/Caddyfile
```

Do not proceed if validation fails.

## Phase 7 — Monitoring and recovery gates

Before cutover is considered complete:

- update the existing Search availability monitor to identify GoreeCloud Search only at the appropriate migration point when the new runtime becomes authoritative;
- verify alert delivery through the approved GoreeCloud monitoring/notification path;
- include authoritative GoreeCloud Search deployment/configuration in the approved backup scope;
- perform a representative application-level restore into an isolated location;
- verify the restored configuration can recreate a healthy GoreeCloud Search instance;
- preserve known-good image, previous runtime configuration, and previous Search route material for rollback;
- record the restore and rollback evidence without including reusable secrets.

A provider-managed VPS backup is useful disaster-recovery protection but does not substitute for an application-level restore test.

Follow `docs/goreecloud/RECOVERY-ACCEPTANCE.md`. After candidate release evidence and target-runtime evidence exist, create the candidate-bound incomplete template:

```bash
python goreecloud/recovery_evidence.py template \
  --release-evidence release-evidence.json \
  --target-runtime-evidence target-runtime-evidence.json \
  --rollback-baseline goreecloud/release_baseline.json \
  --output recovery-evidence.json
```

Only after the actual isolated restore, monitoring check, and rollback-evidence work is complete should the filled artifact pass:

```bash
python goreecloud/recovery_evidence.py validate \
  --evidence recovery-evidence.json \
  --release-evidence release-evidence.json \
  --target-runtime-evidence target-runtime-evidence.json \
  --rollback-baseline goreecloud/release_baseline.json
```

A successful recovery-evidence validation still does not authorize cutover by itself.

## Phase 8 — Controlled cutover

Cutover is permitted only when all mandatory gates are green and the separate release decision explicitly approves the transition.

After changing the Caddy backend to the accepted GoreeCloud Search candidate, validate from an approved NetBird client:

```bash
dig +short search.goreecloud.com
curl -I https://search.goreecloud.com
curl -fsS https://search.goreecloud.com/ | grep -q 'GoreeCloud Search'
```

Repeat representative searches through the real production hostname. Check Caddy and application logs for configuration, certificate, upstream-provider, and repeated connection errors without recording user query history unnecessarily.

## Phase 9 — Rollback

Rollback must remain possible until post-cutover acceptance is complete.

If the candidate fails a mandatory production check:

1. restore the previous verified Caddy Search backend/route;
2. validate the complete Caddyfile;
3. reload/recreate Caddy only as required by the actual change;
4. confirm `https://search.goreecloud.com` reaches the previous verified Search runtime;
5. keep the failed candidate isolated for diagnosis;
6. do not delete previous configuration or recovery artifacts until the failure is understood and documented.

The rollback mode recorded in recovery evidence may be an actual controlled production-route rehearsal or equivalent verified rollback evidence, as defined by `RECOVERY-ACCEPTANCE.md`. The final acceptance decision must evaluate whether the selected evidence is sufficient for the planned cutover.

## Completion criteria

Production acceptance is complete only when all of the following are recorded as successful:

- immutable source/image provenance;
- candidate-bound healthy target-host runtime identity;
- runtime-bound seven-category representative provider suite, including all five first-Stable required categories;
- Glaze UI/browser acceptance against the deployed service;
- private DNS, HTTPS, NetBird and Caddy behavior;
- authorized success and unauthorized denial;
- no unnecessary public backend port;
- privacy/security headers and logging behavior;
- monitoring and alert delivery;
- backup scope and representative isolated application restore;
- rollback test or equivalent verified rollback evidence;
- post-cutover production-hostname validation when cutover is actually performed;
- authoritative GoreeCloud documentation/inventory updates.

Until every applicable gate is complete, the current known-good production image and preserved rollback material remain authoritative for recovery.
