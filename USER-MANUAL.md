# GoreeCloud Search — User Manual

## Current status

GoreeCloud Search is in Development. Version `0.1.0.dev4` is a developer-facing foundation and is not a complete search engine release.

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

## Current limitations

Authenticated live search transport, Browser integration, approved external-provider federation, Private View, AI answers, history, synchronization, and production administration are not implemented yet. The developer core now includes provider execution orchestration and Index pagination, but no authenticated live Index transport or approved external network adapter ships with it.
