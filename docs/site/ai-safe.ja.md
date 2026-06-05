# AI-safe Mode 日本語版

[English](ai-safe.md)

AI-safe mode は、実 credential や本番外部サービス URL を使わずに Mockport を動かすための default safety layer です。

## 使い方

- 通常は `mode: ai-safe` のまま使います。
- CI で unsafe config を失敗にしたい場合は `strict` mode を使います。
- `mockport run --check` で起動前に config を確認できます。
