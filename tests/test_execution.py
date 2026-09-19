import asyncio
import unittest

from goreecloud_search import (
    ExecutionPolicy,
    ProviderDescriptor,
    ProviderExecutionStatus,
    ProviderOrigin,
    QueryDisclosureBudget,
    ProviderSearchBatch,
    ResultCandidate,
    SearchAvailability,
    SearchCategory,
    SearchCore,
    SearchExecutor,
    SourceMode,
    parse_query,
    plan_sources,
)


class FakeProvider:
    def __init__(
        self,
        name: str,
        origin: ProviderOrigin,
        *,
        candidates: tuple[ResultCandidate, ...] = (),
        delay: float = 0.0,
        error: Exception | None = None,
        degraded: bool = False,
        warnings: tuple[str, ...] = (),
        tracker: dict[str, int] | None = None,
    ) -> None:
        self._descriptor = ProviderDescriptor(
            name=name,
            origin=origin,
            categories=frozenset({SearchCategory.GENERAL}),
            priority=10 if origin is not ProviderOrigin.EXTERNAL else 100,
        )
        self._candidates = candidates
        self._delay = delay
        self._error = error
        self._degraded = degraded
        self._warnings = warnings
        self._tracker = tracker
        self.calls = 0
        self.cancelled = False

    @property
    def descriptor(self) -> ProviderDescriptor:
        return self._descriptor

    async def search(self, query, *, limit: int) -> ProviderSearchBatch:
        self.calls += 1
        if self._tracker is not None:
            self._tracker["active"] = self._tracker.get("active", 0) + 1
            self._tracker["max"] = max(self._tracker.get("max", 0), self._tracker["active"])
        try:
            if self._delay:
                await asyncio.sleep(self._delay)
            if self._error is not None:
                raise self._error
            return ProviderSearchBatch(candidates=self._candidates[:limit], degraded=self._degraded, warnings=self._warnings)
        except asyncio.CancelledError:
            self.cancelled = True
            raise
        finally:
            if self._tracker is not None:
                self._tracker["active"] -= 1


def candidate(provider: str, slug: str, title: str | None = None) -> ResultCandidate:
    return ResultCandidate(title=title or slug.title(), url=f"https://example.com/{slug}", snippet=f"result for {slug}", provider=provider)


