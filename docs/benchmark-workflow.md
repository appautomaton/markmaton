# Benchmark Workflow

## Purpose

This workflow exists to keep `markmaton` general.

We do **not** want to tune the parser against one attractive page at a time.
We want a repeatable benchmark loop that:

- samples real pages
- caches their HTML locally
- compares Firecrawl and `markmaton` behaviorally
- attributes gaps to the correct parser layer
- promotes only representative pages into long-term regression fixtures

## Comparison standard

Benchmark comparisons are **behavioral**, not byte-for-byte Markdown comparisons.

For each page, compare:

- title quality
- main content retention
- shell/chrome leakage
- card/list readability
- metadata quality
- link and image normalization
- quality-score honesty
- whether fallback was required

Do **not** treat Firecrawl output as the exact formatting target.
Use it as a mature reference point for behavior.

## HTML acquisition rules

Pick the HTML mode before comparing parser quality.

### Use fetched HTML first for:

- traditional article pages
- wiki pages
- docs pages that are largely server-rendered
- forums/discussion pages that render meaningful HTML without client hydration

### Use rendered HTML first for:

- app-shell pages
- modern marketing/product pages
- job applications and multi-step forms
- card/list pages driven by client UI
- heavy media pages with client-side personalization or shell logic

If a fetched page is obviously incomplete or blocked, promote the page to rendered-first.

## Local benchmark cache layout

Cache all benchmark work under:

```text
tmp/benchmarks/<slug>/
```

Recommended contents:

```text
tmp/benchmarks/<slug>/
  manifest.md
  fetched.html
  rendered.html
  firecrawl-scrape.json
  markmaton.json
  notes.md
```

Rules:

- cache locally first
- do not add live network tests to the automated suite
- do not commit cached benchmark runs
- only promote a page into `testdata/fixtures/regression/` if it exposes a reusable parser failure mode

## Sampling procedure

### 1. Capture page metadata

Record:

- canonical URL
- page class
- chosen HTML mode
- why that HTML mode was chosen

### 2. Capture HTML

Use one or both:

- fetched HTML
- rendered HTML

When using Playwright-rendered HTML, save the raw browser result and, if needed, decode it into a plain `.html` artifact before running `markmaton`.

### 3. Run Firecrawl `/v2/scrape`

Use local Firecrawl as the reference scrape implementation.

Example:

```bash
curl -sS -X POST http://localhost:3002/v2/scrape \
  -H 'Content-Type: application/json' \
  --data '{"url":"https://example.com","formats":["markdown"]}'
```

Store the full JSON response in:

```text
tmp/benchmarks/<slug>/firecrawl-scrape.json
```

### 4. Run `markmaton`

Use cached HTML, not the live page.

Example:

```bash
uv run python -m markmaton.cli convert \
  --html-file tmp/benchmarks/<slug>/rendered.html \
  --url https://example.com \
  --output-format json
```

Store the output in:

```text
tmp/benchmarks/<slug>/markmaton.json
```

### 5. Write findings

For each page, summarize:

- what Firecrawl got right
- what `markmaton` got right
- the biggest gap
- the likely parser layer responsible
- whether the page should become a regression fixture

## Gap attribution rules

Use the narrowest responsible layer.

### `cleanhtml`

Use when the issue is:

- navigation, dialogs, alerts, overlays, banners, or shell residue
- body chrome surviving into Markdown
- obvious main-content mis-scoping

### `postprocess`

Use when the issue is:

- generic list controls surviving
- card entries needing spacing or separation
- label/date or metadata collisions
- low-risk Markdown cleanup after good conversion

### `convert/core`

Use when the issue remains after clean input and cannot be fixed cleanly in postprocess:

- block structure is awkward
- lists/tables/timelines collapse badly
- links/images are wrapped in persistently poor shapes
- the same structural problem appears across multiple sites

### `quality`

Use when the output is produced, but the scoring or fallback guidance is misleading.

## Promotion rules

Promote a page into `testdata/fixtures/regression/` only if all are true:

- it exposes a reusable parser pattern
- the failure is likely to recur on other sites
- a local synthetic fixture would miss the important structure
- the page adds coverage that current fixtures do not already provide

Do **not** promote pages just because they are interesting or high-profile.

## Initial benchmark classes

The benchmark set should cover:

- article
- docs
- wiki
- card/list grid
- careers landing
- job listing
- job detail / application form
- thread/discussion
- repo/app shell
- issue/pr timeline
- product/commerce page

## Exit criteria for a benchmark round

A benchmark round is complete when:

- each selected page has a cached HTML source
- each selected page has a Firecrawl scrape snapshot
- each selected page has a `markmaton` output snapshot
- each page has a gap attribution
- each page has a promotion decision
- the next parser slice is grouped by parser layer, not by site name
