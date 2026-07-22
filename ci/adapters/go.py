#!/usr/bin/env python3
"""Go project adapter for the framework-neutral CI impact planner."""

from __future__ import annotations

import json
import pathlib
import subprocess
from typing import NamedTuple, Sequence


class GoPackage(NamedTuple):
    import_path: str
    directory: str
    imports: list[str]


def decode_go_list(payload: str, repository: pathlib.Path) -> list[GoPackage]:
    """Decode the concatenated JSON objects emitted by `go list -json`."""
    decoder = json.JSONDecoder()
    offset = 0
    packages: list[GoPackage] = []
    while offset < len(payload):
        while offset < len(payload) and payload[offset].isspace():
            offset += 1
        if offset == len(payload):
            break
        value, offset = decoder.raw_decode(payload, offset)
        directory = pathlib.Path(value["Dir"]).resolve().relative_to(repository.resolve())
        imports = list(
            dict.fromkeys(
                value.get("Imports", [])
                + value.get("TestImports", [])
                + value.get("XTestImports", [])
            )
        )
        packages.append(
            GoPackage(value["ImportPath"], directory.as_posix(), imports)
        )
    return packages


def select_packages(
    packages: Sequence[GoPackage], changed_paths: Sequence[str]
) -> list[str]:
    package_by_directory = {
        pathlib.PurePosixPath(package.directory): package for package in packages
    }
    directly_changed: set[str] = set()

    for changed_path in changed_paths:
        if pathlib.PurePosixPath(changed_path).suffix != ".go":
            continue
        directory = pathlib.PurePosixPath(changed_path).parent
        package = package_by_directory.get(directory)
        if package is None:
            raise ValueError(f"cannot map changed Go file to a package: {changed_path}")
        directly_changed.add(package.import_path)

    affected = set(directly_changed)
    while True:
        dependents = {
            package.import_path
            for package in packages
            if any(import_path in affected for import_path in package.imports)
        }
        expanded = affected | dependents
        if expanded == affected:
            return sorted(affected)
        affected = expanded


def analyze(repository: pathlib.Path, changed_paths: Sequence[str]) -> dict[str, list[str]]:
    """Resolve changed Go packages and every in-repository reverse dependent."""
    completed = subprocess.run(
        ["go", "list", "-json", "./..."],
        cwd=repository,
        capture_output=True,
        text=True,
        check=False,
    )
    if completed.returncode != 0:
        raise RuntimeError(f"go list failed: {completed.stderr.strip()}")
    packages = decode_go_list(completed.stdout, repository)
    affected = select_packages(packages, changed_paths)
    projects = ["go"] if packages else []
    return {
        "detectedProjects": projects,
        "affectedProjects": [repository.name] if affected else [],
        "affectedModules": affected,
        "unitTestTargets": affected,
    }
