# GoreeCloud Search Capability Evidence

## Purpose

The native Search status contract exposes a minimized producer-owned description of the `search.query` capability so first-party consumers can reason about the exact interface without inferring acceptance from service reachability alone.

## Current capability

The current source advertises:

- capability ID: `search.query`;
- contract version: `1`;
- endpoint: `/api/v1/search`;
- authority: GoreeCloud Search;
- authoritative: `true`;
- current: `true`;
- production accepted: `false`.

The evidence is additive to the existing status response and does not contain query text, result content, user content, credentials, authorization headers, or producer runtime errors.

## Authority and acceptance

Search is authoritative for its query capability and contract identity. That authority does not authorize a consuming application to treat the current pre-Stable service as production accepted.

`authoritative: true` means the capability statement is produced by the service that owns the query contract. `current: true` means it describes the current native source contract. Neither field is equivalent to deployment acceptance, Stable qualification, or production approval.

`production_accepted` therefore remains `false` until the applicable deployment, platform-integration, security, privacy, continuity, release, and acceptance authorities provide the required evidence for the selected runtime.

## Consumer behavior

A first-party consumer such as GoreeCloud Browser may require the exact capability ID and contract version and may fail closed when production acceptance is absent. Search does not weaken that consumer policy and does not manufacture consumer authorization.

## Relationship to platform status

This capability evidence is separate from `/api/v1/platform/status`. The platform-status endpoint projects bounded Privacy Shield, Wardveil Security, and Everkeep evidence. The `search.query` capability evidence describes Search's own query interface. Neither contract converts the other into production approval.
