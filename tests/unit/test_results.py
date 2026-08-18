# SPDX-License-Identifier: AGPL-3.0-or-later
# pylint: disable=missing-module-docstring,disable=missing-class-docstring,invalid-name

from searx.result_types import LegacyResult
from searx.results import ResultContainer, calculate_score, diversify_top_domains
from tests import SearxTestCase


class ResultContainerTestCase(SearxTestCase):
    # pylint: disable=use-dict-literal

    TEST_SETTINGS = "test_result_container.yml"

    @staticmethod
    def _rank_result(
        url: str,
        title: str,
        content: str = "",
        engines: set[str] | None = None,
        positions: list[int] | None = None,
        score: float = 0.0,
    ) -> LegacyResult:
        result = LegacyResult(url=url, title=title, content=content, engine="google")
        result.normalize_result_fields()
        result.engines = engines or {"google"}
        result.positions = positions or [1]
        result.score = score
        return result

    def test_empty(self):
        container = ResultContainer()
        self.assertEqual(container.get_ordered_results(), [])

    def test_one_result(self):
        result = dict(url="https://example.org", title="title ..", content="Lorem ..")
        container = ResultContainer()
        container.extend("google", [result])
        container.close()
        self.assertEqual(len(container.get_ordered_results()), 1)

        res = LegacyResult(result)
        res.normalize_result_fields()
        self.assertIn(res, container.get_ordered_results())

    def test_one_suggestion(self):
        result = dict(suggestion="lorem ipsum ..")
        container = ResultContainer()
        container.extend("duckduckgo", [result])
        container.close()
        self.assertEqual(len(container.get_ordered_results()), 0)
        self.assertEqual(len(container.suggestions), 1)
        self.assertIn(result["suggestion"], container.suggestions)

    def test_merge_url_result(self):
        result = LegacyResult(
            url="https://example.org", title="very long title, lorem ipsum", content="Lorem ipsum dolor sit amet .."
        )
        result.normalize_result_fields()
        eng1 = dict(url=result.url, title="short title", content=result.content, engine="google")
        eng2 = dict(url="http://example.org", title=result.title, content="lorem ipsum", engine="duckduckgo")

        container = ResultContainer()
        container.extend(None, [eng1, eng2])
        container.close()

        result_list = container.get_ordered_results()
        self.assertEqual(len(container.get_ordered_results()), 1)
        self.assertIn(result, result_list)
        self.assertEqual(result_list[0].title, result.title)
        self.assertEqual(result_list[0].content, result.content)

    def test_exact_title_relevance_breaks_similar_provider_positions(self):
        query = "goreecloud browser privacy"
        exact = self._rank_result(
            "https://docs.example.org/goreecloud-browser",
            "GoreeCloud Browser Privacy",
            "Browser privacy documentation",
            positions=[2],
        )
        weak = self._rank_result(
            "https://example.net/browser",
            "General browser article",
            "A long article that happens to mention GoreeCloud Browser privacy near the end.",
            positions=[2],
        )
        self.assertGreater(calculate_score(exact, "", query), calculate_score(weak, "", query))

    def test_multi_engine_consensus_receives_bounded_bonus(self):
        query = "private metasearch"
        consensus = self._rank_result(
            "https://example.org/search",
            "Private metasearch",
            engines={"google", "duckduckgo"},
            positions=[3, 4],
        )
        single = self._rank_result(
            "https://example.net/search",
            "Private metasearch",
            engines={"google"},
            positions=[3],
        )
        self.assertGreater(calculate_score(consensus, "", query), calculate_score(single, "", query))

    def test_high_and_low_priority_semantics_are_preserved(self):
        result = self._rank_result("https://example.org", "Example", positions=[50])
        normal = calculate_score(result, "", "example")
        self.assertGreater(calculate_score(result, "high", "example"), normal)
        self.assertEqual(calculate_score(result, "low", "example"), 0.0)

    def test_domain_diversity_limits_first_viewport_monopoly(self):
        results = [
            self._rank_result("https://same.example/a", "A", score=10),
            self._rank_result("https://same.example/b", "B", score=9),
            self._rank_result("https://same.example/c", "C", score=8),
            self._rank_result("https://other.example/a", "D", score=7),
            self._rank_result("https://third.example/a", "E", score=6),
        ]
        diversified = diversify_top_domains(results, window=4, max_per_host=2)
        first_four_hosts = [item.parsed_url.hostname for item in diversified[:4]]
        self.assertLessEqual(first_four_hosts.count("same.example"), 2)
        self.assertIn("other.example", first_four_hosts)
        self.assertEqual(len(diversified), len(results))

    def test_query_is_kept_only_in_result_container_memory(self):
        container = ResultContainer(query="synthetic ranking query")
        self.assertEqual(container.query, "synthetic ranking query")
        self.assertFalse(hasattr(container, "query_history"))
