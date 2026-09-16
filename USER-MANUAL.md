# GoreeCloud Search — User Manual

## Current status

GoreeCloud Search is in Development. Version `0.1.0.dev8` is a developer-facing foundation and is not a complete search engine release.

## Install for development

```bash
python -m pip install -e .
```

Python 3.11 or newer is required.

## Parse a query

```bash
goreecloud-search parse 'privacy "search engine" site:example.com category:docs'
```

The command prints normalized query JSON and performs no network request.

## Supported operators

`site:`, `-domain:`, `filetype:`, `ext:`, `before:`, `after:`, `language:`, `region:`, `source:`, `category:`, `lens:`, quoted phrases, and excluded terms.

## Developer Lens support

The Python API can register `Lens` objects with domain, filetype, and language rules using boost, lower, or exclude actions. `lens:<name>` selects a configured Lens. Lens selection is Search-local and is removed before provider execution. Unknown Lens names fail before provider calls.

Developers can use `export_lens(lens)` and `import_lens(text)` for the versioned `goreecloud.search-lens.v1` portable JSON representation. These APIs return/accept text only; they do not read/write files or send data over the network. Untrusted Lens documents are strictly validated before becoming typed configuration.

There is not yet a user-facing Lens editor, automatic file persistence, sharing/discovery service, signature/trust workflow, synchronization, or remote Lens registry.

## Developer content-policy controls

The Python API supports `SafeSearchMode.OFF`, `MODERATE`, and `STRICT`, plus content-policy hooks. Moderate/Strict fails before provider execution unless a configured hook actually enforces it. The built-in domain hook is not a production safety classifier.

## Developer privacy control

`QueryDisclosureBudget(max_third_party_providers=N)` can limit third-party query recipients before provider execution.

## Current limitations

Authenticated live search transport, Browser integration, approved external federation, Private View, AI answers, history, synchronization, production administration, and production deployment are not implemented yet.
