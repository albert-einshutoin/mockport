> **⚠️ アーカイブ注意 — Not maintained, may diverge, do not cite as authoritative.**
>
> 実装開始前の設計アーカイブです。**実装の正本ではありません。**
> 現行仕様は [docs/site/](../../../site/index.ja.md) を参照してください。

# Testing Strategy 日本語版

[English](11_testing_strategy.md)

Testing strategy は、unit test、adapter behavior test、smoke test、Docker/Compose check、public env scan を組み合わせて confidence を作る方針です。

## 見るポイント

- success/failure/auth/rate-limit/timeout scenario。
- state と idempotency の race-free behavior。
- report output と compatibility evidence。
