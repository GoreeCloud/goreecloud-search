import unittest

from goreecloud_search import (
    INDEX_ORIGINATED_DELEGATION_CONTRACT_VERSION,
    INDEX_ORIGINATED_DELEGATION_MODE,
    INDEX_ORIGINATED_FALLBACK_ALLOWED,
    INDEX_ORIGINATED_INDEX_PROVIDER_REENTRY_ALLOWED,
    ProviderDescriptor,
    ProviderOrigin,
    ProviderSearchBatch,
    QueryDisclosureBudget,
    ResultCandidate,
    SearchAvailability,
    SearchCategory,
    SearchCore,
    SourceMode,
    SourcePlanningError,
)


class FakeProvider:
    def __init__(
        self,
        name: str,
        origin: ProviderOrigin,
        *,
        fail: bool = False,
    ) -> None:
        self._descriptor = ProviderDescriptor(
            name=name,
            origin=origin,
            categories=frozenset({SearchCategory.GENERAL}),
            priority=10,
        )
        self.fail = fail
        self.calls = 0

    @property
    def descriptor(self) -> ProviderDescriptor:
        return self._descriptor

    async def search(self, query, *, limit: int) -> ProviderSearchBatch:
        self.calls += 1
        if self.fail:
            raise RuntimeError("provider failed")
        return ProviderSearchBatch(
            candidates=(
                ResultCandidate(
                    title=f"{self._descriptor.name} result",
                    url="https://example.com/result",
                    snippet="cycle-safe result",
                    provider=self._descriptor.name,
                ),
            )[:limit],
        )


class IndexOriginatedDelegationTests(unittest.IsolatedAsyncioTestCase):
    def test_contract_constants_are_explicit_and_fail_closed(self) -> None:
        self.assertEqual(
            INDEX_ORIGINATED_DELEGATION_CONTRACT_VERSION,
            "goreecloud.search-index-delegation.v1",
        )
        self.assertEqual(INDEX_ORIGINATED_DELEGATION_MODE, "external_only")
        self.assertFalse(INDEX_ORIGINATED_INDEX_PROVIDER_REENTRY_ALLOWED)
        self.assertFalse(INDEX_ORIGINATED_FALLBACK_ALLOWED)

    async def test_index_originated_search_executes_only_external_provider(self) -> None:
        index = FakeProvider("index", ProviderOrigin.GOREECLOUD_INDEX)
        service = FakeProvider("service", ProviderOrigin.GOREECLOUD_SERVICE)
        local = FakeProvider("local", ProviderOrigin.LOCAL)
        external = FakeProvider("external", ProviderOrigin.EXTERNAL)
        core = SearchCore(provider_adapters=(index, service, local, external))

        response = await core.search_from_index("current weather", limit=5)

        self.assertEqual(response.plan.mode, SourceMode.EXTERNAL_ONLY)
        self.assertEqual(
            [(step.provider, step.stage, step.origin) for step in response.plan.steps],
            [("external", "primary", ProviderOrigin.EXTERNAL)],
        )
        self.assertFalse(response.execution.fallback_used)
        self.assertEqual(index.calls, 0)
        self.assertEqual(service.calls, 0)
        self.assertEqual(local.calls, 0)
        self.assertEqual(external.calls, 1)
        self.assertEqual(response.execution.availability, SearchAvailability.AVAILABLE)

    async def test_index_originated_search_cannot_fall_back_to_index_after_external_failure(self) -> None:
        index = FakeProvider("index", ProviderOrigin.GOREECLOUD_INDEX)
        external = FakeProvider("external", ProviderOrigin.EXTERNAL, fail=True)
        core = SearchCore(provider_adapters=(index, external))

        response = await core.search_from_index("current weather", limit=5)

        self.assertEqual(external.calls, 1)
        self.assertEqual(index.calls, 0)
        self.assertFalse(response.execution.fallback_used)
        self.assertEqual(response.execution.availability, SearchAvailability.UNAVAILABLE)

    async def test_index_originated_search_fails_before_dispatch_when_no_external_provider_exists(self) -> None:
        index = FakeProvider("index", ProviderOrigin.GOREECLOUD_INDEX)
        core = SearchCore(provider_adapters=(index,))

        with self.assertRaises(SourcePlanningError):
            await core.search_from_index("current weather", limit=5)

        self.assertEqual(index.calls, 0)

    async def test_index_source_filter_cannot_widen_index_originated_scope(self) -> None:
        index = FakeProvider("index", ProviderOrigin.GOREECLOUD_INDEX)
        external = FakeProvider("external", ProviderOrigin.EXTERNAL)
        core = SearchCore(provider_adapters=(index, external))

        with self.assertRaises(SourcePlanningError):
            await core.search_from_index("weather source:index", limit=5)

        self.assertEqual(index.calls, 0)
        self.assertEqual(external.calls, 0)

    async def test_zero_disclosure_budget_blocks_index_originated_external_execution(self) -> None:
        external = FakeProvider("external", ProviderOrigin.EXTERNAL)
        core = SearchCore(provider_adapters=(external,))

        with self.assertRaises(SourcePlanningError):
            await core.search_from_index(
                "current weather",
                limit=5,
                disclosure_budget=QueryDisclosureBudget(max_third_party_providers=0),
            )

        self.assertEqual(external.calls, 0)


if __name__ == "__main__":
    unittest.main()
