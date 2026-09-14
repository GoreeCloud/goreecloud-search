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
- **Privacy Shield** remains authoritative for whether a query may cross the local/remote boundary and for any applicable purpose, minimization, retention, destination, expiry, revocation, and replay constraints.
- **Wardveil Security** remains authoritative for applicable trust/security evidence. Search, Index, and Browser must not manufacture Wardveil status locally.

## Query capability

The initial interoperable capability is:

- Capability ID: `search.query`
- Contract version: `1`
- Discovery endpoint: `/api/v1/status`
- Discovery collection: `capability_evidence`
- Query endpoint: `/api/v1/search`
- Supported machine methods: `POST`, `GET`
- Preferred first-party method: `POST`
- Preferred query transport: `json_body`
- Request media type: `application/json`
- Response media type: `application/json`
- Privacy authorization required: `true`
- Privacy authorization scheme: `privacy_shield_capability_token_reference`
- Privacy authorization header: `X-GoreeCloud-Privacy-Capability`
- Authenticated requester required: `true`
- Maximum request body: `16384` bytes
- Maximum result count: `100`
- Required request fields: normalized query, category, bounded result limit
- Default category: `general`

The machine-readable capability evidence publishes these transport requirements explicitly so each consumer can independently validate discovery, authorization, requester-authentication expectations, and request shape before delegating a query. The presence of an endpoint alone is never sufficient authorization.

The current Development service explicitly advertises Privacy Shield authorization enforcement as `not_enforced_development`. This is intentional evidence of a remaining runtime boundary, not permission to send production traffic. Stable/production consumers must require an accepted enforcement state such as `required`, `authenticated_requester_required=true`, and production acceptance before they may use the advertised authorization header.

`POST` is preferred for first-party remote delegation so query text does not have to appear in the request URL. The native Development service accepts a bounded JSON body and rejects unknown fields, multiple JSON objects, and structurally invalid bodies. `GET` remains an additive Development/compatibility surface; its continued availability does not authorize production consumers to prefer query-bearing URLs. Required Privacy Shield enforcement mode is POST-only and JSON-only even if a capability reference is otherwise valid.

A consumer must fail closed if the advertised capability is missing, duplicated/ambiguous, stale, non-authoritative, version-incompatible, structurally invalid, does not independently prove the expected Privacy Shield authorization transport, or does not require authenticated requester identity.

Development consumers may explicitly opt into non-production or legacy-GET capability evidence only in Development builds and only when that exception is visible in repository-local acceptance evidence. Stable/production consumers must require production-accepted capability evidence, enforced Privacy Shield authorization, authenticated requester identity, and the compatible POST + JSON-body contract.

## Privacy Shield authorization transport

The preferred production model is for authorization to travel with the delegated operation as an opaque Privacy Shield capability reference, rather than as a copied policy decision, signed bearer token, or generic identity claim.

A production consumer must create a unique Privacy Shield `request_id` before authorization, submit the canonical request for resource `goreecloud.search.query`, operation `search.query`, purpose `internet_search`, processing zone `private_goreecloud`, destination `https://search.goreecloud.com`, and retention `none`, and require the returned decision to correlate to that same request. A compatible-looking decision issued for another request is not valid authorization.

Consumers must also ensure that the decision permits the expected processing zone, destination, retention behavior, and supported obligations before transmitting the request. Consumers that cannot enforce returned obligations must reject the decision rather than silently treating it as an unconditional allow.

Only a canonical opaque `psc_*` capability reference should cross the Search boundary through `X-GoreeCloud-Privacy-Capability`. Raw signed tokens, Privacy Shield signing keys, policy documents, browsing state, local Index data, or unrelated evidence must not be attached. Browser and Index also validate the canonical reference at their final Search handoff rather than relying only on an earlier authorization adapter.

Each production Search operation should acquire fresh authorization rather than treating a `psc_*` reference as persistent application state. Index explicitly reacquires authorization for every delegated query. Search's authority-side verification request uses consume semantics so Privacy Shield can enforce single-use/replay policy centrally.

## Privacy Shield reference verification contract

Search verifies opaque references through the authority-owned Privacy Shield capability-reference verification contract rather than interpreting capability tokens itself.

The initial verification envelope is version `1`, defined authoritatively by Privacy Shield at:

```text
contracts/privacy-shield.capability-reference-verification.schema.json
```

Search's v1 request contains only:

- `contract_version: 1`;
- authenticated verification consumer `goreecloud-search`;
- the opaque `capability_reference`;
- exact expected requester ID;
- resource ID;
- purpose;
- operation;
- processing zone;
- destination;
- retention mode;
- `consume: true` for one Search execution.

The Go request/response models pin these JSON field names explicitly. Unsupported verification protocol versions must fail closed before authority lookup or replay-state mutation. Verification-protocol versioning is independent from capability-token signing/key format.

The Privacy Shield verification service returns only contract version, authorization state, the opaque reference, and enforceable constraints. It does not return the signed capability token, raw capability claims, or signing keys.

The verification service's `consumer_id` is not self-authenticating request data. A concrete local/runtime transport must authenticate Search and inject/derive that identity before invoking Privacy Shield. Separately, Search must resolve the original delegated requester identity—such as GoreeCloud Browser or GoreeCloud Index—from an authenticated runtime/transport boundary rather than trusting an arbitrary requester HTTP header.

