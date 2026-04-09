# Changelog

All notable changes to `markmaton` will be documented in this file.

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
