import importlib.util
import json
import pathlib
import subprocess
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[2]


def load_module(name: str, path: pathlib.Path):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


impact = load_module("impact", ROOT / "ci" / "impact.py")


class PlanTests(unittest.TestCase):
    def setUp(self):
        self.config = {
            "fullTestPatterns": ["go.mod", ".github/workflows/**"],
            "pathRules": [
                {"patterns": ["docs/**", "README.md"], "checks": ["docs"]},
            ],
            "smokeTests": ["health"],
        }

    def test_dangerous_shared_change_forces_full_suite(self):
        plan = impact.build_plan(
            self.config,
            [impact.Change("M", "go.mod")],
            impact.AdapterSelection(),
            base_revision="base",
            head_revision="head",
        )

        self.assertEqual("full", plan["strategy"])
        self.assertTrue(plan["fallback"])
        self.assertIn("go.mod", plan["fallbackReason"])

    def test_unclassified_change_forces_full_suite(self):
        plan = impact.build_plan(
            self.config,
            [impact.Change("A", "unknown.asset")],
            impact.AdapterSelection(),
            base_revision="base",
            head_revision="head",
        )

        self.assertEqual("full", plan["strategy"])
        self.assertIn("unclassified", plan["fallbackReason"])

    def test_docs_change_selects_docs_check_and_always_adds_smoke(self):
        plan = impact.build_plan(
            self.config,
            [impact.Change("M", "docs/site/index.md")],
            impact.AdapterSelection(),
            base_revision="base",
            head_revision="head",
        )

        self.assertEqual("selective", plan["strategy"])
        self.assertEqual(["docs"], plan["checkTargets"])
        self.assertEqual(["health"], plan["smokeTestTargets"])

    def test_path_rule_adds_related_integration_and_e2e_targets(self):
        config = {
            "pathRules": [
                {
                    "patterns": ["adapters/**"],
                    "integrationTests": ["./internal/server"],
                    "e2eTests": ["multi-adapter"],
                }
            ],
            "smokeTests": ["health"],
        }
        selection = impact.AdapterSelection(
            detected_projects=("go",),
            affected_projects=("mockport",),
            affected_modules=("example/adapters/stripe",),
            unit_test_targets=("example/adapters/stripe",),
        )

        plan = impact.build_plan(
            config,
            [impact.Change("M", "adapters/stripe/adapter.go")],
            selection,
            base_revision="base",
            head_revision="head",
        )

        self.assertEqual(["./internal/server"], plan["integrationTestTargets"])
        self.assertEqual(["multi-adapter"], plan["e2eTestTargets"])

    def test_deleted_file_is_reported_but_not_passed_as_existing_path(self):
        changes = impact.parse_name_status("D\tinternal/old/removed.go\n")

        self.assertEqual([impact.Change("D", "internal/old/removed.go")], changes)
        self.assertEqual([], impact.existing_changed_paths(changes, ROOT))

    def test_rename_tracks_destination_and_source_for_safety(self):
        changes = impact.parse_name_status(
            "R100\tinternal/old/name.go\tinternal/new/name.go\n"
        )

        self.assertEqual("internal/new/name.go", changes[0].path)
        self.assertEqual("internal/old/name.go", changes[0].old_path)
        self.assertEqual(
            ["internal/new/name.go", "internal/old/name.go"],
            impact.paths_for_classification(changes),
        )

    def test_diff_failure_falls_back_instead_of_skipping_tests(self):
        def failing_runner(*args, **kwargs):
            return subprocess.CompletedProcess(args[0], 128, "", "missing base")

        plan = impact.plan_from_git(
            self.config,
            ROOT,
            "missing-base",
            "HEAD",
            adapter_selector=lambda changes: impact.AdapterSelection(),
            runner=failing_runner,
        )

        self.assertEqual("full", plan["strategy"])
        self.assertIn("diff detection failed", plan["fallbackReason"])

    def test_adapter_failure_preserves_changed_files_in_fallback_plan(self):
        responses = iter(
            [
                subprocess.CompletedProcess([], 0, "base-sha\n", ""),
                subprocess.CompletedProcess([], 0, "M\tinternal/new/value.go\n", ""),
            ]
        )

        def runner(*args, **kwargs):
            return next(responses)

        def failing_adapter(changes):
            raise RuntimeError("dependency graph unavailable")

        plan = impact.plan_from_git(
            self.config,
            ROOT,
            "base",
            "head",
            adapter_selector=failing_adapter,
            runner=runner,
        )

        self.assertEqual("full", plan["strategy"])
        self.assertEqual(["internal/new/value.go"], plan["changedFiles"])
        self.assertIn("dependency graph unavailable", plan["fallbackReason"])

    def test_full_strategy_uses_only_declared_full_commands(self):
        config = {
            "commands": {
                "full": [["go", "test", "./..."]],
                "smoke": {"health": ["go", "test", "./internal/server"]},
            }
        }
        plan = {"strategy": "full"}

        commands = impact.commands_for_plan(config, plan)

        self.assertEqual([["go", "test", "./..."]], commands)

    def test_selective_strategy_expands_packages_without_shell(self):
        config = {
            "commands": {
                "unitPrefix": ["go", "test"],
                "racePrefix": ["go", "test", "-race"],
                "checks": {"docs": ["bash", "scripts/check-doc-links.sh"]},
                "smoke": {
                    "health": [
                        "go",
                        "test",
                        "./internal/server",
                        "-run",
                        "TestHealthReturnsOK",
                    ]
                },
            }
        }
        plan = {
            "strategy": "selective",
            "unitTestTargets": ["./internal/config"],
            "integrationTestTargets": [],
            "e2eTestTargets": [],
            "checkTargets": ["docs"],
            "smokeTestTargets": ["health"],
        }

        commands = impact.commands_for_plan(config, plan)

        self.assertEqual(
            [
                ["go", "test", "./internal/config"],
                ["go", "test", "-race", "./internal/config"],
                ["bash", "scripts/check-doc-links.sh"],
                [
                    "go",
                    "test",
                    "./internal/server",
                    "-run",
                    "TestHealthReturnsOK",
                ],
            ],
            commands,
        )

    def test_command_failure_is_reported_and_stops_execution(self):
        calls = []

        def runner(command, **kwargs):
            calls.append(command)
            return subprocess.CompletedProcess(command, 1 if len(calls) == 1 else 0)

        result = impact.execute_commands(
            [["go", "test", "./bad"], ["go", "test", "./never"]],
            ROOT,
            runner=runner,
            clock=iter([1.0, 2.5]).__next__,
        )

        self.assertEqual(1, result["failed"])
        self.assertEqual(0, result["succeeded"])
        self.assertEqual(1, result["skipped"])
        self.assertEqual(1.5, result["durationSeconds"])
        self.assertEqual([["go", "test", "./bad"]], calls)


if __name__ == "__main__":
    unittest.main()
