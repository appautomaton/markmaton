import pathlib
import unittest


class SkillArtifactsTestCase(unittest.TestCase):
    def test_html_to_markdown_skill_has_pep723_script(self) -> None:
        root = pathlib.Path("skills/html-to-markdown/scripts")
        script = root / "markmaton_convert.py"

        self.assertTrue(script.is_file())

        script_text = script.read_text(encoding="utf-8")
        self.assertIn("# /// script", script_text)
        self.assertIn('requires-python = ">=3.12"', script_text)
        self.assertIn('dependencies = [', script_text)
        self.assertIn('"markmaton"', script_text)


if __name__ == "__main__":
    unittest.main()
