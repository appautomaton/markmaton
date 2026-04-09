# Firecrawl Scrape Traceback

## Why this document exists

We do not want to tune `markmaton` against one pretty page at a time.

We want to trace how Firecrawl gets from a URL to usable Markdown, so we can separate:

- page-fetching concerns
- main-content cleaning concerns
- HTML-to-Markdown concerns
- post-processing concerns
- quality/fallback concerns

That gives us a cleaner answer to:

- which parts of Firecrawl are actually helping
- which parts `markmaton` should imitate
- which parts `markmaton` should avoid
- which changes in `markmaton` are structural, not page-specific

## End-to-end pipeline

At a high level, Firecrawl's `v2/scrape` flow looks like this:

```text
request body
-> route/controller validation
-> scrape engine selection
-> raw or rendered HTML
-> HTML cleaning / main-content transform
-> HTML to Markdown conversion
-> Markdown post-process
-> optional downstream transforms (summary/json/query/extract)
```

For our purposes, the core chain is:

```text
HTML in
-> transformHtml(...)
-> parseMarkdown(...)
-> postProcessMarkdown(...)
-> markdown out
```

Relevant source files:

- [v2 scrape route](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/routes/v2.ts)
- [v2 scrape controller](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/controllers/v2/scrape.ts)
- [scrapeURL engine registry](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/engines/index.ts)
- [scrapeURL transformer chain](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/transformers/index.ts)
- [HTML cleaning wrapper](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/lib/removeUnwantedElements.ts)
- [Markdown wrapper](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/lib/html-to-markdown.ts)

## Layer 1: where the HTML comes from

Firecrawl does not always use the same kind of HTML.

### Fetch engine

The fetch engine returns the HTTP response body more or less as-is.

Relevant file:

- [fetch engine](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/engines/fetch/index.ts)

Important behaviors:

- uses `undici.fetch`
- follows redirects
- re-decodes response text using detected charset when possible
- returns `html: response.body`

This is closest to "raw response HTML".

### Playwright engine

The Playwright engine returns browser-produced page content.

Relevant file:

- [playwright engine](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/engines/playwright/index.ts)

Important behaviors:

- calls the Playwright microservice
- waits after load when needed
- returns `html: response.content`

This is much closer to "rendered DOM HTML".

### Why this matters for markmaton

This is the first important lesson:

**When we compare `markmaton` to Firecrawl, we must compare on the same HTML class.**

That means:

- static pages can be compared on fetched HTML
- JS-heavy apps should be compared on rendered HTML

Otherwise we end up blaming the parser for a fetch-layer mismatch.

This explains why:

- raw `curl` against some OpenAI pages gave `403`
- Firecrawl still succeeded
- Playwright-rendered HTML gave `markmaton` a fairer input

## Layer 2: HTML cleaning and main-content extraction

Firecrawl's JS wrapper calls `htmlTransform(...)`, which is mostly delegated to the Rust package `@mendable/firecrawl-rs`.

Wrapper:

- [removeUnwantedElements.ts](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/lib/removeUnwantedElements.ts)

Rust implementation:

- [html.rs](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/native/src/html.rs)

### What it does

The Rust transform:

- resolves `<base href>`
- removes:
  - `head`
  - `meta`
  - `noscript`
  - `style`
  - `script`
- supports explicit `include_tags`
- supports explicit `exclude_tags`
- supports `only_main_content`
- normalizes `img[srcset]` to the biggest candidate
- absolutizes:
  - `img[src]`
  - `a[href]`

The wrapper has the same shape and a cheerio fallback:

- hard-coded non-main selectors
- force-include selectors
- relative URL fixing
- `srcset` handling

### What "only main content" really means here

This is not a semantic article extractor in the strong sense.

The core behavior is mostly:

- remove obvious shell regions with a fixed selector list
- optionally drop nodes matching OMCE signatures
- keep some forced-inclusion islands

More precisely:

