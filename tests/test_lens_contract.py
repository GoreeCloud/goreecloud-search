import json
import unittest

from goreecloud_search import (
    LENS_FORMAT_VERSION,
    MAX_LENS_DOCUMENT_BYTES,
    Lens,
    LensFormatError,
    LensRule,
    LensRuleAction,
    LensRuleTarget,
    export_lens,
    import_lens,
)


class LensContractTests(unittest.TestCase):
    def setUp(self) -> None:
        self.lens = Lens(
            name="Development",
            description="Prefer official developer material.",
            rules=(
                LensRule(LensRuleTarget.DOMAIN, "docs.example", LensRuleAction.BOOST, 8),
                LensRule(LensRuleTarget.FILETYPE, "pdf", LensRuleAction.LOWER, 2),
                LensRule(LensRuleTarget.DOMAIN, "spam.example", LensRuleAction.EXCLUDE),
            ),
        )

    def test_round_trip_is_lossless_and_deterministic(self) -> None:
        first = export_lens(self.lens)
        imported = import_lens(first)
        second = export_lens(imported)
        self.assertEqual(imported, self.lens)
        self.assertEqual(first, second)
        self.assertTrue(first.endswith("\n"))

    def test_export_contains_versioned_plain_configuration_only(self) -> None:
        document = json.loads(export_lens(self.lens))
        self.assertEqual(document["format_version"], LENS_FORMAT_VERSION)
        self.assertEqual(set(document), {"format_version", "name", "description", "rules"})
        self.assertNotIn("provider", export_lens(self.lens).casefold())

    def test_rejects_unknown_format_version(self) -> None:
        payload = export_lens(self.lens).replace(LENS_FORMAT_VERSION, "goreecloud.search-lens.v99")
        with self.assertRaises(LensFormatError):
            import_lens(payload)

    def test_rejects_unknown_fields(self) -> None:
        document = json.loads(export_lens(self.lens))
        document["tracking_id"] = "not-allowed"
        with self.assertRaises(LensFormatError):
            import_lens(json.dumps(document))

    def test_rejects_duplicate_json_keys(self) -> None:
        payload = '{"format_version":"goreecloud.search-lens.v1","name":"A","name":"B","description":null,"rules":[{"target":"domain","value":"example.com","action":"boost","weight":1}]}'
        with self.assertRaises(LensFormatError):
            import_lens(payload)

    def test_rejects_nonfinite_weight(self) -> None:
        payload = '{"format_version":"goreecloud.search-lens.v1","name":"A","description":null,"rules":[{"target":"domain","value":"example.com","action":"boost","weight":NaN}]}'
        with self.assertRaises(LensFormatError):
            import_lens(payload)

    def test_rejects_unsupported_rule_enum(self) -> None:
        document = json.loads(export_lens(self.lens))
        document["rules"][0]["target"] = "behavior_profile"
        with self.assertRaises(LensFormatError):
            import_lens(json.dumps(document))

    def test_rejects_oversized_document_before_parse(self) -> None:
        payload = " " * (MAX_LENS_DOCUMENT_BYTES + 1)
        with self.assertRaises(LensFormatError):
            import_lens(payload)

    def test_exclude_weight_is_normalized_by_model(self) -> None:
        document = json.loads(export_lens(self.lens))
        exclude = next(rule for rule in document["rules"] if rule["action"] == "exclude")
        exclude["weight"] = 42
        imported = import_lens(json.dumps(document))
        imported_exclude = next(rule for rule in imported.rules if rule.action is LensRuleAction.EXCLUDE)
        self.assertEqual(imported_exclude.weight, 0.0)


if __name__ == "__main__":
    unittest.main()
