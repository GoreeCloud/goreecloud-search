import json
from threading import Thread
import unittest
from urllib.error import HTTPError
from urllib.request import Request, urlopen

from goreecloud_search import (
    ProviderDescriptor,
    ProviderOrigin,
    ProviderSearchBatch,
    ResultCandidate,
    SearchCategory,
    SearchCore,
    SourceMode,
    create_local_server,
)


class FakeProvider:
    def __init__(self) -> None:
        self._descriptor = ProviderDescriptor(
            name="fake-external",
            origin=ProviderOrigin.EXTERNAL,
            categories=frozenset({SearchCategory.GENERAL}),
        )
        self.queries = []

    @property
    def descriptor(self):
        return self._descriptor

    async def search(self, query, *, limit):
        self.queries.append(query)
        return ProviderSearchBatch(
            candidates=(
                ResultCandidate(
                    title="GoreeCloud Search",
                    url="https://example.com/search",
                    snippet="Native search result",
                    provider=self.descriptor.name,
                ),
            )[:limit]
        )


class LocalHTTPAPITests(unittest.TestCase):
    def setUp(self) -> None:
        self.provider = FakeProvider()
        core = SearchCore(provider_adapters=(self.provider,))
        self.server = create_local_server(
            core,
            port=0,
            mode=SourceMode.EXTERNAL_ONLY,
        )
        self.thread = Thread(
            target=self.server.serve_forever,
            daemon=True,
        )
        self.thread.start()
        self.base = f"http://127.0.0.1:{self.server.server_port}"

    def tearDown(self) -> None:
        self.server.shutdown()
        self.server.server_close()
        self.thread.join(timeout=2)

    def _get(self, path):
        with urlopen(f"{self.base}{path}", timeout=2) as response:
            return response.status, response.headers, json.loads(response.read())

    def _error(self, path, *, method="GET"):
        request = Request(f"{self.base}{path}", method=method)
        with self.assertRaises(HTTPError) as context:
            urlopen(request, timeout=2)
        error = context.exception
        return error.code, error.headers, json.loads(error.read())

    def test_health_is_loopback_only_no_store_and_no_cors(self):
        self.assertEqual(self.server.server_address[0], "127.0.0.1")
        status, headers, payload = self._get("/healthz")
        self.assertEqual(status, 200)
        self.assertEqual(payload["status"], "ok")
        self.assertEqual(payload["lifecycle"], "development")
        self.assertEqual(headers["Cache-Control"], "no-store")
        self.assertEqual(headers["Referrer-Policy"], "no-referrer")
        self.assertIsNone(headers["Access-Control-Allow-Origin"])

    def test_search_runs_existing_pipeline(self):
        status, headers, payload = self._get(
            "/api/v1/search?q=privacy%20search&limit=5"
        )
        self.assertEqual(status, 200)
        self.assertEqual(headers["X-Content-Type-Options"], "nosniff")
        self.assertEqual(payload["api_version"], 1)
        results = payload["response"]["results"]
        self.assertEqual(len(results), 1)
        self.assertEqual(
            results[0]["result"]["canonical_url"],
            "https://example.com/search",
        )
        self.assertEqual(self.provider.queries[0].terms, ("privacy", "search"))

    def test_missing_query_is_rejected_without_provider_execution(self):
        status, _, payload = self._error("/api/v1/search?limit=5")
        self.assertEqual(status, 400)
        self.assertEqual(payload["error"], "invalid_search_request")
        self.assertEqual(self.provider.queries, [])

    def test_unknown_parameter_is_rejected(self):
        status, _, payload = self._error(
            "/api/v1/search?q=privacy&callback=example"
        )
        self.assertEqual(status, 400)
        self.assertEqual(payload["error"], "unsupported_query_parameter")
        self.assertEqual(self.provider.queries, [])

    def test_excessive_limit_is_rejected(self):
        status, _, payload = self._error(
            "/api/v1/search?q=privacy&limit=21"
        )
        self.assertEqual(status, 400)
        self.assertEqual(payload["error"], "invalid_search_request")
        self.assertEqual(self.provider.queries, [])

    def test_health_rejects_parameters(self):
        status, _, payload = self._error("/healthz?verbose=1")
        self.assertEqual(status, 400)
        self.assertEqual(
            payload["error"],
            "health_endpoint_accepts_no_parameters",
        )

    def test_mutating_methods_are_rejected(self):
        status, _, payload = self._error(
            "/api/v1/search",
            method="POST",
        )
        self.assertEqual(status, 405)
        self.assertEqual(payload["error"], "method_not_allowed")


if __name__ == "__main__":
    unittest.main()
