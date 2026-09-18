# GoreeCloud Search

GoreeCloud Search is the privacy-first, self-hostable search and information-discovery service for the GoreeCloud ecosystem.

> **Lifecycle:** Development  
> **Version:** `0.1.0.dev11`  
> **License:** `AGPL-3.0-or-later`  
> **Current scope:** Native query parsing, privacy-aware source planning, bounded provider execution, query-disclosure budgeting, content-policy hooks, transparent local Lenses with a strict portable v1 format, a versioned GoreeCloud Index contract with pagination, Search-owned normalization/deduplication, deterministic explainable ranking, bounded snippet generation, and an opt-in Brave Web Search API adapter. Production provider acceptance and production safety classification remain open.

## What exists now

- Typed query model and parser for the initial operator set.
- Deterministic source modes and provider eligibility planning.
- Optional per-query third-party disclosure budgets enforced before execution.
- Replaceable provider contracts plus bounded asynchronous execution, timeouts, cancellation, failure isolation, fallback behavior, and explicit availability state.
- Versioned Search ↔ GoreeCloud Index v1 models, transport-injected adapter boundary, cursor pagination, and degraded/warning propagation.
- Conservative URL canonicalization, deduplication, source agreement, and provenance.
- Local content-policy hooks applied after normalization and before ranking.
- Typed SafeSearch intent modes: Off, Moderate, and Strict with fail-closed enforcement availability checks.
- Built-in administrator domain allowlist/blocklist policy with subdomain matching. This domain hook is not a content or threat classifier.
- Deterministic ranking with inspectable signals and “Why this result?” explanations.
- Search-local GoreeCloud Lenses with transparent domain, filetype, and language boost/lower/exclude rules.
- Lens explanations and exclusions are explicit; the selected Lens is removed from the provider-facing query before provider execution.
- Versioned `goreecloud.search-lens.v1` JSON import/export with deterministic serialization and strict schema/size validation.
- Unit tests and pull-request CI.

The repository now ships one opt-in external network provider, a loopback-only Development HTTP API, and a separate Docker-network Development listener intended for the verified VPS reverse-proxy topology. The container path preserves the current `searxng-core:8080` Caddy backend contract without publishing a host port, supports a runtime file for the Brave credential, and includes a hardened non-root container/Compose candidate. This is not deployment acceptance: no image from this native repository has been published or deployed to `goreecloud-vps-01`, no live provider credential has been exercised there, and no user-facing Glaze UI, production rate/abuse controls, Identity acceptance, SearXNG-derived dependency retirement, or Stable qualification is established.

## Development use

Requires Python 3.11 or newer.

```bash
python -m pip install -e .
goreecloud-search parse 'privacy "search engine" site:example.com category:docs'
python -m unittest discover -s tests -v
```

The Docker candidate is built from the repository root. Its default command is `serve-container --port 8080 --service-hostname search.goreecloud.com`. The container listener is intentionally separate from `serve`, which remains fixed to `127.0.0.1`. The repository Compose example under `deploy/vps/` is a service candidate only; it must not overwrite the live `/srv/docker/stacks/searxng/docker-compose.yml` without authoritative readback and controlled reconciliation.

## Architecture direction

```text
Browser / AI / API clients
          |
          v
   GoreeCloud Search
   - query parser
   - source planner / privacy budget
   - provider executor
   - normalization / deduplication
   - content policy
   - baseline ranking
   - local Lens reranking
          |
     +----+----+
     |         |
     v         v
GoreeCloud   optional
  Index      providers
```

## Privacy and policy boundaries

The Brave adapter is opt-in and external: when it is selected by the source plan, the minimized provider-facing query is disclosed to Brave. Search-local `lens:` and `source:` controls are not forwarded in the Brave query string, the API credential is sent only to the fixed Brave HTTPS endpoint, and redirects are refused. The planner can cap third-party query recipients before execution. The executor runs only providers admitted by that plan. Content-policy hooks then evaluate normalized results before ranking. Non-Off SafeSearch intent fails before provider execution if no configured hook can enforce it, preventing a falsely protected state from disclosing the query first.

Lenses are a Search-local ranking layer in this candidate. A requested `lens:` value is resolved before provider execution and then removed from the query passed to provider adapters. Portable Lens export/import is a pure local data transformation; this repository does not upload, publish, synchronize, or remotely fetch Lens documents.

The built-in `DomainPolicyHook` enforces explicit administrator domain policy only. It does not classify adult content, malware, phishing, misinformation, or any other semantic/safety category. Those classifications require separately verified sources and platform integrations.

## License

GoreeCloud-owned code in this repository is licensed under the **GNU Affero General Public License v3.0 or later (`AGPL-3.0-or-later`)**, unless a specific file or third-party component states otherwise.

This license was selected specifically for GoreeCloud Search because Search is intended to be a self-hostable, network-interactive service and GoreeCloud wants modified hosted versions to preserve reciprocal source availability. See `LICENSE`, `LICENSING.md`, and `NOTICE`.

Third-party material remains governed by its own applicable license terms.

## Status integrity

A branch, pull request, passing CI run, configuration declaration, or documented plan does not mean a feature is released, deployed, production-accepted, or Stable.
