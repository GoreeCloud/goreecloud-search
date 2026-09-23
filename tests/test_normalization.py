import unittest

from goreecloud_search import (
    ProviderDescriptor,
    ProviderOrigin,
    ResultCandidate,
    ResultNormalizationError,
    SearchCategory,
    SearchCore,
    canonicalize_url,
)


class NormalizationTests(unittest.TestCase):
    def setUp(self) -> None:
        self.index = ProviderDescriptor(
            name="goreecloud-index",
            origin=ProviderOrigin.GOREECLOUD_INDEX,
            categories=frozenset({SearchCategory.GENERAL}),
            priority=10,
        )
        self.external = ProviderDescriptor(
            name="external-example",
            origin=ProviderOrigin.EXTERNAL,
            categories=frozenset({SearchCategory.GENERAL}),
            priority=100,
        )
        self.core = SearchCore((self.index, self.external))

    def test_canonicalize_url_removes_fragment_trackers_and_default_port(self) -> None:
        self.assertEqual(
            canonicalize_url(
                "HTTPS://Example.COM:443/path?utm_source=newsletter&a=1&fbclid=x#section"
            ),
            "https://example.com/path?a=1",
        )

    def test_credential_bearing_result_url_is_rejected(self) -> None:
        with self.assertRaises(ResultNormalizationError):
            canonicalize_url("https://user:pass@example.com/private")

    def test_control_bearing_result_url_is_rejected_before_parsing(self) -> None:
        with self.assertRaises(ResultNormalizationError):
            canonicalize_url("https://example.com/private\npath")

    def test_bidi_control_bearing_result_url_is_rejected(self) -> None:
        with self.assertRaises(ResultNormalizationError):
            canonicalize_url("https://example.com/invoice\u202efdp.exe")

    def test_safe_canonical_alias_cannot_launder_unsafe_open_url(self) -> None:
        with self.assertRaises(ResultNormalizationError):
            self.core.normalize(
                (
                    ResultCandidate(
                        title="Unsafe target",
                        url="https://user:pass@example.com/private",
                        canonical_url="https://example.com/private",
                        snippet="unsafe",
                        provider="external-example",
                    ),
                )
            )

    def test_provider_display_text_is_sanitized_and_whitespace_normalized(self) -> None:
        results = self.core.normalize(
            (
                ResultCandidate(
                    title="Invoice\u202efdp.exe\u2066\n ready",
                    url="https://example.com/invoice",
                    snippet="  First\tline\x00 second   line  ",
                    provider="external-example",
                ),
            )
        )

        self.assertEqual(results[0].title, "Invoice fdp.exe ready")
        self.assertEqual(results[0].snippet, "First line second line")
        self.assertNotIn("\u202e", results[0].title)
        self.assertNotIn("\x00", results[0].snippet)

    def test_provider_display_text_is_bounded(self) -> None:
        results = self.core.normalize(
            (
                ResultCandidate(
                    title="T" * 700,
                    url="https://example.com/long",
                    snippet="S" * 5000,
                    provider="external-example",
                ),
            )
        )

        self.assertEqual(len(results[0].title), 512)
        self.assertTrue(results[0].title.endswith("…"))
        self.assertEqual(len(results[0].snippet), 4096)
        self.assertTrue(results[0].snippet.endswith("…"))

    def test_title_that_becomes_empty_after_sanitization_fails_closed(self) -> None:
        with self.assertRaises(ResultNormalizationError):
            self.core.normalize(
                (
                    ResultCandidate(
                        title="\x00\n\u202e",
                        url="https://example.com/empty-title",
                        snippet="safe",
                        provider="external-example",
                    ),
                )
            )

    def test_same_canonical_url_merges_provider_provenance(self) -> None:
        results = self.core.normalize(
            (
                ResultCandidate(
                    title="First",
                    url="https://example.com/page?utm_medium=x",
                    snippet="one",
                    provider="goreecloud-index",
                    source_id="doc-1",
                    content_hash="hash-a",
                    provider_contract_version="goreecloud.search-index.v1",
                ),
                ResultCandidate(
                    title="Duplicate",
                    url="https://example.com/page#fragment",
                    snippet="two",
                    provider="external-example",
                ),
            )
        )
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0].canonical_url, "https://example.com/page")
        self.assertEqual(results[0].source_agreement, 2)
        self.assertEqual(
            {item.provider for item in results[0].provenance},
            {"goreecloud-index", "external-example"},
        )
        self.assertTrue(any(item.indexed_by_goreecloud for item in results[0].provenance))

    def test_content_hash_merges_mirror_urls(self) -> None:
        results = self.core.normalize(
            (
                ResultCandidate(
                    title="Original",
                    url="https://example.com/a",
                    snippet="one",
                    provider="goreecloud-index",
                    content_hash="same-content",
                ),
                ResultCandidate(
                    title="Mirror",
                    url="https://mirror.example.net/a",
                    snippet="two",
                    provider="external-example",
                    content_hash="same-content",
                ),
            )
        )
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0].source_agreement, 2)

    def test_distinct_urls_remain_distinct_without_shared_hash(self) -> None:
        results = self.core.normalize(
            (
                ResultCandidate("A", "https://example.com/a", "a", "goreecloud-index"),
                ResultCandidate("B", "https://example.com/b", "b", "external-example"),
            )
        )
        self.assertEqual(len(results), 2)

    def test_undeclared_provider_fails_closed(self) -> None:
        with self.assertRaises(ResultNormalizationError):
            self.core.normalize(
                (ResultCandidate("A", "https://example.com/a", "a", "unknown"),)
            )


if __name__ == "__main__":
    unittest.main()
