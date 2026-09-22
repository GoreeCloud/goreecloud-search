from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[1]


class RepositoryFeatureTrackingGovernanceTest(unittest.TestCase):
    def test_required_repository_native_feature_records_exist(self) -> None:
        for name in (
            "IMPLEMENTED-FEATURES.md",
            "PLANNED-FEATURES.md",
            "CHANGELOGS.md",
        ):
            with self.subTest(name=name):
                self.assertTrue((ROOT / name).is_file(), f"missing required root record: {name}")

    def test_retired_root_feature_tracking_files_do_not_return(self) -> None:
        for name in ("FEATURE-ROADMAP.md", "CHANGELOG.md"):
            with self.subTest(name=name):
                self.assertFalse((ROOT / name).exists(), f"retired root record returned: {name}")


if __name__ == "__main__":
    unittest.main()
