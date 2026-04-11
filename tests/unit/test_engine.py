import json
import os
import tempfile
import unittest
from pathlib import Path
from unittest import mock

from markmaton.engine import EngineNotFoundError, convert_html, discover_engine
from markmaton.models import ConvertOptions, ConvertRequest


class EngineTestCase(unittest.TestCase):
    def test_discover_engine_uses_env_var(self) -> None:
        with tempfile.NamedTemporaryFile() as handle:
            with mock.patch.dict(os.environ, {"MARKMATON_ENGINE": handle.name}):
                discovered = discover_engine()
        self.assertIsInstance(discovered, Path)

    def test_discover_engine_raises_when_missing(self) -> None:
        with mock.patch.dict(os.environ, {}, clear=True):
            with mock.patch("markmaton.engine.shutil.which", return_value=None):
                with mock.patch("markmaton.engine.Path.is_file", return_value=False):
                    with self.assertRaises(EngineNotFoundError):
                        discover_engine()

    @mock.patch("markmaton.engine.discover_engine")
    @mock.patch("markmaton.engine.subprocess.run")
    def test_convert_html_invokes_engine(self, run_mock: mock.Mock, discover_mock: mock.Mock) -> None:
        discover_mock.return_value = Path("/tmp/markmaton-engine")
        run_mock.return_value = mock.Mock(
            returncode=0,
            stdout=json.dumps(
                {
                    "markdown": "Hello",
                    "html_clean": "<p>Hello</p>",
                    "metadata": {"title": "Hello"},
                    "links": [],
                    "images": [],
                    "quality": {"text_length": 5},
                }
            ),
            stderr="",
        )

        response = convert_html(ConvertRequest(html="<p>Hello</p>"))

        self.assertEqual(response.markdown, "Hello")
        run_mock.assert_called_once()

    @mock.patch("markmaton.engine.discover_engine")
    @mock.patch("markmaton.engine.subprocess.run")
    def test_convert_html_preserves_explicit_false_content_mode(self, run_mock: mock.Mock, discover_mock: mock.Mock) -> None:
        discover_mock.return_value = Path("/tmp/markmaton-engine")
        run_mock.return_value = mock.Mock(
            returncode=0,
            stdout=json.dumps(
                {
                    "markdown": "Hello",
                    "html_clean": "<p>Hello</p>",
                    "metadata": {"title": "Hello"},
                    "links": [],
                    "images": [],
                    "quality": {"text_length": 5},
                }
            ),
            stderr="",
        )

        convert_html(
            ConvertRequest(
                html="<p>Hello</p>",
                options=ConvertOptions(only_main_content=False),
            )
        )

        payload = json.loads(run_mock.call_args.kwargs["input"])
        self.assertFalse(payload["options"]["only_main_content"])


if __name__ == "__main__":
    unittest.main()
