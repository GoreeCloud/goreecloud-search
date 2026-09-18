# GoreeCloud Search

GoreeCloud Search is the privacy-first, self-hostable search and information-discovery service for the GoreeCloud ecosystem.

> **Lifecycle:** Development  
> **Version:** `0.1.0.dev13`  
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

The repository now ships one opt-in external network provider, a loopback-only HTTP/API and server-rendered web UI, OpenSearch discovery, and a non-root Zorin/systemd user-service profile for Development use. The API is not a production service boundary: it has no GoreeCloud Identity integration, no production rate/abuse controls, no user-facing Glaze UI, and no deployment acceptance. The web surface targets current Stable Glaze UI 1.5.1 semantics but has not completed repository-local rendered/human/target-runtime acceptance and therefore does not claim Glaze conformance. The laptop profile does not stop, disable, remove, or reconfigure SearXNG; retirement remains a separate target-host cutover step after live-provider, browser/UI, monitoring, rollback, and replacement acceptance. It does not make Brave a production-accepted provider, add an authenticated live GoreeCloud Index transport, establish a production SafeSearch classifier or Wardveil feed, or establish Stable qualification.

## Development use

Requires Python 3.10 or newer.

```bash
python -m pip install -e .
goreecloud-search parse 'privacy "search engine" site:example.com category:docs'
python -m unittest discover -s tests -v
```

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
