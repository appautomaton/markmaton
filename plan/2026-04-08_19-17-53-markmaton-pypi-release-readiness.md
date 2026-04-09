---
mode: plan
task: markmaton PyPI release readiness
created_at: "2026-04-08T19:17:53-04:00"
complexity: complex
status: completed
---

# Plan: markmaton PyPI release readiness

## Goal
- Make `markmaton` ready for real PyPI/TestPyPI publication by shipping platform-specific wheels that bundle the Go engine, using GitHub Actions and PyPI Trusted Publishing without changing the public Python/CLI contract.

## Scope
- In: wheel/sdist packaging strategy, binary bundling layout, local packaging smoke, GitHub Actions CI/release workflows, Trusted Publishing setup guidance, install-state verification, and release docs.
- Out: parser logic changes, crawler/page-visiting work, LLM features, PyPI publication itself, Homebrew distribution, and automatic release notes generation beyond basic workflow metadata.

## Assumptions / Dependencies
- `markmaton` remains a Python package with a bundled Go binary, not a Python extension module.
- Binary discovery contract in `markmaton/engine.py` remains the source of truth.
- Future installers should get `markmaton/bin/markmaton-engine` inside the wheel.
- Wheels should be platform-specific and not claim `py3-none-any`.
- GitHub Actions release publishing should use PyPI Trusted Publishing (`id-token: write`) rather than long-lived API tokens.
- The eventual PyPI/TestPyPI trusted publisher config must bind to the exact release workflow filename.
- Existing parser/test baselines must remain green during packaging work.

## Phases
1. Phase 1 — Lock packaging architecture and local build contract
2. Phase 2 — Implement wheel/sdist packaging for bundled Go binaries
3. Phase 3 — Replace placeholder GitHub workflow with CI and release workflows
4. Phase 4 — Add install-state smoke checks and release docs
5. Phase 5 — Validate the full release-readiness path without publishing

## Tests & Verification
- Python metadata and packaging config are valid -> `python3 -m build --sdist --wheel`
- Bundled binary is present in built wheel -> manual: inspect wheel contents and verify `markmaton/bin/markmaton-engine*`
- Local install-state wrapper contract works -> manual: install built wheel in a clean venv, run `markmaton convert ...`
- Go and Python project baselines remain green -> `env GOCACHE=/tmp/markmaton-go-build GOMODCACHE=/tmp/markmaton-go-mod go test ./...` and `PYTHONPYCACHEPREFIX=/tmp/markmaton-pyc python3 -m unittest discover -s tests -p 'test_*.py'`
- GitHub Actions config is coherent -> manual: review workflow triggers, artifact flow, matrix, and Trusted Publishing job permissions
- Release path is dry-run ready -> manual: confirm TestPyPI/PyPI environment guidance matches workflow filenames and job names

## Issue CSV
- Path: `issues/2026-04-08_19-17-53-markmaton-pypi-release-readiness.csv`
- Must share the same timestamp/slug as this plan.
- Column spec: `references/issue-csv-spec.md`

## Acceptance Checklist
- [x] Platform-specific wheel strategy is explicit and documented
- [x] Packaging config can build sdist and wheel with bundled binary layout
- [x] GitHub Actions CI workflow replaces the placeholder workflow
- [x] GitHub Actions release workflow is compatible with PyPI Trusted Publishing
- [x] Local install-state smoke path is documented and testable
- [x] Go and Python automated tests still pass after packaging changes

## Risks / Blockers
- Hatch/wheel configuration can accidentally emit `py3-none-any` wheels if binary tagging is not handled correctly.
- Bundling a Go binary into wheels may require build-hook or pre-build staging details that differ across platforms.
- Linux wheel compatibility can be undermined if the Go build/release path ignores platform tags or libc expectations.
- Trusted Publishing can fail at publish time if the workflow filename, environment name, or repository binding drifts from the PyPI publisher configuration.

## References
- `/Users/ac/dev/agents/firecrawl/markmaton/pyproject.toml`
- `/Users/ac/dev/agents/firecrawl/markmaton/markmaton/engine.py`
- `/Users/ac/dev/agents/firecrawl/markmaton/docs/packaging-layout.md`
- `/Users/ac/dev/agents/firecrawl/markmaton/docs/local-smoke.md`
- `/Users/ac/dev/agents/firecrawl/markmaton/.github/workflows/workflow.yml`
- https://docs.pypi.org/trusted-publishers/using-a-publisher/
- https://docs.pypi.org/trusted-publishers/adding-a-publisher/
- https://github.com/pypa/gh-action-pypi-publish
- https://github.com/pypa/cibuildwheel
- https://hatch.pypa.io/
- https://packaging.python.org/en/latest/specifications/binary-distribution-format/

## Tools / MCP
- none

## Rollback / Recovery
- Restore the placeholder workflow if CI/release workflows break basic repository checks.
- Revert packaging-target changes in `pyproject.toml` and any build hook files if wheels become invalid or stop bundling the binary.
- Keep release and CI workflows separate so packaging failures do not block ordinary unit-test feedback.

## Checkpoints
- Commit after: packaging config and local wheel strategy are working
- Commit after: CI and release workflows are in place
- Commit after: install-state smoke and release docs are complete
