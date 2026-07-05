# Reports 日本語版

[English](reports.md)

reports は、Mockport の test run が何を実行し、どの safety check に引っかかったかを確認するための user-facing evidence です。

## 確認方法

- HTTP では `/_mockport/report` を参照します。
- CLI では `mockport report` または JSON format を使います。
- CI では report を artifact として保存すると、adapter coverage と safety status を追いやすくなります。

## JSON ペイロードの主要フィールド

`/_mockport/report` と `mockport report --format json` は同じ JSON を返します。トップレベルフィールド:

| フィールド | 説明 |
| --- | --- |
| `mode` | 設定された Mockport の実行モード（例: `ai-safe`、`strict`）。 |
| `safety` | AI-safe サマリー: 実行モード、総合的な安全状態、real-looking secret と external URL の件数、public env の安全性。 |
| `adapters` | 有効な adapter の base path、scenario、maturity、capabilities。 |
| `requests` | 処理済みトラフィックの replay-safe なリクエストメタデータ。本文や secret header は保存されません。 |
| `safety_warnings` | `safety` サマリーに寄与した個別の safety warning。 |
| `scenario_coverage` | adapter ごとの対応 scenario。 |
| `behavior_matrix` | 対応 endpoint と、その endpoint を exercise する scenario。 |
| `compatibility` | 測定された compatibility level、score、provider version、SDK/client evidence、known gap。空の場合は省略されます。 |
| `state_coverage` | adapter ごとの stateful resource coverage、idempotency、reset 対応。空の場合は省略されます。 |
| `unsupported_endpoints` | unsupported-endpoint 分類を返したリクエスト。`requests` から導出されます。 |

## リクエスト履歴

リクエスト履歴は、実行中に記録された最新の1000件のリクエストのメタデータを保持します。この制限を超えると、古いエントリから順に削除（プルーニング）され、レポートには常に最新のリクエストが時系列順に返されます。この制限付きの履歴は、レポートのペイロード内の `unsupported_endpoints` にも同様に適用されます。

CLI 出力:

```bash
mockport report --format text
mockport report --format json
```