- `include_tags` is a hard scope, not a hint
- `exclude_tags` runs before the hardcoded shell-removal pass
- invalid `exclude_tags` selectors quietly no-op
- invalid `include_tags` selectors error
- there is no independent readability/text-density algorithm here

Important constants:

- `EXCLUDE_NON_MAIN_TAGS` in [html.rs](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/native/src/html.rs)
- `FORCE_INCLUDE_MAIN_TAGS` in [html.rs](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/native/src/html.rs)

So the main-content layer is best described as:

**selector-based shell suppression with a few extra heuristics**, not magical content understanding.

### What this means for markmaton

This gives us a concrete direction.

We should not treat `only_main_content` as a vague quality goal.

We should treat it as a deterministic subsystem with:

- explicit shell-removal rules
- optional user selectors
- optional future site signatures
- conservative fallback to full-content mode

## Layer 3: HTML to Markdown conversion

Firecrawl's Node wrapper does not do the real conversion itself.

Relevant file:

- [html-to-markdown.ts](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/lib/html-to-markdown.ts)

It tries, in order:

1. HTTP markdown service
2. Go shared library via `koffi`
3. JS fallback using:
   - `turndown`
   - `joplin-turndown-plugin-gfm`

Every path ends with:

- `postProcessMarkdown(...)` from `@mendable/firecrawl-rs`

### Important consequence

Firecrawl's Node layer is orchestration.

The real Markdown behavior mostly lives in:

- [html-to-markdown repo](/Users/ac/dev/agents/firecrawl/html-to-markdown)

That repo is where headings, lists, links, tables, code fences, and whitespace behavior actually come from.

One useful boundary to keep in mind:

`postProcessMarkdown(...)` in the Rust crate is narrow. It mostly:

- escapes multiline link text
- removes `Skip to Content` anchor links

So if Firecrawl output looks well-structured, most of that structure came from:

- the HTML cleaning layer
- the Go Markdown conversion core

not from a giant magic post-process phase.

## Layer 4: what the Go markdown core actually does

The `html-to-markdown` repo is not a trivial wrapper around a generic package.

Key files:

- [from.go](/Users/ac/dev/agents/firecrawl/html-to-markdown/from.go)
- [commonmark.go](/Users/ac/dev/agents/firecrawl/html-to-markdown/commonmark.go)
- [utils.go](/Users/ac/dev/agents/firecrawl/html-to-markdown/utils.go)
- [plugin_test.go](/Users/ac/dev/agents/firecrawl/html-to-markdown/plugin_test.go)
- [list_regression_test.go](/Users/ac/dev/agents/firecrawl/html-to-markdown/list_regression_test.go)
- [code_block_test.go](/Users/ac/dev/agents/firecrawl/html-to-markdown/code_block_test.go)

### Structural features

The converter has:

- a rule system per tag
- before hooks
- after hooks
- plugin support
- snapshot-based configuration for conversion

This is much closer to a parser engine than a one-shot helper.

### The default conversion model

The CommonMark rules in [commonmark.go](/Users/ac/dev/agents/firecrawl/html-to-markdown/commonmark.go) already encode a lot of subtle behavior:

- list container handling
- multiline list indentation
- heading style choices
- inline spacing around emphasized text
- absolute URL resolution in links and images
- fenced or indented code blocks
- blockquote handling

Some especially relevant details:

- `<p>` and `<div>` are treated as block-ish containers with surrounding blank lines
- headings collapse internal newlines and become either ATX or setext headings
- link contents are escaped for multiline behavior
- inline text spacing is repaired with neighbor-aware helpers
- list indentation is precomputed in a DOM pre-pass

### Tests that matter

The test surface is much richer than Firecrawl's API-layer tests.

Particularly relevant:

- plugin golden tests:
  - [plugin_test.go](/Users/ac/dev/agents/firecrawl/html-to-markdown/plugin_test.go)
- list regressions:
  - [list_regression_test.go](/Users/ac/dev/agents/firecrawl/html-to-markdown/list_regression_test.go)
- code-block regressions:
  - [code_block_test.go](/Users/ac/dev/agents/firecrawl/html-to-markdown/code_block_test.go)
