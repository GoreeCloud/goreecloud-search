import unittest

from goreecloud_search import (
    Lens,
    LensError,
    LensRegistry,
    LensRule,
    LensRuleAction,
    LensRuleTarget,
    NormalizedResult,
    ProviderDescriptor,
    ProviderOrigin,
    ProviderSearchBatch,
    ResultCandidate,
    ResultProvenance,
    SearchCategory,
    SearchCore,
    apply_lens,
    parse_query,
    rank_results,
)


def result(result_id: str, url: str, *, language: str | None = None, title: str = "Result") -> NormalizedResult:
    return NormalizedResult(
        result_id=result_id,
        title=title,
        url=url,
        canonical_url=url,
        snippet="result",
        published_at=None,
        content_type=None,
        language=language,
        source_agreement=1,
        provenance=(
            ResultProvenance(
                provider="index",
                origin=ProviderOrigin.GOREECLOUD_INDEX,
                provider_rank=None,
                source_id=None,
                contract_version=None,
                indexed_by_goreecloud=True,
                content_hash=None,
                last_crawled_at=None,
            ),
        ),
    )


class LensTests(unittest.TestCase):
    def test_domain_boost_reranks_and_explains(self) -> None:
        query = parse_query("result")
        baseline = rank_results(
            query,
            (
                result("a", "https://a.example/page", title="Result"),
                result("b", "https://docs.example/page", title="Result"),
            ),
        )
        lens = Lens(
            "Development",
            (LensRule(LensRuleTarget.DOMAIN, "docs.example", LensRuleAction.BOOST, 12),),
        )
        report = apply_lens(baseline, lens)
        self.assertEqual(report.results[0].result.result_id, "b")
        signal = next(s for s in report.results[0].signals if s.code == "lens_domain_boost")
        self.assertEqual(signal.score, 12)
        self.assertIn("Development", signal.explanation)

    def test_domain_lower_is_negative_and_transparent(self) -> None:
        query = parse_query("result")
        baseline = rank_results(query, (result("a", "https://a.example/"), result("b", "https://b.example/")))
        lens = Lens("Lower A", (LensRule(LensRuleTarget.DOMAIN, "a.example", LensRuleAction.LOWER, 4),))
        report = apply_lens(baseline, lens)
        self.assertEqual(report.results[0].result.result_id, "b")
        lowered = next(item for item in report.results if item.result.result_id == "a")
        self.assertEqual(next(s for s in lowered.signals if s.code == "lens_domain_lower").score, -4)

    def test_exclusion_matches_subdomains_and_removes_result(self) -> None:
        query = parse_query("result")
        baseline = rank_results(query, (result("a", "https://news.example.net/page"), result("b", "https://example.org/")))
        lens = Lens("No Example", (LensRule(LensRuleTarget.DOMAIN, "example.net", LensRuleAction.EXCLUDE),))
        report = apply_lens(baseline, lens)
        self.assertEqual([item.result.result_id for item in report.results], ["b"])
        self.assertEqual(report.excluded_count, 1)
        self.assertEqual(report.exclusions[0].result.result_id, "a")

    def test_filetype_and_language_rules_apply(self) -> None:
        query = parse_query("result")
        baseline = rank_results(
            query,
            (
                result("pdf", "https://docs.example/report.pdf", language="en"),
                result("html", "https://docs.example/page", language="fr"),
            ),
        )
        lens = Lens(
            "Docs",
            (
                LensRule(LensRuleTarget.FILETYPE, "pdf", LensRuleAction.BOOST, 3),
                LensRule(LensRuleTarget.LANGUAGE, "en", LensRuleAction.BOOST, 2),
            ),
        )
        report = apply_lens(baseline, lens)
        self.assertEqual(report.results[0].result.result_id, "pdf")
        codes = {s.code for s in report.results[0].signals}
        self.assertTrue({"lens_filetype_boost", "lens_language_boost"}.issubset(codes))

    def test_duplicate_lens_names_are_rejected_case_insensitively(self) -> None:
        first = Lens("Dev", (LensRule(LensRuleTarget.DOMAIN, "a.example", LensRuleAction.BOOST),))
        second = Lens("dev", (LensRule(LensRuleTarget.DOMAIN, "b.example", LensRuleAction.BOOST),))
        with self.assertRaises(LensError):
            LensRegistry((first, second))

    def test_invalid_weights_are_rejected(self) -> None:
        with self.assertRaises(LensError):
            LensRule(LensRuleTarget.DOMAIN, "a.example", LensRuleAction.BOOST, 0)
        with self.assertRaises(LensError):
            LensRule(LensRuleTarget.DOMAIN, "a.example", LensRuleAction.LOWER, 101)


class RecordingProvider:
    def __init__(self) -> None:
        self.calls = 0
        self.last_lens = "unset"
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
        self.last_lens = query.filters.lens
        return ProviderSearchBatch(
            candidates=(
                ResultCandidate(
                    title="Result",
                    url="https://docs.example/page",
                    snippet="result",
                    provider="index",
                ),
            )[:limit]
        )


class LensPipelineTests(unittest.IsolatedAsyncioTestCase):
    async def test_requested_lens_is_applied_but_not_disclosed_to_provider(self) -> None:
        provider = RecordingProvider()
        lens = Lens("Development", (LensRule(LensRuleTarget.DOMAIN, "docs.example", LensRuleAction.BOOST, 5),))
        core = SearchCore(provider_adapters=(provider,), lenses=(lens,))
        response = await core.search("result lens:Development")
        self.assertEqual(provider.calls, 1)
        self.assertIsNone(provider.last_lens)
        self.assertEqual(response.query.filters.lens, "Development")
        self.assertEqual(response.lens.lens_name, "Development")
        self.assertIn("lens_domain_boost", {s.code for s in response.results[0].signals})

    async def test_unknown_lens_fails_before_provider_execution(self) -> None:
        provider = RecordingProvider()
        core = SearchCore(provider_adapters=(provider,))
        with self.assertRaises(LensError):
            await core.search("result lens:missing")
        self.assertEqual(provider.calls, 0)

    async def test_lens_exclusion_occurs_after_content_policy_and_before_final_results(self) -> None:
        provider = RecordingProvider()
        lens = Lens("Exclude Docs", (LensRule(LensRuleTarget.DOMAIN, "docs.example", LensRuleAction.EXCLUDE),))
        core = SearchCore(provider_adapters=(provider,), lenses=(lens,))
        response = await core.search("result lens:exclude-docs")
        self.assertEqual(response.lens.excluded_count, 1)
        self.assertEqual(response.results, ())


if __name__ == "__main__":
    unittest.main()
