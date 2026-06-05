# Reporting 日本語版

[English](reporting.md)

Mockport の report は、local test run でどの adapter/scenario が実行され、安全性 warning が出たかを確認するための evidence です。

## 見るポイント

- `/_mockport/report` と `mockport report` の出力。
- adapter workflow、request count、scenario、safety summary。
- CI artifact や compatibility report と組み合わせた確認。
