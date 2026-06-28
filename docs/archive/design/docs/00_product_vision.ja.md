> **⚠️ アーカイブ注意 — Not maintained, may diverge, do not cite as authoritative.**
>
> 実装開始前の設計アーカイブです。**実装の正本ではありません。**
> 現行仕様は [docs/site/](../../../site/index.ja.md) を参照してください。

# Product Vision 日本語版

[English](00_product_vision.md)

Mockport の製品ビジョンは、AI-native development と CI で外部サービス連携を secret-free に検証できる local emulator を提供することです。

## 要点

- 実 provider secret を使わずに integration path を確認します。
- Docker-first で導入し、Go binary は補助経路にします。
- full clone ではなく、selected workflow の再現性と safety を優先します。
