from threading import Thread
import unittest
from urllib.error import HTTPError
from urllib.parse import quote
from urllib.request import urlopen

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


class WebFakeProvider:
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
                    title="<script>alert('title')</script>",
                    url="https://example.com/result?utm_source=test",
                    snippet="<img src=x onerror=alert('snippet')>",
                    provider=self.descriptor.name,
                ),
            )[:limit]
        )


class LocalWebUITests(unittest.TestCase):
    def setUp(self) -> None:
        self.provider = WebFakeProvider()
        self.server = create_local_server(
            SearchCore(provider_adapters=(self.provider,)),
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

    def _text(self, path):
        with urlopen(f"{self.base}{path}", timeout=2) as response:
            return (
                response.status,
                response.headers,
                response.read().decode("utf-8"),
            )

    def test_home_is_local_self_contained_search_surface(self):
        status, headers, body = self._text("/")
        self.assertEqual(status, 200)
        self.assertTrue(headers["Content-Type"].startswith("text/html"))
        self.assertIn("GoreeCloud Search", body)
        self.assertIn('action="/search"', body)
        self.assertIn('href="/assets/search.css"', body)
        self.assertIn('href="/opensearch.xml"', body)
        self.assertNotIn("<script", body.casefold())
        self.assertIn("style-src 'self'", headers["Content-Security-Policy"])
        self.assertNotIn("script-src", headers["Content-Security-Policy"])
        self.assertIsNone(headers["Access-Control-Allow-Origin"])

    def test_html_search_escapes_query_and_provider_content(self):
        query = "<svg onload=alert('query')>"
        status, headers, body = self._text(
            "/search?q=" + quote(query)
        )
        self.assertEqual(status, 200)
        self.assertIn(
            "External provider query disclosure active",
            body,
        )
        self.assertNotIn("<svg", body.casefold())
        self.assertNotIn("<script", body.casefold())
        self.assertNotIn("<img", body.casefold())
        self.assertIn("&lt;svg", body)
        self.assertIn("&lt;script&gt;", body)
        self.assertIn("&lt;img", body)
        self.assertIn(
            'href="https://example.com/result"',
            body,
        )
        self.assertIn('rel="noreferrer noopener"', body)
        self.assertEqual(headers["Referrer-Policy"], "no-referrer")

    def test_opensearch_descriptor_uses_actual_loopback_port(self):
        status, headers, body = self._text("/opensearch.xml")
        self.assertEqual(status, 200)
        self.assertTrue(
            headers["Content-Type"].startswith(
                "application/opensearchdescription+xml"
            )
        )
        self.assertIn(
            f"http://127.0.0.1:{self.server.server_port}/search?"
            "q={searchTerms}",
            body,
        )
        self.assertNotIn("localhost", body)

    def test_css_has_accessibility_fallbacks_and_no_remote_imports(self):
        status, headers, body = self._text("/assets/search.css")
        self.assertEqual(status, 200)
        self.assertTrue(headers["Content-Type"].startswith("text/css"))
        self.assertIn("prefers-reduced-motion", body)
        self.assertIn("forced-colors: active", body)
        self.assertNotIn("@import", body.casefold())
        self.assertNotIn("url(http", body.casefold())

    def test_invalid_human_search_returns_html_without_provider_call(self):
        with self.assertRaises(HTTPError) as context:
            urlopen(
                f"{self.base}/search?callback=x",
                timeout=2,
            )
        error = context.exception
        self.assertEqual(error.code, 400)
        self.assertTrue(
            error.headers["Content-Type"].startswith("text/html")
        )
        body = error.read().decode("utf-8")
        self.assertIn("Search unavailable", body)
        self.assertIn("unsupported query parameter", body)
        self.assertEqual(self.provider.queries, [])


if __name__ == "__main__":
    unittest.main()
