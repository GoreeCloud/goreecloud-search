.. SPDX-License-Identifier: AGPL-3.0-or-later

.. _SearXNG development guide: https://docs.searxng.org/dev/index.html
.. _SearXNG commit guide: https://docs.searxng.org/dev/commits.html
.. _SearXNG repository: https://github.com/searxng/searxng
.. _SearXNG Weblate: https://translate.codeberg.org/projects/searxng/searxng/

Contributing to GoreeCloud Search
=================================

Thank you for your interest in GoreeCloud Search.

GoreeCloud Search is an original GoreeCloud-owned native application under
active migration. The repository still contains an inherited SearXNG-derived
tree for continuity, compatibility, feature-preservation decisions, rollback,
and applicable upstream security maintenance until controlled retirement.
Contributions must preserve the GoreeCloud product, privacy, security,
continuity, documentation, platform-integration, and current Stable Glaze UI
contracts. See ``GORECLOUD.md``, ``SPECIFICATIONS.md``,
``docs/goreecloud/UPSTREAM.md``, ``docs/goreecloud/READINESS.md``,
``AI_POLICY.rst``, and ``SECURITY.md`` before making material changes.

Where should a change go?
=========================

GoreeCloud-owned changes
------------------------

New first-party product behavior should normally be implemented in the native
application under ``native/``. Submit changes to this repository when they
primarily affect GoreeCloud Search, including:

- native search orchestration, provider contracts, ranking, result handling, and
  first-party APIs;
- Glaze UI implementation, accessibility, responsive behavior, and product UX;
- GoreeCloud naming, browser metadata, icons, manifests, and secondary surfaces;
- GoreeCloud runtime and deployment defaults;
- Privacy Shield, Wardveil Security, Everkeep, GoreeCloud Identity, GoreeCloud
  Mesh, monitoring, recovery, and release-readiness controls applicable to
  Search;
- GoreeCloud CI, acceptance tooling, provider validation, and documentation;
- GoreeCloud application, Browser, AI, automation, Sync, or research
  integrations.

Do not add new product behavior to the inherited SearXNG-derived tree merely
because that path is familiar. A temporary compatibility change in inherited
code should have a clear migration reason and must not silently redefine the
native target architecture.

Upstream-capable transitional changes
-------------------------------------

A generally useful SearXNG engine, parser, networking, performance, security, or
backend fix that is required while the transitional tree remains in service and
does not depend on GoreeCloud-specific behavior should normally be considered
for the `SearXNG repository`_ as well. Keeping broadly useful inherited-engine
changes upstream reduces transitional divergence and makes security maintenance
safer until retirement.

If a GoreeCloud deployment needs an urgent inherited-runtime fix before an
upstream change is accepted, the fix may be carried here with clear provenance,
a transition/rollback rationale, and an upstream-sync note in the relevant
documentation.

The upstream SearXNG project controls its own contribution and AI policies.
GoreeCloud's ``AI_POLICY.rst`` applies to this repository and does not override
requirements that upstream SearXNG may apply to a contribution submitted there.

Translations
============

Retained upstream SearXNG interface strings continue to use the upstream
`SearXNG Weblate`_ translation process where appropriate while that interface
remains part of the transitional runtime.

GoreeCloud-owned product copy, branding, About content, Glaze UI terminology,
and private-service guidance require GoreeCloud review. A translation must not
silently reintroduce SearXNG as the target product identity or restore public-
instance guidance that conflicts with the GoreeCloud Search private-service
model. Until a GoreeCloud-maintained translation exists, the canonical
GoreeCloud product content may intentionally fall back to English.

Development workflow
====================

Use an isolated branch and a pull request for material work. Keep commits
focused, documented, and reviewable. Large architectural changes or changes that
materially alter migration, platform trust, provider behavior, privacy,
security, or recovery should explain the purpose, maintenance cost, migration
impact, and rollback strategy.

The retained SearXNG development tooling and conventions remain useful when
working in inherited code. Refer to the `SearXNG development guide`_ and
`SearXNG commit guide`_ for that scope unless a GoreeCloud rule or
project-specific contract is stricter.

AI-assisted and agent-authored contributions
--------------------------------------------

AI-assisted and agent-authored contributions are permitted under
``AI_POLICY.rst``. An AI agent may be the primary implementation author and may
also help prepare tests, documentation, pull-request descriptions, and review
responses. GoreeCloud does not impose a human-written-majority requirement.

