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

The Python wrapper looks for the Go engine in this order:

1. explicit path passed by the caller
2. `MARKMATON_ENGINE`
3. packaged binary in `markmaton/bin/`
4. local development binary in `./bin/`
5. `markmaton-engine` on `PATH`

This keeps local development and packaged installs on the same interface.

## When to pass `url`

Always pass `url` if you have it.

It improves:

- canonical URL fallback
- relative link resolution
- image source normalization
- source attribution in downstream workflows
