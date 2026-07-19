# Reports 日本語版

[English](reports.md)

reports は、Mockport の test run が何を実行し、どの safety check に引っかかったかを確認するための user-facing evidence です。

## 確認方法

- HTTP では `/_mockport/report` を参照します。
- CLI では text の `mockport report`、または機械可読な `mockport report --json` を使います。既存の `--format json` も後方互換として利用できます。
- CI では report を artifact として保存すると、adapter coverage と safety status を追いやすくなります。
- JSON フィールドの詳細は英語版 [Report schema](report-schema.md) を参照してください。

## リクエスト履歴

リクエスト履歴は、実行中に記録された最新の500件のリクエストのメタデータを保持します。この制限を超えると、古いエントリから順に削除（プルーニング）され、レポートには常に最新のリクエストが時系列順に返されます。この制限付きの履歴は、レポートのペイロード内の `unsupported_endpoints` にも同様に適用されます。

`MOCKPORT_REQUEST_HISTORY` に正の整数を指定すると、デフォルト上限を上書きできます。無効な値（空、0、負数、非数値、オーバーフロー）は 500 にフォールバックします。

履歴が切り詰められた場合、レポートには `request_history` サマリー（`limit`、`retained`、`evicted`、`truncated`）が含まれます。テキストレポートは `truncated` が true のときだけ切り詰め行を表示します。
