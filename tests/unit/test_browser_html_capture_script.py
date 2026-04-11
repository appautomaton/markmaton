import contextlib
import importlib.util
import io
import pathlib
import unittest


SCRIPT_PATH = pathlib.Path(
    "skills/html-to-markdown/scripts/capture_html.py"
)


def load_module():
    spec = importlib.util.spec_from_file_location(
        "browser_html_capture_script", SCRIPT_PATH
    )
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


class BrowserHtmlCaptureScriptTestCase(unittest.TestCase):
    def setUp(self) -> None:
        self.module = load_module()

    def test_build_parser_defaults(self) -> None:
        parser = self.module.build_parser()
        args = parser.parse_args(["https://example.com"])

        self.assertEqual(args.url, "https://example.com")
        self.assertEqual(args.output_format, "json")
        self.assertEqual(args.timeout, 10.0)
        self.assertIsNone(args.wait_selector)
        self.assertIsNone(args.wait_text)

    def test_build_parser_has_no_headless_flag(self) -> None:
        parser = self.module.build_parser()
        args = parser.parse_args(["https://example.com"])
        self.assertFalse(hasattr(args, "headless"))

    def test_render_output_html_mode(self) -> None:
        payload = {
            "html": "<html><body>ok</body></html>",
            "url": "https://example.com",
            "final_url": "https://example.com",
            "content_type": "text/html",
        }

        self.assertEqual(
            self.module.render_output(payload, "html"),
            "<html><body>ok</body></html>",
        )

    def test_main_emits_json_with_injected_capture_impl(self) -> None:
        calls = []

        def fake_capture(url, **kwargs):
            calls.append((url, kwargs))
            return {
                "html": "<html><body>ok</body></html>",
                "url": url,
                "final_url": url,
                "content_type": "text/html",
                "title": "Example",
                "rendered": True,
            }

        stdout = io.StringIO()
        with contextlib.redirect_stdout(stdout):
            rc = self.module.main(
                ["https://example.com", "--wait-selector", "main"],
                capture_impl=fake_capture,
            )

        self.assertEqual(rc, 0)
        self.assertIn('"html": "<html><body>ok</body></html>"', stdout.getvalue())
        self.assertEqual(calls[0][0], "https://example.com")
        self.assertEqual(calls[0][1]["wait_selector"], "main")
        self.assertNotIn("headless_mode", calls[0][1])

    def test_main_emits_html_with_injected_capture_impl(self) -> None:
        def fake_capture(url, **kwargs):
            return {
                "html": "<html><body>ok</body></html>",
                "url": url,
                "final_url": url,
                "content_type": "text/html",
            }

        stdout = io.StringIO()
        with contextlib.redirect_stdout(stdout):
            rc = self.module.main(
                ["https://example.com", "--output-format", "html"],
                capture_impl=fake_capture,
            )

        self.assertEqual(rc, 0)
        self.assertEqual(stdout.getvalue(), "<html><body>ok</body></html>\n")

    def test_main_reports_capture_errors(self) -> None:
        def failing_capture(url, **kwargs):
            raise RuntimeError("boom")

        stderr = io.StringIO()
        with contextlib.redirect_stderr(stderr):
            rc = self.module.main(
                ["https://example.com"],
                capture_impl=failing_capture,
            )

        self.assertEqual(rc, 1)
        self.assertIn("capture failed for https://example.com: boom", stderr.getvalue())

    def test_find_chrome_is_callable(self) -> None:
        self.assertTrue(callable(self.module.find_chrome))


if __name__ == "__main__":
    unittest.main()
