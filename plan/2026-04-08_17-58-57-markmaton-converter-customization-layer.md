---
mode: plan
task: markmaton converter customization layer
created_at: "2026-04-08T17:58:57-04:00"
complexity: complex
status: completed
---

# Plan: markmaton converter customization layer

## Goal
- Build a `markmaton`-owned converter customization layer on top of the existing Go core so structural Markdown output improves for general page patterns such as threads, timelines, repo/app-shell pages, dense card grids, and comparison/product layouts.

## Scope
- In: converter builder surface, hooks/plugins/rule-extension points, a small promoted regression set for structure-heavy pages, and first general converter-level improvements.
- Out: LLM features, page visiting/crawling, Python packaging redesign, FFI changes, fallback-policy redesign, and site-specific parser rules.

## Assumptions / Dependencies
- `markmaton` remains a general HTML-to-Markdown parser.
- The existing Go core from `github.com/firecrawl/html-to-markdown` remains the base.
- Public request/response shapes in the Go engine and Python wrapper stay unchanged in this epic.
- Current benchmark findings in `docs/benchmark-matrix.md` are the input for prioritization.
- Existing strong classes such as careers landing pages, job detail/application detail pages, and clean docs pages must remain protected.

## Phases
1. Phase 1 — Create a converter control surface in `internal/convert`
2. Phase 2 — Promote a small set of structure-heavy regression fixtures from the benchmark cache
3. Phase 3 — Implement the first general converter-level customizations for thread/timeline/repo/app-shell structure
4. Phase 4 — Harden with converter-level tests and engine regression coverage

## Tests & Verification
- Converter builder/hook/plugin surface behaves deterministically -> `env GOCACHE=/tmp/markmaton-go-build GOMODCACHE=/tmp/markmaton-go-mod go test ./internal/convert ./internal/engine`
- Existing parser baselines are preserved -> `env GOCACHE=/tmp/markmaton-go-build GOMODCACHE=/tmp/markmaton-go-mod go test ./...`
- Python wrapper contract remains unchanged -> `PYTHONPYCACHEPREFIX=/tmp/markmaton-pyc python3 -m unittest discover -s tests -p 'test_*.py'`
- Promoted structure-heavy fixtures improve readability without site-specific hacks -> manual: rerun selected benchmark rows and compare opening quality, block grouping, and structural readability

## Issue CSV
- Path: issues/2026-04-08_17-58-57-markmaton-converter-customization-layer.csv
- Must share the same timestamp/slug as this plan.
- Column spec: `references/issue-csv-spec.md`

## Acceptance Checklist
- [x] `internal/convert` exposes a clear builder/hook/plugin surface
- [x] At least one discussion/timeline-style regression fixture is promoted from benchmark cache
- [x] At least one app-shell/repo-style regression fixture is promoted from benchmark cache
- [x] First converter-level improvements land without site-name rules
- [x] Existing strong page classes do not regress
- [x] Go and Python automated tests pass

## Risks / Blockers
- Converter-level changes can accidentally regress already-good page classes if structure rules are too broad.
- Some difficult pages may still need better HTML acquisition before converter work can help.
- Real-page fixtures can become too site-shaped if promotion rules are not strict.

## References
- docs/benchmark-matrix.md — benchmark findings and layer attribution
- docs/benchmark-workflow.md — benchmark workflow and promotion rules
- internal/convert/convert.go — current thin converter entry point
- internal/engine/process_test.go — current engine-level regression coverage
- tmp/benchmarks/markmaton-benchmark-driven-refinement/ — local benchmark cache for promoted fixture selection

## Tools / MCP
- none

## Rollback / Recovery
- Revert converter-surface changes in `internal/convert` and rerun the full Go/Python test suite.
- Drop newly promoted regression fixtures if they prove too site-specific and replace them with a better benchmark-backed sample.
- Keep the benchmark cache as the source of truth so fixture promotion can be re-done safely.

## Checkpoints
- Commit after: converter builder surface is in place
- Commit after: regression fixture promotion is complete
- Commit after: first converter-level customizations pass regression
