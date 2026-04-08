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

## Current shape

- Go engine: `cmd/markmaton-engine`
- Python wrapper: `markmaton/`
- Architecture docs: `docs/`
- Plans and issue CSVs: `plan/` and `issues/`

## Testing policy

- automated tests should be unit-test-first
- parser module tests should use local fixtures and golden files
- Python wrapper tests should mock the engine boundary
- real engine checks stay manual unless there is a strong reason to automate them

## Local smoke

See:

- `docs/local-smoke.md`
- `docs/packaging-layout.md`
