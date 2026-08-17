# GoreeCloud Search Target-Environment Acceptance

## Status

Production-acceptance procedure. This document does not authorize cutover merely because a container starts.

The current target is `goreecloud-vps-01`. The existing SearXNG runtime remains the production rollback source until GoreeCloud Search completes every mandatory gate below.

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

## Phase 1 — Record rollback state before change

Before modifying the current runtime, record enough evidence to recreate or restore it:

```bash
sudo docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}'
sudo docker inspect searxng > /tmp/searxng-pre-cutover-inspect.json
sudo docker image inspect "$(sudo docker inspect -f '{{.Config.Image}}' searxng)" \
  > /tmp/searxng-pre-cutover-image.json
```

If the current container uses another name, inspect the live runtime and substitute the verified name rather than guessing.

Inspect and record the active SearXNG Compose file, settings/configuration path, environment source, Docker networks, mounts, image reference, health state, and current Caddy backend. Do not print secret values into logs or documentation.

Back up the active Caddyfile before any material route change. The currently documented active path on `goreecloud-vps-01` is `/srv/docker/caddy/Caddyfile`; inspect the live host first and use the active path actually present.

## Phase 2 — Select an immutable GoreeCloud Search candidate

Production acceptance must identify:

- exact Git commit SHA;
- exact OCI image reference;
- immutable image digest when a registry image is used;
- build/CI evidence for that source revision;
- the GoreeCloud runtime settings derived from `goreecloud/settings.yml.example`.

Do not use an unreviewed mutable `latest` tag as production provenance.

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

The staged service must remain separate from the existing SearXNG production container and route.

## Phase 4 — Run target-environment acceptance

From the exact reviewed source checkout, run:

```bash
./goreecloud/target_acceptance.sh \
  --base-url http://127.0.0.1:8888 \
  --container goreecloud-search
```

The harness is read-only. It validates the application identity, Preferences/About surfaces, health endpoint, privacy/security headers, Docker running/health state when Docker is available, absence of non-loopback published ports, and the representative provider suite.

Provider acceptance covers:

- `general`;
- `images`;
- `news`;
- `videos`;
- `it`;
- `science`.

Provider failures must be classified. A rate limit, captcha, provider block, engine initialization failure, or provider-specific empty result is not automatically an application failure, but it must be recorded and evaluated before acceptance.

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

Use `/srv/docker/stacks/goreecloud-search/` as the stable stack name and `/srv/docker/appdata/goreecloud-search/` for required application-controlled persistent state. Use `goreecloud/compose.production.yml.example` as the reviewed production starting point.

The active production Compose file must identify the immutable candidate image and join only networks required for its role. Do not attach GoreeCloud Search to unrelated Docker networks.

Before Caddy cutover:

1. validate the GoreeCloud Search container is healthy;
2. verify Caddy can reach the backend on the `proxy` network;
3. back up the active Caddyfile;
4. edit only the verified Search route/backend;
5. validate the complete Caddyfile before reload or recreation;
6. preserve the previous SearXNG route/configuration as rollback evidence.

Caddy validation is mandatory:

```bash
sudo docker exec caddy \
  caddy validate --config /etc/caddy/Caddyfile
```

Do not proceed if validation fails.

## Phase 7 — Monitoring and recovery gates

Before cutover is considered complete:

- update the existing Search availability monitor to identify GoreeCloud Search rather than legacy SearXNG only after the new runtime becomes authoritative;
- verify alert delivery through the approved GoreeCloud monitoring/notification path;
- include authoritative GoreeCloud Search deployment/configuration in the approved backup scope;
- perform a representative restore into an isolated location;
- verify the restored configuration can recreate a healthy GoreeCloud Search instance;
- record the restore evidence and rollback procedure.

A provider-managed VPS backup is useful disaster-recovery protection but does not substitute for an application-level restore test.

## Phase 8 — Controlled cutover

Cutover is permitted only when all mandatory gates are green.

After changing the Caddy backend to GoreeCloud Search, validate from an approved NetBird client:

```bash
dig +short search.goreecloud.com
curl -I https://search.goreecloud.com
curl -fsS https://search.goreecloud.com/ | grep -q 'GoreeCloud Search'
```

Repeat representative searches through the real production hostname. Check Caddy and application logs for configuration, certificate, upstream-provider, and repeated connection errors without recording user query history unnecessarily.

## Phase 9 — Rollback

Rollback must remain possible until post-cutover acceptance is complete.

If GoreeCloud Search fails a mandatory production check:

1. restore the previous verified Caddy Search backend/route;
2. validate the complete Caddyfile;
3. reload/recreate Caddy only as required by the actual change;
4. confirm `https://search.goreecloud.com` reaches the previous SearXNG runtime;
5. keep the failed GoreeCloud Search candidate isolated for diagnosis;
6. do not delete previous configuration or recovery artifacts until the failure is understood and documented.

## Completion criteria

Production acceptance is complete only when all of the following are recorded as successful:

- immutable source/image provenance;
- healthy target-host runtime;
- six-category representative provider suite;
- Glaze UI/browser acceptance against the deployed service;
- private DNS, HTTPS, NetBird and Caddy behavior;
- authorized success and unauthorized denial;
- no unnecessary public backend port;
- privacy/security headers and logging behavior;
- monitoring and alert delivery;
- backup scope and representative restore;
- rollback test or equivalent verified rollback evidence;
- post-cutover production-hostname validation;
- authoritative GoreeCloud documentation/inventory updates.

Until then, the existing SearXNG production runtime remains the rollback authority.
