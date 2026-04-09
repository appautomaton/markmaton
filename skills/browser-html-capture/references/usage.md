# Usage

The primary entrypoint is the bundled PEP 723 script:

```bash
uv run --script scripts/capture_html.py ...
```

## Capture to JSON

This is the default mode and the preferred output for chaining:

```bash
uv run --script scripts/capture_html.py \
  https://example.com/article \
  --output-format json
```

The JSON envelope includes:

- `html`
- `url`
- `final_url`
- `content_type`

It may also include narrow helper fields such as `title` or `rendered`.

## Capture to raw HTML

Use this mode when you want to pipe rendered HTML directly into another step:

```bash
uv run --script scripts/capture_html.py \
  https://example.com/article \
  --output-format html
```

## Wait controls

Use `--wait-selector` when the page is loaded only after a stable DOM hook appears:

```bash
uv run --script scripts/capture_html.py \
  https://example.com/article \
  --wait-selector article
```

Use `--wait-text` when the page has no stable selector but a clear visible string:

```bash
uv run --script scripts/capture_html.py \
  https://example.com/article \
  --wait-text "Read more"
```

Use `--timeout` to control the maximum wait time in seconds:

```bash
uv run --script scripts/capture_html.py \
  https://example.com/article \
  --wait-selector main \
  --timeout 15
```

## Headless mode

The script defaults to `--headless auto`.

Override only when needed:

```bash
uv run --script scripts/capture_html.py https://example.com --headless on
uv run --script scripts/capture_html.py https://example.com --headless off
```

## Composition with `html-to-markdown`

### Pipe raw HTML directly

```bash
uv run --script scripts/capture_html.py \
  https://example.com/article \
  --output-format html | \
uv run --script ../html-to-markdown/scripts/markmaton_convert.py \
  --url https://example.com/article \
  --output-format json
```

### Save the JSON envelope first

```bash
uv run --script scripts/capture_html.py \
  https://example.com/article \
  --output-format json > capture.json
```

Then pass `capture.json["html"]` and its context fields into the companion conversion skill.