Material AI or agent involvement should be identified in the pull request or
another repository record when useful for auditability. More importantly, all
claims about tests, review, release state, deployment, and acceptance must remain
truthful and evidence-backed. AI involvement never bypasses the normal branch,
pull-request, validation, security, privacy, licensing, release, or production
boundaries.

Commit messages
---------------

- Write descriptive commit titles.
- Use the imperative mood and present tense.
- Keep the first line concise; approximately 72 characters or fewer is
  preferred.
- Explain non-obvious security, privacy, compatibility, recovery, migration, or
  upstream-synchronization decisions in the commit body or project
  documentation.

Code and configuration expectations
===================================

Changes should be understandable without undocumented assumptions. Preserve
logical structure, readable formatting, comments that explain important
implementation decisions, and clear boundaries between transitional upstream-
derived code and GoreeCloud-owned native behavior.

For Go code, use ``gofmt`` and keep package boundaries narrow and testable. For
Python code retained by the transitional tree or acceptance tooling, continue
to follow the project's PEP 8/PEP 20 conventions and applicable lint/type/test
tooling. Do not commit secrets, credentials, private keys, production addresses,
private DNS data, user information, or other sensitive GoreeCloud infrastructure
details.

Glaze UI
========

GLAZE UI V1.1 / 1.1.0 is the current declared Stable consumer target for
GoreeCloud-controlled Search surfaces. Final Search conformance must use the
current authoritative immutable Stable Glaze contract and evidence available at
the time of acceptance. A superseded, reset-baseline, Candidate, RC, or other
non-Stable line must not be used to claim current Stable consumer conformance.

Changes to the native interface must preserve the current semantic/material
hierarchy, a 48px general interaction-target floor, intentional light and dark
behavior, accessible focus, Reduced Motion, Reduced Transparency, Increased
Contrast, Forced Colors/effects-free resilience, and Compact/Medium/Expanded/
Wide adaptive behavior. Stable conformance remains exact-candidate-bound and
requires the application-specific browser, physical-device where applicable,
accessibility, resilience, and human visual-review evidence defined by
``docs/goreecloud/READINESS.md``.

The inherited SearXNG-derived Glaze layers remain transitional compatibility
material. Their historical Glaze version markers must not be presented as the
current Stable Search consumer target.

A visual change is not complete merely because it renders. It must remain
coherent, usable, accessible, privacy-conscious, resilient, and recognizably
GoreeCloud.

Validation
==========

A material native pull request should keep the applicable deterministic gates
green, including:

- ``GoreeCloud Search Native Foundation``;
- ``GoreeCloud API v1 service contract`` when API/readiness behavior changes;
- ``GoreeCloud native results browser acceptance`` (currently expanded to cover
  the core native application surfaces);
- ``GoreeCloud native development artifact`` when native source or release-
  packaging boundaries change;
- ``GoreeCloud platform integrations``;
- ``GoreeCloud workflow supply-chain guard``;
- applicable container/runtime checks; and
- retained transitional ``GoreeCloud foundation`` and upstream ``Integration``
  checks while those trees remain required for migration safety.

Run candidate-bound live-provider acceptance before Stable promotion and when a
change affects production provider behavior, category routing, result parsing,
provider metadata authority, or target-environment provider readiness. Every
category selected for the native release must have accepted executable provider
coverage. External throttling or blocking must be classified separately from an
application defect.

Passing CI establishes source confidence; it does not by itself authorize a
production deployment. Production promotion remains subject to the readiness
requirements in ``docs/goreecloud/READINESS.md`` and the governing GoreeCloud
release-lifecycle and production-readiness standards.

Security
========

Follow ``SECURITY.md`` for vulnerability reporting. Do not disclose a sensitive
GoreeCloud Search vulnerability in a public issue. Vulnerabilities that are
reproducible in unchanged transitional upstream SearXNG code should be routed
through the upstream SearXNG security process as applicable.

License and attribution
=======================

The inherited SearXNG-derived code remains subject to the GNU Affero General
Public License v3.0 or later and all applicable upstream copyright,
source-availability, and attribution obligations. GoreeCloud-owned changes in
this repository must preserve those obligations wherever derivative or
inherited code is involved. Native ownership does not erase the licensing
requirements of transitional upstream-derived material that remains distributed
or deployed.