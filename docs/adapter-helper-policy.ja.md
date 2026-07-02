# Adapter Helper Policy 日本語版

[English](adapter-helper-policy.md)

adapter helper は、provider ごとの差異を隠しすぎず、Mockport の supported workflow を明確に保つための補助層です。

## 要点

- helper は共通化のために使い、provider 固有の contract を曖昧にしません。
- error handling、state mutation、request validation は adapter spec と test で追えるようにします。
- helper を増やす場合は、複数 adapter で実際の重複を減らすことを条件にします。

## 重複 helper 名の追跡

[`scripts/check-adapter-helpers.sh`](../scripts/check-adapter-helpers.sh) は、built-in adapter パッケージ内の unexported helper 名（`writeJSON` や `normalizeScenario` など）の重複を機械的に一覧化します。

この script は追跡用であり、即時の共通化を義務づけるものではありません。

- 名前の重複は、adapter 間で同一挙動であることの証明にはなりません。
- 共通化の前に、provider 固有の response shape、headers、status code、scenario default を regression test で保護する必要があります。
- 既定では重複を報告して正常終了します。

`DUPLICATE_ADAPTER_THRESHOLD` を超える adapter 数で同じ helper 名が見つかった場合のみ失敗します。既定の閾値は現在の built-in adapter パッケージ数と同じで、通常の重複は CI を止めずに可視化するための値です。