## Search-side authorization gate

Search contains a request-handler authorization gate. In required mode, the gate executes before Search request parsing or engine execution and:

- requires a real Privacy Shield verifier;
- requires an authenticated requester-identity resolver;
- rejects missing, blank, ambiguous, malformed, whitespace-bearing, or oversized capability references;
- requires POST and `application/json`;
- binds the verifier request to the authenticated requester identity plus exact resource, operation, purpose, processing zone, destination, and retention context;
- sanitizes authority rejection details from HTTP responses;
- refuses to claim readiness/enforcement when required dependencies are absent.

The Development runtime deliberately instantiates this gate with enforcement disabled because a real authenticated Privacy Shield verification transport and requester resolver are not yet connected. Search therefore continues to advertise `not_enforced_development`; the existence of the gate and v1 client model is not production verification evidence and must not be used to change that state prematurely.

## Data minimization

The delegated request must not include unrelated local state. In particular, the Search query boundary must not receive:

- local Index results;
- installed-application inventory;
- contacts, calendar, file, media, or clipboard content;
- Browser history or open-tab inventory unless a separately authorized feature explicitly requires it;
- GoreeCloud Identity identifiers merely because Identity was used to authenticate the caller;
- raw Privacy Shield or Wardveil internal evidence payloads beyond the minimum opaque reference required by accepted transport contracts.

## Result contract

Search results returned to Index or Browser must contain only fields required by the consuming experience, including a title, canonical HTTP(S) URL, optional snippet, and bounded Search-owned ranking metadata.

Consumers must independently validate executable destinations. HTTP(S) results containing invalid hosts, unsupported schemes, embedded user-info credentials, malformed URLs, or other disallowed destination forms must not become executable actions.

The machine response carries API version identity in both the `X-GoreeCloud-API-Version` header and the successful response body so consumers can reject incompatible or mismatched contracts.

## Degradation

Partial provider failure inside Search may produce a degraded response if Search can still return valid results. Consumers may preserve those results while presenting degraded availability. Degraded status must never be transformed into a healthy/fully-available claim.

A consumer may also apply its own bounded result cap and independently reject unsafe result actions without suppressing valid siblings. Consumer-side degradation and invalid-result evidence remain authoritative for that consumer even when Search itself reports a successful response.

## Cancellation and replacement

Interactive consumers must cancel superseded query work when the user replaces the active query. Search treats caller cancellation or a caller-owned deadline as operation cancellation, not provider failure: it returns without manufacturing partial provider-timeout evidence and the HTTP layer does not render or encode a synthetic Search error after the request context is canceled.

Search's own bounded engine deadline remains a separate concern. When that Search-owned deadline expires while the caller is still active, valid results from providers that completed in time may be preserved and unresolved providers are represented as bounded `provider_timeout` degradation. This distinction prevents user-driven cancellation from corrupting provider-health evidence.

## Browser behavior

Browser omnibox classification must remain explicit:

1. Valid navigable URL / accepted navigation intent → Browser navigation path.
2. Non-URL query → local Search intent only; no query-bearing remote URL is constructed.
3. One unambiguous compatible Search capability + correlated accepted Privacy Shield authorization + canonical `psc_*` reference + authenticated-requester requirement → a separate transport adapter may construct the bounded POST/JSON request.
4. Ambiguous, unsafe, unauthorized, expired, constrained-but-unsupported, request-mismatched, unauthenticated, noncanonical-reference, or incompatible input → no remote execution.

The classifier must not encode free-text queries into a `?q=` destination as an intermediate convenience. Query text remains transport-neutral until the authorization and capability gates have independently accepted the operation.

Browser may open Search result URLs, but Search does not gain tab/session authority.

## Index behavior

Index may dispatch Search concurrently with eligible local providers only when:

- remote processing is allowed for the current query;
- Search is in the exact provider allowlist;
- applicable Privacy Shield authority evidence is present and enforceable;
- the advertised Search capability passes compatibility checks.

Production-mode Index consumers must additionally require production acceptance plus the exact discovery, POST, `json_body`, media-type, privacy-authorization scheme/header/enforcement, authenticated-requester, and request-size evidence published by the Search capability. Before calling the Search client, Index creates a unique Privacy Shield request, accepts only the matching decision, obtains a canonical capability reference, and carries only that reference with the delegated request.

Index acquires fresh authorization for each production Search query. It must not cache a previous `psc_*` reference as application/provider state.

Development mode may retain an explicit compatibility exception for older evidence without converting that exception into production acceptance.

Local-only mode must not preflight or call Search. Index must not compare Search-owned raw score magnitudes against unrelated local-provider score scales; source ordering may be retained only through an explicit bounded normalization/tie-break contract.

## Glaze UI

All Search-owned user-facing surfaces must track the latest approved Stable Glaze UI release. Browser and Index remain independently responsible for their own Glaze adoption and acceptance; Search capability availability must not be used as design-system conformance evidence for another product.

## Stability rule

No Search–Index–Browser integration may be described as Stable merely because the API compiles, CI passes, or a source adapter exists. Stable acceptance requires current contracts, supported runtime evidence, a real authenticated Privacy Shield reference-verification transport, authenticated requester identity, accessibility, error/degradation behavior, cancellation behavior, and repository-local release acceptance for every participating product.
