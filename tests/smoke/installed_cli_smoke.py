from __future__ import annotations

import json
import subprocess
import sys
import tempfile
from pathlib import Path


def main() -> int:
    with tempfile.TemporaryDirectory() as tmpdir:
        html_path = Path(tmpdir) / "sample.html"
        html_path.write_text("<html><body><main><h1>Hello</h1><p>World</p></main></body></html>", encoding="utf-8")
        scripts_dir = Path(sys.prefix) / ("Scripts" if sys.platform.startswith("win") else "bin")
        cli_path = scripts_dir / ("markmaton.exe" if sys.platform.startswith("win") else "markmaton")

        completed = subprocess.run(
            [
                str(cli_path),
                "convert",
                "--html-file",
                str(html_path),
                "--url",
                "https://example.com/sample",
            ],
            capture_output=True,
            text=True,
            check=False,
        )
        if completed.returncode != 0:
            raise SystemExit(completed.stderr or completed.stdout or "markmaton CLI failed")

        payload = json.loads(completed.stdout)
        if "markdown" not in payload or "Hello" not in payload["markdown"]:
            raise SystemExit("markmaton CLI smoke did not produce the expected markdown payload")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
