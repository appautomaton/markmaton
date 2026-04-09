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
- release track: GitHub Actions + Trusted Publishing

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

## Testing layout

- Go package unit tests live beside each package under `internal/*`.
- Shared Go fixture/golden helpers live in `internal/testutil/`.
- Stable parser fixtures live under `testdata/fixtures/core/`.
- Real-world regression fixtures live under `testdata/fixtures/regression/`.
- Golden markdown outputs for stable core fixtures live under `testdata/golden/core/`.
- Python wrapper tests live under `tests/unit/`.

## Local smoke

See:

- `docs/local-smoke.md`
- `docs/packaging-layout.md`
- `docs/pypi-release.md`
