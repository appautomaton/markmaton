# Usage Guide

## What `markmaton` expects

`markmaton` does not fetch pages itself.

It expects:

- `html`: the page HTML you want to parse
- `url`: the source URL, used as parsing context

Typical upstream sources are:

- Playwright-rendered HTML
- raw HTML from `fetch`
- HTML returned by another scrape layer such as Firecrawl

## CLI

### Parse a local HTML file into Markdown

```bash
markmaton convert \
  --html-file page.html \
  --url https://example.com/article \
  --output-format markdown
```

### Parse a local HTML file into JSON

```bash
markmaton convert \
  --html-file page.html \
  --url https://example.com/article \
  --output-format json
```

### Read HTML from stdin

```bash
cat page.html | markmaton convert \
  --url https://example.com/article \
  --output-format markdown
```

### Use full-content mode

By default, `markmaton` prefers main-content extraction.

To keep more of the page:

```bash
markmaton convert \
  --html-file page.html \
  --url https://example.com/article \
  --full-content \
  --output-format json
```

`--full-content` disables the main-content narrowing step, but it does **not** return raw page HTML as-is. Global cleanup still removes hidden nodes, modal chrome, skip links, and similar shell noise.

### Content-mode semantics

- Default mode: start with main-content extraction.
- Automatic fallback: if the default result looks too weak, the engine reruns with broader content.
- Explicit `--full-content`: skip main-content narrowing up front and run the broader cleanup path directly.

### Force include or exclude selectors

```bash
markmaton convert \
  --html-file page.html \
  --url https://example.com/article \
  --include-selector article \
  --exclude-selector ".cookie-banner" \
  --exclude-selector ".newsletter-modal" \
  --output-format json
```

### Additional context flags

`--final-url` passes the post-redirect URL and `--content-type` passes a MIME-type hint. Both sharpen canonical-URL resolution and metadata extraction when available.

## Python API

```python
from markmaton import ConvertOptions, ConvertRequest, convert_html

request = ConvertRequest(
    html="<article><h1>Hello</h1><p>World</p></article>",
    url="https://example.com/article",
    options=ConvertOptions(
        only_main_content=True,
        include_selectors=[],
        exclude_selectors=[],
    ),
)

response = convert_html(request)

print(response.markdown)
print(response.metadata.title)
print(response.quality.quality_score)
```

## Response shape

### Markdown mode

Returns only the Markdown body.

### JSON mode

Returns:

- `markdown`
- `html_clean`
- `metadata`
- `links`
- `images`
- `quality`

## Engine discovery

The Python wrapper searches for the Go engine binary in a defined order. See [packaging layout](packaging-layout.md#binary-discovery-contract) for the full lookup sequence.

## When to pass `url`

Always pass `url` if you have it.

It improves:

- canonical URL fallback
- relative link resolution
- image source normalization
- source attribution in downstream workflows
