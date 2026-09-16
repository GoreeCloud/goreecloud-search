# GoreeCloud Search — User Manual

## Current status

GoreeCloud Search is in Development. Version `0.1.0.dev7` is a developer-facing foundation and is not a complete search engine release.

## Install for development

```bash
python -m pip install -e .
```

Python 3.11 or newer is required.

## Parse a query

```bash
goreecloud-search parse 'privacy "search engine" site:example.com category:docs'
```

The command prints a normalized JSON representation of the query. It does not contact the Internet or any search provider.

## Supported operators

- `site:example.com`
- `-domain:example.com`
- `filetype:pdf`
- `ext:pdf`
- `before:2026-09-16`
- `after:2026-01-01`
- `language:en`
- `region:US`
- `source:index`
- `category:news`
- `lens:official`
- `"quoted phrase"`
- `-excluded-term`

## Developer Lens support

The Python API can register `Lens` objects with domain, filetype, and language rules using boost, lower, or exclude actions. A query can select a configured Lens with `lens:<name>`. Lens selection is Search-local in this candidate and is removed before provider execution. Unknown Lens names fail before provider calls.

There is not yet a user-facing Lens editor, persistence format, import/export, sharing, synchronization, or remote Lens registry.

## Developer content-policy controls

The Python API supports `SafeSearchMode.OFF`, `MODERATE`, and `STRICT`, plus content-policy hooks. The built-in `DomainPolicyHook` can enforce explicit domain allowlists/blocklists. Moderate or Strict SafeSearch fails before provider execution unless at least one configured hook declares actual SafeSearch enforcement; this candidate does not ship a production safety classifier.

## Developer privacy control

The Python API can apply `QueryDisclosureBudget(max_third_party_providers=N)` when planning or executing a search. The budget is enforced before provider execution. There is not yet a user-facing preferences interface for it.

## Current limitations

Authenticated live search transport, Browser integration, approved external-provider federation, Private View, AI answers, history, synchronization, production administration, and production deployment are not implemented yet.
