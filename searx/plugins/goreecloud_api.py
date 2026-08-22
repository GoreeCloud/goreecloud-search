# SPDX-License-Identifier: AGPL-3.0-or-later
"""GoreeCloud Search integration API.

This module exposes a deliberately small, versioned machine-readable service
contract for GoreeCloud clients. It does not expose search queries, results,
provider configuration, credentials, or user preferences.
"""

from flask import jsonify

from searx.version import VERSION_STRING

from ._core import Plugin, PluginInfo


class GoreeCloudAPIPlugin(Plugin):
    """Always-on first-party API boundary for GoreeCloud integrations."""

    id = "goreecloud_api"

    def __init__(self, plg_cfg):
        super().__init__(plg_cfg)
        self.info = PluginInfo(
            id=self.id,
            name="GoreeCloud API",
            description="Versioned integration status contract for GoreeCloud clients.",
            preference_section=None,
        )

    def init(self, app):
        """Register the versioned read-only GoreeCloud API status endpoint."""
        app.add_url_rule(
            "/api/v1/status",
            endpoint="goreecloud_api_status",
            view_func=self.status,
            methods=["GET"],
        )
        return True

    @staticmethod
    def status():
        """Return non-sensitive service identity and capability metadata."""
        response = jsonify(
            {
                "api_version": "1",
                "product": "GoreeCloud Search",
                "service": "search",
                "status": "ok",
                "foundation": {
                    "name": "SearXNG",
                    "version": VERSION_STRING,
                },
                "capabilities": {
                    "html_search": True,
                    "opensearch": True,
                    "machine_readable_search_api": False,
                },
                "endpoints": {
                    "health": "/healthz",
                    "opensearch": "/opensearch.xml",
                    "search": "/search",
                },
            }
        )
        response.headers["Cache-Control"] = "no-store"
        response.headers["Pragma"] = "no-cache"
        response.headers["X-GoreeCloud-API-Version"] = "1"
        return response
