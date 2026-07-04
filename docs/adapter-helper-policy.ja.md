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

同じ helper 名が `DUPLICATE_ADAPTER_THRESHOLD` を超える adapter 数で見つかった場合のみ失敗します。script は `>` 判定のため、閾値 `3` では 4 つ以上の adapter で同じ helper 名が見つかったときに失敗します。

### `DUPLICATE_ADAPTER_THRESHOLD` の調整

`DUPLICATE_ADAPTER_THRESHOLD` は、重複 helper 名が情報表示から CI 失敗に切り替わる境界を決めます。共通化すべきかどうかは決めません。名前の重複は追跡用のシグナルにとどまります。

| 用途 | 値 | 挙動 |
| --- | --- | --- |
| 通常の追跡 | 未設定（既定） | 現在の built-in adapter パッケージ数（現時点で 6）が使われます。重複は報告され、CI は通過します。 |
| 厳しめの共通化レビュー | `3` | 同じ helper 名が 4 つ以上の adapter に現れたとき CI が失敗します。上記の「4 つ以上の adapter で同一挙動が繰り返される」レビュー条件と一致します。 |

通常の adapter 作業では未設定のままにしてください。4 adapter 以上での共通化レビューを CI で強制したいときだけ `DUPLICATE_ADAPTER_THRESHOLD=3` を設定します。

閾値を超えても、helper 名の重複だけでは共通化の根拠にはなりません。共通化には adapter 間での同一挙動と、provider 固有の response shape、headers、status code、scenario default を保護する regression test が依然として必要です。
