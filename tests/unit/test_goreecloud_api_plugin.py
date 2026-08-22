# SPDX-License-Identifier: AGPL-3.0-or-later
"""GoreeCloud Search API status-contract tests."""

import searx.webapp  # pylint: disable=unused-import
from tests import SearxTestCase


class GoreeCloudAPIPluginTestCase(SearxTestCase):
    """Validate the first-party, read-only integration status contract."""

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
        self.assertEqual(payload["endpoints"]["search"], "/search")

        self.assertNotIn("engines", payload)
        self.assertNotIn("plugins", payload)
        self.assertNotIn("preferences", payload)

    def test_status_ignores_query(self):
        marker = "goreecloud-private-query-marker"
        result = self.client.get(f"/api/v1/status?q={marker}")

        self.assertEqual(result.status_code, 200)
        self.assertNotIn(marker, result.get_data(as_text=True))
