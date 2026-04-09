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

| ID | Page class | URL | HTML mode | Firecrawl `/v2/scrape` snapshot | `markmaton` current status | Likely gap layer | Fixture decision |
| --- | --- | --- | --- | --- | --- | --- | --- |
| BM-01 | Card/list grid | [OpenAI Engineering](https://openai.com/news/engineering/) | rendered | Good coverage, but still opens with site shell and packs card metadata densely | Strong content retention; opens directly on `Engineering`; generic list controls are gone, but card blocks still feel dense | `postprocess` first, then `convert/core` if repeated card-block awkwardness remains | Promote / already represented by `regression/card_grid.html` |
| BM-02 | Careers landing | [OpenAI Careers](https://openai.com/careers/) | rendered | Strong on hero and feature sections, though it still opens with brand shell | Stable baseline; hero copy and CTA survive well; only a light `Company` shell line remains | baseline protection only; no new parser work driven by this row | Keep as regression baseline |
| BM-03 | Job detail / application | [Ashby application page](https://jobs.ashbyhq.com/openai/49a16d46-bf3e-4806-a8af-a0e48c26336c/application) | rendered | Strong on job detail and compensation sections | Stable baseline; preserves section headings, location, compensation, and benefits; still carries a top wordmark block | baseline protection only; no new parser work driven by this row | Keep as regression baseline |
| BM-04 | Jobs landing | [Thermo Fisher jobs landing](https://jobs.thermofisher.com/global/en) | rendered | Good on recruiting message and search-entry framing | Main headline survives, but output is image-led and tile labels are thin; weaker than OpenAI careers | `convert/core` first, with possible `postprocess` cleanup later | Keep local cache first |
| BM-05 | Application form | [Thermo Fisher apply](https://jobs.thermofisher.com/global/en/apply?jobSeqNo=TFSCGLOBALR01347309EXTERNALENGLOBAL) | rendered | Weak: almost empty markdown, banner-heavy, weak title metadata | Still weak: skip-link noise is gone and quality is now scored more honestly, but the page still falls back to header chrome, duplicate logos, loading text, and minimal useful form content | `cleanhtml` first, then `quality`; likely revisit HTML acquisition before deeper parser work | Keep local cache first |
| BM-06 | Discussion thread | [Hacker News thread](https://news.ycombinator.com/item?id=40508445) | fetched-first | Weak: table-heavy chrome and comment structure leak heavily into Markdown | Still weak: opening is table markup and nav chrome, not a readable discussion thread | `convert/core` after shell cleanup; thread/timeline layout is not salvageable with light postprocess alone | Candidate for promotion as a thread/timeline fixture |
| BM-07 | News article with heavy shell | [Yahoo News article](https://www.yahoo.com/news/articles/officials-surprised-impact-customer-habits-210000558.html?guccounter=1&guce_referrer=aHR0cHM6Ly93d3cuZ29vZ2xlLmNvbS8&guce_referrer_sig=AQAAAISaELtFiuTfiMrkD6MQ4LMAWD72pO_pLYCdVgO7G1NnfSbfJHJt__y0io-X2d_cZzBXMRIHurNfGZqDLB5Dk6okL3W27RpOS15soJx80eaYBv0avIJmiWlcTfSV1Z5Z6wpNOZTh5EKvQv-iIAr36nt-bD9AH8ddWNjrvRlh5ug1) | rendered-first | Weak opening: ads, homepage return link, top stories, promos | Improved materially: article content now surfaces near the top instead of opening on `Top Stories`, but heavy-shell article quality still scores too optimistically and article presentation still begins with image-heavy promo framing | `cleanhtml` first, then `quality` | Promote if we want a real heavy-shell article regression fixture |
| BM-08 | Wiki | [Wikipedia LLM page](https://en.wikipedia.org/wiki/Large_language_model) | fetched | Usable but retains wiki chrome, tabs, and maintenance notices | Article body is present, and quality now reflects the shell leakage better, but the page still opens with `Birthday mode`, article/talk tabs, and maintenance-table chrome | `cleanhtml` first; `postprocess` only for small cleanup if needed | Keep local cache first |
| BM-09 | Docs | [MDN HTTP Overview](https://developer.mozilla.org/en-US/docs/Web/HTTP/Overview) | fetched | Good body capture, but still keeps skip links and minor docs chrome | Strong result: opens directly on `# Overview of HTTP` and reads like real docs Markdown | baseline protection only; useful as a clean docs benchmark | Keep local cache first |
| BM-10 | Repo/app shell | [zellij repo page](https://github.com/zellij-org/zellij) | rendered-first | Good coverage, but still noisy and structurally dense | Better than the old shell-heavy state, but still opens with repo chrome and a very dense file table before the README body | `convert/core` first; `postprocess` can only shave small edges here | Keep as local rich benchmark; existing simplified repo shell fixture remains regression |
| BM-11 | Issue timeline | [VS Code issue #286040](https://github.com/microsoft/vscode/issues/286040) | rendered-first | Weak opening: skip links, stale-session alerts, repo chrome, auth chrome | Improved materially in the converter epic: it now opens on the issue title with a single status line, and duplicate issue actions/title echo are gone; assignee/label metadata is still dense before the main body | `convert/core` next for timeline metadata grouping; only light cleanup remains in outer layers | Promote for this converter-layer epic |
| BM-12 | Product/commerce | [Apple iPhone page](https://www.apple.com/iphone-16-pro/) | rendered-first | Strong page access, but output opens with dense promo/commerce blocks | Main product sections survive, but the page still opens with compressed promo text and commerce CTA density; title/canonical also skew to the broader iPhone hub | `cleanhtml` and `postprocess` first; escalate to `convert/core` only if hero/commerce blocks remain structurally awkward | Keep local cache first |

## Current takeaways

### Firecrawl is not uniformly clean

Current samples show:

- good results on careers landing and some docs pages
- middling results on heavy article shells, wiki chrome, and forms
- weak results on thread/timeline pages and GitHub-style app pages
- good page access does not automatically mean good article or product-page prioritization

This reinforces the right comparison stance:

- use Firecrawl as a mature behavioral reference
- do not chase exact output parity

### `markmaton` is already competitive on several classes

Current local observations show:

- careers landing pages are already strong
- application/detail pages are already strong
- docs pages are already close to “good enough” for the current parser shape
- shell-heavy pages have improved materially after cleanhtml hardening
- heavy-shell article pages now respond to better root selection, which is a sign that this layer is still worth improving
- card/list pages are improving, but still need a more mature organization layer
- thread/timeline and repo/app pages are now clearly the strongest argument for a later `convert/core` customization layer

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

| ID | Page class | URL | HTML mode | Firecrawl `/v2/scrape` snapshot | `markmaton` current status | Likely gap layer | Fixture decision |
| --- | --- | --- | --- | --- | --- | --- | --- |
| BM-13 | Q&A / answer thread | [Stack Overflow question](https://stackoverflow.com/questions/1732348/regex-match-open-tags-except-xhtml-self-contained-tags) | fetched | Very noisy opening with Collectives shell, promo copy, and community chrome before the question body | Improved materially in the converter epic: the Markdown now opens directly on the question body and strips the most obvious interaction controls, but answer-score/metadata lines are still dense and quality remains too optimistic | `quality` first; later `convert/core` if answer hierarchy remains hard to read | Promote for this converter-layer epic |
| BM-14 | PR diff / files changed | [zellij PR files changed](https://github.com/zellij-org/zellij/pull/5012/files) | rendered | Opens with skip links, stale-session chrome, auth chrome, and repo shell before the PR body | Still opens with repo chrome, notifications, and repo stats before the PR title; diff/files structure is not yet expressed cleanly | `cleanhtml` first, then `convert/core` | Strong candidate for a richer timeline/app-shell benchmark |
| BM-15 | API docs with tabs and code samples | [Stripe API docs](https://docs.stripe.com/api/payment_intents/create) | rendered | Noisy docs shell with search and docs navigation before endpoint content | Good body prioritization: opens directly on the endpoint section, but docs affordances like `Was this section helpful?YesNo` still leak through | `cleanhtml` first, then `postprocess` | Keep local cache first |
| BM-16 | Product comparison / commerce | [Apple iPhone compare](https://www.apple.com/iphone/compare/) | rendered | Accessible, but very long and commerce-heavy; opens on promo/trade-in copy before comparison content | Similar behavior: comparison title survives, but the page still opens with promo/trade-in copy and dense commerce scaffolding | `postprocess` first, then `convert/core` only if comparison blocks remain structurally awkward | Keep local cache first |

## What the harder tier changes

The second-tier set sharpens the next-phase picture:

- `BM-13` shows that a general parser can beat Firecrawl on opening priority while still being weak on thread/answer structure.
- `BM-14` is a stronger signal than the plain repo page that app-shell and timeline-like views are now a converter-layer problem.
- `BM-15` confirms that docs pages can already be strong without site-specific rules, as long as shell cleanup keeps improving.
- `BM-16` shows that long commerce/comparison pages are not only shell-heavy; they also challenge block organization and ranking of what should appear first.

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
