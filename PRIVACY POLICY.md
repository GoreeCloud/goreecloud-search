# GoreeCloud Search — Privacy Policy

## Current development implementation

Version `0.1.0.dev1` contains a local query parser and source planner. It:

- Does not send search queries over the network.
- Does not contact third-party search providers.
- Does not persist search history.
- Does not contain advertising or tracking code.
- Does not build behavioral profiles.
- Does not require third-party analytics.

## Source planning

The source planner marks whether a plan would disclose a query to a third-party provider. `goreecloud_only` and `offline_local` fail closed against third-party query disclosure.

## Future network features

Future provider execution, suggestions, synchronization, account storage, analytics, AI, Browser, Index, or other network features must update this policy to match verified behavior and apply relevant Identity, Privacy Shield, Wardveil, retention, consent, minimization, and disclosure controls before production acceptance.

This document describes repository behavior only; it does not establish the behavior of any future hosted deployment until separately verified.
