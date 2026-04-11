from __future__ import annotations

import json
import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


class CLIIntegrationTestCase(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.repo_root = Path(__file__).resolve().parents[2]
        cls.python = sys.executable
        binary_name = "markmaton-engine.exe" if sys.platform.startswith("win") else "markmaton-engine"
        explicit_engine = os.environ.get("MARKMATON_ENGINE")
        cls._engine_tmpdir: tempfile.TemporaryDirectory[str] | None = None

        if explicit_engine:
            cls.engine_path = Path(explicit_engine)
            if not cls.engine_path.is_file():
                raise unittest.SkipTest(
                    f"integration tests require a built engine binary at {cls.engine_path}"
                )
            return

        if shutil.which("go") is None:
            raise unittest.SkipTest("integration tests require Go or MARKMATON_ENGINE")

        cls._engine_tmpdir = tempfile.TemporaryDirectory()
        cls.engine_path = Path(cls._engine_tmpdir.name) / binary_name
        env = os.environ.copy()
        env["GOCACHE"] = str(Path(cls._engine_tmpdir.name) / "gocache")
        completed = subprocess.run(
            ["go", "build", "-o", str(cls.engine_path), "./cmd/markmaton-engine"],
            cwd=cls.repo_root,
            env=env,
            capture_output=True,
            text=True,
            check=False,
        )
        if completed.returncode != 0:
            raise RuntimeError(completed.stderr or completed.stdout or "failed to build markmaton-engine")

    @classmethod
    def tearDownClass(cls) -> None:
        if cls._engine_tmpdir is not None:
            cls._engine_tmpdir.cleanup()

    def run_cli(self, *args: str) -> subprocess.CompletedProcess[str]:
        env = os.environ.copy()
        env["MARKMATON_ENGINE"] = str(self.engine_path)
        pythonpath = env.get("PYTHONPATH")
        env["PYTHONPATH"] = (
            f"{self.repo_root}{os.pathsep}{pythonpath}"
            if pythonpath
            else str(self.repo_root)
        )

        return subprocess.run(
            [self.python, "-m", "markmaton.cli", *args],
            cwd=self.repo_root,
            env=env,
            capture_output=True,
            text=True,
            check=False,
        )

    def read_fixture(self, relative_path: str) -> Path:
        return self.repo_root / "testdata" / "fixtures" / relative_path

    def read_golden(self, relative_path: str) -> str:
        return (self.repo_root / "testdata" / "golden" / relative_path).read_text(
            encoding="utf-8"
        )

    def test_convert_json_matches_fixture_and_metadata(self) -> None:
        completed = self.run_cli(
            "convert",
            "--html-file",
            str(self.read_fixture("core/article.html")),
            "--url",
            "https://example.com/articles/harnessing-parsers",
        )

        self.assertEqual(completed.returncode, 0, completed.stderr or completed.stdout)

        payload = json.loads(completed.stdout)
        self.assertEqual(payload["markdown"].rstrip("\n"), self.read_golden("core/article.md").rstrip("\n"))
        self.assertEqual(payload["metadata"]["title"], "Harnessing Parsers")
        self.assertEqual(
            payload["metadata"]["canonical_url"],
            "https://example.com/articles/harnessing-parsers",
        )
        self.assertEqual(
            payload["links"],
            ["https://example.com/guides/parser-design"],
        )
        self.assertEqual(
            payload["images"],
            ["https://example.com/images/cover-large.jpg"],
        )
        self.assertTrue(payload["quality"]["title_present"])
        self.assertGreater(payload["quality"]["text_length"], 0)

    def test_convert_markdown_matches_fixture_golden_output(self) -> None:
        completed = self.run_cli(
            "convert",
            "--html-file",
            str(self.read_fixture("core/docs.html")),
            "--output-format",
            "markdown",
        )

        self.assertEqual(completed.returncode, 0, completed.stderr or completed.stdout)
        self.assertEqual(
            completed.stdout.rstrip("\n"),
            self.read_golden("core/docs.md").rstrip("\n"),
        )

    def test_convert_markdown_keeps_multi_article_links(self) -> None:
        completed = self.run_cli(
            "convert",
            "--html-file",
            str(self.read_fixture("core/news.html")),
            "--url",
            "https://example.com/newsroom",
            "--output-format",
            "markdown",
        )

        self.assertEqual(completed.returncode, 0, completed.stderr or completed.stdout)
        self.assertEqual(
            completed.stdout.rstrip("\n"),
            self.read_golden("core/news.md").rstrip("\n"),
        )

    def test_full_content_mode_changes_extraction_scope(self) -> None:
        html = """
        <html>
          <head><title>Main Story</title></head>
          <body>
            <article>
              <h1>Main Story</h1>
              <p>Primary article body with enough detail to look like meaningful page content instead of a thin fragment.</p>
              <p>Second paragraph adds more context so the quality heuristics can treat main-content mode as a real success path.</p>
              <p>Third paragraph keeps the article comfortably above the fallback threshold while staying easy to inspect in the test.</p>
            </article>
            <aside>
              <h2>Visible sidebar chrome</h2>
              <p>Keep this only in full-content mode.</p>
            </aside>
          </body>
        </html>
        """

        with tempfile.NamedTemporaryFile("w+", suffix=".html") as handle:
            handle.write(html)
            handle.flush()

            default_completed = self.run_cli(
                "convert",
                "--html-file",
                handle.name,
                "--output-format",
                "json",
            )
            full_completed = self.run_cli(
                "convert",
                "--html-file",
                handle.name,
                "--full-content",
                "--output-format",
                "json",
            )

        self.assertEqual(default_completed.returncode, 0, default_completed.stderr or default_completed.stdout)
        self.assertEqual(full_completed.returncode, 0, full_completed.stderr or full_completed.stdout)
        default_payload = json.loads(default_completed.stdout)
        full_payload = json.loads(full_completed.stdout)
        self.assertTrue(default_payload["quality"]["used_main_content"])
        self.assertFalse(default_payload["quality"]["fallback_used"])
        self.assertNotIn("Visible sidebar chrome", default_payload["markdown"])
        self.assertFalse(full_payload["quality"]["used_main_content"])
        self.assertFalse(full_payload["quality"]["fallback_used"])
        self.assertIn("Visible sidebar chrome", full_payload["markdown"])


if __name__ == "__main__":
    unittest.main()
