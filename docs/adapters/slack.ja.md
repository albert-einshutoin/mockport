# Slack Adapter 日本語版

[English](slack.md)

Slack adapter は、messaging workflow と Events API の selected subset を local で検証するための adapter です。

## 対応範囲

- `auth.test`、conversation list/history、message post/update/delete。
- URL verification と message callback subset。
- real delivery、Block Kit validation、files、enterprise policy、workspace directory の完全再現は対象外です。

詳細な endpoint と error model は英語版を正とします。
