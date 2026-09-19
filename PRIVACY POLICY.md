# GoreeCloud Search — Privacy Policy

## Current development implementation

Version `0.1.0.dev3` contains a local query parser, source planner, versioned Search ↔ Index contract models, a transport-injected Index adapter boundary, and local result normalization/deduplication. The repository code in this version:

- Does not send search queries over the network. The Index adapter has no built-in network transport.
- Does not contact third-party search providers.
- Does not persist search history.
- Does not contain advertising or tracking code.
- Does not build behavioral profiles.
- Does not require third-party analytics.
- Does not use click history, cross-query behavioral profiles, advertising identifiers, or paid placement for the implemented ranking baseline.

## Source planning

The source planner marks whether an execution plan would disclose a query to a third-party provider.

`goreecloud_only` and `offline_local` are required to fail closed against third-party query disclosure.

## Future network features

Any future provider execution, remote suggestion, synchronization, account storage, analytics, AI, Browser, Index, or other network-capable feature must update this policy to match verified behavior and must apply the relevant GoreeCloud Identity, Privacy Shield, Wardveil Security, retention, consent, minimization, and disclosure controls before production acceptance.

## Development warning

This document describes the current repository implementation. It does not claim that a future hosted deployment has the same data flows until that deployment is separately verified.
