import os
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import hatch_build


class HatchBuildTestCase(unittest.TestCase):
    def test_binary_filename_matches_platform(self) -> None:
        self.assertEqual(hatch_build.binary_filename("Windows"), "markmaton-engine.exe")
        self.assertEqual(hatch_build.binary_filename("Darwin"), "markmaton-engine")

    def test_platform_wheel_tag_is_not_any(self) -> None:
        tag = hatch_build.platform_wheel_tag()
        self.assertTrue(tag.startswith("py3-none-"))
        self.assertNotEqual(tag, "py3-none-any")

    def test_resolve_engine_source_prefers_explicit_path(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            explicit = Path(tmpdir) / "custom-engine"
            explicit.write_text("bin")
            discovered = hatch_build.resolve_engine_source(explicit_path=str(explicit))
            self.assertEqual(discovered, explicit.resolve())

    def test_resolve_engine_source_uses_environment_override(self) -> None:
        with tempfile.NamedTemporaryFile() as handle:
            with mock.patch.dict(os.environ, {"MARKMATON_ENGINE": handle.name}):
                discovered = hatch_build.resolve_engine_source()
            self.assertEqual(discovered, Path(handle.name).resolve())

    def test_packaged_binary_path_uses_package_bin_directory(self) -> None:
        path = hatch_build.packaged_binary_path()
        self.assertIn(str(Path("markmaton") / "bin"), str(path))


if __name__ == "__main__":
    unittest.main()
