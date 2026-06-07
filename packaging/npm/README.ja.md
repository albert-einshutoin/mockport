# npm Packaging 日本語版

[English](README.md)

npm package は Go binary/Docker を補助する wrapper として扱います。Mockport の主配布経路は Docker と GitHub release archive です。

## 注意

- npm wrapper は experimental です。
- binary fallback、platform detection、install script の安全性を確認します。
- Docker fallback は `MOCKPORT_IMAGE` があればそれを使い、未設定なら `ghcr.io/albert-einshutoin/mockport:0.1.0-alpha` を使います。
- 再現性が必要な CI では explicit release image tag を使います。`latest` は default branch image に追従し、preview release contract ではありません。
- real secret や production URL を package examples に含めません。