class ExecutionTests(unittest.IsolatedAsyncioTestCase):
    async def test_index_first_skips_fallback_when_primary_fills_target(self) -> None:
        native = FakeProvider("index", ProviderOrigin.GOREECLOUD_INDEX, candidates=(candidate("index", "one"), candidate("index", "two")))
        external = FakeProvider("external", ProviderOrigin.EXTERNAL, candidates=(candidate("external", "three"),))
        core = SearchCore(provider_adapters=(native, external))
        response = await core.search("result", mode=SourceMode.INDEX_FIRST, limit=2)
        self.assertEqual(external.calls, 0)
        self.assertFalse(response.execution.fallback_used)
        self.assertEqual(response.execution.availability, SearchAvailability.AVAILABLE)
        self.assertEqual({a.provider: a.status for a in response.execution.attempts}["external"], ProviderExecutionStatus.SKIPPED)
        self.assertEqual(len(response.results), 2)

    async def test_index_first_uses_fallback_to_fill_short_primary(self) -> None:
        native = FakeProvider("index", ProviderOrigin.GOREECLOUD_INDEX, candidates=(candidate("index", "one"),))
        external = FakeProvider("external", ProviderOrigin.EXTERNAL, candidates=(candidate("external", "two"),))
        core = SearchCore(provider_adapters=(native, external))
        response = await core.search("result", mode=SourceMode.INDEX_FIRST, limit=2)
        self.assertTrue(response.execution.fallback_used)
        self.assertEqual(external.calls, 1)
        self.assertEqual(response.execution.availability, SearchAvailability.AVAILABLE)
        self.assertEqual(len(response.results), 2)

    async def test_zero_disclosure_budget_prevents_external_fallback_execution(self) -> None:
        native = FakeProvider("index", ProviderOrigin.GOREECLOUD_INDEX, candidates=(candidate("index", "one"),))
        external = FakeProvider("external", ProviderOrigin.EXTERNAL, candidates=(candidate("external", "two"),))
        core = SearchCore(provider_adapters=(native, external))
        response = await core.search(
            "result", mode=SourceMode.INDEX_FIRST, limit=2,
            disclosure_budget=QueryDisclosureBudget(max_third_party_providers=0),
        )
        self.assertEqual(external.calls, 0)
        self.assertFalse(response.execution.fallback_used)
        self.assertEqual(response.plan.third_party_provider_count, 0)
        self.assertEqual(response.plan.third_party_providers_omitted, 1)
        self.assertEqual(len(response.results), 1)

    async def test_timeout_isolated_as_degraded_partial_result(self) -> None:
        fast = FakeProvider("fast", ProviderOrigin.GOREECLOUD_SERVICE, candidates=(candidate("fast", "fast"),))
        slow = FakeProvider("slow", ProviderOrigin.EXTERNAL, delay=0.2, candidates=(candidate("slow", "slow"),))
        core = SearchCore(provider_adapters=(fast, slow), execution_policy=ExecutionPolicy(per_provider_timeout_seconds=0.02, max_concurrency=2))
        response = await core.search("result", mode=SourceMode.FEDERATED, limit=10)
        self.assertEqual(response.execution.availability, SearchAvailability.DEGRADED)
        statuses = {a.provider: a.status for a in response.execution.attempts}
        self.assertEqual(statuses["fast"], ProviderExecutionStatus.SUCCESS)
        self.assertEqual(statuses["slow"], ProviderExecutionStatus.TIMEOUT)
        self.assertEqual(len(response.results), 1)

    async def test_provider_reported_degraded_state_is_preserved(self) -> None:
        provider = FakeProvider("index", ProviderOrigin.GOREECLOUD_INDEX, candidates=(candidate("index", "one"),), degraded=True, warnings=("partial shard availability",))
        core = SearchCore(provider_adapters=(provider,))
        response = await core.search("result", limit=10)
        attempt = response.execution.attempts[0]
        self.assertEqual(response.execution.availability, SearchAvailability.DEGRADED)
        self.assertEqual(attempt.status, ProviderExecutionStatus.DEGRADED)
        self.assertEqual(attempt.warnings, ("partial shard availability",))

    async def test_all_provider_errors_report_unavailable_without_raising(self) -> None:
        first = FakeProvider("one", ProviderOrigin.GOREECLOUD_SERVICE, error=RuntimeError("x"))
        second = FakeProvider("two", ProviderOrigin.EXTERNAL, error=RuntimeError("y"))
        core = SearchCore(provider_adapters=(first, second))
        response = await core.search("result", mode=SourceMode.FEDERATED, limit=10)
        self.assertEqual(response.execution.availability, SearchAvailability.UNAVAILABLE)
        self.assertEqual(response.results, ())
        self.assertTrue(all(a.status is ProviderExecutionStatus.ERROR for a in response.execution.attempts))

    async def test_goreecloud_only_never_executes_registered_external_provider(self) -> None:
        native = FakeProvider("native", ProviderOrigin.GOREECLOUD_SERVICE, candidates=(candidate("native", "one"),))
        external = FakeProvider("external", ProviderOrigin.EXTERNAL, candidates=(candidate("external", "two"),))
        core = SearchCore(provider_adapters=(native, external))
        response = await core.search("result", mode=SourceMode.GOREECLOUD_ONLY, limit=10)
        self.assertEqual(external.calls, 0)
        self.assertFalse(response.plan.third_party_query_disclosure)
        self.assertEqual(len(response.results), 1)

    async def test_max_concurrency_is_enforced(self) -> None:
        tracker = {"active": 0, "max": 0}
        providers = tuple(FakeProvider(f"p{i}", ProviderOrigin.GOREECLOUD_SERVICE, delay=0.03, candidates=(candidate(f"p{i}", f"r{i}"),), tracker=tracker) for i in range(4))
        query = parse_query("result")
        plan = plan_sources(query, SourceMode.FEDERATED, tuple(p.descriptor for p in providers))
        executor = SearchExecutor(providers, policy=ExecutionPolicy(per_provider_timeout_seconds=1.0, max_concurrency=2))
        report = await executor.execute(query, plan, limit=10)
        self.assertEqual(tracker["max"], 2)
        self.assertEqual(report.availability, SearchAvailability.AVAILABLE)

    async def test_outer_cancellation_propagates_and_cancels_provider_work(self) -> None:
        slow = FakeProvider("slow", ProviderOrigin.GOREECLOUD_SERVICE, delay=1.0, candidates=(candidate("slow", "one"),))
        query = parse_query("result")
        plan = plan_sources(query, SourceMode.FEDERATED, (slow.descriptor,))
        executor = SearchExecutor((slow,), policy=ExecutionPolicy(per_provider_timeout_seconds=5.0, max_concurrency=1))
        task = asyncio.create_task(executor.execute(query, plan, limit=10))
        await asyncio.sleep(0.02)
        task.cancel()
        with self.assertRaises(asyncio.CancelledError):
            await task
        self.assertTrue(slow.cancelled)

    async def test_spoofed_provider_provenance_is_rejected_as_failure(self) -> None:
        provider = FakeProvider("declared", ProviderOrigin.GOREECLOUD_SERVICE, candidates=(candidate("other-provider", "one"),))
        core = SearchCore(provider_adapters=(provider,))
        response = await core.search("result", mode=SourceMode.FEDERATED, limit=10)
        self.assertEqual(response.execution.availability, SearchAvailability.UNAVAILABLE)
        self.assertEqual(response.execution.attempts[0].status, ProviderExecutionStatus.ERROR)
        self.assertEqual(response.results, ())


if __name__ == "__main__":
    unittest.main()
