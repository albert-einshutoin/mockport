# Contributing 日本語版

[English](CONTRIBUTING.md)

Mockport への contribution では、secret-free なローカル統合テストという製品境界を保ちつつ、adapter surface、fixture、report、CI evidence をそろえることを重視します。

## セットアップ

Go 1.26.4 を使用します。mise、asdf、Homebrew、公式インストーラーなどで Go toolchain を導入し、PATH 上で `go` が使えることを確認してください。

## 進め方

- 変更前に issue または既存 roadmap/task の意図を確認します。
- adapter 追加や互換性変更では、human-readable docs と machine-checkable evidence を両方更新します。
- adapter 変更や shared state の変更では、race detector 付きテスト `go test -race ./...` を実行してください。
- public examples には実 credential、production URL、顧客 payload を含めないでください。
- PR では実行した test、残した known gap、互換性への影響を明記します。

詳細な branch/PR 手順、commit style、review policy は英語版を正とします。
