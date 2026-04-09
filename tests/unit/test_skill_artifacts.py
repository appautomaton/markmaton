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

    def test_browser_html_capture_skill_has_pep723_script(self) -> None:
        root = pathlib.Path("skills/browser-html-capture/scripts")
        script = root / "capture_html.py"

        self.assertTrue(script.is_file())

        script_text = script.read_text(encoding="utf-8")
        self.assertIn("# /// script", script_text)
        self.assertIn('requires-python = ">=3.12"', script_text)
        self.assertIn('dependencies = [', script_text)
        self.assertIn('"nodriver"', script_text)

    def test_skills_cross_reference_each_other_lightly(self) -> None:
        html_to_markdown = pathlib.Path(
            "skills/html-to-markdown/SKILL.md"
        ).read_text(encoding="utf-8")
        browser_capture = pathlib.Path(
            "skills/browser-html-capture/SKILL.md"
        ).read_text(encoding="utf-8")

        self.assertIn("browser-html-capture", html_to_markdown)
        self.assertIn("html-to-markdown", browser_capture)


if __name__ == "__main__":
    unittest.main()
