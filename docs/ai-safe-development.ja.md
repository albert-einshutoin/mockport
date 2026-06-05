# AI-safe Development 日本語版

[English](ai-safe-development.md)

Mockport は、AI coding tool に実 secret を渡さずに local/CI integration test を動かすための安全側の default を提供します。

## 要点

- `ai-safe` mode は real-looking secret や live provider URL を警告します。
- `strict` mode は unsafe field があると startup 前に失敗します。
- CLI output と report は secret 全文を出さず、カテゴリと field 名中心で知らせます。
- safe fake value は `mockport_`、`local_`、`fake_`、`dummy_` などの local prefix を使います。