- performance smoke tests:
  - `perf_*_test.go` in the repo root

This is a strong signal that Firecrawl's robustness is not just in selector cleaning.
It also comes from a real Markdown conversion core with regression history.

## Layer 5: fallback behavior

One simple but important behavior in Firecrawl:

- if `onlyMainContent=true`
- and Markdown comes back empty
- it reruns using full content

Relevant code:

- [deriveMarkdownFromHTML fallback](/Users/ac/dev/agents/firecrawl/firecrawl/apps/api/src/scraper/scrapeURL/transformers/index.ts)

This fallback is too narrow for our needs because it only checks "empty or not".

But the architectural idea is still correct:

**main-content extraction should be reversible when the result quality is poor.**

## What our page comparisons taught us

Looking at real pages gave a clearer pattern than code alone.

### Pattern A: marketing / newsroom / job detail pages

Examples:

- `openai.com/news/engineering`
- `openai.com/careers`
- Ashby job detail pages

What we saw:

- `markmaton` is already close on these
- content coverage is good
- the remaining differences are mostly formatting:
  - card spacing
  - inline text glue
  - heading style polish

This means our current architecture is already viable here.

### Pattern B: large card search / listing pages

Example:

- `openai.com/careers/search`

What we saw:

- content is mostly there
- but block boundaries are weak
- cards compress into dense inline paragraphs
- "apply now" type links get glued to neighboring content

This points to a missing **block-boundary / card-boundary layer**, not a missing fetch layer.

### Pattern C: app-shell / docs-shell pages

Examples:

- `developers.openai.com/blog`
- GitHub repository pages

What we saw:

- shell content leaks in heavily
- search / nav / suggestion / session UI survives too long
- quality scoring is too optimistic
- on GitHub pages, the file table and shell are both still present

This points to a missing **shell suppression layer** and a more discriminating **quality layer**.

## What this means for markmaton

We should group changes by subsystem.

### Change group 1: input discipline

We should evaluate `markmaton` against the right HTML source:

- fetched HTML for static pages
- rendered HTML for JS-heavy pages

This is a workflow rule, not a parser change.

### Change group 2: stronger cleanhtml

We need a more deliberate shell-removal layer:

- expand deterministic shell selectors carefully
- support better include/exclude selector combinations
- consider optional signature-based shell removal later
- keep fallback to full-content mode

This is the most important area for docs/blog/app-shell pages.

### Change group 3: block-aware postprocess

We need a stronger post-process layer for:

- splitting glued inline text
- separating card boundaries
- keeping CTA links from collapsing into adjacent content
- handling repeated hero/card image patterns

This is the most important area for listing pages and card grids.

### Change group 4: markdown-structure improvements

Firecrawl gets a lot from a real Markdown core.

If we stay with our current lighter converter path, we need to deliberately close the gap in:

- heading rendering
- list indentation
- table handling
- code-block preservation
- multiline link behavior

This may eventually justify:

- either a stronger internal block tree
- or a more capable conversion core behind our Go engine

### Change group 5: quality heuristics that actually mean something

Our current score is too forgiving.

It should penalize things like:

- shell keywords dominating the opening section
- repeated nav/search/menu patterns
- suspicious error-state strings
- low information density despite high character count
- table-heavy output with mostly empty cells
- too many links relative to usable text

We need the quality layer to detect "long but bad".

## What not to copy from Firecrawl

We should not blindly inherit:

- multi-runtime orchestration complexity
- service + shared library + JS fallback all at once
- remote-engine assumptions
- LLM-adjacent transforms mixed into parser decisions

Those are useful in Firecrawl as a product, but they are not our parser core.

## Practical next step

The right next move is not "tune another page".

The right next move is:

1. turn the change groups above into parser work items
2. add fixtures that match those page classes
3. improve one subsystem at a time:
   - cleanhtml
   - postprocess
   - quality
   - conversion core

That gives us a general parser, not a page-by-page patch set.
