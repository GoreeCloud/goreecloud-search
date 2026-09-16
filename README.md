# GoreeCloud Search

GoreeCloud Search is the privacy-first, self-hostable search and information-discovery service for the GoreeCloud ecosystem.

> **Lifecycle:** Development  
> **Version:** `0.1.0.dev6`  
> **Current scope:** Native query parsing, privacy-aware source planning, bounded provider execution, query-disclosure budgeting, content-policy hooks, a versioned GoreeCloud Index contract with pagination, Search-owned normalization/deduplication, and deterministic explainable ranking. No authenticated live provider transport or production safety classifier ships in this revision.

## What exists now

- Typed query model and parser for the initial operator set.
- Deterministic source modes and provider eligibility planning.
- Optional per-query third-party disclosure budgets enforced before execution.
- Replaceable provider contracts plus bounded asynchronous execution, timeouts, cancellation, failure isolation, fallback behavior, and explicit availability state.
- Versioned Search ↔ GoreeCloud Index v1 models, transport-injected adapter boundary, cursor pagination, and degraded/warning propagation.
- Conservative URL canonicalization, deduplication, source agreement, and provenance.
- Local content-policy hooks applied after normalization and before ranking.
- Typed SafeSearch intent modes: Off, Moderate, and Strict.
- Fail-closed pre-execution rejection when Moderate/Strict is requested but no configured hook can enforce it.
- Built-in administrator domain allowlist/blocklist policy with subdomain matching. This domain hook is not a content or threat classifier.
- Deterministic ranking with inspectable signals and “Why this result?” explanations.
- Unit tests and pull-request CI.

The repository ships no authenticated live network provider, no production SafeSearch classifier, no Wardveil safety feed, and no user-facing UI. The development CLI performs no network access.

## Development use

Requires Python 3.11 or newer.

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
   - ranking / explanations
          |
     +----+----+
     |         |
     v         v
GoreeCloud   optional
  Index      providers
```

## Privacy and policy boundaries

The planner can cap third-party query recipients before execution. The executor runs only providers admitted by that plan. Content-policy hooks then evaluate normalized results before ranking. Non-Off SafeSearch intent fails before provider execution if no configured hook can enforce it, preventing a falsely protected state from disclosing the query first.

The built-in `DomainPolicyHook` enforces explicit administrator domain policy only. It does not classify adult content, malware, phishing, misinformation, or any other semantic/safety category. Those classifications require separately verified sources and platform integrations.

## Status integrity

A branch, pull request, passing CI run, configuration declaration, or documented plan does not mean a feature is released, deployed, production-accepted, or Stable.
