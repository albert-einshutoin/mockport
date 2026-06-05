# Multi Adapter Example 日本語版

[English](README.md)

この example は、複数 adapter を同じ Mockport instance で起動し、アプリケーションが複数外部サービスを local endpoint に向けられることを確認します。

## 対象

- Stripe、OpenAI、GitHub OAuth、Slack、LINE などの同時設定。
- `.env.mockport` の fake value。
- `/_mockport/report` による実行確認。
