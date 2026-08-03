# Benchmark and Regression Workflow

## Purpose

Benchmarking exists to improve Markmaton by reusable page pattern, not by site-specific patching. The workflow separates exploratory page captures from durable automated regression coverage.

## Working directories

Keep exploratory captures outside the committed corpus:

```text
tmp/benchmarks/<slug>/
  manifest.md
  fetched.html
  rendered.html
  baseline.json
  candidate.json
  notes.md
```

The `tmp/` tree is local-only. Commit a page only after it has exposed a repeatable parser failure.

## Capture rules

Use fetched HTML for server-rendered articles, documentation, wikis, and conventional discussion pages. Use rendered HTML for app shells, client-driven lists, job applications, media-heavy pages, and pages whose fetched HTML lacks the visible content.

Record:

- source URL
- capture method
- page class
- whether the input is fetched or rendered
- the parser behavior under investigation
- any privacy or licensing cleanup applied before promotion

## Evaluation loop

### 1. Capture stable HTML

Use a browser or fetch layer outside Markmaton. The bundled skill can capture rendered HTML:

```bash
uv run python skills/html-to-markdown/scripts/capture_html.py \
  https://example.com \
  --output-format html \
  > tmp/benchmarks/example/rendered.html
```

Never add live-network requests to the automated parser suite.

### 2. Run the current engine

```bash
uv run python -m markmaton.cli convert \
  --html-file tmp/benchmarks/example/rendered.html \
  --url https://example.com \
  --output-format json \
  > tmp/benchmarks/example/candidate.json
```

Compare against the last known-good Markmaton output when one exists. External converters may provide research context, but they are not an executable oracle or a required local dependency.

### 3. Attribute the gap

Use the narrowest responsible layer:

- `cleanhtml` — shell, dialogs, navigation, hidden content, or wrong main-content scope
- `resolve` / `safeurl` — relative URLs, `srcset`, media destinations, or unsafe schemes
- `convert/native` — block structure, inline structure, lists, tables, code, links, images, or media
- conversion policies — generic control lines, duplicate opening content, or low-risk Markdown cleanup
- `postprocess` — final spacing and formatting normalization
- `metadata`, `links`, `images` — extraction-specific defects
- `quality` — misleading scores or fallback decisions

### 4. Minimize before promotion

Prefer a small synthetic fixture when it reproduces the failure. Preserve a real-world snapshot only when surrounding DOM structure is essential and cannot be represented faithfully by a minimized sample.

Remove unrelated scripts, personal data, volatile tokens, and irrelevant page content when doing so does not destroy the failure mode.

## Promotion destinations

- `testdata/fixtures/regression/` and `testdata/golden/regression/` for Markmaton-owned end-to-end failures
- `testdata/fixtures/compatibility/` and `testdata/golden/compatibility/` for reusable HTML-to-Markdown contract coverage
- `testdata/fixtures/performance/` for deterministic large inputs

Every promoted file must be declared by the corresponding manifest or engine fixture table. Orphan detection is part of the automated suite.

## Golden policy

Use exact Goldens for deterministic Markmaton output. Use semantic comparison only for an external historical reference, never as a substitute for Markmaton's own exact `expected.md`.

Refresh conversion Goldens only after reviewing the renderer change:

```bash
MARKMATON_UPDATE_CONVERSION_GOLDENS=1 \
  go test ./internal/convert -run '^TestConversionCorpus$' -count=1
```

## Promotion criteria

Promote a case only when it:

- represents a reusable parser pattern
- protects content that could otherwise be silently lost or corrupted
- adds behavior not already covered by a smaller fixture
- is deterministic and network-independent
- has a clear expected result or explicit behavioral assertions

Do not promote a page merely because it is prominent, visually complex, or currently popular.

## Verification

Run the normal suite after every promotion:

```bash
go test ./...
go vet ./...
uv run python -m unittest discover -s tests -p 'test_*.py'
```

Run the large corpus explicitly when conversion or rendering behavior changes:

```bash
MARKMATON_RUN_LARGE_CORPUS=1 \
  go test ./internal/convert/native -run '^TestLargePerformanceCorpus$' -count=1
```
