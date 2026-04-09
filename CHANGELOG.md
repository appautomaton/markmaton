# Changelog

All notable changes to `markmaton` will be documented in this file.

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
