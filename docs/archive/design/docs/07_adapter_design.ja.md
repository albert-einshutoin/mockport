# Adapter Design 日本語版

[English](07_adapter_design.md)

adapter design は、provider-like API surface と Mockport の scenario-driven behavior を分けて扱うための初期設計です。

## 方針

- adapter ごとに supported workflow と known gap を明示します。
- common helper は contract を曖昧にしない範囲で使います。
- docs、fixtures、tests、report を合わせて更新します。
