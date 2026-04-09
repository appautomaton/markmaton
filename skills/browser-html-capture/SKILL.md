---
name: browser-html-capture
description: Use when you do not have page HTML yet and need a real browser to capture rendered HTML before handing it to a downstream parser such as html-to-markdown. Best for JS-heavy pages, app-shell pages, form-heavy flows, and other cases where a simple HTTP fetch is not enough.
---

# Browser HTML Capture

Use this skill when:

- you need rendered HTML from a real browser
- the page depends on JavaScript, hydration, or client-side rendering
- you want a narrow one-shot capture step, not a full browser toolkit
- you want to hand the captured HTML into a downstream parser such as `html-to-markdown`

Do not use this skill to convert HTML into Markdown.
This skill stops at HTML capture and page context.

Use the bundled PEP 723 script as the default execution path:

- `scripts/capture_html.py`

Run it with:

- `uv run --script scripts/capture_html.py ...`

This keeps the skill self-contained and isolates dependencies from the caller's runtime.

## Workflow

1. Decide whether the page really needs a real browser.
2. Capture rendered HTML with the bundled script.
3. Prefer the default `json` output when another step needs `html`, `url`, `final_url`, or `content_type`.
4. Use `html` output only when you want to pipe raw HTML directly into a downstream step.
5. Pass the captured HTML into the companion `html-to-markdown` skill when you need Markdown or parser metadata.

## Read these references as needed

- For command syntax and script behavior:
  - `references/usage.md`
- For capture defaults, the handoff contract, and when to prefer a browser over a simple fetch:
  - `references/integration-patterns.md`

Only read the reference file you need for the current task.
