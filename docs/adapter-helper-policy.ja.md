# Adapter Helper Policy 日本語版

[English](adapter-helper-policy.md)

adapter helper は、provider ごとの差異を隠しすぎず、Mockport の supported workflow を明確に保つための補助層です。

## 要点

- helper は共通化のために使い、provider 固有の contract を曖昧にしません。
- error handling、state mutation、request validation は adapter spec と test で追えるようにします。
- helper を増やす場合は、複数 adapter で実際の重複を減らすことを条件にします。
