import asyncio
import unittest

from goreecloud_search import (
    INDEX_CONTRACT_VERSION,
    GoreeCloudIndexProvider,
    IndexCapabilities,
    IndexContractError,
    IndexDocumentCandidate,
    IndexSearchRequest,
    IndexSearchResponse,
    SearchCategory,
    parse_query,
)


class FakeIndexTransport:
    def __init__(self, response: IndexSearchResponse) -> None:
        self.response = response
        self.requests: list[IndexSearchRequest] = []

    async def search(self, request: IndexSearchRequest) -> IndexSearchResponse:
        self.requests.append(request)
        return self.response


class IndexContractTests(unittest.TestCase):
    def test_request_preserves_structured_query_filters(self) -> None:
        query = parse_query(
            'privacy "search engine" site:example.com -domain:ads.example '
            'filetype:pdf after:2026-01-01 before:2026-09-16 '
            'language:en region:us category:docs lens:official -tracking'
        )
        request = IndexSearchRequest.from_query(query, page_size=25, cursor="next")
        self.assertEqual(request.contract_version, INDEX_CONTRACT_VERSION)
        self.assertEqual(request.terms, ("privacy",))
        self.assertEqual(request.phrases, ("search engine",))
        self.assertEqual(request.excluded_terms, ("tracking",))
        self.assertEqual(request.sites, ("example.com",))
        self.assertEqual(request.before, "2026-09-16")
        self.assertEqual(request.after, "2026-01-01")
        self.assertEqual(request.category, SearchCategory.DOCUMENTATION)
        self.assertEqual(request.page_size, 25)
        self.assertEqual(request.cursor, "next")

    def test_capabilities_reject_incompatible_contract(self) -> None:
        with self.assertRaises(IndexContractError):
            IndexCapabilities(
                contract_version="goreecloud.search-index.v999",
                categories=frozenset({SearchCategory.GENERAL}),
            )

    def test_provider_maps_index_response_and_caps_page_size(self) -> None:
        capabilities = IndexCapabilities(
            contract_version=INDEX_CONTRACT_VERSION,
            categories=frozenset({SearchCategory.GENERAL}),
            max_page_size=20,
        )
        response = IndexSearchResponse(
            contract_version=INDEX_CONTRACT_VERSION,
            candidates=(
                IndexDocumentCandidate(
                    document_id="doc-1",
                    title="Private Search",
                    url="https://example.com/article?utm_source=test",
                    canonical_url="https://example.com/article",
                    snippet="A result",
                    rank=3,
                    content_hash="ABC123",
                    language="en",
                    last_crawled_at="2026-09-16T00:00:00Z",
                ),
            ),
        )
        transport = FakeIndexTransport(response)
        provider = GoreeCloudIndexProvider(transport, capabilities)
        results = asyncio.run(provider.search(parse_query("private search"), limit=50))

        self.assertEqual(len(transport.requests), 1)
        self.assertEqual(transport.requests[0].page_size, 20)
        self.assertEqual(results[0].source_id, "doc-1")
        self.assertEqual(results[0].canonical_url, "https://example.com/article")
        self.assertEqual(results[0].content_hash, "ABC123")
        self.assertEqual(results[0].provider_contract_version, INDEX_CONTRACT_VERSION)

    def test_provider_rejects_response_version_mismatch(self) -> None:
        capabilities = IndexCapabilities(
            contract_version=INDEX_CONTRACT_VERSION,
            categories=frozenset({SearchCategory.GENERAL}),
        )
        transport = FakeIndexTransport(
            IndexSearchResponse(contract_version="old.contract", candidates=())
        )
        provider = GoreeCloudIndexProvider(transport, capabilities)
        with self.assertRaises(IndexContractError):
            asyncio.run(provider.search(parse_query("test"), limit=10))


if __name__ == "__main__":
    unittest.main()
