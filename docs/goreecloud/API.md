# GoreeCloud Search API Boundary

## Status

Design boundary only. No GoreeCloud-facing machine API is enabled by this document.

## Purpose

GoreeCloud Search will eventually provide a stable, documented interface for approved GoreeCloud applications and local research workflows. Consumers should depend on a GoreeCloud-owned contract rather than on incidental SearXNG internals.

## Initial consumers

Potential consumers include GoreeCloud Research Library, GoreeCloud Manager, browser integrations, local AI systems, research agents, and approved automation workflows.

## Contract principles

The future interface must:

- preserve a versioned or otherwise governed response contract;
- normalize provider-specific result details where practical;
- expose source and engine provenance needed for research and debugging;
- distinguish partial-provider failure from complete request failure;
- avoid exposing secrets or administrative configuration;
- apply bounded request limits and operational safeguards;
- avoid retaining query history unless separately approved;
- remain replaceable if SearXNG is later reduced or removed from the backend.

## Candidate response model

A future response may include:

- query and normalized search parameters;
- result category and result type;
- title, URL, snippet, and published date when available;
- source engine or engines;
- score or ordering metadata only when its meaning can be documented;
- request timing and partial-failure information suitable for diagnostics;
- pagination or continuation data;
- optional structured answers or infobox data.

## Security boundary

The current baseline exposes HTML only in the GoreeCloud runtime example. JSON, RSS, and CSV formats remain disabled until the API access model, authentication or network restriction, rate limiting, versioning, abuse controls, and monitoring expectations are accepted.

## Compatibility

No application should be written against undocumented SearXNG template objects, internal Python classes, or provider-specific structures when a GoreeCloud-owned abstraction can reasonably be introduced.
