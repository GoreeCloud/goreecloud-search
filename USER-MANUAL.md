# GoreeCloud Search — User Manual

## Current status

GoreeCloud Search is in Development. Version `0.1.0.dev5` is a developer-facing foundation and is not a complete search engine release.

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

## Developer privacy control

The Python API can apply `QueryDisclosureBudget(max_third_party_providers=N)` when planning or executing a search. The budget is enforced before provider execution. This is a development API capability; there is not yet a user-facing preferences interface for it.

## Current limitations

Authenticated live search transport, Browser integration, approved external-provider federation, Private View, AI answers, history, synchronization, and production administration are not implemented yet. The developer core includes provider execution orchestration, Index pagination, and disclosure budgeting, but no authenticated live Index transport or approved external network adapter ships with it.
