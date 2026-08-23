# SPDX-License-Identifier: AGPL-3.0-or-later
"""GoreeCloud Search API service-contract tests."""

from unittest import mock

import searx
import searx.webapp  # pylint: disable=unused-import
from tests import SearxTestCase


class GoreeCloudAPIPluginTestCase(SearxTestCase):
    """Validate first-party, read-only integration service contracts."""

    @staticmethod
    def _setting_override(instance_name):
        """Return a settings lookup with an explicit test instance identity."""
        real_get_setting = searx.get_setting

        def get_setting(name, default=None):
            if name == "general.instance_name":
                return instance_name
            return real_get_setting(name, default)

        return get_setting

    def test_status_contract(self):
        result = self.client.get("/api/v1/status")

        self.assertEqual(result.status_code, 200)
        self.assertEqual(result.mimetype, "application/json")
        self.assertEqual(result.headers["Cache-Control"], "no-store")
        self.assertEqual(result.headers["Pragma"], "no-cache")
        self.assertEqual(result.headers["X-GoreeCloud-API-Version"], "1")

        payload = result.get_json()
        self.assertEqual(payload["api_version"], "1")
        self.assertEqual(payload["product"], "GoreeCloud Search")
        self.assertEqual(payload["service"], "search")
        self.assertEqual(payload["status"], "ok")
        self.assertEqual(payload["foundation"]["name"], "SearXNG")
        self.assertTrue(payload["foundation"]["version"])
        self.assertTrue(payload["capabilities"]["html_search"])
        self.assertTrue(payload["capabilities"]["opensearch"])
        self.assertFalse(payload["capabilities"]["machine_readable_search_api"])
        self.assertEqual(payload["endpoints"]["health"], "/healthz")
        self.assertEqual(payload["endpoints"]["opensearch"], "/opensearch.xml")
        self.assertEqual(payload["endpoints"]["readiness"], "/api/v1/readiness")
        self.assertEqual(payload["endpoints"]["search"], "/search")

        self.assertNotIn("engines", payload)
        self.assertNotIn("plugins", payload)
        self.assertNotIn("preferences", payload)

    def test_status_ignores_query(self):
        marker = "goreecloud-private-query-marker"
        result = self.client.get(f"/api/v1/status?q={marker}")

        self.assertEqual(result.status_code, 200)
        self.assertNotIn(marker, result.get_data(as_text=True))

    def test_readiness_contract(self):
        with mock.patch(
            "searx.plugins.goreecloud_api.searx.get_setting",
            side_effect=self._setting_override("GoreeCloud Search"),
        ):
            result = self.client.get("/api/v1/readiness")

        self.assertEqual(result.status_code, 200)
        self.assertEqual(result.mimetype, "application/json")
        self.assertEqual(result.headers["Cache-Control"], "no-store")
        self.assertEqual(result.headers["Pragma"], "no-cache")
        self.assertEqual(result.headers["X-GoreeCloud-API-Version"], "1")

        payload = result.get_json()
        self.assertEqual(payload["api_version"], "1")
        self.assertEqual(payload["product"], "GoreeCloud Search")
        self.assertEqual(payload["service"], "search")
        self.assertEqual(payload["status"], "ready")
        self.assertTrue(payload["ready"])
        self.assertEqual(payload["readiness_scope"], "local_application")
        self.assertTrue(all(payload["checks"].values()))
        self.assertEqual(
            payload["not_evaluated"],
            [
                "external_search_providers",
                "dns",
                "reverse_proxy",
                "monitoring_and_alert_delivery",
                "backup_restore_and_rollback",
            ],
        )
        for forbidden in ("engines", "plugins", "preferences", "query", "results"):
            self.assertNotIn(forbidden, payload)

    def test_readiness_failure(self):
        with mock.patch(
            "searx.plugins.goreecloud_api.searx.get_setting",
            side_effect=self._setting_override("Unexpected Search"),
        ):
            result = self.client.get("/api/v1/readiness")

        self.assertEqual(result.status_code, 503)
        payload = result.get_json()
        self.assertEqual(payload["status"], "not_ready")
        self.assertFalse(payload["ready"])
        self.assertFalse(payload["checks"]["service_identity"])

    def test_readiness_ignores_query(self):
        marker = "goreecloud-private-query-marker"
        with mock.patch(
            "searx.plugins.goreecloud_api.searx.get_setting",
            side_effect=self._setting_override("GoreeCloud Search"),
        ):
            result = self.client.get(f"/api/v1/readiness?q={marker}")

        self.assertEqual(result.status_code, 200)
        self.assertNotIn(marker, result.get_data(as_text=True))

    def test_advertised_health_path(self):
        result = self.client.get("/healthz")

        self.assertEqual(result.status_code, 200)
        self.assertEqual(result.mimetype, "text/plain")
        self.assertEqual(result.get_data(as_text=True), "OK")
