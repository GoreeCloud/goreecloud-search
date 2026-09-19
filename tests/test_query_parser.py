from datetime import date
import unittest

from goreecloud_search import QueryParseError, SearchCategory, parse_query


class QueryParserTests(unittest.TestCase):
    def test_parses_terms_phrases_and_filters(self) -> None:
        parsed = parse_query(
            'privacy "search engine" site:example.com -domain:ads.example '
            'filetype:pdf after:2026-01-01 before:2026-09-16 '
            'language:en region:us source:index category:docs lens:official -tracking'
        )
        self.assertEqual(parsed.terms, ("privacy",))
        self.assertEqual(parsed.phrases, ("search engine",))
        self.assertEqual(parsed.excluded_terms, ("tracking",))
        self.assertEqual(parsed.filters.sites, ("example.com",))
        self.assertEqual(parsed.filters.excluded_domains, ("ads.example",))
        self.assertEqual(parsed.filters.filetypes, ("pdf",))
        self.assertEqual(parsed.filters.after, date(2026, 1, 1))
        self.assertEqual(parsed.filters.before, date(2026, 9, 16))
        self.assertEqual(parsed.filters.language, "en")
        self.assertEqual(parsed.filters.region, "US")
        self.assertEqual(parsed.filters.sources, ("index",))
        self.assertEqual(parsed.filters.category, SearchCategory.DOCUMENTATION)
        self.assertEqual(parsed.filters.lens, "official")

    def test_deduplicates_repeatable_filters(self) -> None:
        parsed = parse_query("query site:example.com site:example.com ext:.PDF filetype:pdf")
        self.assertEqual(parsed.filters.sites, ("example.com",))
        self.assertEqual(parsed.filters.filetypes, ("pdf",))

    def test_rejects_invalid_date_range(self) -> None:
        with self.assertRaises(QueryParseError):
            parse_query("query after:2026-09-16 before:2026-01-01")

    def test_rejects_unknown_category(self) -> None:
        with self.assertRaises(QueryParseError):
            parse_query("query category:not-a-category")

    def test_rejects_empty_query(self) -> None:
        with self.assertRaises(QueryParseError):
            parse_query("   ")

    def test_allows_site_only_query(self) -> None:
        parsed = parse_query("site:example.com")
        self.assertEqual(parsed.filters.sites, ("example.com",))


if __name__ == "__main__":
    unittest.main()
