# Architecture 日本語版

[English](03_architecture.md)

Mockport は Go server、adapter layer、scenario/state、reporting、CLI/Docker distribution で構成されます。

## 境界

- server は routing と lifecycle を持ちます。
- adapter は provider-like API と workflow behavior を持ちます。
- state は deterministic fake data と idempotency を扱います。
- report は実行 evidence と safety summary を返します。
