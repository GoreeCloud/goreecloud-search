.. SPDX-License-Identifier: AGPL-3.0-or-later

.. _SearXNG development guide: https://docs.searxng.org/dev/index.html
.. _SearXNG commit guide: https://docs.searxng.org/dev/commits.html
.. _SearXNG repository: https://github.com/searxng/searxng
.. _SearXNG Weblate: https://translate.codeberg.org/projects/searxng/searxng/

Contributing to GoreeCloud Search
=================================

Thank you for your interest in GoreeCloud Search.

GoreeCloud Search is a GoreeCloud-maintained fork of SearXNG. Contributions
should preserve the fork's upstream maintainability while respecting the
GoreeCloud product, privacy, security, documentation, and Glaze UI contracts.
See ``GORECLOUD.md``, ``docs/goreecloud/UPSTREAM.md``,
``docs/goreecloud/READINESS.md``, and ``SECURITY.md`` before making material
changes.

Where should a change go?
=========================

GoreeCloud-owned changes
------------------------

Submit changes to this repository when they primarily affect GoreeCloud Search,
including:

- Glaze UI implementation, accessibility, responsive behavior, and product UX;
- GoreeCloud naming, templates, browser metadata, icons, manifests, and
  secondary surfaces;
- GoreeCloud runtime and deployment defaults;
- privacy, security, recovery, monitoring, and release-readiness controls that
  are specific to this maintained fork;
- GoreeCloud CI, acceptance tooling, provider validation, and documentation;
- GoreeCloud application, browser, AI, automation, or research integrations.

Upstream-capable changes
------------------------

A generally useful SearXNG engine, parser, networking, performance, security, or
backend fix that does not depend on GoreeCloud-specific behavior should normally
be considered for the `SearXNG repository`_ as well. Keeping broadly useful
engine changes upstream reduces long-term fork divergence and makes future
security and engine synchronization safer.

If a GoreeCloud release needs an urgent fix before an upstream change is
accepted, the fix may be carried here with clear provenance and an upstream-sync
note in the relevant documentation.

Translations
============

Retained upstream SearXNG interface strings continue to use the upstream
`SearXNG Weblate`_ translation process where appropriate.

GoreeCloud-owned product copy, branding, About content, Glaze UI terminology,
and private-service guidance require GoreeCloud review. A translation must not
silently reintroduce SearXNG as the visible product identity or restore public-
instance guidance that conflicts with the GoreeCloud Search private-service
model. Until a GoreeCloud-maintained translation exists, the canonical
GoreeCloud product content may intentionally fall back to English.

Development workflow
====================

Use an isolated branch and a pull request for material work. Keep commits
focused, documented, and reviewable. Large architectural changes or changes that
materially increase upstream divergence should begin with an issue or proposal
that explains the purpose, maintenance cost, migration impact, and rollback
strategy.

The retained SearXNG development tooling and conventions remain useful for the
shared engine foundation. Refer to the `SearXNG development guide`_ and
`SearXNG commit guide`_ when working in inherited code, unless a GoreeCloud rule
or project-specific contract is stricter.

Commit messages
---------------

- Write descriptive commit titles.
- Use the imperative mood and present tense.
- Keep the first line concise; approximately 72 characters or fewer is
  preferred.
- Explain non-obvious security, privacy, compatibility, recovery, or upstream-
  synchronization decisions in the commit body or project documentation.

Code and configuration expectations
===================================

Changes should be understandable without undocumented assumptions. Preserve
logical structure, readable formatting, comments that explain important
implementation decisions, and clear boundaries between upstream-derived code
and GoreeCloud-owned behavior.

For Python code, continue to follow the project's PEP 8/PEP 20 conventions and
the retained SearXNG lint/type/test tooling. Do not commit secrets, credentials,
private keys, production addresses, private DNS data, user information, or
other sensitive GoreeCloud infrastructure details.

Glaze UI
========

Glaze UI 1.0 is mandatory for GoreeCloud-controlled user-facing surfaces.
Changes to the web interface must preserve the semantic token contract,
intentional surface hierarchy, light and dark behavior, accessible focus and
practical target sizes, reduced-motion and contrast/transparency resilience,
and Compact/Medium/Expanded/Wide adaptive behavior.

A visual change is not complete merely because it renders. It should remain
coherent, usable, accessible, privacy-conscious, and recognizably GoreeCloud.

Validation
==========

A material pull request should keep the applicable deterministic gates green:

- ``GoreeCloud foundation``;
- ``GoreeCloud runtime smoke``;
- ``GoreeCloud container build``;
- ``GoreeCloud browser acceptance``;
- retained upstream ``Integration`` checks.

Run the manual ``GoreeCloud provider acceptance`` workflow when the change can
affect real search-provider behavior, category routing, result parsing, engine
configuration, or target-environment provider readiness. External throttling or
blocking must be classified separately from an application defect.

Passing CI establishes source confidence; it does not by itself authorize a
production deployment. Production promotion remains subject to the readiness
requirements in ``docs/goreecloud/READINESS.md``.

Security
========

Follow ``SECURITY.md`` for vulnerability reporting. Do not disclose a sensitive
GoreeCloud Search vulnerability in a public issue. Vulnerabilities that are
reproducible in unchanged upstream SearXNG should be routed through the upstream
SearXNG security process.

License and attribution
=======================

GoreeCloud Search remains licensed under the GNU Affero General Public License
v3.0 or later, consistent with the upstream source baseline. Preserve required
copyright, license, source-availability, and attribution information when
modifying inherited or GoreeCloud-owned files.