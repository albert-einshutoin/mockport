# Contributing 日本語版

[English](CONTRIBUTING.md)

## セットアップ

Go 1.26.5 を使用します。mise、asdf、Homebrew、公式インストーラーなどで Go toolchain を導入し、PATH 上で `go` が使えることを確認してください。

```bash
go test ./...
go vet ./...
go build ./cmd/mockport
```

## Spec-First TDD

本番コードの変更は spec-first TDD に従う必要があります。書かれた Mockport contract から始め、本番コードを変更する前に失敗するテストでギャップを証明してください。

1. 関連する spec を先に更新します。adapter の振る舞いは `docs/adapters/<adapter>.md` に、互換性の根拠は fixture、manifest、SDK/client contract、report に記載します。
2. 最小の有用なスライスに対する失敗するテストを書きます。
3. 狭い範囲のテストを実行し、期待どおりの理由で失敗することを確認します。
4. テストを通す最小の変更を実装します。
5. 狭い範囲のテストを再実行します。
6. metadata、docs、fixture、report、known gap を更新し、公開 claim が runtime の振る舞いと一致するようにします。
7. 触れた Phase のフル検証を実行します。

例を通すために provider surface を広げないでください。Mockport は選択された決定的なローカル workflow をサポートし、provider の内部全体はサポートしません。

### 新規 adapter のオンボーディング

adapter PR では、実装前にオンボーディングガイドに従ってください。

- [`docs/adding-an-adapter.md`](docs/adding-an-adapter.md)
- [`docs/scenario-policy.md`](docs/scenario-policy.md)
- [`docs/compatibility-model.md`](docs/compatibility-model.md)
- [`docs/adapter-helper-policy.ja.md`](docs/adapter-helper-policy.ja.md)

## Public Trust Checks

PR を開く前に以下を実行してください。標準的なローカルゲートでは、`make verify` が下記のコアチェックを連鎖実行します（distribution や release の検証は含みません）。

```bash
make verify
```

個別に実行する場合の基礎コマンド:

```bash
go vet ./...
go test ./...
go test -race ./...
bash scripts/check-public-trust.sh
bash scripts/check-adapter-completeness.sh
bash scripts/check-distribution.sh
```

adapter 変更や shared state の変更では、race detector 付きテストが期待されます。
`check-distribution.sh` は release/distribution 向けの追加確認であり、`make verify` には含まれません。

## Adapter Changes

adapter の受け入れ基準:

adapter の pull request には以下を含めてください。

- Spec-first TDD の根拠: spec 変更、失敗する regression または contract test、実装、最終検証。
- 成功・失敗シナリオの HTTP テスト。
- adapter の `Metadata()`、`mockport add <adapter>`、`mockport help <service>` と一致する metadata/report のカバレッジ。
- ユーザー向けの場合は example config または docs。
- fake credential とローカル base URL に対する AI-safe な振る舞い。
- report または docs に unsupported な振る舞いを明記。
- 実 provider secret、production URL、顧客 payload、未サニタイズの fixture を含めないこと。

## Pull Requests

PR 本文にテストの根拠を含めてください。実 provider secret、production URL、顧客 payload、未サニタイズの fixture を含めないでください。
