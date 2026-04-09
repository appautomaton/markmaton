---
mode: plan
task: markmaton implementation bootstrap
created_at: "2026-04-08T11:04:40-04:00"
complexity: complex
status: completed
---

# Plan: markmaton implementation bootstrap

## Goal
- Build the first implementation slice of `markmaton`: a parser core that turns input HTML into clean Markdown plus metadata and quality signals, without taking on page fetching or any LLM features.

## Scope
- In:
  - Build the Go engine skeleton
  - Define request/response models
  - Implement first-pass `cleanhtml`, `resolve`, `convert`, and `postprocess`
  - Implement first-pass `metadata`, `links`, and `images`
  - Add fixtures and golden-file tests
  - Add the first Python wrapper and CLI
- Out:
  - No Playwright / fetch / no-driver integration
  - No LLM extract / summary / query
  - No large site-specific plugin system
  - No full release automation yet

## Assumptions / Dependencies
- The architecture decision is captured in `/Users/ac/dev/agents/firecrawl/markmaton/docs/architecture-brief.md`
- The Firecrawl reference audit is captured in `/Users/ac/dev/agents/firecrawl/markmaton/docs/firecrawl-reference-audit.md`
- The implementation handoff is captured in `/Users/ac/dev/agents/firecrawl/markmaton/docs/implementation-handoff.md`
- The Go engine and Python wrapper communicate over JSON stdin/stdout
- v1 is optimized for article, docs, and news-style pages, not universal coverage
- Automated testing should favor unit tests with mocked boundaries
- Integration checks should stay manual unless a real need appears

## Phases
1. Phase 1 — Engine shell and models
2. Phase 2 — Clean + convert happy path
3. Phase 3 — Resolve + metadata + extraction helpers
4. Phase 4 — Quality heuristics and fallback
5. Phase 5 — Python wrapper and CLI
6. Phase 6 — Fixtures, golden tests, regression pass

## Tests & Verification
- Go engine logic -> unit tests
- HTML cleaning / resolving / postprocessing -> unit tests against local fixtures
- Python wrapper behavior -> unit tests with mocked subprocess boundaries
- Golden outputs -> manual review of representative fixture outputs
- Real engine smoke test -> manual and only when needed
- Architecture boundaries preserved -> manual: verify no fetch/playwright/LLM code enters the parser core

## Issue CSV
- Path: `issues/2026-04-08_11-04-40-markmaton-implementation-bootstrap.csv`
- Must share the same timestamp/slug as this plan.
- Column spec: `references/issue-csv-spec.md`

## Acceptance Checklist
- [x] Go engine can accept JSON input and produce JSON output
- [x] First-pass Markdown conversion works on representative fixtures
- [x] Metadata, links, images, and quality signals are present
- [x] Python wrapper and CLI can invoke the engine
- [x] Golden tests exist for core page types
- [x] Automated coverage is unit-test-first and does not rely on broad integration suites

## Risks / Blockers
- Main-content heuristics may over-prune real content
- Markdown conversion may look fine on toy inputs and break on real pages
- Python packaging can drift if binary discovery is unclear
- Test scope can expand too early if integration tests creep in

## References
- `/Users/ac/dev/agents/firecrawl/markmaton/docs/firecrawl-reference-audit.md`
- `/Users/ac/dev/agents/firecrawl/markmaton/docs/architecture-brief.md`
- `/Users/ac/dev/agents/firecrawl/markmaton/docs/implementation-handoff.md`

## Tools / MCP
- none

## Rollback / Recovery
- Keep implementation slices small and reviewable
- Revert per phase if parser output regresses badly on golden fixtures

## Checkpoints
- Commit after: engine shell
- Commit after: clean + convert first pass
- Commit after: Python wrapper + CLI
- Commit after: golden tests and regression pass
