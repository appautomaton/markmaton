# markmaton

Lightweight HTML-to-Markdown tooling for agent workflows.

## Status

This repository is intentionally starting small.

The current goal is to build a clean, fast parser core that can:

- take normalized page HTML from tools like Playwright, `fetch`, or no-driver
- clean the page structure
- return robust Markdown and page metadata

## Direction

- parser core: Go
- distribution: Python packaging / PyPI
- first focus: library and CLI for local agent use

