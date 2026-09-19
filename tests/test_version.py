from pathlib import Path
import unittest

from goreecloud_search import __version__


class VersionTests(unittest.TestCase):
    def test_root_version_matches_package_version(self) -> None:
        root = Path(__file__).resolve().parents[1]
        self.assertEqual((root / "VERSION").read_text().strip(), __version__)


if __name__ == "__main__":
    unittest.main()
