import pathlib
import re
import unittest

import markmaton


class PackageMetadataTestCase(unittest.TestCase):
    def test_package_version_matches_pyproject(self) -> None:
        pyproject = pathlib.Path("pyproject.toml").read_text(encoding="utf-8")
        match = re.search(r'^version = "([^"]+)"$', pyproject, re.MULTILINE)
        self.assertIsNotNone(match)
        self.assertEqual(markmaton.__version__, match.group(1))

    def test_python_version_contract_is_pinned_to_3_12(self) -> None:
        pyproject = pathlib.Path("pyproject.toml").read_text(encoding="utf-8")
        requires_python = re.search(r'^requires-python = "([^"]+)"$', pyproject, re.MULTILINE)
        self.assertIsNotNone(requires_python)
        self.assertEqual(requires_python.group(1), ">=3.12")
        self.assertEqual(pathlib.Path(".python-version").read_text(encoding="utf-8").strip(), "3.12")

    def test_public_api_exports(self) -> None:
        self.assertTrue(callable(markmaton.convert_html))
        self.assertTrue(callable(markmaton.discover_engine))
        self.assertIsNotNone(markmaton.ConvertRequest)
        self.assertIsNotNone(markmaton.ConvertOptions)
        self.assertIsNotNone(markmaton.ConvertResponse)


if __name__ == "__main__":
    unittest.main()
