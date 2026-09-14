# GoreeCloud Search — Index and Browser Integration Contract

**Status:** Development contract  
**Scope:** GoreeCloud Search, GoreeCloud Index, and GoreeCloud Browser  
**Authority:** Search remains authoritative for Internet/web/current-information retrieval.

## Purpose

This contract defines how GoreeCloud Browser and GoreeCloud Index may delegate Internet search to GoreeCloud Search without collapsing product authority boundaries, leaking unrelated local state, or treating service availability as blanket authorization.

## Product authority boundaries

- **GoreeCloud Search** owns Internet/web/current-information query execution, provider orchestration, result normalization, Search preferences, and Search-owned research experiences.
- **GoreeCloud Index** owns universal/local federated discovery and result composition. It may include Search as an explicitly authorized remote provider, but it does not become the Internet-search authority.
- **GoreeCloud Browser** owns URL navigation, browser chrome, tabs, page lifecycle, and web-result opening. Non-URL omnibox input delegates search authority to GoreeCloud Search rather than implementing a parallel search engine.
- **Privacy Shield** remains authoritative for whether a query may cross the local/remote boundary and for any applicable purpose, minimization, retention, or destination constraints.
- **Wardveil Security** remains authoritative for applicable trust/security evidence. Search, Index, and Browser must not manufacture Wardveil status locally.

## Query capability

The initial interoperable capability is:

- Capability ID: `search.query`
- Contract version: `1`
- Endpoint: `/api/v1/search`
- Supported machine methods: `POST`, `GET`
- Preferred first-party method: `POST`
- POST media type: `application/json`
- Required request fields: normalized query, category, bounded result limit
- Default category: `general`

`POST` is preferred for first-party remote delegation so query text does not have to appear in the request URL. The native Development service accepts a bounded JSON body and rejects unknown fields, multiple JSON objects, and structurally invalid bodies. `GET` remains an additive Development/compatibility surface; its continued availability does not authorize production consumers to prefer query-bearing URLs.

A consumer must fail closed if the advertised capability is missing, duplicated/ambiguous, stale, non-authoritative, version-incompatible, or structurally invalid.

Development consumers may explicitly opt into non-production or legacy-GET capability evidence only in Development builds and only when that exception is visible in repository-local acceptance evidence. Stable/production consumers must require production-accepted capability evidence and a compatible POST-capable transport.

## Data minimization

The delegated request must not include unrelated local state. In particular, the Search query boundary must not receive:

- local Index results;
- installed-application inventory;
- contacts, calendar, file, media, or clipboard content;
- Browser history or open-tab inventory unless a separately authorized feature explicitly requires it;
- GoreeCloud Identity identifiers merely because Identity was used to authorize the operation;
- raw Privacy Shield or Wardveil internal evidence payloads beyond the minimum reference/token required by their accepted transport contracts.

## Result contract

Search results returned to Index or Browser must contain only fields required by the consuming experience, including a title, canonical HTTP(S) URL, optional snippet, and bounded Search-owned ranking metadata.

Consumers must independently validate executable destinations. HTTP(S) results containing invalid hosts, unsupported schemes, embedded user-info credentials, malformed URLs, or other disallowed destination forms must not become executable actions.

The machine response carries API version identity in both the `X-GoreeCloud-API-Version` header and the successful response body so consumers can reject incompatible or mismatched contracts.

## Degradation

Partial provider failure inside Search may produce a degraded response if Search can still return valid results. Consumers may preserve those results while presenting degraded availability. Degraded status must never be transformed into a healthy/fully-available claim.

## Cancellation and replacement

Interactive consumers must cancel superseded query work when the user replaces the active query. Search should treat client cancellation as cancellation rather than a provider failure and avoid unnecessary downstream work where possible.

## Browser behavior

Browser omnibox classification must remain explicit:

1. Valid navigable URL / accepted navigation intent → Browser navigation path.
2. Non-URL query → Search delegation path.
3. Ambiguous or unsafe input → no silent remote execution until classification and applicable Privacy Shield authorization succeed.

Browser may open Search result URLs, but Search does not gain tab/session authority.

## Index behavior

Index may dispatch Search concurrently with eligible local providers only when:

- remote processing is allowed for the current query;
- Search is in the exact provider allowlist;
- applicable Privacy Shield authority evidence is present and enforceable;
- the advertised Search capability passes compatibility checks.

Local-only mode must not preflight or call Search. Index must not compare Search-owned raw score magnitudes against unrelated local-provider score scales; source ordering may be retained only through an explicit bounded normalization/tie-break contract.

## Glaze UI

All Search-owned user-facing surfaces must track the latest approved Stable Glaze UI release. Browser and Index remain independently responsible for their own Glaze adoption and acceptance; Search capability availability must not be used as design-system conformance evidence for another product.

## Stability rule

No Search–Index–Browser integration may be described as Stable merely because the API compiles, CI passes, or a source adapter exists. Stable acceptance requires current contracts, supported runtime evidence, privacy/security enforcement, accessibility, error/degradation behavior, and repository-local release acceptance for every participating product.
