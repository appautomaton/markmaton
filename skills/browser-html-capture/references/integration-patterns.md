# Integration Patterns

## Output contract

The capture skill produces a narrow handoff contract:

- `html`
- `url`
- `final_url`
- `content_type`

Downstream steps should rely on these fields, not on browser-specific internals.

`url` is the requested source URL.
`final_url` is the browser's final location after redirects or navigation.
`content_type` is best-effort browser-side context from `document.contentType`.

## When to use a real browser

Prefer this skill for:

- app-shell pages
- commerce pages
- form-heavy pages
- client-side card/list pages
- modern landing pages with substantial hydration
- pages where a simple HTTP fetch does not produce meaningful content

Prefer a simple fetch for:

- static article pages
- wiki pages
- docs pages that are mostly server-rendered
- direct JSON or API endpoints

## Good defaults

- Start with `--output-format json`.
- Keep `--headless auto` unless the task explicitly needs a visible browser.
- Add `--wait-selector` or `--wait-text` only when the page needs a stronger readiness signal than `<body>`.
- Avoid turning capture into a site-specific automation stack unless the task explicitly requires that.

## Intended pipeline shape

```text
URL
-> browser-html-capture
-> HTML envelope
-> html-to-markdown
-> Markdown or JSON
```

## Relationship to `html-to-markdown`

This skill does not convert HTML.
Its job is to get a high-fidelity browser-rendered HTML snapshot.

Use the companion `html-to-markdown` skill when you need:

- Markdown
- parser metadata
- links
- images
- quality signals
