import importlib.util
import json
import pathlib
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[2]


def load_module(name: str, path: pathlib.Path):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


go_adapter = load_module("go_adapter", ROOT / "ci" / "adapters" / "go.py")


class GoAdapterTests(unittest.TestCase):
    def test_changed_package_includes_reverse_dependencies(self):
        packages = [
            go_adapter.GoPackage("example/shared", "internal/shared", []),
            go_adapter.GoPackage(
                "example/service", "internal/service", ["example/shared"]
            ),
            go_adapter.GoPackage("example/cmd", "cmd/app", ["example/service"]),
        ]

        selected = go_adapter.select_packages(
            packages, ["internal/shared/value.go"]
        )

        self.assertEqual(
            ["example/cmd", "example/service", "example/shared"], selected
        )

    def test_changed_test_file_selects_its_package(self):
        packages = [go_adapter.GoPackage("example/service", "internal/service", [])]

        selected = go_adapter.select_packages(
            packages, ["internal/service/service_test.go"]
        )

        self.assertEqual(["example/service"], selected)

    def test_deleted_unmappable_go_package_is_unsafe(self):
        packages = [go_adapter.GoPackage("example/service", "internal/service", [])]

        with self.assertRaisesRegex(ValueError, "cannot map"):
            go_adapter.select_packages(packages, ["internal/removed/value.go"])

    def test_decodes_concatenated_go_list_json_and_test_imports(self):
        payload = "".join(
            [
                json.dumps(
                    {
                        "ImportPath": "example/shared",
                        "Dir": str(ROOT / "internal/shared"),
                        "Imports": ["fmt"],
                        "TestImports": ["example/testutil"],
                    }
                ),
                json.dumps(
                    {
                        "ImportPath": "example/service",
                        "Dir": str(ROOT / "internal/service"),
                        "Imports": ["example/shared"],
                    }
                ),
            ]
        )

        packages = go_adapter.decode_go_list(payload, ROOT)

        self.assertEqual("internal/shared", packages[0].directory)
        self.assertEqual(["fmt", "example/testutil"], packages[0].imports)


if __name__ == "__main__":
    unittest.main()
