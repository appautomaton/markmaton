---
mode: plan
task: browser-html-capture skill for markmaton workflows
created_at: "2026-04-09T14:01:19-04:00"
complexity: medium
status: completed
---

# Plan: browser-html-capture skill for markmaton workflows

## Goal
- Add a new sibling skill that captures rendered HTML from a real browser and hands it off cleanly to `html-to-markdown` through a stable, minimal contract.

## Scope
- In:
  - a new skill next to `html-to-markdown`
  - a self-contained PEP 723 script for real-browser HTML capture
  - a stable handoff contract to `html-to-markdown`
  - progressive-disclosure docs and tests
- Out:
  - no changes to `markmaton` parser core
  - no merger of browser logic into `html-to-markdown`
  - no site-specific rules
  - no full persistent browser-daemon subsystem in v1 unless planning proves it is strictly necessary

## Assumptions / Dependencies
- `html-to-markdown` remains the conversion skill and stays HTML-in only.
- The new skill should be general in purpose, not implementation-branded; the backend may use `nodriver`, but the skill should not be named after the library.
- The skill should be self-contained and agent-friendly, using a single PEP 723 script as the primary entrypoint.
- Robustness matters more than feature breadth; the first version should prefer a narrow, deterministic one-shot capture flow over a highly stateful browser automation surface.
- The handoff contract should align with `ConvertRequest` fields:
  - `html`
  - `url`
  - `final_url`
  - `content_type`

## Phases
1. Phase 1 — Define the cross-skill contract
   - Decide the exact capture output schema.
   - Decide which fields are required vs optional.
   - Decide the default output mode:
     - `json` envelope for chaining
     - optional raw `html` mode for direct piping
   - Decide how `html-to-markdown` should refer to the companion skill without coupling to local paths or repo runtime.

2. Phase 2 — Define the new skill boundary and artifact layout
   - Choose a general name such as `browser-html-capture`.
   - Create the skill structure:
     - `SKILL.md`
     - `references/usage.md`
     - `references/integration-patterns.md`
     - `scripts/<capture_script>.py`
   - Keep `SKILL.md` short and decision-oriented.
   - Put implementation details and usage patterns into `references/`.

3. Phase 3 — Design the capture script interface
   - Primary entrypoint: one PEP 723 script using `uv run --script`.
   - Proposed core arguments:
     - positional `url`
     - `--wait-selector`
     - `--wait-text`
     - `--timeout`
     - `--headless auto|on|off`
     - `--output-format json|html`
   - Optional debug-only arguments if justified:
     - `--screenshot-path`
     - `--full-page`
   - Keep the v1 interface narrow; do not expose a huge browser API.

4. Phase 4 — Define seamless composition with `html-to-markdown`
   - Make the capture skill produce a stable JSON envelope with fields that map directly to `markmaton`.
   - Document the two preferred compositions:
     - capture -> save envelope -> feed `html` and context into `html-to-markdown`
     - capture -> emit raw HTML -> pipe into `html-to-markdown`
   - Add only a light cross-reference in both skills; do not make one skill hard-call the other.

5. Phase 5 — Verification and robustness hardening
   - Add unit tests for:
     - skill artifact presence
     - PEP 723 metadata
     - argument parsing
     - output contract shape
   - Add at least one manual smoke path for:
     - a JS-heavy page
     - a static page
   - Confirm the new skill can be used without repo-local `.venv` knowledge.

## Tests & Verification
- Skill artifact structure is complete -> `python -m unittest` skill artifact tests
- PEP 723 script metadata is valid -> unit test
- Output contract includes `html/url/final_url/content_type` -> unit test
- `html` mode can be piped into `html-to-markdown` -> manual smoke
- `json` mode can be consumed by downstream tooling -> manual smoke
- `html-to-markdown` remains unchanged in behavior -> existing skill artifact tests remain green

## Issue CSV
- Path: `issues/2026-04-09_14-01-19-browser-html-capture-skill.csv`
- Must share the same timestamp/slug as this plan.
- Column spec: `references/issue-csv-spec.md`

## Acceptance Checklist
- [x] A new sibling browser-capture skill exists with a narrow, self-contained scope
- [x] The new skill uses progressive disclosure correctly
- [x] The capture script is PEP 723-based and does not assume a local repo runtime
- [x] The output contract composes cleanly with `html-to-markdown`
- [x] `html-to-markdown` stays self-contained and does not absorb browser responsibilities
- [x] Tests and manual smoke checks verify the composition path

## Risks / Blockers
- `nodriver` can introduce browser-state complexity if we overexpose its capabilities in v1.
- If the capture skill tries to be both a browser automation toolkit and an HTML handoff tool, it will become noisy and fragile.
- Anti-bot-heavy sites may require interaction patterns that exceed a narrow one-shot capture interface; those should be deferred rather than bloating v1.

## References
- Existing conversion skill:
  - [SKILL.md](/Users/ac/dev/agents/firecrawl/markmaton/skills/html-to-markdown/SKILL.md)
  - [usage.md](/Users/ac/dev/agents/firecrawl/markmaton/skills/html-to-markdown/references/usage.md)
  - [integration-patterns.md](/Users/ac/dev/agents/firecrawl/markmaton/skills/html-to-markdown/references/integration-patterns.md)
- Existing nodriver reference skill:
  - [/Users/ac/dev/agents/skills/search/webmaton/skills/nodriver/SKILL.md](/Users/ac/dev/agents/skills/search/webmaton/skills/nodriver/SKILL.md)
- PEP 723:
  - https://peps.python.org/pep-0723/
- uv scripts guidance:
  - https://docs.astral.sh/uv/guides/scripts/

## Checkpoints
- Commit after: contract and skill skeleton are settled
- Commit after: PEP 723 capture script and tests are green
- Commit after: cross-skill composition docs are complete
