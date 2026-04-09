# Usage

The primary entrypoint is the bundled PEP 723 script:

```bash
uv run --script skills/html-to-markdown/scripts/markmaton_convert.py ...
```

## CLI-style usage through the script

### Parse an HTML file to Markdown

```bash
uv run --script skills/html-to-markdown/scripts/markmaton_convert.py \
  --html-file page.html \
  --url https://example.com/article \
  --output-format markdown
```

### Parse an HTML file to JSON

```bash
uv run --script skills/html-to-markdown/scripts/markmaton_convert.py \
  --html-file page.html \
  --url https://example.com/article \
  --output-format json
```

### Parse HTML from stdin

```bash
cat page.html | uv run --script skills/html-to-markdown/scripts/markmaton_convert.py \
  --url https://example.com/article \
  --output-format json
```

This is the preferred shape when another tool already produced the HTML string.

## Output modes

Use `markdown` when you only need readable Markdown.

Use `json` when you need:

- `markdown`
- `metadata`
- `links`
- `images`
- `quality`

## Notes

- The script uses the published `markmaton` package as an inline dependency.
- Because it uses PEP 723 metadata, `uv run` will manage an isolated environment for the script.
- The script does not require the caller to know about a local repo checkout, local `.venv`, or a preinstalled `markmaton` CLI.
