> **⚠️ アーカイブ注意 — Not maintained, may diverge, do not cite as authoritative.**
>
> 実装開始前の設計アーカイブです。**実装の正本ではありません。**
> 現行仕様は [docs/site/](../../../site/index.ja.md) を参照してください。

# Config Spec 日本語版

[English](06_config_spec.md)

Mockport の config は、mode、adapter、base path、scenario、安全性設定を local/CI で再現可能にするための入力です。

## 方針

- default は safe な fake/local value を使います。
- real-looking secret や external provider URL は警告または失敗にします。
- config check により startup 前に問題を検出します。
