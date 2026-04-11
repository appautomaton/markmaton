import importlib.util
import io
import json
import pathlib
import sys
import unittest
from unittest import mock


SCRIPT_PATH = pathlib.Path(
    "skills/html-to-markdown/scripts/markmaton_convert.py"
)


def load_module():
    spec = importlib.util.spec_from_file_location(
        "markmaton_convert_script", SCRIPT_PATH
    )
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


class ConvertScriptTestCase(unittest.TestCase):
    def setUp(self) -> None:
        self.module = load_module()

    def test_build_parser_defaults(self) -> None:
        parser = self.module.build_parser()
        args = parser.parse_args([])

        self.assertEqual(args.output_format, "json")
        self.assertFalse(args.from_capture)
        self.assertFalse(args.full_content)
        self.assertIsNone(args.html_file)
        self.assertIsNone(args.url)
        self.assertIsNone(args.final_url)
        self.assertIsNone(args.content_type)

    def test_build_parser_from_capture_flag(self) -> None:
        parser = self.module.build_parser()
        args = parser.parse_args(["--from-capture"])

        self.assertTrue(args.from_capture)

    def test_read_capture_envelope_extracts_fields(self) -> None:
        envelope = json.dumps({
            "html": "<html><body>ok</body></html>",
            "url": "https://example.com",
            "final_url": "https://www.example.com/page",
            "content_type": "text/html",
            "title": "Example",
            "rendered": True,
        })

        parser = self.module.build_parser()
        args = parser.parse_args(["--from-capture"])

        with mock.patch("sys.stdin", io.StringIO(envelope)):
            html, url, final_url, content_type = (
                self.module._read_capture_envelope(args)
            )

        self.assertEqual(html, "<html><body>ok</body></html>")
        self.assertEqual(url, "https://example.com")
        self.assertEqual(final_url, "https://www.example.com/page")
        self.assertEqual(content_type, "text/html")

    def test_read_capture_envelope_cli_overrides_envelope(self) -> None:
        envelope = json.dumps({
            "html": "<html><body>ok</body></html>",
            "url": "https://example.com",
            "final_url": "https://www.example.com/page",
            "content_type": "text/html",
        })

        parser = self.module.build_parser()
        args = parser.parse_args([
            "--from-capture",
            "--url", "https://override.com",
        ])

        with mock.patch("sys.stdin", io.StringIO(envelope)):
            html, url, final_url, content_type = (
                self.module._read_capture_envelope(args)
            )

        self.assertEqual(url, "https://override.com")
        self.assertEqual(final_url, "https://www.example.com/page")

    def test_read_capture_envelope_handles_minimal_json(self) -> None:
        envelope = json.dumps({"html": "<p>hello</p>"})

        parser = self.module.build_parser()
        args = parser.parse_args(["--from-capture"])

        with mock.patch("sys.stdin", io.StringIO(envelope)):
            html, url, final_url, content_type = (
                self.module._read_capture_envelope(args)
            )

        self.assertEqual(html, "<p>hello</p>")
        self.assertIsNone(url)
        self.assertIsNone(final_url)
        self.assertIsNone(content_type)


if __name__ == "__main__":
    unittest.main()
