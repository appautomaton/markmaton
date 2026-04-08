# Local Smoke Flow

## Purpose

This is the manual smoke path for checking the real Go engine end to end.

It is intentionally manual.

Automated coverage for `markmaton` should stay unit-test-first.

## Build the engine

From the repo root:

```bash
mkdir -p bin
go build -o bin/markmaton-engine ./cmd/markmaton-engine
```

## Run the Go tests

```bash
go test ./...
```

## Run the Python unit tests

```bash
python3 -m unittest discover -s tests -p 'test_*.py'
```

## Smoke the real engine with a fixture

```bash
python3 - <<'EOF'
import json
import pathlib
import subprocess

html = pathlib.Path("testdata/fixtures/article.html").read_text()
payload = {
    "url": "https://example.com/articles/harnessing-parsers",
    "html": html,
}

result = subprocess.run(
    ["./bin/markmaton-engine"],
    input=json.dumps(payload),
    text=True,
    capture_output=True,
    check=True,
)

print(result.stdout)
EOF
```

The easier real check is through the Python CLI:

```bash
MARKMATON_ENGINE=./bin/markmaton-engine \
python3 -m markmaton.cli convert \
  --html-file testdata/fixtures/article.html \
  --url https://example.com/articles/harnessing-parsers \
  --output-format markdown
```

## What to look for

- the command exits successfully
- markdown is non-empty
- links are absolute when a URL is provided
- obvious layout noise is removed
- code blocks stay readable
- quality fields are present in JSON mode

## When to use this

Use the smoke flow when:

- a unit test passes but the real binary behavior is still suspicious
- the packaging path changes
- the CLI or engine contract changes

Do not make this the default automated test path.
