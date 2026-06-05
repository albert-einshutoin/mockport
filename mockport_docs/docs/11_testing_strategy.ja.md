# Testing Strategy 日本語版

[English](11_testing_strategy.md)

Testing strategy は、unit test、adapter behavior test、smoke test、Docker/Compose check、public env scan を組み合わせて confidence を作る方針です。

## 見るポイント

- success/failure/auth/rate-limit/timeout scenario。
- state と idempotency の race-free behavior。
- report output と compatibility evidence。
