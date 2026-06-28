> **⚠️ アーカイブ注意 — Not maintained, may diverge, do not cite as authoritative.**
>
> 実装開始前の設計アーカイブです。**実装の正本ではありません。**
> 現行仕様は [docs/site/](../../../site/index.ja.md) を参照してください。

# Adapter Design 日本語版

[English](07_adapter_design.md)

adapter design は、provider-like API surface と Mockport の scenario-driven behavior を分けて扱うための初期設計です。

## 方針

- adapter ごとに supported workflow と known gap を明示します。
- common helper は contract を曖昧にしない範囲で使います。
- docs、fixtures、tests、report を合わせて更新します。
