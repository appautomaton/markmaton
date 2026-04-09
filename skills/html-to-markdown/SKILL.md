---
name: html-to-markdown
description: Use when you already have page HTML and need to convert it into clean Markdown with metadata using markmaton. Best for workflows where Playwright, fetch, Firecrawl, or another upstream tool has already captured the HTML, and you want a deterministic HTML-to-Markdown step rather than LLM extraction.
---

# HTML to Markdown

Use this skill when:

- you already have page HTML
- you want deterministic HTML-to-Markdown conversion
- you want Markdown plus metadata, links, images, or quality signals
- you want to use `markmaton` as a parser step, not as a crawler

Do not use this skill to visit URLs directly.
Get HTML first, then pass it into `markmaton`.

Use the bundled PEP 723 script as the default execution path:

- `scripts/markmaton_convert.py`

Run it with:

- `uv run --script skills/html-to-markdown/scripts/markmaton_convert.py ...`

This keeps the skill self-contained and isolates dependencies from the caller's runtime.

## Workflow

1. Decide whether the upstream page source should be fetched HTML or rendered HTML.
2. Pass the HTML into the bundled script.
3. Prefer including the source `url`.
4. Choose `markdown` or `json` output based on whether downstream steps need metadata and quality fields.

## Read these references as needed

- For command syntax and script behavior:
  - `references/usage.md`
- For choosing fetched vs rendered HTML, and for parser defaults:
  - `references/integration-patterns.md`

Only read the reference file you need for the current task.
