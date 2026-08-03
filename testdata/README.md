# Test data

Markmaton owns and classifies every retained test artifact. Test files are grouped by the product behavior they protect rather than by the layout of an external project.

## Layout

- `fixtures/core/` — small end-to-end product examples with exact response Markdown.
- `golden/core/` — canonical Markdown for core examples.
- `fixtures/regression/` — minimized product regressions and larger real-world page snapshots.
- `golden/regression/` — exact Markdown contracts for minimized product regressions.
- `fixtures/compatibility/` — selected HTML conversion cases covering CommonMark, GFM, and real-world structures.
- `golden/compatibility/` — conversion contracts and retained references:
  - every case has an `expected.md` containing Markmaton's exact current contract.
  - semantic and intentional-divergence cases also have a `reference.md` retained for external comparison.
  - exact cases do not duplicate `expected.md` as a second reference file.
- `fixtures/performance/` — large deterministic inputs used by benchmarks and the opt-in large-corpus test.

## Compatibility corpus provenance

The compatibility inputs and references were selected from `github.com/firecrawl/html-to-markdown` version `v0.0.0-20260312013131-1af9901a5d61` during the Native converter rewrite. They are retained as regression evidence only. Markmaton has no runtime, build, or module dependency on that project.

The original mirror layout, converter API examples, option-specific output variants, frontmatter-plugin cases, and Keep/Remove API case were intentionally not carried forward. They either tested an API outside Markmaton's product contract or duplicated behavior already covered by the selected references.

Every compatibility input has one manifest entry in `internal/convert/corpus_manifest_test.go`. Every input also has an exact Markmaton `expected.md`, so semantic comparison against a historical reference cannot hide output drift.

To intentionally refresh Markmaton conversion goldens after reviewing a renderer change:

```bash
MARKMATON_UPDATE_CONVERSION_GOLDENS=1 go test ./internal/convert -run '^TestConversionCorpus$' -count=1
```

## Adding regressions

Prefer a minimized, self-contained HTML fixture that reproduces one product failure. Add an exact golden when byte-level Markdown is the contract. Use content-presence and content-absence assertions for large real-world snapshots whose irrelevant shell markup may vary.

Large performance fixtures are run explicitly with:

```bash
MARKMATON_RUN_LARGE_CORPUS=1 go test ./internal/convert/native -run '^TestLargePerformanceCorpus$' -count=1
```
