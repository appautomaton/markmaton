# Integration Patterns

## Input contract

`markmaton` expects:

- HTML content
- optional `url`
- optional `final_url`
- optional `content_type`

The parser works best when `url` is provided.
It improves canonical fallback and absolute link normalization.

## HTML source choice

### Prefer fetched HTML for:

- article pages
- wiki pages
- docs pages that are mostly server-rendered
- discussion pages that already render meaningful HTML on the server

### Prefer rendered HTML for:

- app-shell pages
- commerce pages
- form-heavy pages
- client-side card/list pages
- modern landing pages with substantial hydration

## Good defaults

- Start with main-content mode.
- Use `--full-content` only when main-content extraction is clearly too aggressive.
- Use `--include-selector` or `--exclude-selector` only when a page has a stable structural reason for it.
- Do not turn parser usage into site-specific patching unless the task explicitly calls for that.

## Intended pipeline shape

```text
upstream capture tool
-> HTML
-> markmaton
-> Markdown or JSON
```

Typical upstream tools:

- the companion `browser-html-capture` skill
- Playwright
- fetch / requests / httpx
- Firecrawl
- another browser automation or scraping layer
