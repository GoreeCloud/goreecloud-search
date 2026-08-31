# GoreeCloud Search User Manual

## Current availability

GoreeCloud Search is in native migration and is not Stable. Depending on the environment, users may encounter the transitional Search service, native development surfaces, or controlled test environments. A source capability described in this manual is not automatically available on every deployment.

## Searching

For a Search environment exposed to you:

1. Open GoreeCloud Search directly or use a GoreeCloud Browser search entry point configured for GoreeCloud Search.
2. Enter a query.
3. Choose an available category when the current environment supports it.
4. Review result titles, destination URLs, source/provider information, and any degraded-provider status shown by the interface.

The native source defines General, Images, Videos, News, and Files categories. General has a development execution path even with no configured providers. Other categories remain unavailable unless the deployed native provider layer has an approved executable adapter for that category.

## Search authority and fallback behavior

GoreeCloud Search is intended to be the configured GoreeCloud search authority. Browser searches should not silently fall back to an unrelated external search provider when Search is unavailable.

Direct website navigation in the Browser is separate from Search and may continue to work during a Search outage.

## Result safety and privacy

The native engine accepts only HTTP and HTTPS result destinations. URL fragments are removed during normalization, and result URLs containing embedded username/password-style user information are rejected before entering the result set.

Provider failures are represented with bounded status values such as unavailable or timeout. Raw provider error messages are not intended to be shown to users because those errors may contain operational or secret information.

External search providers may still observe requests originating from GoreeCloud infrastructure. Using GoreeCloud Search improves control and reduces direct Browser-to-provider coupling; it is not a claim of anonymity from external providers.

## Preferences

The native Search implementation contains first-party preference foundations. Preference availability depends on the deployed native experience and accepted release scope.

Where available, use Preferences to control supported Search behavior such as appearance or search choices. Controls must not be assumed to synchronize to an account unless that environment explicitly provides accepted GoreeCloud Identity and Sync integration.

## Search history and synchronization

Native source contains an application-owned `search.history` GoreeCloud Sync contract. This includes bounded records/cursors, authenticated submission, exact schema validation, payload-free deletion tombstones, and signed-record preflight.

This is not a guarantee that account history synchronization is enabled in your environment. If history synchronization is not explicitly presented as available and accepted, treat search history as local/environment-specific.

## Privacy expectations

GoreeCloud Search is designed around these user-facing expectations:

- no GoreeCloud advertising or sponsored-result placement;
- no behavioral-profiling business model;
- minimized query/history retention;
- no hidden external search fallback bypassing the configured Search authority;
- visible, understandable degraded-provider behavior where the interface exposes it.

Privacy Shield remains the authority for implemented data-use controls. Do not interpret a planned privacy control as active unless it appears in the current application and has corresponding release/runtime evidence.

## Accessibility and Glaze UI

GoreeCloud-owned Search surfaces are required to follow the latest approved Stable Glaze UI contract for production acceptance, including applicable keyboard, focus, contrast, reduced-motion, reduced-transparency, text scaling, and adaptive-layout behavior.

Native source or CI success alone is not accessibility acceptance for every device or deployment.

## Troubleshooting

### No results appear

The query may have no matching results, no native providers may be configured for the current environment, or configured providers may be degraded. Check provider/degraded status if the interface exposes it.

### A category is unavailable

The native engine fails closed for specialized categories unless an executable provider adapter is configured. Use an available category or contact the GoreeCloud administrator responsible for that environment.

### Search is unavailable but websites still open

This is expected separation between Search and direct Browser navigation. Search outages should not require the Browser to redirect searches to an unrelated external engine.

### Preferences or history sync are missing

Those capabilities depend on the active native release and its accepted Identity/Sync integration. Their source foundations do not imply that they are enabled in the current runtime.

## Data recovery and migration

Everkeep is the GoreeCloud authority for backup, restore, rollback, preservation, and migration. Search is still transitioning away from the inherited runtime, so migration and recovery behavior must be evaluated against the exact environment and release rather than inferred from source code.

## Security reporting

Do not place provider credentials, private tokens, or sensitive search data in public issue reports. Use the repository's documented security-reporting path for security-sensitive findings.

## Current limitations

The native application is not yet a complete Stable replacement for the transitional runtime. Production-approved provider adapters, full native feature parity, complete platform-system acceptance, deployment/recovery evidence, and controlled migration/cutover remain work in progress.
