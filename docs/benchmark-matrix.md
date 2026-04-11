# Benchmark Matrix

## Purpose

This matrix is the working baseline for parser refinement.

It records:

- which page classes we care about
- how we should acquire HTML for comparison
- how Firecrawl behaves today
- how `markmaton` behaves today
- where the likely parser gap lives
- whether the page should stay a local benchmark or become a regression fixture

## Initial benchmark set

First-pass benchmark artifacts now exist under:

```text
tmp/benchmarks/markmaton-benchmark-driven-refinement/
```

For this first pass, every benchmark row has:

- a local Firecrawl `/v2/scrape` snapshot
- a cached HTML artifact derived from Firecrawl `rawHtml`
- a local `markmaton` JSON output snapshot

This means the benchmark set is now grounded in cached artifacts rather than in-memory observations.
For rendered-first rows, we can still add an independent Playwright-rendered capture later if a remaining gap is ambiguous.

| ID | Page class | URL | HTML mode | Firecrawl | `markmaton` | Likely gap layer | Fixture |
| --- | --- | --- | --- | --- | --- | --- | --- |
| BM-01 | Card/list grid | [OpenAI Engineering](https://openai.com/news/engineering/) | rendered | Good coverage; opens with shell, dense card metadata | Strong retention; opens on `Engineering`; cards still dense | `postprocess`, then `convert/core` | Promote / `regression/card_grid.html` |
| BM-02 | Careers landing | [OpenAI Careers](https://openai.com/careers/) | rendered | Strong hero/features; opens with brand shell | Stable; hero and CTA survive; light shell line remains | Baseline — no parser work needed | Regression baseline |
| BM-03 | Job detail | [Ashby application](https://jobs.ashbyhq.com/openai/49a16d46-bf3e-4806-a8af-a0e48c26336c/application) | rendered | Strong on detail and compensation | Stable; headings, location, comp preserved; top wordmark remains | Baseline — no parser work needed | Regression baseline |
| BM-04 | Jobs landing | [Thermo Fisher jobs](https://jobs.thermofisher.com/global/en) | rendered | Good recruiting message and search framing | Headline survives; image-led, thin tile labels | `convert/core`, then `postprocess` | Local cache |
| BM-05 | Application form | [Thermo Fisher apply](https://jobs.thermofisher.com/global/en/apply?jobSeqNo=TFSCGLOBALR01347309EXTERNALENGLOBAL) | rendered | Weak: near-empty markdown, banner-heavy | Still weak: skip-link noise gone, but falls back to chrome and logos | `cleanhtml`, then `quality`; revisit HTML acquisition | Local cache |
| BM-06 | Discussion thread | [HN thread](https://news.ycombinator.com/item?id=40508445) | fetched-first | Weak: table chrome and comments leak into Markdown | Still weak: opens on table markup and nav chrome | `convert/core` after shell cleanup | Promote candidate (thread fixture) |
| BM-07 | Heavy-shell article | [Yahoo News](https://www.yahoo.com/news/articles/officials-surprised-impact-customer-habits-210000558.html?guccounter=1&guce_referrer=aHR0cHM6Ly93d3cuZ29vZ2xlLmNvbS8&guce_referrer_sig=AQAAAISaELtFiuTfiMrkD6MQ4LMAWD72pO_pLYCdVgO7G1NnfSbfJHJt__y0io-X2d_cZzBXMRIHurNfGZqDLB5Dk6okL3W27RpOS15soJx80eaYBv0avIJmiWlcTfSV1Z5Z6wpNOZTh5EKvQv-iIAr36nt-bD9AH8ddWNjrvRlh5ug1) | rendered-first | Weak: ads, promos, top stories before content | Improved: article surfaces near top; quality still scores too high | `cleanhtml`, then `quality` | Promote candidate (heavy-shell fixture) |
| BM-08 | Wiki | [Wikipedia LLM](https://en.wikipedia.org/wiki/Large_language_model) | fetched | Usable; retains wiki chrome, tabs, notices | Body present; still opens with tabs and maintenance chrome | `cleanhtml` first | Local cache |
| BM-09 | Docs | [MDN HTTP Overview](https://developer.mozilla.org/en-US/docs/Web/HTTP/Overview) | fetched | Good body; keeps skip links and minor chrome | Strong: opens on `# Overview of HTTP`, reads like real docs | Baseline — clean docs benchmark | Local cache |
| BM-10 | Repo/app shell | [zellij repo](https://github.com/zellij-org/zellij) | rendered-first | Good coverage; noisy and structurally dense | Better; still opens with repo chrome and dense file table | `convert/core` first | Local rich benchmark |
| BM-11 | Issue timeline | [VS Code issue #286040](https://github.com/microsoft/vscode/issues/286040) | rendered-first | Weak: skip links, stale-session alerts, repo chrome | Improved: opens on issue title; duplicate actions gone; metadata still dense | `convert/core` for timeline metadata grouping | Promote (converter epic) |
| BM-12 | Product/commerce | [Apple iPhone](https://www.apple.com/iphone-16-pro/) | rendered-first | Strong access; dense promo/commerce opening | Product sections survive; opens with promo text and CTA density | `cleanhtml` and `postprocess` first | Local cache |

## Current takeaways

### Firecrawl is not uniformly clean

- Strong on careers and docs pages
- Middling on heavy-shell articles, wiki chrome, forms
- Weak on thread/timeline and GitHub app pages

Use Firecrawl as a behavioral reference, not an output-parity target.

### `markmaton` is already competitive

- Careers, job detail, and docs pages are strong
- Shell-heavy pages improved after `cleanhtml` hardening
- Card/list pages improving but need a more mature organization layer
- Thread/timeline and repo/app pages are the strongest case for `convert/core` work

## Promotion candidates for the next parser slice

Highest-value candidates for curated regression fixtures after local capture:

1. a richer heavy-shell article page
2. a discussion/timeline page
3. one richer app-shell page beyond the simplified repo fixture
4. optionally, a failed or degraded application flow page if we want one “bad form page” benchmark

These are the places where synthetic fixtures are most likely to miss the real parser failure mode.

## Harder second-tier benchmark set

These rows deliberately raise the difficulty level.
They are still evaluated as **general parser patterns**, not as site-specific targets.

| ID | Page class | URL | HTML mode | Firecrawl | `markmaton` | Likely gap layer | Fixture |
| --- | --- | --- | --- | --- | --- | --- | --- |
| BM-13 | Q&A thread | [Stack Overflow](https://stackoverflow.com/questions/1732348/regex-match-open-tags-except-xhtml-self-contained-tags) | fetched | Noisy: Collectives shell and community chrome before body | Improved: opens on question body; answer metadata still dense | `quality`, then `convert/core` | Promote (converter epic) |
| BM-14 | PR diff | [zellij PR files](https://github.com/zellij-org/zellij/pull/5012/files) | rendered | Skip links, stale-session, auth chrome before PR body | Repo chrome and notifications before PR title; diff not clean | `cleanhtml`, then `convert/core` | Promote candidate (app-shell) |
| BM-15 | API docs | [Stripe API](https://docs.stripe.com/api/payment_intents/create) | rendered | Noisy docs shell before endpoint content | Good: opens on endpoint; minor affordance leaks remain | `cleanhtml`, then `postprocess` | Local cache |
| BM-16 | Product comparison | [Apple compare](https://www.apple.com/iphone/compare/) | rendered | Long, commerce-heavy; promo before comparison | Title survives; promo and commerce scaffolding still lead | `postprocess`, then `convert/core` | Local cache |

## What the harder tier changes

- `BM-13`: General parser beats Firecrawl on opening priority but is weak on answer structure.
- `BM-14`: Confirms app-shell/timeline views are a `convert/core` problem.
- `BM-15`: Docs pages are strong without site-specific rules if shell cleanup holds.
- `BM-16`: Commerce/comparison pages challenge block organization, not just shell removal.

## Promotion decision for the current converter epic

Promote these rows into repo-backed regression fixtures now:

1. `BM-13` Stack Overflow question
2. `BM-11` GitHub issue timeline

Keep these rows as local-only hard benchmarks for now:

1. `BM-14` GitHub PR files changed
2. `BM-15` Stripe API docs
3. `BM-16` Apple iPhone compare

This keeps the first converter epic focused on:

- one discussion/Q&A structure
- one app-shell/timeline structure

without pulling diff-table handling and commerce-comparison ranking into the first converter pass.

## Converter-layer epic update

The first converter customization pass has now landed for the two promoted rows:

- `BM-13` now opens directly on the question body and no longer carries the worst control-line clutter at the top.
- `BM-11` now opens directly on the issue title, with issue-action noise and the redundant linked title echo removed.

What remains after this pass is more clearly structural than shell-related:

- answer/timeline metadata density
- grouping of assignees/labels/status blocks
- quality scoring that is still too optimistic on structure-heavy pages
