#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.12"
# dependencies = [
#   "nodriver",
# ]
# ///

from __future__ import annotations

import argparse
import asyncio
import contextlib
import io
import json
import os
import sys
from typing import Any, Callable


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="browser-html-capture")
    parser.add_argument("url", help="URL to capture in a real browser")
    parser.add_argument(
        "--wait-selector",
        help="CSS selector to wait for before capture",
    )
    parser.add_argument(
        "--wait-text",
        help="Visible text to wait for before capture",
    )
    parser.add_argument(
        "--timeout",
        type=float,
        default=10.0,
        help="Maximum seconds to wait for body and optional readiness signals",
    )
    parser.add_argument(
        "--headless",
        choices=("auto", "on", "off"),
        default="auto",
        help="Browser visibility mode",
    )
    parser.add_argument(
        "--output-format",
        choices=("json", "html"),
        default="json",
        help="Emit a JSON capture envelope or raw HTML only",
    )
    return parser


def resolve_headless(mode: str) -> tuple[bool, str]:
    if mode == "on":
        return True, "cli"
    if mode == "off":
        return False, "cli"

    if sys.platform == "darwin":
        if os.environ.get("SSH_CONNECTION") or os.environ.get("SSH_TTY"):
            return True, "auto-detect (darwin ssh)"
        return False, "auto-detect (darwin)"

    if os.name == "nt":
        return False, "auto-detect (windows)"

    if os.environ.get("DISPLAY") or os.environ.get("WAYLAND_DISPLAY"):
        return False, "auto-detect (display)"
    return True, "auto-detect (headless)"


async def js(tab: Any, expr: str) -> Any:
    raw = await tab.evaluate(f"JSON.stringify({expr})")
    if raw is None:
        return None
    if isinstance(raw, str):
        try:
            return json.loads(raw)
        except json.JSONDecodeError:
            return raw
    return raw


async def wait_for_capture_ready(
    tab: Any,
    *,
    wait_selector: str | None,
    wait_text: str | None,
    timeout: float,
) -> None:
    body = await tab.select("body", timeout=timeout)
    if body is None:
        raise RuntimeError("timed out waiting for the page body")

    if wait_selector:
        target = await tab.select(wait_selector, timeout=timeout)
        if target is None:
            raise RuntimeError(
                f"timed out waiting for selector: {wait_selector}"
            )

    if wait_text:
        try:
            target = await tab.find(wait_text, timeout=timeout, best_match=True)
        except TypeError:
            target = await tab.find(wait_text, timeout=timeout)
        if target is None:
            raise RuntimeError(f"timed out waiting for text: {wait_text}")

    await asyncio.sleep(0.25)


async def capture_once(
    url: str,
    *,
    wait_selector: str | None,
    wait_text: str | None,
    timeout: float,
    headless: bool,
) -> dict[str, Any]:
    import nodriver as uc

    browser = None
    try:
        browser = await uc.start(headless=headless)
        tab = await browser.get(url)
        await wait_for_capture_ready(
            tab,
            wait_selector=wait_selector,
            wait_text=wait_text,
            timeout=timeout,
        )
        html = await tab.get_content()
        page_state = await js(
            tab,
            """(() => ({
                final_url: location.href,
                title: document.title || null,
                content_type: document.contentType || null
            }))()""",
        )

        final_url = url
        title = None
        content_type = None
        if isinstance(page_state, dict):
            final_url = page_state.get("final_url") or url
            title = page_state.get("title")
            content_type = page_state.get("content_type")

        return {
            "html": html or "",
            "url": url,
            "final_url": final_url,
            "content_type": content_type,
            "title": title,
            "rendered": True,
        }
    finally:
        if browser is not None:
            try:
                with contextlib.redirect_stdout(io.StringIO()):
                    with contextlib.redirect_stderr(io.StringIO()):
                        browser.stop()
            except Exception:
                pass


def capture_page(
    url: str,
    *,
    wait_selector: str | None,
    wait_text: str | None,
    timeout: float,
    headless_mode: str,
) -> dict[str, Any]:
    import nodriver as uc

    headless, _source = resolve_headless(headless_mode)
    return uc.loop().run_until_complete(
        capture_once(
            url,
            wait_selector=wait_selector,
            wait_text=wait_text,
            timeout=timeout,
            headless=headless,
        )
    )


def render_output(payload: dict[str, Any], output_format: str) -> str:
    if output_format == "html":
        return payload["html"]
    return json.dumps(payload, ensure_ascii=False)


def main(
    argv: list[str] | None = None,
    *,
    capture_impl: Callable[..., dict[str, Any]] | None = None,
    stdout: Any | None = None,
    stderr: Any | None = None,
) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    capture = capture_impl or capture_page
    stdout = stdout or sys.stdout
    stderr = stderr or sys.stderr

    try:
        payload = capture(
            args.url,
            wait_selector=args.wait_selector,
            wait_text=args.wait_text,
            timeout=args.timeout,
            headless_mode=args.headless,
        )
    except Exception as exc:
        stderr.write(f"capture failed for {args.url}: {exc}\n")
        return 1

    stdout.write(render_output(payload, args.output_format))
    if args.output_format == "html":
        if payload["html"] and not payload["html"].endswith("\n"):
            stdout.write("\n")
    else:
        stdout.write("\n")
    return 0


if __name__ == "__main__":
    _stdout = sys.stdout
    _stderr = sys.stderr
    sys.stdout = io.StringIO()
    sys.stderr = io.StringIO()
    raise SystemExit(main(stdout=_stdout, stderr=_stderr))
