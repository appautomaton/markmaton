# markmaton

[![CI](https://github.com/appautomaton/markmaton/actions/workflows/ci.yml/badge.svg)](https://github.com/appautomaton/markmaton/actions/workflows/ci.yml)
[![Release](https://github.com/appautomaton/markmaton/actions/workflows/workflow.yml/badge.svg)](https://github.com/appautomaton/markmaton/actions/workflows/workflow.yml)
[![PyPI version](https://img.shields.io/pypi/v/markmaton)](https://pypi.org/project/markmaton/)
[![Python versions](https://img.shields.io/pypi/pyversions/markmaton)](https://pypi.org/project/markmaton/)
[![License: MIT](https://img.shields.io/github/license/appautomaton/markmaton)](LICENSE)

`markmaton` is a lightweight HTML-to-Markdown parser core built for agent workflows.

It solves the last-mile parsing problem in a web pipeline: you already have page HTML,
but it is still too noisy and awkward for downstream agent use. Feed `markmaton`
HTML from a fetcher or browser layer and get back cleaner Markdown, metadata, links,
images, and quality signals.

> [!NOTE]
> `markmaton` is a general parser, not a crawler.
> Feed it HTML from Playwright, `fetch`, Firecrawl, or another upstream page-visit tool.

## Example

In: page HTML with nav, a cookie banner, related links, a footer, and scripts.

```html
<html lang="en">
  <head>
    <title>Shipping Faster With Queues · Acme Engineering</title>
    <meta name="description" content="How Acme cut job latency with a queue-first design." />
  </head>
  <body>
    <header class="topbar"><nav><a href="/">Acme</a> <a href="/blog">Blog</a></nav></header>
    <div class="cookie-banner">We use cookies. <button>Accept all</button></div>
    <main>
      <article>
        <h1>Shipping Faster With Queues</h1>
        <p>We cut p95 job latency by 60% after moving webhook delivery to a queue-first design.</p>
        <p>The full breakdown is in our <a href="/posts/queue-first-design">queue-first design post</a>.</p>
        <pre><code class="language-python">def enqueue(job):
    queue.push(job, delay=backoff(job.attempts))</code></pre>
        <img src="/static/latency-small.png"
             srcset="/static/latency-small.png 1x, /static/latency-chart.png 2x"
             alt="Latency chart" />
      </article>
      <aside class="related"><a href="/posts/retry-storms">Taming retry storms</a></aside>
    </main>
    <footer>© 2026 Acme Corp</footer>
    <script>window.analytics.track("pageview");</script>
  </body>
</html>
```

```bash
markmaton convert \
  --html-file page.html \
  --url https://engineering.acme.com/posts/shipping-faster-with-queues \
  --output-format markdown
```

Out: main content only, as Markdown.

````markdown
# Shipping Faster With Queues

We cut p95 job latency by 60% after moving webhook delivery to a queue-first design.

The full breakdown is in our [queue-first design post](https://engineering.acme.com/posts/shipping-faster-with-queues).

```python
def enqueue(job):
    queue.push(job, delay=backoff(job.attempts))
```

![Latency chart](https://engineering.acme.com/static/latency-chart.png)
````

Nav, banner, aside, footer, and script are stripped; the relative link and the
2x `srcset` image resolve to absolute URLs. JSON mode adds metadata, links,
images, and quality signals — see [Output](#output).

## Why it exists

- Raw page HTML is usually not directly useful for downstream agent workflows.
- Modern pages often mix the real content with navigation, overlays, cards, and app shell chrome.
- `markmaton` keeps that cleanup and conversion step deterministic and separate from crawling.
- The project stays narrow by design: no crawling, browser control, network, or LLM features.
- The user-facing entrypoint is a Python CLI and API wrapped around a fast Go engine.

## How it compares

- **`markdownify`** converts HTML to Markdown but does no main-content
  extraction or metadata collection. `markmaton` strips page chrome, converts,
  and returns metadata, links, images, and quality signals in one step.
- **`readability-lxml`** distills main content as cleaned HTML; you still need
  a separate HTML-to-Markdown converter and metadata layer on top.
  `markmaton` returns the full structured response in one call.
- **`trafilatura`** is a broader extraction framework with its own fetching and
  discovery pipelines. `markmaton` is deliberately narrower: a parser core you
  embed behind your own fetcher or browser layer.

## Install

### `pip`

```bash
pip install markmaton
```

### `uv tool`

```bash
uv tool install markmaton
```

> [!TIP]
> The installed package works through plain `pip`.
> Local development uses `uv` with Python 3.12.

## Quickstart

### CLI

```bash
markmaton convert \
  --html-file page.html \
  --url https://example.com/article \
  --output-format markdown
```

To get the full structured response:

```bash
markmaton convert \
  --html-file page.html \
  --url https://example.com/article \
  --output-format json
```

### Python API

```python
from markmaton import ConvertOptions, ConvertRequest, convert_html

html = "<article><h1>Hello</h1><p>World</p></article>"

response = convert_html(
    ConvertRequest(
        html=html,
        url="https://example.com/article",
        options=ConvertOptions(only_main_content=True),
    )
)

print(response.markdown)
print(response.metadata.title)
```

> [!TIP]
> Pass `url` whenever you can.
> `markmaton` uses it as parsing context for canonical metadata and absolute link resolution.

## Output

JSON mode returns `markdown`, `html_clean`, `metadata`, `links`, `images`, and `quality`. See [response shape](docs/usage.md#response-shape) for details.

## Project shape

- Go engine: `cmd/markmaton-engine`
- Python wrapper and CLI: `markmaton/`
- Parser fixtures and golden files: `testdata/`
- Architecture, benchmark, and release docs: `docs/`

## Documentation

- [Landing page](https://appautomaton.renocrypt.com/markmaton/)
- [Documentation index](docs/README.md)
- [Usage guide](docs/usage.md)
- [Packaging layout](docs/packaging-layout.md)
- [PyPI release path](docs/pypi-release.md)
- [Benchmark and regression workflow](docs/benchmark-workflow.md)
- [Regression corpus](testdata/README.md)
- [AI agent skill](skills/html-to-markdown/SKILL.md) — for using `markmaton` inside an agent workflow

## Development

Set up the local development environment:

```bash
uv sync --group dev
```

Run the core test suites:

```bash
uv run python -m unittest discover -s tests -p 'test_*.py'
go test ./...
```

For a manual end-to-end smoke:

- [Local smoke flow](docs/local-smoke.md)

The repo is pinned to:

- Python `3.12` via [`.python-version`](.python-version)
- a committed `uv.lock`

> [!IMPORTANT]
> Automated tests are unit-test-first. Live page visits and benchmarks are manual.

## Release notes

- [Changelog](CHANGELOG.md)
- [GitHub Releases](https://github.com/appautomaton/markmaton/releases)
