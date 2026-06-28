> **⚠️ アーカイブ注意 — Not maintained, may diverge, do not cite as authoritative.**
>
> 実装開始前の設計アーカイブです。**実装の正本ではありません。**
> 現行仕様は [docs/site/](../../../site/index.ja.md) を参照してください。

# CLI Spec 日本語版

[English](08_cli_spec.md)

CLI は `mockport init`、`mockport run`、`mockport report` などの developer workflow を提供します。

## 目的

- 空 directory から local emulator を起動できるようにします。
- generated files をデフォルトで保護します。
- check/report により安全性と実行結果を確認します。
