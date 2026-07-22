#!/usr/bin/env python3
"""Framework-neutral change classification for selective CI test execution."""

from __future__ import annotations

import argparse
import fnmatch
import importlib.util
import json
import pathlib
import subprocess
import sys
import time
from typing import Any, NamedTuple, Sequence


class Change(NamedTuple):
    status: str
    path: str
    old_path: str | None = None


class AdapterSelection(NamedTuple):
    detected_projects: tuple[str, ...] = ()
    affected_projects: tuple[str, ...] = ()
    affected_modules: tuple[str, ...] = ()
    unit_test_targets: tuple[str, ...] = ()
    integration_test_targets: tuple[str, ...] = ()
    e2e_test_targets: tuple[str, ...] = ()
    check_targets: tuple[str, ...] = ()


def parse_name_status(output: str) -> list[Change]:
    """Parse git's NUL-free --name-status output, retaining rename sources."""
    changes: list[Change] = []
    for line in output.splitlines():
        if not line:
            continue
        fields = line.split("\t")
        status = fields[0]
        kind = status[0]
        if kind in {"R", "C"}:
            if len(fields) != 3:
                raise ValueError(f"invalid rename/copy record: {line!r}")
            changes.append(Change(kind, fields[2], fields[1]))
        elif kind in {"A", "M", "D", "T", "U"}:
            if len(fields) != 2:
                raise ValueError(f"invalid change record: {line!r}")
            changes.append(Change(kind, fields[1]))
        else:
            raise ValueError(f"unsupported git status: {status}")
    return changes


def paths_for_classification(changes: Sequence[Change]) -> list[str]:
    paths: list[str] = []
    for change in changes:
        paths.append(change.path)
        if change.old_path is not None:
            paths.append(change.old_path)
    return sorted(set(paths))


def existing_changed_paths(
    changes: Sequence[Change], repository: pathlib.Path
) -> list[pathlib.Path]:
    """Return only current paths; deleted and renamed-away files must not leak downstream."""
    return [
        repository / change.path
        for change in changes
        if change.status != "D" and (repository / change.path).exists()
    ]


def _matches(path: str, patterns: Sequence[str]) -> bool:
    return any(fnmatch.fnmatchcase(path, pattern) for pattern in patterns)


def _full_plan(
    reason: str,
    changes: Sequence[Change],
    base_revision: str,
    head_revision: str,
    selection: AdapterSelection,
) -> dict[str, Any]:
    return {
        "strategy": "full",
        "baseRevision": base_revision,
        "headRevision": head_revision,
        "changedFiles": paths_for_classification(changes),
        "detectedProjects": list(selection.detected_projects),
        "affectedProjects": list(selection.affected_projects),
        "affectedModules": list(selection.affected_modules),
        "unitTestTargets": [],
        "integrationTestTargets": [],
        "e2eTestTargets": [],
        "smokeTestTargets": [],
        "checkTargets": [],
        "fallback": True,
        "fallbackReason": reason,
    }


