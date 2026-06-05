# Go Engineering Guide 日本語版

[English](04_go_engineering_guide.md)

Go 実装では、単純な構造、明確な interface、context、explicit error、race-free state を重視します。

## 方針

- adapter/server/state/report の責務を分けます。
- zero-value safety と deep copy を意識します。
- test、race detector、static analysis を readiness gate として使います。
