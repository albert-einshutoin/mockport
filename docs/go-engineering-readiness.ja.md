# Go Engineering Readiness 日本語版

[English](go-engineering-readiness.md)

この文書は、Mockport の Go 実装が production-quality に近い構造を保てているかを確認する readiness guide です。

## 見るポイント

- typed boundary、explicit error handling、context propagation、graceful shutdown。
- zero-value safety、deep copy、race-free state management。
- `go test -race`、staticcheck、govulncheck、smoke test などの gate。
- adapter、state、report、server boundary の責務分離。

## 低優先 polish の扱い

Issue #23 の server lifecycle 項目は、run command の HTTP server に `ReadHeaderTimeout`、`ReadTimeout`、`IdleTimeout`、`MaxHeaderBytes` を設定して対応します。`WriteTimeout` は streaming 風の local test flow を途中で切る可能性があるため、現時点では意図的に未設定です。

JSON response write は、response 送出中の best-effort write として扱う場合のみ ignored error policy の範囲に残します。`LevelClient` の scoring/modeling と manifest の method/path/unsupported behavior 検証強化は compatibility semantics を変えるため、別 implementation issue として扱う方針です。
