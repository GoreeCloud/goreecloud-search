from threading import Thread
import unittest
from urllib.error import HTTPError
from urllib.parse import quote, urlencode
from urllib.request import Request, urlopen

from goreecloud_search import (
    ProviderDescriptor,
    ProviderOrigin,
    ProviderSearchBatch,
    ResultCandidate,
    SearchCategory,
    SearchCore,
    SourceMode,
    create_container_server,
    create_local_server,
)


class WebFakeProvider:
    def __init__(self) -> None:
        self._descriptor = ProviderDescriptor(
            name="fake-external",
            origin=ProviderOrigin.EXTERNAL,
            categories=frozenset({SearchCategory.GENERAL}),
            third_party_query_disclosure=True,
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

    def _post_form(self, values):
        request = Request(
            f"{self.base}/search",
            data=urlencode(values).encode("utf-8"),
            method="POST",
        )
        with urlopen(request, timeout=2) as response:
            return (
                response.status,
                response.headers,
                response.read().decode("utf-8"),
            )

    def test_home_is_self_contained_post_search_surface(self):
        status, headers, body = self._text("/")
        self.assertEqual(status, 200)
        self.assertTrue(headers["Content-Type"].startswith("text/html"))
        self.assertIn("GoreeCloud Search", body)
        self.assertIn('action="/search"', body)
        self.assertIn('method="post"', body)
        self.assertIn('href="/assets/search.css"', body)
        self.assertIn('href="/opensearch.xml"', body)
        self.assertNotIn("<script", body.casefold())
        self.assertNotIn("autofocus", body.casefold())
        self.assertIn("style-src 'self'", headers["Content-Security-Policy"])
        self.assertIn("form-action 'self'", headers["Content-Security-Policy"])
        self.assertNotIn("script-src", headers["Content-Security-Policy"])
        self.assertIsNone(headers["Access-Control-Allow-Origin"])

    def test_post_search_escapes_query_and_provider_content(self):
        query = "<svg onload=alert('query')>"
        status, headers, body = self._post_form({"q": query})
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
        self.assertEqual(self.provider.queries[0].terms, ("<svg", "onload=alert('query')>"))

    def test_get_search_remains_available_for_opensearch_integration(self):
        status, _, body = self._text(
            "/search?q=" + quote("privacy search")
        )
        self.assertEqual(status, 200)
        self.assertIn("GoreeCloud Search", body)
        self.assertEqual(
            self.provider.queries[0].terms,
            ("privacy", "search"),
        )

    def test_local_opensearch_descriptor_uses_actual_loopback_port(self):
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
        self.assertIn(":focus-visible", body)
        self.assertNotIn("@import", body.casefold())
        self.assertNotIn("url(http", body.casefold())

    def test_invalid_post_form_returns_html_without_provider_call(self):
        request = Request(
            f"{self.base}/search",
            data=urlencode({"callback": "x"}).encode("utf-8"),
            method="POST",
        )
        with self.assertRaises(HTTPError) as context:
            urlopen(request, timeout=2)
        error = context.exception
        self.assertEqual(error.code, 400)
        self.assertTrue(
            error.headers["Content-Type"].startswith("text/html")
        )
        body = error.read().decode("utf-8")
        self.assertIn("Search unavailable", body)
        self.assertIn("unsupported query parameter", body)
        self.assertEqual(self.provider.queries, [])

    def test_search_form_rejects_wrong_content_type(self):
        request = Request(
            f"{self.base}/search",
            data=b"q=privacy",
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        with self.assertRaises(HTTPError) as context:
            urlopen(request, timeout=2)
        error = context.exception
        self.assertEqual(error.code, 415)
        self.assertTrue(
            error.headers["Content-Type"].startswith("text/html")
        )
        self.assertEqual(self.provider.queries, [])


class ContainerWebUITests(unittest.TestCase):
    def setUp(self) -> None:
        self.server = create_container_server(
            SearchCore(provider_adapters=(WebFakeProvider(),)),
            port=0,
            mode=SourceMode.EXTERNAL_ONLY,
            service_hostname="search.goreecloud.com",
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

    def test_container_opensearch_advertises_private_https_service(self):
        request = Request(
            f"{self.base}/opensearch.xml",
            headers={"Host": "search.goreecloud.com"},
        )
        with urlopen(request, timeout=2) as response:
            body = response.read().decode("utf-8")
        self.assertEqual(response.status, 200)
        self.assertIn(
            "https://search.goreecloud.com/search?q={searchTerms}",
            body,
        )
        self.assertIn(
            "<SearchForm>https://search.goreecloud.com/</SearchForm>",
            body,
        )


if __name__ == "__main__":
    unittest.main()
