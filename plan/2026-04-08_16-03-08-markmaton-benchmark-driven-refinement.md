---
mode: plan
task: markmaton benchmark-driven refinement
created_at: "2026-04-08T16:03:08-04:00"
complexity: complex
status: completed
---

# Plan: markmaton benchmark-driven refinement

## Summary
- Establish a formal benchmark workflow that compares `markmaton` and Firecrawl behaviorally across representative page classes.
- Use real pages for discovery, cache their HTML locally, and only promote representative samples into regression fixtures.
- Turn benchmark findings into a parser-layered next slice, grouped by `cleanhtml`, `postprocess`, `convert/core`, and `quality`.
- Keep parser interfaces unchanged in this round.

## Key changes

### Benchmark artifacts
- Add an English benchmark workflow doc that defines sampling, local cache layout, HTML acquisition rules, comparison method, gap attribution, and fixture-promotion rules.
- Add an English benchmark matrix with the initial page set and current observations for Firecrawl and `markmaton`.
- Keep benchmark caches under `tmp/benchmarks/markmaton-benchmark-driven-refinement/` and out of version control.

### Benchmark scope
- Lock the first benchmark set to cover: article, docs, wiki, card/list grid, careers landing, jobs landing, application form, discussion/thread, repo/app shell, issue timeline, and product/commerce pages.
- For each page, define whether comparison should use fetched HTML or rendered HTML.
- Use Firecrawl `/v2/scrape` as the reference scrape implementation, but compare behaviorally rather than chasing byte-for-byte output parity.

### Next refinement slice definition
- Derive the next parser work from benchmark patterns, not site names.
- Keep `fallback` frozen for the next slice.
- Prefer `cleanhtml` or `postprocess` when the problem is shell leakage or obvious UI residue.
- Escalate to `convert/core` only when the same structural issue persists after clean input and low-risk postprocessing.

## Tests & verification
- Validate the new issue CSV against the repository CSV specification.
- Confirm benchmark docs and matrix are in `docs/` and written in English.
- Preserve existing Go and Python automated tests; no new live network automation is added.
- Manual verification for this planning round is complete when each benchmark row has an HTML mode, current Firecrawl behavior summary, current `markmaton` status or explicit pending capture note, gap attribution, and promotion decision.

## Assumptions
- Public parser request/response shapes stay unchanged.
- Real URLs are discovery inputs only; cached HTML is used for repeated analysis.
- Only a curated subset of benchmark pages should become regression fixtures.
- Firecrawl is a mature behavioral reference, not the exact formatting target.

## Current benchmark outcome
- First-pass benchmark artifacts now exist under `tmp/benchmarks/markmaton-benchmark-driven-refinement/`.
- Each benchmark row now has a local Firecrawl `/v2/scrape` snapshot, a cached HTML artifact, and a local `markmaton` JSON output.
- The benchmark set confirms that `markmaton` is already strong on careers landing pages, application/detail pages, and clean docs pages.
- The benchmark set also confirms that thread/timeline pages, heavy-shell article pages, repo/app shell pages, and some product/commerce pages still expose general parser gaps.

## Next slice recommendation by parser layer

### `cleanhtml`
- Prioritize article-shell cleanup for media pages that still open on promos, top stories, or non-body blocks.
- Tighten wiki and timeline shell suppression where auxiliary tabs, stale-session chrome, and repo shell still survive.
- Keep form/detail pages conservative so valid application content is not over-pruned.

### `postprocess`
- Continue low-risk work on card/list readability and product-page promo spacing.
- Improve compressed inline promo text and dense CTA runs when the block structure is otherwise usable.
- Do not use `postprocess` to fake full thread/timeline structure; that problem is too structural.

### `convert/core`
- Start planning a `markmaton`-owned converter customization layer for repeated structural failures.
- The strongest triggers are discussion threads, issue timelines, and repo/app-shell pages where clean input still falls into awkward tables or dense block layouts.
- Product and jobs-grid pages may also need converter-level help if card/tile organization remains image-led after cleanhtml and postprocess improvements.

### `quality`
- Make scoring less optimistic for shell-heavy articles, timeline pages, and degraded form/application pages.
- Keep `fallback` frozen for now, but use the benchmark set to identify where current scores are clearly overstating output quality.
