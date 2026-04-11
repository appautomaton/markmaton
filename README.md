# markmaton

[![CI](https://github.com/appautomaton/markmaton/actions/workflows/ci.yml/badge.svg)](https://github.com/appautomaton/markmaton/actions/workflows/ci.yml)
[![Release](https://github.com/appautomaton/markmaton/actions/workflows/workflow.yml/badge.svg)](https://github.com/appautomaton/markmaton/actions/workflows/workflow.yml)
[![PyPI version](https://img.shields.io/pypi/v/markmaton)](https://pypi.org/project/markmaton/)
[![Python versions](https://img.shields.io/pypi/pyversions/markmaton)](https://pypi.org/project/markmaton/)

`markmaton` is a lightweight HTML-to-Markdown parser core built for agent workflows.

It solves the last-mile parsing problem in a web pipeline: you already have page HTML,
but it is still too noisy and awkward for downstream agent use. Feed `markmaton`
HTML from a fetcher or browser layer and get back cleaner Markdown, metadata, links,
images, and quality signals.

> [!NOTE]
> `markmaton` is a general parser, not a crawler.
> Feed it HTML from Playwright, `fetch`, Firecrawl, or another upstream page-visit tool.

## Why it exists

- Raw page HTML is usually not directly useful for downstream agent workflows.
- Modern pages often mix the real content with navigation, overlays, cards, and app shell chrome.
- `markmaton` keeps that cleanup and conversion step deterministic and separate from crawling.
- The project stays narrow by design: no crawling, browser control, network, or LLM features.
- The user-facing entrypoint is a Python CLI and API wrapped around a fast Go engine.

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
- Research, benchmark, and release docs: `docs/`

## Documentation

- [Documentation index](docs/README.md)
- [Usage guide](docs/usage.md)
- [Packaging layout](docs/packaging-layout.md)
- [PyPI release path](docs/pypi-release.md)
- [Benchmark workflow](docs/benchmark-workflow.md)
- [Benchmark matrix](docs/benchmark-matrix.md)
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
