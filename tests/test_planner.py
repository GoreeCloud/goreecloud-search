import unittest

from goreecloud_search import (
    ProviderDescriptor,
    ProviderOrigin,
    QueryDisclosureBudget,
    SearchCategory,
    SearchCore,
    SourceMode,
    SourcePlanningError,
    parse_query,
    plan_sources,
)


def provider(name: str, origin: ProviderOrigin, *, local_only: bool = False, categories: frozenset[SearchCategory] | None = None, priority: int = 100) -> ProviderDescriptor:
    return ProviderDescriptor(name=name, origin=origin, categories=categories or frozenset({SearchCategory.GENERAL}), local_only=local_only, priority=priority)


class PlannerTests(unittest.TestCase):
    def setUp(self) -> None:
        self.providers = (
            provider("index", ProviderOrigin.GOREECLOUD_INDEX, priority=10),
            provider("local", ProviderOrigin.LOCAL, local_only=True, priority=20),
            provider("external-a", ProviderOrigin.EXTERNAL, priority=30),
        )

    def test_index_first_uses_external_as_fallback(self) -> None:
        plan = plan_sources(parse_query("privacy search"), SourceMode.INDEX_FIRST, self.providers)
        self.assertEqual([(s.provider, s.stage) for s in plan.steps], [("index", "primary"), ("local", "primary"), ("external-a", "fallback")])
        self.assertTrue(plan.third_party_query_disclosure)
        self.assertEqual(plan.third_party_provider_count, 1)

    def test_goreecloud_only_excludes_external(self) -> None:
        plan = plan_sources(parse_query("privacy search"), SourceMode.GOREECLOUD_ONLY, self.providers)
        self.assertEqual([s.provider for s in plan.steps], ["index", "local"])
        self.assertFalse(plan.third_party_query_disclosure)
        self.assertEqual(plan.third_party_provider_count, 0)

    def test_offline_local_uses_only_local_provider(self) -> None:
        plan = plan_sources(parse_query("privacy search"), SourceMode.OFFLINE_LOCAL, self.providers)
        self.assertEqual([s.provider for s in plan.steps], ["local"])
        self.assertFalse(plan.third_party_query_disclosure)

    def test_external_only_uses_external_provider(self) -> None:
        plan = plan_sources(parse_query("privacy search"), SourceMode.EXTERNAL_ONLY, self.providers)
        self.assertEqual([s.provider for s in plan.steps], ["external-a"])
        self.assertTrue(plan.third_party_query_disclosure)

    def test_source_filter_is_enforced(self) -> None:
        with self.assertRaises(SourcePlanningError):
            plan_sources(parse_query("privacy source:external-a"), SourceMode.GOREECLOUD_ONLY, self.providers)

    def test_category_capability_is_enforced(self) -> None:
        news = provider("news-index", ProviderOrigin.GOREECLOUD_INDEX, categories=frozenset({SearchCategory.NEWS}))
        plan = plan_sources(parse_query("privacy category:news"), SourceMode.GOREECLOUD_ONLY, (news,))
        self.assertEqual([s.provider for s in plan.steps], ["news-index"])

    def test_disclosure_budget_caps_external_providers_by_priority(self) -> None:
        providers = self.providers + (
            provider("external-b", ProviderOrigin.EXTERNAL, priority=40),
            provider("external-c", ProviderOrigin.EXTERNAL, priority=50),
        )
        plan = plan_sources(
            parse_query("privacy search"), SourceMode.FEDERATED, providers,
            disclosure_budget=QueryDisclosureBudget(max_third_party_providers=2),
        )
        self.assertEqual([s.provider for s in plan.steps], ["index", "local", "external-a", "external-b"])
        self.assertEqual(plan.third_party_provider_count, 2)
        self.assertEqual(plan.disclosure_budget, 2)
        self.assertEqual(plan.third_party_providers_omitted, 1)

    def test_zero_disclosure_budget_keeps_native_sources_and_omits_external(self) -> None:
        plan = plan_sources(
            parse_query("privacy search"), SourceMode.INDEX_FIRST, self.providers,
            disclosure_budget=QueryDisclosureBudget(max_third_party_providers=0),
        )
        self.assertEqual([s.provider for s in plan.steps], ["index", "local"])
        self.assertFalse(plan.third_party_query_disclosure)
        self.assertEqual(plan.third_party_provider_count, 0)
        self.assertEqual(plan.third_party_providers_omitted, 1)

    def test_zero_disclosure_budget_blocks_external_only_plan(self) -> None:
        with self.assertRaises(SourcePlanningError) as ctx:
            plan_sources(
                parse_query("privacy search"), SourceMode.EXTERNAL_ONLY, self.providers,
                disclosure_budget=QueryDisclosureBudget(max_third_party_providers=0),
            )
        self.assertIn("disclosure_budget=0", str(ctx.exception))

    def test_unlimited_budget_preserves_existing_behavior(self) -> None:
        providers = self.providers + (provider("external-b", ProviderOrigin.EXTERNAL, priority=40),)
        plan = plan_sources(
            parse_query("privacy search"), SourceMode.FEDERATED, providers,
            disclosure_budget=QueryDisclosureBudget(),
        )
        self.assertEqual(plan.third_party_provider_count, 2)
        self.assertEqual(plan.third_party_providers_omitted, 0)
        self.assertIsNone(plan.disclosure_budget)

    def test_negative_disclosure_budget_is_rejected(self) -> None:
        with self.assertRaises(ValueError):
            QueryDisclosureBudget(max_third_party_providers=-1)

    def test_search_core_exposes_budget_to_planning(self) -> None:
        providers = self.providers + (provider("external-b", ProviderOrigin.EXTERNAL, priority=40),)
        core = SearchCore(providers)
        _, plan = core.plan(
            "privacy search", mode=SourceMode.FEDERATED,
            disclosure_budget=QueryDisclosureBudget(max_third_party_providers=1),
        )
        self.assertEqual(plan.third_party_provider_count, 1)
        self.assertEqual(plan.third_party_providers_omitted, 1)


if __name__ == "__main__":
    unittest.main()
