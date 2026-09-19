# GoreeCloud Search — User Manual

## Current status

GoreeCloud Search is in Development. Version `0.1.0.dev6` is a developer-facing foundation, not a complete search-engine release.

## Install for development

```bash
python -m pip install -e .
```

Python 3.11 or newer is required.

## Parse a query

```bash
goreecloud-search parse 'privacy "search engine" site:example.com category:docs'
```

The CLI prints normalized query JSON and performs no network access.

## Developer privacy control

The Python API can apply `QueryDisclosureBudget(max_third_party_providers=N)` during planning or search execution. The cap is enforced before provider execution.

## Developer content-policy controls

The Python API supports `SafeSearchMode.OFF`, `MODERATE`, and `STRICT` plus content-policy hooks. `DomainPolicyHook` enforces explicit domain allowlists/blocklists. Moderate or Strict fails before provider execution unless at least one configured hook declares actual SafeSearch enforcement. This candidate does not ship a production safety classifier.

## Current limitations

Authenticated live search transport, production SafeSearch classification, Wardveil safety feeds, Browser integration, approved external-provider federation, Private View, AI answers, history, synchronization, and production administration are not implemented yet.
