# Repository Structure 日本語版

[English](05_repository_structure.md)

この文書は、Mockport の repository layout と各 directory の責務を説明する初期設計資料です。

## 見るポイント

- `adapters/` は provider-specific behavior。
- `internal/` は server、state、config、report などの実装境界。
- `docs/` と `docs/site/` は public docs と detailed specs。
- `examples/` は導入確認用の最小構成。
