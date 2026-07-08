# Adapter Helper Policy 日本語版

[English](adapter-helper-policy.md)

Stripe、OpenAI、GitHub OAuth、Slack、LINE、Zoho OAuth は、現在各 adapter パッケージ内に `writeJSON` や `normalizeScenario` などの小さな local helper を保持しています。

adapter helper は、provider ごとの差異を隠しすぎず、Mockport の supported workflow を明確に保つための補助層です。

## 要点

- helper は共通化のために使い、provider 固有の contract を曖昧にしません。
- error handling、state mutation、request validation は adapter spec と test で追えるようにします。
- helper を増やす場合は、複数 adapter で実際の重複を減らすことを条件にします。

## 重複 helper 名の追跡

[`scripts/check-adapter-helpers.sh`](../scripts/check-adapter-helpers.sh) は、built-in adapter パッケージ内の unexported helper 名（`writeJSON` や `normalizeScenario` など）の重複を機械的に一覧化します。

この script は追跡用であり、即時の共通化を義務づけるものではありません。

- 重複 helper 名はレビュー用の根拠であり、今すぐ refactor すべきという意味ではありません。
- 名前の重複は、adapter 間で同一挙動であることの証明にはなりません。
- 共通化の前に、provider 固有の response shape、headers、status code、scenario default を regression test で保護する必要があります。test なしに provider 固有の挙動を shared helper の裏に隠さないでください。
- 既定では重複を報告して正常終了します。

### 出力例

重複した helper 名が 1 行ずつ出力され、続けて summary 行と pass/fail 行が出ます。

```text
duplicate helper: bearerToken (2 adapters: githuboauth line)
duplicate helper: clientIDMatches (3 adapters: githuboauth line zohooauth)
duplicate helper: decodePayload (2 adapters: line openai)
duplicate helper: firstNonEmpty (2 adapters: line zohooauth)
duplicate helper: newWithWebhookTimeout (2 adapters: line stripe)
duplicate helper: redirectWithQuery (3 adapters: githuboauth line zohooauth)
duplicate helper: writeError (2 adapters: openai zohooauth)
duplicate helper: writeOAuthError (2 adapters: githuboauth line)
check-adapter-helpers: 8 duplicated helper name(s) tracked (threshold=6)
check-adapter-helpers passed
```

helper が `DUPLICATE_ADAPTER_THRESHOLD` を超える adapter 数で見つかった場合、stderr に閾値警告が出て exit status 1 になります。

```text
duplicate helper: clientIDMatches (3 adapters: githuboauth line zohooauth)
  exceeds DUPLICATE_ADAPTER_THRESHOLD=2
...
duplicate helper: redirectWithQuery (3 adapters: githuboauth line zohooauth)
  exceeds DUPLICATE_ADAPTER_THRESHOLD=2
...
check-adapter-helpers: 8 duplicated helper name(s) tracked (threshold=2)
check-adapter-helpers failed: one or more helpers exceed DUPLICATE_ADAPTER_THRESHOLD=2
```

### `DUPLICATE_ADAPTER_THRESHOLD`

同じ helper 名が `DUPLICATE_ADAPTER_THRESHOLD` を超える adapter 数で見つかった場合のみ失敗します。既定値は現在の built-in adapter パッケージ数と同じで、通常の重複は CI を止めずに可視化するための値です。

- 通常の開発や CI では既定値のままにします。
- より広い重複を早めに検知してレビューで扱いたい場合は、閾値を下げます。
- 広い重複が一時的に想定される場合だけ、理由を文書化したうえで閾値を上げます。

例:

```bash
DUPLICATE_ADAPTER_THRESHOLD=4 bash scripts/check-adapter-helpers.sh
```
