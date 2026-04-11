import pathlib
import unittest


SKILL_ROOT = pathlib.Path("skills/html-to-markdown")


class SkillArtifactsTestCase(unittest.TestCase):
    def test_convert_script_has_pep723_metadata(self) -> None:
        script = SKILL_ROOT / "scripts" / "markmaton_convert.py"

        self.assertTrue(script.is_file())

        script_text = script.read_text(encoding="utf-8")
        self.assertIn("# /// script", script_text)
        self.assertIn('requires-python = ">=3.12"', script_text)
        self.assertIn('dependencies = [', script_text)
        self.assertIn('"markmaton"', script_text)

    def test_capture_script_has_pep723_metadata(self) -> None:
        script = SKILL_ROOT / "scripts" / "capture_html.py"

        self.assertTrue(script.is_file())

        script_text = script.read_text(encoding="utf-8")
        self.assertIn("# /// script", script_text)
        self.assertIn('requires-python = ">=3.12"', script_text)
        self.assertIn('dependencies = [', script_text)
        self.assertIn('"nodriver"', script_text)

    def test_skill_references_both_scripts(self) -> None:
        skill_md = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8")

        self.assertIn("capture_html.py", skill_md)
        self.assertIn("markmaton_convert.py", skill_md)

    def test_skill_documents_both_paths(self) -> None:
        skill_md = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8")

        self.assertIn("## From a URL", skill_md)
        self.assertIn("## From HTML", skill_md)
        self.assertIn("--from-capture", skill_md)


if __name__ == "__main__":
    unittest.main()
