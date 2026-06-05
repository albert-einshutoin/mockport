# Fixture Policy 日本語版

[English](fixture-policy.md)

fixture は Mockport の supported behavior を再現可能にするための test data です。実 provider から取得した秘密情報や顧客データを保存する場所ではありません。

## 方針

- fixture は sanitized で、commit 可能な fake data のみを含めます。
- adapter spec、contract harness、compatibility report と対応を持たせます。
- 不明確な production-like value は public env safety の観点で避けます。
