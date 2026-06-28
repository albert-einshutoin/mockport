> **⚠️ アーカイブ注意 — Not maintained, may diverge, do not cite as authoritative.**
>
> 実装開始前の設計アーカイブです。**実装の正本ではありません。**
> 現行仕様は [docs/site/](../../../site/index.ja.md) を参照してください。

# Problem Statement 日本語版

[English](01_problem_statement.md)

外部サービス連携の開発では、実 API key、sandbox account、webhook setup、CI secret 管理が摩擦になります。AI coding workflow では secret exposure のリスクも上がります。

## Mockport が解く問題

- local/CI で外部サービス風 API を安全に使う。
- success だけでなく failure や auth error を再現する。
- report により、何を検証したかを可視化する。
