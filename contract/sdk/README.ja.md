# SDK Contract Harness 日本語版

[English](README.md)

この directory は、公式または代表的な SDK/client が Mockport の local endpoint に対して期待通り動くかを検証するための contract harness を扱います。

## 役割

- selected workflow に限定して SDK 呼び出しの互換性を確認します。
- adapter の fake state、validation error、idempotency、streaming などを evidence として残します。
- provider の全機能を clone する目的ではありません。
