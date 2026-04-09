# PyPI Release Path

## Purpose

This document fixes the release contract for `markmaton`.

It answers:

- how artifacts are built
- which GitHub workflows matter
- how TestPyPI and PyPI publishing are triggered
- what GitHub/PyPI Trusted Publishing must be configured to match

## Workflow split

`markmaton` uses two workflows:

- `.github/workflows/ci.yml`
  - ordinary push / pull request validation
  - Go tests
  - Python unit tests
  - local-dev binary smoke

- `.github/workflows/workflow.yml`
  - builds platform wheels and sdist
  - uploads artifacts
  - publishes only through explicit release paths

Do not rename `workflow.yml` casually.

PyPI Trusted Publishing binds to the exact workflow file path.

## Publishing paths

### TestPyPI

Use:

- `workflow_dispatch`
- input: `publish_target=testpypi`

This path builds artifacts first, then publishes to TestPyPI through Trusted Publishing.

### PyPI

Use:

- a git tag that matches `v*`

or:

- `workflow_dispatch`
- input: `publish_target=pipy`

This path builds artifacts first, then publishes to PyPI through Trusted Publishing.

## GitHub environments

Create these GitHub environments:

- `testpypi`
- `pipy`

These should match the environment names used in `workflow.yml`.

## Trusted Publisher setup

For both TestPyPI and PyPI, configure a Trusted Publisher that matches:

- owner: `appautomaton`
- repository: `markmaton`
- workflow filename: `workflow.yml`

If the workflow filename or environment names change, update the publisher configuration too.

## Artifact strategy

Release artifacts are:

- platform-specific wheels containing `markmaton/bin/markmaton-engine*`
- one sdist

The wheel contract is:

- not `py3-none-any`
- platform-specific
- same Python CLI entrypoint on every platform

The current release workflow builds wheels natively on each GitHub runner.

It does **not** use `cibuildwheel` in the first release track.

That is intentional:

- `markmaton` bundles an external Go binary
- the first release path favors explicit runner-native control over a more abstract wheel orchestration layer
- this keeps the workflow easier to debug while the packaging contract settles

## Local pre-publish checks

Before a real publish, do all of these locally:

1. `python3 -m build --sdist --wheel`
2. Inspect wheel contents and confirm `markmaton/bin/markmaton-engine*`
3. Install the wheel in a clean venv
4. Run `markmaton convert ...` and confirm the CLI finds the bundled binary
5. Run Go and Python test suites

## Current support matrix

Initial release-readiness target:

- Linux x86_64
- macOS x86_64
- macOS arm64
- Windows x86_64

More targets can be added later, but the first release track should stay narrow and reliable.