def build_plan(
    config: dict[str, Any],
    changes: Sequence[Change],
    selection: AdapterSelection,
    *,
    base_revision: str,
    head_revision: str,
) -> dict[str, Any]:
    paths = paths_for_classification(changes)
    for path in paths:
        if _matches(path, config.get("fullTestPatterns", [])):
            return _full_plan(
                f"full-test rule matched: {path}",
                changes,
                base_revision,
                head_revision,
                selection,
            )

    checks = set(selection.check_targets)
    integration_tests = set(selection.integration_test_targets)
    e2e_tests = set(selection.e2e_test_targets)
    classified = set()
    for rule in config.get("pathRules", []):
        for path in paths:
            if _matches(path, rule.get("patterns", [])):
                classified.add(path)
                checks.update(rule.get("checks", []))
                integration_tests.update(rule.get("integrationTests", []))
                e2e_tests.update(rule.get("e2eTests", []))

    adapter_classified = {
        path for path in paths if pathlib.PurePosixPath(path).suffix == ".go"
    }
    unclassified = sorted(set(paths) - classified - adapter_classified)
    if unclassified:
        return _full_plan(
            f"unclassified change: {unclassified[0]}",
            changes,
            base_revision,
            head_revision,
            selection,
        )

    has_target = any(
        (
            selection.unit_test_targets,
            integration_tests,
            e2e_tests,
            checks,
        )
    )
    if changes and not has_target:
        return _full_plan(
            "changes produced no test or validation targets",
            changes,
            base_revision,
            head_revision,
            selection,
        )

    return {
        "strategy": "selective",
        "baseRevision": base_revision,
        "headRevision": head_revision,
        "changedFiles": paths,
        "detectedProjects": list(selection.detected_projects),
        "affectedProjects": list(selection.affected_projects),
        "affectedModules": list(selection.affected_modules),
        "unitTestTargets": sorted(set(selection.unit_test_targets)),
        "integrationTestTargets": sorted(integration_tests),
        "e2eTestTargets": sorted(e2e_tests),
        "smokeTestTargets": sorted(set(config.get("smokeTests", []))),
        "checkTargets": sorted(checks),
        "fallback": False,
        "fallbackReason": None,
    }


def plan_from_git(
    config: dict[str, Any],
    repository: pathlib.Path,
    base_revision: str,
    head_revision: str,
    *,
    adapter_selector,
    runner=subprocess.run,
) -> dict[str, Any]:
    """Build a plan from a merge-base diff, failing closed on every git error."""
    merge_base = runner(
        ["git", "merge-base", base_revision, head_revision],
        cwd=repository,
        capture_output=True,
        text=True,
        check=False,
    )
    if merge_base.returncode != 0 or not merge_base.stdout.strip():
        return _full_plan(
            f"diff detection failed: {merge_base.stderr.strip() or 'merge base unavailable'}",
            [],
            base_revision,
            head_revision,
            AdapterSelection(),
        )

    resolved_base = merge_base.stdout.strip()
    diff = runner(
        [
            "git",
            "diff",
            "--name-status",
            "--find-renames",
            "--find-copies",
            resolved_base,
            head_revision,
        ],
        cwd=repository,
        capture_output=True,
        text=True,
        check=False,
    )
    if diff.returncode != 0:
        return _full_plan(
            f"diff detection failed: {diff.stderr.strip() or 'git diff failed'}",
            [],
            resolved_base,
            head_revision,
            AdapterSelection(),
        )

    changes: list[Change] = []
    try:
        changes = parse_name_status(diff.stdout)
        selection = adapter_selector(changes)
        return build_plan(
            config,
            changes,
            selection,
            base_revision=resolved_base,
            head_revision=head_revision,
        )
    except Exception as error:  # A planner error must broaden, never skip, CI coverage.
        return _full_plan(
            f"impact analysis failed: {error}",
            changes,
            resolved_base,
            head_revision,
            AdapterSelection(),
        )


def commands_for_plan(
    config: dict[str, Any], plan: dict[str, Any]
) -> list[list[str]]:
    """Translate a trusted plan to argv arrays so changed paths never reach a shell."""
    commands = config["commands"]
    if plan["strategy"] == "full":
        return [list(command) for command in commands["full"]]

    selected: list[list[str]] = []
    packages = sorted(
        set(plan.get("unitTestTargets", []))
        | set(plan.get("integrationTestTargets", []))
    )
    if packages:
        selected.append(list(commands["unitPrefix"]) + packages)
        selected.append(list(commands["racePrefix"]) + packages)

    for name in plan.get("checkTargets", []):
        selected.append(list(commands["checks"][name]))
    for name in plan.get("e2eTestTargets", []):
        selected.append(list(commands["e2e"][name]))
    for name in plan.get("smokeTestTargets", []):
        selected.append(list(commands["smoke"][name]))
    return selected


