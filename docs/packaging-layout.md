# Packaging Layout

## Purpose

This document fixes one thing early:

- how the Python package finds the Go engine in local development
- how that same contract should map to future PyPI / `uv tool` installs

It does **not** implement release automation.

## Binary discovery contract

`markmaton` uses a small Python wrapper around a standalone Go binary.

The wrapper looks for the engine in this order:

1. explicit path passed by the caller
2. `MARKMATON_ENGINE`
3. packaged binary inside `markmaton/bin/markmaton-engine`
4. local development binary in `./bin/markmaton-engine`
5. `markmaton-engine` on `PATH`

That contract is implemented in:

- `/Users/ac/dev/agents/firecrawl/markmaton/markmaton/engine.py`

## Local development layout

For local work, the expected shape is:

```text
markmaton/
  bin/
    markmaton-engine
  markmaton/
    engine.py
```

The Python CLI does not build the Go binary for you.
It assumes the binary already exists at one of the known locations.

## Future packaged layout

For packaged installs, the intended layout is:

```text
site-packages/
  markmaton/
    bin/
      markmaton-engine
```

That means the Python wheel should eventually carry:

- the Python package
- the platform-specific Go binary

The wheel strategy is explicitly:

- platform-specific wheels, not `py3-none-any`
- a bundled engine at `markmaton/bin/markmaton-engine*`
- the same Python CLI contract across local dev and packaged installs

The initial target matrix for release-readiness is:

- Linux x86_64
- macOS x86_64
- macOS arm64
- Windows x86_64

This is why packaging must be treated as a separate concern from parser logic.

## Why this is explicit now

We do not want a hidden rule like:

- “the wrapper magically knows where the binary is”

That always becomes brittle later.

Instead, the rule is fixed now:

- local dev path is explicit
- packaged path is explicit
- environment override is explicit

## What this implies for release work later

When we get to release work, we will need:

- platform-specific wheels
- a build matrix for supported targets
- a wheel layout that places the binary in `markmaton/bin`
- a check that `uv tool` installs expose the Python CLI while the wrapper can still locate the bundled binary
- a stable release workflow filename for PyPI Trusted Publishing
- GitHub environments for `testpypi` and `pipy`

That is release-track work, not bootstrap parser work.
