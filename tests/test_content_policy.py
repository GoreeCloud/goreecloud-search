import unittest

from goreecloud_search import (
    ContentPolicyAction,
    ContentPolicyDecision,
    ContentPolicyEngine,
    ContentPolicyError,
    DomainPolicyHook,
    ProviderDescriptor,
    ProviderOrigin,
    ProviderSearchBatch,
    ResultCandidate,
    SafeSearchMode,
    SearchCategory,
    SearchCore,
)


class FakeSafeSearchHook:
    name = "fake-safesearch"
    supports_safe_search = True

    def evaluate(self, result, *, safe_search: SafeSearchMode):
        if safe_search is SafeSearchMode.OFF:
            return None
        if "sensitive" in result.title.casefold():
            return ContentPolicyDecision(
                hook=self.name,
                action=ContentPolicyAction.BLOCK,
                reason=f"Fixture classified as blocked for {safe_search.value} SafeSearch.",
            )
        return None


class BrokenHook:
    name = "broken"
    supports_safe_search = False

    def evaluate(self, result, *, safe_search: SafeSearchMode):
        raise RuntimeError("boom")


class SpoofingHook:
    name = "declared"
    supports_safe_search = False

    def evaluate(self, result, *, safe_search: SafeSearchMode):
        return ContentPolicyDecision(
            hook="other",
            action=ContentPolicyAction.BLOCK,
            reason="bad provenance",
        )


class PipelineProvider:
    def __init__(self) -> None:
        self.calls = 0
        self._descriptor = ProviderDescriptor(
            name="index",
            origin=ProviderOrigin.GOREECLOUD_INDEX,
            categories=frozenset({SearchCategory.GENERAL}),
        )

    @property
    def descriptor(self):
        return self._descriptor

    async def search(self, query, *, limit: int):
        self.calls += 1
        return ProviderSearchBatch(
            candidates=(
                ResultCandidate(
                    title="Sensitive Fixture",
                    url="https://blocked.example.net/page",
                    snippet="sensitive",
                    provider="index",
                ),
                ResultCandidate(
                    title="Allowed Result",
                    url="https://docs.example.com/page",
                    snippet="allowed",
                    provider="index",
                ),
            )[:limit]
        )


def normalized_results():
    provider = ProviderDescriptor(
        name="index",
        origin=ProviderOrigin.GOREECLOUD_INDEX,
        categories=frozenset({SearchCategory.GENERAL}),
    )
    core = SearchCore((provider,))
    return core.normalize(
        (
            ResultCandidate(
                title="Allowed Result",
                url="https://docs.example.com/page",
                snippet="allowed",
                provider="index",
            ),
            ResultCandidate(
                title="Sensitive Fixture",
                url="https://blocked.example.net/page",
                snippet="sensitive",
                provider="index",
            ),
        )
    )


class ContentPolicyTests(unittest.TestCase):
    def test_domain_blocklist_blocks_subdomains(self) -> None:
        results = normalized_results()
        report = ContentPolicyEngine(
            (DomainPolicyHook(blocked_domains=("example.net",)),)
        ).apply(results)
        self.assertEqual(report.blocked_count, 1)
        self.assertEqual(len(report.visible_results), 1)

    def test_domain_allowlist_blocks_outside_domains(self) -> None:
        report = ContentPolicyEngine(
            (DomainPolicyHook(allowed_domains=("example.com",)),)
        ).apply(normalized_results())
        self.assertEqual([r.title for r in report.visible_results], ["Allowed Result"])

    def test_non_off_safesearch_fails_if_no_hook_can_enforce_it(self) -> None:
        engine = ContentPolicyEngine((DomainPolicyHook(),))
        with self.assertRaises(ContentPolicyError):
            engine.apply(normalized_results(), safe_search=SafeSearchMode.STRICT)

    def test_safesearch_intent_is_enforced_by_declared_hook(self) -> None:
        report = ContentPolicyEngine((FakeSafeSearchHook(),)).apply(
            normalized_results(), safe_search=SafeSearchMode.MODERATE
        )
        self.assertTrue(report.safe_search_enforced)
        self.assertEqual([r.title for r in report.visible_results], ["Allowed Result"])

    def test_hook_failure_fails_closed_instead_of_silently_bypassing_policy(self) -> None:
        with self.assertRaises(ContentPolicyError):
            ContentPolicyEngine((BrokenHook(),)).apply(normalized_results())

    def test_hook_cannot_spoof_policy_decision_provenance(self) -> None:
        with self.assertRaises(ContentPolicyError):
            ContentPolicyEngine((SpoofingHook(),)).apply(normalized_results())

    def test_duplicate_hook_names_are_rejected(self) -> None:
        with self.assertRaises(ValueError):
            ContentPolicyEngine((DomainPolicyHook(), DomainPolicyHook()))

    def test_off_mode_is_truthfully_marked_enforced_without_classifier(self) -> None:
        report = ContentPolicyEngine().apply(
            normalized_results(), safe_search=SafeSearchMode.OFF
        )
        self.assertTrue(report.safe_search_enforced)
        self.assertEqual(report.blocked_count, 0)


class ContentPolicyPipelineTests(unittest.IsolatedAsyncioTestCase):
    async def test_domain_policy_filters_before_ranking(self) -> None:
        provider = PipelineProvider()
        core = SearchCore(
            provider_adapters=(provider,),
            content_policy_hooks=(DomainPolicyHook(blocked_domains=("example.net",)),),
        )
        response = await core.search("result", limit=10)
        self.assertEqual(provider.calls, 1)
        self.assertEqual(response.content_policy.blocked_count, 1)
        self.assertEqual([item.result.title for item in response.results], ["Allowed Result"])

    async def test_unenforceable_safesearch_fails_before_provider_execution(self) -> None:
        provider = PipelineProvider()
        core = SearchCore(
            provider_adapters=(provider,),
            content_policy_hooks=(DomainPolicyHook(),),
        )
        with self.assertRaises(ContentPolicyError):
            await core.search("result", safe_search=SafeSearchMode.STRICT)
        self.assertEqual(provider.calls, 0)

    async def test_declared_safesearch_hook_filters_pipeline(self) -> None:
        provider = PipelineProvider()
        core = SearchCore(
            provider_adapters=(provider,),
            content_policy_hooks=(FakeSafeSearchHook(),),
        )
        response = await core.search("result", safe_search=SafeSearchMode.MODERATE)
        self.assertEqual(provider.calls, 1)
        self.assertTrue(response.content_policy.safe_search_enforced)
        self.assertEqual([item.result.title for item in response.results], ["Allowed Result"])


if __name__ == "__main__":
    unittest.main()
