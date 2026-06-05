# Compatibility Fixtures 日本語版

[English](README.md)

この directory は、adapter の互換性 claims を検証するための sanitized fixture を置く場所です。fixture は provider behavior をそのまま公開するためではなく、Mockport が対応する local workflow を再現可能にするために使います。

## 方針

- 実 credential、顧客 payload、production response を含めません。
- 仕様書、manifest、contract test と対応が追える形にします。
- provider-compatible claim を上げる場合は、fixture と known gap を一緒に更新します。
