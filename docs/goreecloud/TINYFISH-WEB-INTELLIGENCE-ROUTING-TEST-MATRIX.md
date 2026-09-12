# GoreeCloud Search — TinyFish Web Intelligence Routing Test Matrix

## Status

Development verification companion for FR-008. This file records source-level expectations for the provider-neutral routing and budget controller. It is not production acceptance evidence.

## Required source behaviors

The `native/internal/webintelligence` controller must demonstrate all of the following through automated tests and normal repository CI:

| Area | Required behavior |
| --- | --- |
| Capability boundary | A plan returns only providers implementing the exact requested capability. |
| Free preference | Free providers are ordered before metered providers within one capability. |
| Metered opt-in | Metered providers are excluded unless the caller explicitly permits them. |
| Budget prerequisite | Any plan containing a metered attempt requires a configured total budget, per-operation budget, and non-zero estimated cost. |
| Per-operation ceiling | The conservative reservation for all planned metered attempts cannot exceed the per-operation ceiling. |
| Total ceiling | New reservations cannot exceed total remaining budget after recorded spend and outstanding reservations. |
| Accounting | Committing a reservation records actual spend; releasing a reservation returns unused capacity. |
| Overrun evidence | Actual cost above the reservation is still recorded and returns a budget-overrun error. |
| Health | Unavailable providers are excluded; degraded providers remain eligible. |
| Fallback | Fallback is limited to the same capability and the configured attempt bound. |
| Observability | Observations aggregate normalized outcome count and duration only. |
| Privacy minimization | The observation model contains no query, URL, page-content, citation, credential, or provider-body field. |
| Fail closed | Invalid capabilities, provider IDs, duplicate providers, malformed budgets, unknown reservations, and unreserved metered costs return stable errors. |

## CI boundary

Passing source tests demonstrates only that the Development implementation behaves as encoded in this repository revision. It does not demonstrate live TinyFish billing accuracy, provider-wallet reconciliation, production credentials, multi-instance durability, Manager UI acceptance, runtime Privacy Shield authorization, Wardveil acceptance, Everkeep recovery, or production release qualification.
