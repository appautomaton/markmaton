# Changelog

All notable changes to `markmaton` will be documented in this file.

## [0.1.8] - 2026-08-03

### Added

- an independent Native DOM-to-IR-to-Markdown converter with deterministic rendering, cancellation, typed limits, immutable document transforms, and malformed-input safety
- project-owned CommonMark, GFM, real-world, product-regression, and large performance corpora with exact Markmaton Goldens and retained semantic references
- safe URL policy enforcement for links, images, media sources, `srcset`, and document base URLs
- readable iframe, video, and audio link preservation without network-dependent plugins

### Changed

- Native is now the only production conversion path; conversion cleanup policies are applied directly around the Native converter
- regression fixtures, Goldens, performance inputs, and provenance are organized under Markmaton's `testdata/fixtures` and `testdata/golden` taxonomy
- CI now runs the large performance corpus, `go vet`, a CGO-free production build, and the local CLI smoke path
- benchmark guidance now uses cached HTML and project-owned regressions instead of requiring a local reference service

### Fixed

- block structure inside custom elements and links, definition-list separation, reversed and explicitly numbered lists, quoted fenced code, and nested syntax-highlighter markup
- unsafe `javascript:`, `data:`, and `vbscript:` destinations leaking into Markdown, cleaned HTML, links, or images
- meaningful opening links being mistaken for repository chrome and removed by Markdown cleanup
- audio and video `<source>` elements being classified as images

### Removed

- the Legacy converter, Firecrawl runtime dependency, Builder/Hook/Plugin compatibility layer, and obsolete migration-only planning records

## [0.1.7] - 2026-04-13

### Changed

- renamed the capture-script test file and class to match the consolidated `html-to-markdown` skill naming
- the Python API usage example now shows the optional `final_url` and `content_type` fields so the Python and CLI surfaces line up

## [0.1.6] - 2026-04-11

### Added

- end-to-end CLI integration coverage that exercises content-mode behavior against the current Go engine build
- a unified `html-to-markdown` skill flow that can capture browser HTML and convert it in one pipeline

### Changed

- the capture workflow is now folded into `skills/html-to-markdown` instead of living as a separate `browser-html-capture` skill
- skill and package docs now describe the tighter parser contract and the unified HTML capture path more clearly

### Fixed

- explicit `--full-content` / `only_main_content=False` requests now survive the Python-to-Go boundary instead of being silently reset
- docs now distinguish explicit full-content mode from automatic fallback behavior

## [0.1.5] - 2026-04-09

### Added

- `uv`-managed local development with a committed `uv.lock`
- Python version pinning via `.python-version`
- documentation index and usage guide for installed and local workflows
- GitHub release-note defaults in `.github/release.yml`

### Changed

- local development is now `uv`-first and pinned to Python `3.12`
- CI now uses `uv sync --group dev --locked` for Python-side validation
- README and docs now present install, usage, smoke, benchmark, and release paths more clearly
- workflow actions were updated to current major versions to reduce old runtime warnings

## [0.1.4] - 2026-04-09

### Added

- First published PyPI release with bundled Go engine wheels for:
  - Linux x86_64
  - macOS x86_64
  - macOS arm64
  - Windows x86_64
- Python CLI and Python API for running the parser from installed packages
- Benchmark-driven parser hardening and converter customization groundwork
- Packaging and release documentation

### Changed

- Release workflow now enforces tag/version alignment before publishing
- Linux wheel publishing now uses a PyPI-accepted `manylinux2014_x86_64` tag
- README and docs now reflect install, usage, benchmark, and release paths

### Fixed

- PyPI publish flow now succeeds through GitHub Trusted Publishing
- Packaged install smoke path now verifies the bundled engine lookup contract
