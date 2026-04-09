#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.12"
# dependencies = [
#   "markmaton",
# ]
# ///

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

from markmaton import ConvertOptions, ConvertRequest, convert_html


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="markmaton-html-to-markdown")
    parser.add_argument("--html-file", type=Path, help="Path to an HTML file")
    parser.add_argument("--url", help="Source URL used as parsing context")
    parser.add_argument("--final-url", help="Final URL after redirects")
    parser.add_argument("--content-type", help="Optional content type hint")
    parser.add_argument(
        "--output-format",
        choices=("json", "markdown"),
        default="json",
        help="Choose between full JSON output or markdown only",
    )
    parser.add_argument(
        "--full-content",
        action="store_true",
        help="Disable main-content-only cleaning",
    )
    parser.add_argument(
        "--include-selector",
        action="append",
        default=[],
        help="CSS selector to force-include before conversion",
    )
    parser.add_argument(
        "--exclude-selector",
        action="append",
        default=[],
        help="CSS selector to remove before conversion",
    )
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)

    html = _read_html(args.html_file)
    response = convert_html(
        ConvertRequest(
            html=html,
            url=args.url,
            final_url=args.final_url,
            content_type=args.content_type,
            options=ConvertOptions(
                only_main_content=not args.full_content,
                include_selectors=list(args.include_selector),
                exclude_selectors=list(args.exclude_selector),
            ),
        )
    )

    if args.output_format == "markdown":
        sys.stdout.write(response.markdown)
        if response.markdown and not response.markdown.endswith("\n"):
            sys.stdout.write("\n")
        return 0

    sys.stdout.write(
        json.dumps(
            {
                "markdown": response.markdown,
                "html_clean": response.html_clean,
                "metadata": response.metadata.__dict__,
                "links": response.links,
                "images": response.images,
                "quality": response.quality.__dict__,
            },
            ensure_ascii=False,
        )
    )
    sys.stdout.write("\n")
    return 0


def _read_html(path: Path | None) -> str:
    if path is None:
        return sys.stdin.read()
    return path.read_text(encoding="utf-8")


if __name__ == "__main__":
    raise SystemExit(main())
