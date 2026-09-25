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

    def test_ipv6_result_url_preserves_brackets_and_default_port_semantics(self) -> None:
        self.assertEqual(
            "https://[2001:db8::1]/path",
            canonicalize_url("https://[2001:db8::1]:443/path"),
        )
        self.assertEqual(
            "http://[2001:db8::1]:8080/path",
            canonicalize_url("http://[2001:db8::1]:8080/path"),
        )

    def test_internationalized_result_host_is_canonicalized_to_ascii_alabel(self) -> None:
        self.assertEqual(
            "https://xn--bcher-kva.example/path",
            canonicalize_url("https://bücher.example/path"),
        )

    def test_canonical_ipv4_result_host_is_preserved(self) -> None:
        self.assertEqual(
            "https://127.0.0.1/path",
            canonicalize_url("https://127.0.0.1/path"),
        )

    def test_ambiguous_numeric_result_hosts_fail_closed(self) -> None:
        for url in (
            "http://127.1/path",
            "http://2130706433/path",
            "http://0177.0.0.1/path",
            "http://0x7f.0.0.1/path",
        ):
            with self.subTest(url=url):
                with self.assertRaises(ResultNormalizationError):
                    canonicalize_url(url)

    def test_invalid_dns_label_and_ipv6_zone_hosts_fail_closed(self) -> None:
        for url in (
            "https://exa_mple.com/path",
            "https://-example.com/path",
            "https://example-.com/path",
            "https://[fe80::1%25eth0]/path",
        ):
            with self.subTest(url=url):
                with self.assertRaises(ResultNormalizationError):
                    canonicalize_url(url)

    def test_result_url_port_zero_is_rejected(self) -> None:
        with self.assertRaises(ResultNormalizationError):
            canonicalize_url("https://example.com:0/path")

    def test_credential_bearing_result_url_is_rejected(self) -> None:
        with self.assertRaises(ResultNormalizationError):
            canonicalize_url("https://user:pass@example.com/private")

    def test_control_bearing_result_url_is_rejected_before_parsing(self) -> None:
        with self.assertRaises(ResultNormalizationError):
            canonicalize_url("https://example.com/private\npath")

    def test_bidi_control_bearing_result_url_is_rejected(self) -> None:
        with self.assertRaises(ResultNormalizationError):
            canonicalize_url("https://example.com/invoice\u202efdp.exe")

    def test_raw_whitespace_format_and_backslash_result_urls_are_rejected(self) -> None:
        for url in (
            "https://example.com/a b",
            "https://example.com/\u200bhidden",
            "https://example.com\\evil.com/path",
        ):
            with self.subTest(url=repr(url)):
                with self.assertRaises(ResultNormalizationError):
                    canonicalize_url(url)

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
