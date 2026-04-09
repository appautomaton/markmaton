from __future__ import annotations

import os
import platform
import shutil
import stat
import subprocess
from pathlib import Path
from typing import Optional

from packaging.tags import sys_tags

try:
    from hatchling.builders.hooks.plugin.interface import BuildHookInterface
except ImportError:  # pragma: no cover - allows unit tests without hatchling installed
    class BuildHookInterface:  # type: ignore[override]
        pass


PROJECT_ROOT = Path(__file__).resolve().parent
LOCAL_BIN_DIR = PROJECT_ROOT / "bin"
STAGING_BIN_DIR = PROJECT_ROOT / ".hatch-build" / "bin"


def binary_filename(system: Optional[str] = None) -> str:
    name = (system or platform.system()).lower()
    return "markmaton-engine.exe" if name.startswith("win") else "markmaton-engine"


def packaged_binary_path(root: Path = PROJECT_ROOT, system: Optional[str] = None) -> Path:
    return root / "markmaton" / "bin" / binary_filename(system)


def platform_wheel_tag() -> str:
    for tag in sys_tags():
        if tag.platform != "any":
            return f"py3-none-{tag.platform}"
    raise RuntimeError("Could not determine a platform-specific wheel tag for this build.")


def resolve_engine_source(root: Path = PROJECT_ROOT, explicit_path: Optional[str] = None) -> Optional[Path]:
    candidates = []
    if explicit_path:
        candidates.append(Path(explicit_path))
    if env_path := os.environ.get("MARKMATON_ENGINE"):
        candidates.append(Path(env_path))
    candidates.append(root / "bin" / binary_filename())

    for candidate in candidates:
        if candidate.is_file():
            return candidate.resolve()

    return None


def ensure_engine_binary(root: Path = PROJECT_ROOT, explicit_path: Optional[str] = None) -> Path:
    staged = STAGING_BIN_DIR / binary_filename()
    staged.parent.mkdir(parents=True, exist_ok=True)

    source = resolve_engine_source(root=root, explicit_path=explicit_path)
    if source is not None:
        shutil.copy2(source, staged)
        _make_executable(staged)
        return staged

    if shutil.which("go") is None:
        raise RuntimeError(
            "Could not find a Go toolchain or a prebuilt markmaton engine. "
            "Set MARKMATON_ENGINE or build ./bin/markmaton-engine first."
        )

    env = os.environ.copy()
    env.setdefault("CGO_ENABLED", "0")
    subprocess.run(
        ["go", "build", "-o", str(staged), "./cmd/markmaton-engine"],
        cwd=root,
        check=True,
        env=env,
    )
    _make_executable(staged)
    return staged


def _make_executable(path: Path) -> None:
    if path.suffix.lower() == ".exe":
        return
    current_mode = path.stat().st_mode
    path.chmod(current_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)


class CustomBuildHook(BuildHookInterface):
    PLUGIN_NAME = "custom"

    def initialize(self, version, build_data) -> None:  # pragma: no cover - exercised via package builds
        if getattr(self, "target_name", None) != "wheel":
            return

        binary = ensure_engine_binary()
        build_data["pure_python"] = False
        build_data["tag"] = platform_wheel_tag()
        force_include = build_data.setdefault("force_include", {})
        force_include[str(binary)] = f"markmaton/bin/{binary.name}"
