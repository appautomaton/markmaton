import unittest

from markmaton.models import ConvertOptions, ConvertRequest, ConvertResponse


class ModelsTestCase(unittest.TestCase):
    def test_request_to_payload(self) -> None:
        request = ConvertRequest(
            html="<p>Hello</p>",
            url="https://example.com",
            options=ConvertOptions(include_selectors=["article"]),
        )

        payload = request.to_payload()

        self.assertEqual(payload["url"], "https://example.com")
        self.assertEqual(payload["options"]["include_selectors"], ["article"])

    def test_response_from_dict(self) -> None:
        response = ConvertResponse.from_dict(
            {
                "markdown": "Hello",
                "html_clean": "<p>Hello</p>",
                "metadata": {"title": "Hello"},
                "links": ["https://example.com"],
                "images": [],
                "quality": {"text_length": 5, "quality_score": 0.9},
            }
        )

        self.assertEqual(response.metadata.title, "Hello")
        self.assertEqual(response.quality.text_length, 5)


if __name__ == "__main__":
    unittest.main()
