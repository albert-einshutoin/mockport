# Reporting 日本語版

[English](reporting.md)

Mockport の report は、local test run でどの adapter/scenario が実行され、安全性 warning が出たかを確認するための evidence です。

## 見るポイント

- `/_mockport/report` と `mockport report` の出力。
- adapter workflow、request count、scenario、safety summary。
- `request_history` の `truncated` と `evicted`（履歴が上限を超えた場合）。
- CI artifact や compatibility report と組み合わせた確認。

機械可読な JSON report は `mockport report --json` で出力します。既存の `mockport report --format json` も後方互換として利用できます。通常の `mockport report` は従来どおり text を出力します。

リクエスト履歴はデフォルトで最新 500 件を保持します。`MOCKPORT_REQUEST_HISTORY` で正の整数のみ上書きできます。無効な値は 500 にフォールバックします。
