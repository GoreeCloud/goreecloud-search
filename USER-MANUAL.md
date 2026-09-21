# GoreeCloud Search — User Manual

## Current status

GoreeCloud Search is in Development. Version `0.1.0.dev14` is a developer-facing foundation and is not a complete search engine release.

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

The repository now includes a developer-facing authenticated Index HTTP source boundary, but it is not a user-facing search surface and is not production accepted. It requires injected GoreeCloud Identity and Privacy Shield verifier implementations and an eligible external provider before query execution can succeed. Browser integration, approved external-provider federation, live authority-service connectivity, Private View, AI answers, history, synchronization, production administration, deployment, and Stable qualification remain unimplemented or unaccepted.