def execute_commands(
    commands: Sequence[Sequence[str]],
    repository: pathlib.Path,
    *,
    runner=subprocess.run,
    clock=time.monotonic,
) -> dict[str, Any]:
    """Execute argv-only commands and stop at the first failure."""
    started = clock()
    succeeded = 0
    failed = 0
    for command in commands:
        print(f"::group::{' '.join(command)}", flush=True)
        completed = runner(list(command), cwd=repository, check=False)
        print("::endgroup::", flush=True)
        if completed.returncode != 0:
            failed = 1
            break
        succeeded += 1
    duration = round(clock() - started, 3)
    return {
        "total": len(commands),
        "succeeded": succeeded,
        "failed": failed,
        "skipped": len(commands) - succeeded - failed,
        "durationSeconds": duration,
    }


def _load_module(path: pathlib.Path):
    spec = importlib.util.spec_from_file_location("ci_impact_adapter", path)
    if spec is None or spec.loader is None:
        raise ValueError(f"cannot load adapter: {path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def _adapter_selection(
    config: dict[str, Any], repository: pathlib.Path, changes: Sequence[Change]
) -> AdapterSelection:
    adapter_path = repository / config["adapter"]["path"]
    adapter = _load_module(adapter_path)
    result = adapter.analyze(repository, paths_for_classification(changes))
    return AdapterSelection(
        detected_projects=tuple(result["detectedProjects"]),
        affected_projects=tuple(result["affectedProjects"]),
        affected_modules=tuple(result["affectedModules"]),
        unit_test_targets=tuple(result["unitTestTargets"]),
    )


def _requested_full_plan(base_revision: str, head_revision: str) -> dict[str, Any]:
    plan = _full_plan(
        "full verification requested",
        [],
        base_revision,
        head_revision,
        AdapterSelection(),
    )
    plan["fallback"] = False
    return plan


def _log_plan(plan: dict[str, Any]) -> None:
    print(f"Test strategy: {plan['strategy']}")
    print(f"Base revision: {plan['baseRevision']}")
    print(f"Head revision: {plan['headRevision']}")
    print(f"Detected projects: {len(plan['detectedProjects'])}")
    print(f"Affected projects: {len(plan['affectedProjects'])}")
    print(f"Changed files: {len(plan['changedFiles'])}")
    print(f"Unit test targets: {len(plan['unitTestTargets'])}")
    print(f"Integration test targets: {len(plan['integrationTestTargets'])}")
    print(f"E2E test targets: {len(plan['e2eTestTargets'])}")
    print(f"Fallback: {str(plan['fallback']).lower()}")
    if plan["fallbackReason"]:
        print(f"Fallback reason: {plan['fallbackReason']}")


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--config", default="ci/config.json")
    subparsers = parser.add_subparsers(dest="command", required=True)

    plan_parser = subparsers.add_parser("plan")
    plan_parser.add_argument("--base", required=True)
    plan_parser.add_argument("--head", default="HEAD")
    plan_parser.add_argument("--output", default="ci-plan.json")
    plan_parser.add_argument("--force-full", action="store_true")

    run_parser = subparsers.add_parser("run")
    run_parser.add_argument("--plan", required=True)

    args = parser.parse_args(argv)
    repository = pathlib.Path.cwd().resolve()
    config = json.loads((repository / args.config).read_text())

    if args.command == "plan":
        if args.force_full:
            plan = _requested_full_plan(args.base, args.head)
        else:
            plan = plan_from_git(
                config,
                repository,
                args.base,
                args.head,
                adapter_selector=lambda changes: _adapter_selection(
                    config, repository, changes
                ),
            )
        (repository / args.output).write_text(json.dumps(plan, indent=2) + "\n")
        _log_plan(plan)
        return 0

    plan = json.loads((repository / args.plan).read_text())
    commands = commands_for_plan(config, plan)
    print(f"Test commands: {len(commands)}")
    result = execute_commands(commands, repository)
    print(json.dumps(result, sort_keys=True))
    return 1 if result["failed"] else 0


if __name__ == "__main__":
    sys.exit(main())
