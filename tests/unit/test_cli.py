import io
import json
import tempfile
import unittest
from contextlib import redirect_stdout
from unittest import mock

from markmaton.cli import main
from markmaton.models import ConvertResponse, Metadata, Quality


class CLITestCase(unittest.TestCase):
    @mock.patch("markmaton.cli.convert_html")
    def test_cli_outputs_json(self, convert_mock: mock.Mock) -> None:
        convert_mock.return_value = ConvertResponse(
            markdown="Hello",
            html_clean="<p>Hello</p>",
            metadata=Metadata(title="Hello"),
            links=[],
            images=[],
            quality=Quality(text_length=5),
        )

        with tempfile.NamedTemporaryFile("w+", suffix=".html") as handle:
            handle.write("<p>Hello</p>")
            handle.flush()
            output = io.StringIO()
            with redirect_stdout(output):
                exit_code = main(["convert", "--html-file", handle.name])

        self.assertEqual(exit_code, 0)
        payload = json.loads(output.getvalue())
        self.assertEqual(payload["markdown"], "Hello")

    @mock.patch("markmaton.cli.convert_html")
    def test_cli_outputs_markdown(self, convert_mock: mock.Mock) -> None:
        convert_mock.return_value = ConvertResponse(
            markdown="Hello",
            html_clean="<p>Hello</p>",
            metadata=Metadata(),
            links=[],
            images=[],
            quality=Quality(),
        )

        with tempfile.NamedTemporaryFile("w+", suffix=".html") as handle:
            handle.write("<p>Hello</p>")
            handle.flush()
            output = io.StringIO()
            with redirect_stdout(output):
                exit_code = main(["convert", "--html-file", handle.name, "--output-format", "markdown"])

        self.assertEqual(exit_code, 0)
        self.assertEqual(output.getvalue(), "Hello\n")

    @mock.patch("markmaton.cli.convert_html")
    def test_cli_passes_full_content_flag_through_request(self, convert_mock: mock.Mock) -> None:
        convert_mock.return_value = ConvertResponse(
            markdown="Hello",
            html_clean="<p>Hello</p>",
            metadata=Metadata(),
            links=[],
            images=[],
            quality=Quality(),
        )

        with tempfile.NamedTemporaryFile("w+", suffix=".html") as handle:
            handle.write("<p>Hello</p>")
            handle.flush()

            output = io.StringIO()
            with redirect_stdout(output):
                exit_code = main(["convert", "--html-file", handle.name, "--full-content"])

        self.assertEqual(exit_code, 0)
        request = convert_mock.call_args.args[0]
        self.assertFalse(request.options.only_main_content)


if __name__ == "__main__":
    unittest.main()
