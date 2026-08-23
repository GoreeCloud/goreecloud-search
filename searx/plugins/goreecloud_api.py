# SPDX-License-Identifier: AGPL-3.0-or-later
"""GoreeCloud Search integration API.

This module exposes deliberately small, versioned machine-readable service
contracts for GoreeCloud clients. It does not expose search queries, results,
provider configuration, credentials, or user preferences.
"""

import searx
from flask import current_app, jsonify

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
            description="Versioned integration status and readiness contracts for GoreeCloud clients.",
            preference_section=None,
        )

    def init(self, app):
        """Register the versioned read-only GoreeCloud API service endpoints."""
        if "goreecloud_api_status" not in app.view_functions:
            app.add_url_rule(
                "/api/v1/status",
                endpoint="goreecloud_api_status",
                view_func=self.status,
                methods=["GET"],
            )
        if "goreecloud_api_readiness" not in app.view_functions:
            app.add_url_rule(
                "/api/v1/readiness",
                endpoint="goreecloud_api_readiness",
                view_func=self.readiness,
                methods=["GET"],
            )
        return True

    @staticmethod
    def _finalize_response(response):
        """Apply shared non-cacheable API response controls."""
        response.headers["Cache-Control"] = "no-store"
        response.headers["Pragma"] = "no-cache"
        response.headers["X-GoreeCloud-API-Version"] = "1"
        return response

    @classmethod
    def status(cls):
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
                    "readiness": "/api/v1/readiness",
                    "search": "/search",
                },
            }
        )
        return cls._finalize_response(response)

    @classmethod
    def readiness(cls):
        """Report whether the local GoreeCloud Search application is ready."""
        routes = {rule.rule for rule in current_app.url_map.iter_rules()}
        checks = {
            "service_identity": searx.get_setting("general.instance_name", "") == "GoreeCloud Search",
            "html_search_enabled": "html" in searx.get_setting("search.formats", []),
            "health_route_registered": "/healthz" in routes,
            "opensearch_route_registered": "/opensearch.xml" in routes,
            "search_route_registered": "/search" in routes,
            "status_route_registered": "/api/v1/status" in routes,
        }
        ready = all(checks.values())
        response = jsonify(
            {
                "api_version": "1",
                "product": "GoreeCloud Search",
                "service": "search",
                "status": "ready" if ready else "not_ready",
                "ready": ready,
                "readiness_scope": "local_application",
                "checks": checks,
                "not_evaluated": [
                    "external_search_providers",
                    "dns",
                    "reverse_proxy",
                    "monitoring_and_alert_delivery",
                    "backup_restore_and_rollback",
                ],
            }
        )
        response.status_code = 200 if ready else 503
        return cls._finalize_response(response)
