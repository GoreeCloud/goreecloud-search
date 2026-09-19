import unittest

from goreecloud_search import (
    NormalizedResult,
    ResultProvenance,
    ProviderOrigin,
    parse_query,
    rank_results,
)


def result(
    result_id: str,
    title: str,
    url: str,
    snippet: str,
    *,
    source_agreement: int = 1,
    language: str | None = None,
    indexed: bool = False,
    provider_rank: int | None = None,
) -> NormalizedResult:
    provenance = tuple(
        ResultProvenance(
            provider=f"source-{index}",
            origin=ProviderOrigin.GOREECLOUD_INDEX if indexed and index == 0 else ProviderOrigin.EXTERNAL,
            provider_rank=provider_rank,
            source_id=None,
            contract_version=None,
            indexed_by_goreecloud=indexed and index == 0,
            content_hash=None,
            last_crawled_at=None,
        )
        for index in range(source_agreement)
    )
    return NormalizedResult(
        result_id=result_id,
        title=title,
        url=url,
        canonical_url=url,
        snippet=snippet,
        published_at=None,
        content_type=None,
        language=language,
        source_agreement=source_agreement,
        provenance=provenance,
    )


class RankingTests(unittest.TestCase):
    def test_title_phrase_match_outranks_snippet_only_match(self) -> None:
        query = parse_query('privacy "search engine"')
        ranked = rank_results(
            query,
            [
                result("snippet", "Privacy tools", "https://b.example/", "A search engine for privacy"),
                result("title", "Private search engine", "https://a.example/", "Privacy tools"),
            ],
        )
        self.assertEqual(ranked[0].result.result_id, "title")
        self.assertIn("phrase_title_match", {signal.code for signal in ranked[0].signals})

    def test_explicit_site_filetype_and_language_add_explainable_signals(self) -> None:
        query = parse_query("report site:example.com filetype:pdf language:en")
        ranked = rank_results(
            query,
            [result("doc", "Report", "https://docs.example.com/report.pdf", "Annual report", language="en")],
        )
        codes = {signal.code for signal in ranked[0].signals}
        self.assertTrue({"site_restriction_match", "filetype_match", "language_match"}.issubset(codes))

    def test_source_agreement_bonus_is_bounded(self) -> None:
        query = parse_query("example")
        many = rank_results(query, [result("many", "Other", "https://many.example/", "", source_agreement=20)])[0]
        agreement = next(signal for signal in many.signals if signal.code == "source_agreement")
        self.assertEqual(agreement.score, 6.0)

    def test_index_provenance_is_visible_but_has_zero_weight(self) -> None:
        query = parse_query("example")
        ranked = rank_results(query, [result("indexed", "Other", "https://a.example/", "", indexed=True)])[0]
        signal = next(signal for signal in ranked.signals if signal.code == "indexed_by_goreecloud")
        self.assertEqual(signal.score, 0.0)

    def test_provider_rank_does_not_change_baseline_score(self) -> None:
        query = parse_query("example")
        first = result("a", "Example", "https://a.example/", "", provider_rank=1)
        second = result("b", "Example", "https://b.example/", "", provider_rank=999)
        ranked = rank_results(query, [second, first])
        self.assertEqual(ranked[0].score, ranked[1].score)
        self.assertEqual([item.result.result_id for item in ranked], ["a", "b"])


if __name__ == "__main__":
    unittest.main()
