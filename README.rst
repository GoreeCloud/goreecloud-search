.. SPDX-License-Identifier: AGPL-3.0-or-later

GoreeCloud Search
=================

GoreeCloud Search is the privacy-first metasearch and research gateway for the GoreeCloud personal cloud platform.

This repository is a GoreeCloud-maintained fork of `SearXNG <https://github.com/searxng/searxng>`_. It keeps SearXNG's mature metasearch engine foundation while establishing a distinct GoreeCloud product experience, Glaze UI presentation layer, privacy-oriented defaults, operational contract, and future GoreeCloud integrations.

Project status
--------------

Development is active on the ``agent/stable-foundation`` branch. The current work is a source-development foundation and is not yet an approved production replacement for the existing SearXNG deployment at ``search.goreecloud.com``.

Current upstream baseline
-------------------------

The initial GoreeCloud development branch was created from upstream commit::

    b2da6b90f2f8446557c91f67d6be5064ab785ecd

Upstream repository:

- https://github.com/searxng/searxng

GoreeCloud repository:

- https://github.com/GoreeCloud/goreecloud-search

Product direction
-----------------

GoreeCloud Search is intended to provide:

- Private and self-hosted metasearch.
- A GoreeCloud-owned search experience using Glaze UI.
- Web, image, video, news, technical, academic, software, and other supported search categories.
- Clear engine and source visibility.
- Privacy-oriented search defaults with minimal retained user data.
- Browser and OpenSearch integration.
- A stable GoreeCloud-facing search API for approved applications and local AI research workflows.
- Failure isolation when individual external search providers are unavailable.
- Documented backup, restore, upgrade, rollback, and upstream-maintenance procedures.

Architecture principle
----------------------

**GoreeCloud Search is the product. SearXNG is the initial search foundation.**

The GoreeCloud-facing UI and integration boundaries should remain sufficiently independent that the backend can evolve later without forcing every GoreeCloud consumer to depend directly on SearXNG internals.

Development boundaries
----------------------

The project intentionally preserves upstream provenance and minimizes unnecessary divergence. Changes should be isolated, documented, testable, and reviewable so that relevant upstream security fixes and engine improvements can continue to be incorporated safely.

Production deployment, DNS changes, Caddy changes, NetBird policy changes, and migration of the existing SearXNG service require separate validation and acceptance.

Upstream SearXNG
----------------

SearXNG is a privacy-respecting open metasearch engine. The original project documentation is available at https://docs.searxng.org/.

This fork preserves SearXNG copyright, license, source-availability, and attribution requirements. See ``LICENSE`` and ``GORECLOUD.md`` for additional project-specific information.

License
-------

This repository is licensed under the GNU Affero General Public License v3.0 or later, consistent with the upstream SearXNG source baseline. See ``LICENSE`` for the complete terms.
