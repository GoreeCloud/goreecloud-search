import unittest

from goreecloud_search import (
    BraveSearchProviderError,
    BraveWebSearchProvider,
    SearchCore,
    SourceMode,
    parse_query,
)


class FakeTransport:
    def __init__(self, payload):
        self.payload = payload
        self.calls = []

    async def search(self, request, *, api_key):
        self.calls.append((request, api_key))
        return self.payload


class BraveWebSearchProviderTests(unittest.IsolatedAsyncioTestCase):
    async def test_search_maps_bounded_provider_query_and_results(self):
        transport = FakeTransport(
            {
                "web": {
                    "results": [
                        {
                            "title": "Example",
                            "url": "https://example.com/page?utm_source=test",
                            "description": "Privacy search result",
                        }
                    ]
                }
            }
        )
        provider = BraveWebSearchProvider(
            api_key="secret-token",
            transport=transport,
            country="US",
            search_lang="en",
        )
        core = SearchCore(provider_adapters=(provider,))
        response = await core.search(
            'privacy "search engine" site:example.com -tracker '
            'source:brave lens:local',
            mode=SourceMode.EXTERNAL_ONLY,
            limit=5,
        )

        self.assertEqual(len(response.results), 1)
        request, api_key = transport.calls[0]
        self.assertEqual(api_key, "secret-token")
        self.assertEqual(request.count, 5)
        self.assertEqual(request.country, "US")
        self.assertEqual(request.search_lang, "en")
        self.assertEqual(request.safesearch, "moderate")
        self.assertIn('"search engine"', request.query)
        self.assertIn("site:example.com", request.query)
        self.assertIn("-tracker", request.query)
        self.assertNotIn("source:", request.query)
        self.assertNotIn("lens:", request.query)
        self.assertEqual(
            response.results[0].result.canonical_url,
            "https://example.com/page",
        )

    async def test_provider_caps_one_request_to_brave_maximum(self):
        transport = FakeTransport({"web": {"results": []}})
        provider = BraveWebSearchProvider(
            api_key="secret-token",
            transport=transport,
        )
        await provider.search(parse_query("privacy"), limit=99)
        request, _ = transport.calls[0]
        self.assertEqual(request.count, 20)

    async def test_open_ended_date_filter_fails_before_transport(self):
        transport = FakeTransport({"web": {"results": []}})
        provider = BraveWebSearchProvider(
            api_key="secret-token",
            transport=transport,
        )
        with self.assertRaises(BraveSearchProviderError):
            await provider.search(
                parse_query("privacy after:2026-01-01"),
                limit=10,
            )
        self.assertEqual(transport.calls, [])

    async def test_bounded_date_range_maps_to_freshness(self):
        transport = FakeTransport({"web": {"results": []}})
        provider = BraveWebSearchProvider(
            api_key="secret-token",
            transport=transport,
        )
        await provider.search(
            parse_query(
                "privacy after:2026-01-01 before:2026-02-01"
            ),
            limit=10,
        )
        request, _ = transport.calls[0]
        self.assertEqual(
            request.freshness,
            "2026-01-01to2026-02-01",
        )

    async def test_malformed_provider_items_degrade(self):
        transport = FakeTransport(
            {
                "web": {
                    "results": [
                        {
                            "title": "Missing URL",
                            "description": "bad",
                        },
                        {
                            "title": "Good",
                            "url": "https://example.org/",
                            "description": "ok",
                        },
                    ]
                }
            }
        )
        provider = BraveWebSearchProvider(
            api_key="secret-token",
            transport=transport,
        )
        batch = await provider.search(
            parse_query("privacy"),
            limit=10,
        )
        self.assertTrue(batch.degraded)
        self.assertEqual(len(batch.candidates), 1)
        self.assertEqual(len(batch.warnings), 1)

    def test_secret_not_stored_in_descriptor(self):
        transport = FakeTransport({"web": {"results": []}})
        provider = BraveWebSearchProvider(
            api_key="secret-token",
            transport=transport,
        )
        self.assertNotIn(
            "secret-token",
            repr(provider.descriptor),
        )


if __name__ == "__main__":
    unittest.main()
