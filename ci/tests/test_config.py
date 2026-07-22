import json
import pathlib
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[2]


class ConfigTests(unittest.TestCase):
    def test_all_rule_targets_resolve_to_trusted_commands(self):
        config = json.loads((ROOT / "ci/config.json").read_text())
        commands = config["commands"]

        for rule in config["pathRules"]:
            for target in rule.get("checks", []):
                self.assertIn(target, commands["checks"])
            for target in rule.get("e2eTests", []):
                self.assertIn(target, commands["e2e"])
        for target in config["smokeTests"]:
            self.assertIn(target, commands["smoke"])

    def test_trusted_commands_are_nonempty_argv_arrays(self):
        config = json.loads((ROOT / "ci/config.json").read_text())
        commands = config["commands"]
        declared = list(commands["full"])
        declared.extend(commands["checks"].values())
        declared.extend(commands["e2e"].values())
        declared.extend(commands["smoke"].values())

        for command in declared:
            self.assertTrue(command)
            self.assertTrue(all(isinstance(argument, str) for argument in command))

    def test_ci_and_dependency_changes_require_full_suite(self):
        config = json.loads((ROOT / "ci/config.json").read_text())

        self.assertIn("ci/**", config["fullTestPatterns"])
        self.assertIn(".github/**", config["fullTestPatterns"])
        self.assertIn("go.mod", config["fullTestPatterns"])
        self.assertIn("go.sum", config["fullTestPatterns"])

    def test_full_suite_keeps_every_public_validation_gate(self):
        config = json.loads((ROOT / "ci/config.json").read_text())
        full_commands = {tuple(command) for command in config["commands"]["full"]}

        for script in (
            "check-adapter-completeness.sh",
            "check-distribution.sh",
            "check-doc-links.sh",
            "check-maintenance-policy.sh",
            "check-public-env.sh",
            "check-public-trust.sh",
        ):
            self.assertIn(("bash", f"scripts/{script}"), full_commands)

    def test_emergency_full_runner_does_not_depend_on_impact_config(self):
        runner = (ROOT / "ci/run-full.sh").read_text()

        self.assertNotIn("ci/impact.py", runner)
        self.assertNotIn("ci/config.json", runner)
        self.assertNotIn("ci/tests", runner)


if __name__ == "__main__":
    unittest.main()
