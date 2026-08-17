.. SPDX-License-Identifier: AGPL-3.0-or-later

GoreeCloud Search
=================

GoreeCloud Search is the privacy-first metasearch and research gateway for the GoreeCloud personal cloud platform.

This repository is a GoreeCloud-maintained fork of `SearXNG <https://github.com/searxng/searxng>`_. It keeps SearXNG's mature metasearch engine foundation while establishing a distinct GoreeCloud product experience, canonical Glaze UI 1.0 presentation layer, privacy-oriented defaults, operational contract, and future GoreeCloud integrations.

Project status
--------------

The maintained fork includes GoreeCloud-specific source validation, direct application runtime smoke testing, custom container build/runtime acceptance, retained upstream integration testing, and Chromium acceptance across the Glaze UI Compact, Medium, Expanded, and Wide adaptive layout classes.

Source validation is intentionally separate from production authorization. A source revision can be suitable for merge while the deployed service still requires representative real-provider search validation, target private-access integration, monitoring and alerting, backup/restore validation, integration acceptance, and a tested rollback path. See ``docs/goreecloud/READINESS.md`` for the complete release boundary.

Current upstream baseline
-------------------------

The initial GoreeCloud development line was created from upstream commit::

    b2da6b90f2f8446557c91f67d6be5064ab785ecd

Upstream repository:

- https://github.com/searxng/searxng

GoreeCloud repository:

- https://github.com/GoreeCloud/goreecloud-search

Product direction
-----------------

GoreeCloud Search is intended to provide:

- Private and self-hosted metasearch.
- A GoreeCloud-owned search experience using canonical Glaze UI 1.0 semantics.
- Web, image, video, news, technical, academic, software, and other supported search categories.
- Clear engine and source visibility.
- Privacy-oriented search defaults with minimal retained user data.
- GoreeCloud-owned browser artwork, web-app manifest, and OpenSearch integration.
- A governed GoreeCloud-facing search API for approved applications and local AI research workflows when its access contract is accepted.
- Failure isolation when individual external search providers are unavailable.
- Documented backup, restore, upgrade, rollback, and upstream-maintenance procedures.

Architecture principle
----------------------

**GoreeCloud Search is the product. SearXNG is the initial search foundation.**

The GoreeCloud-facing UI and integration boundaries should remain sufficiently independent that the backend can evolve later without forcing every GoreeCloud consumer to depend directly on SearXNG internals.

Validation model
----------------

The maintained fork uses four complementary GoreeCloud validation layers plus the retained upstream Integration workflow:

- ``goreecloud-foundation.yml`` for source, product-contract, Glaze UI 1.0, privacy, deployment, provenance, syntax, and AGPL checks.
- ``goreecloud-runtime-smoke.yml`` for direct application startup, privacy/security configuration, rendered product identity, and HTTP behavior.
- ``goreecloud-container-build.yml`` for OCI image build and container-runtime acceptance.
- ``goreecloud-browser-acceptance.yml`` for Compact/Medium/Expanded/Wide browser identity, GoreeCloud manifest/OpenSearch artwork and metadata, target sizing, keyboard behavior, About/Preferences/recovery coverage, and responsive-layout checks.
- upstream ``integration.yml`` for the SearXNG lint, unit, Robot, Python-version, and theme contracts.

Real external providers are tested through the manual ``goreecloud-provider-acceptance.yml`` workflow so third-party throttling does not make deterministic pull-request CI unreliable. Its representative suite covers the SearXNG ``general``, ``images``, ``news``, ``videos``, ``it``, and ``science`` categories, with single-category mode available for diagnosis.

These gates validate the maintained-fork source foundation. They do not replace target-environment, provider, private-access, monitoring, recovery, or rollback acceptance.

Privacy and security baseline
-----------------------------

The GoreeCloud runtime example keeps public-instance mode and SearXNG usage metrics disabled, proxies result images, keeps query text out of browser page titles, sends ``noindex, nofollow`` and ``no-referrer`` directives, denies frame embedding, disables camera/microphone/geolocation browser capabilities, advertises private-search robots metadata, and enables HTML only by default.

Machine-readable formats remain disabled until an approved integration contract defines the requesting service, private access boundary, expected volume, query sensitivity, logging and retention, failure behavior, monitoring, and disablement procedure.

Development boundaries
----------------------

The project intentionally preserves upstream provenance and minimizes unnecessary divergence. Changes should be isolated, documented, testable, and reviewable so that relevant upstream security fixes and engine improvements can continue to be incorporated safely.

Production deployment, DNS changes, Caddy changes, NetBird policy changes, firewall changes, and migration of the existing SearXNG service require separate validation and acceptance.

Upstream SearXNG
----------------

SearXNG is a privacy-respecting open metasearch engine. The original project documentation is available at https://docs.searxng.org/.

This fork preserves SearXNG copyright, license, source-availability, and attribution requirements. See ``LICENSE`` and ``GORECLOUD.md`` for additional project-specific information.

License
-------

This repository is licensed under the GNU Affero General Public License v3.0 or later, consistent with the upstream SearXNG source baseline. See ``LICENSE`` for the complete terms.