from __future__ import annotations

import json
import os
import platform
import shutil
import subprocess
from pathlib import Path
from typing import Optional

from .models import ConvertRequest, ConvertResponse


class EngineNotFoundError(RuntimeError):
    pass


def convert_html(request: ConvertRequest, binary_path: Optional[str] = None) -> ConvertResponse:
    engine = discover_engine(binary_path)
    completed = subprocess.run(
        [str(engine)],
        input=json.dumps(request.to_payload()),
        capture_output=True,
        text=True,
        check=False,
    )
    if completed.returncode != 0:
        message = completed.stderr.strip() or completed.stdout.strip() or "markmaton engine failed"
        raise RuntimeError(message)

    payload = json.loads(completed.stdout)
    return ConvertResponse.from_dict(payload)


def discover_engine(explicit_path: Optional[str] = None) -> Path:
    candidates = []
    if explicit_path:
        candidates.append(Path(explicit_path))

    if env_path := os.environ.get("MARKMATON_ENGINE"):
        candidates.append(Path(env_path))

    package_bin = Path(__file__).resolve().parent / "bin" / _binary_name()
    candidates.append(package_bin)

    repo_bin = Path(__file__).resolve().parent.parent / "bin" / _binary_name()
    candidates.append(repo_bin)

    if which := shutil.which(_binary_name()):
        candidates.append(Path(which))

    for candidate in candidates:
        if candidate.is_file():
            return candidate

    raise EngineNotFoundError(
        "Could not find markmaton-engine. Set MARKMATON_ENGINE or place the binary in markmaton/bin or ./bin."
    )


def _binary_name() -> str:
    return "markmaton-engine.exe" if platform.system().lower().startswith("win") else "markmaton-engine"
