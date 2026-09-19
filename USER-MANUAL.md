# GoreeCloud Search — User Manual

## Current status

GoreeCloud Search is in Development. Version `0.1.0.dev1` is a developer-facing foundation, not a complete search-engine release.

## Install for development

```bash
python -m pip install -e .
```

Python 3.11 or newer is required.

## Parse a query

```bash
goreecloud-search parse 'privacy "search engine" site:example.com category:docs'
```

The command prints normalized JSON and performs no network access.

## Supported operators

`site:`, `-domain:`, `filetype:`, `ext:`, `before:`, `after:`, `language:`, `region:`, `source:`, `category:`, `lens:`, quoted phrases, and `-excluded-term`.

## Current limitations

Live search, Browser/Index integration, provider federation, Private View, AI answers, history, synchronization, and production administration are not implemented yet.
