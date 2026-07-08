# Scenario Policy 日本語版

[English](scenario-policy.md)

Mockport は built-in scenario を優先します。user-defined scenario は将来の機能であり、fixture と compatibility の昇格プロセスを経るまで provider compatibility の主張から分離しておく必要があります。

built-in scenario は `workflow-compatible` なローカルテスト向けの **安定したローカル契約** です。Mockport 内の決定的な挙動を定義しますが、provider の完全再現を意味しません。将来 runtime support が実装された場合、user YAML scenario はアプリ固有テスト向けにローカル挙動を積み増せますが、provider compatibility を意味したり、文書化済みの gap を隠したりしてはいけません。

## Built-in Scenarios

built-in scenario は Mockport が保守します。各 scenario は adapter spec（`docs/adapters/<adapter>.md`）に決定的な契約として公開されます。built-in adapter がローカルで再現できる範囲の正とします。次を満たす必要があります。

- 安定した scenario 名
- adapter metadata
- test
- public-safe な example
- compatibility scoring に使う場合は fixture または documentation evidence

built-in scenario は source metadata と report で可視化された場合にのみ compatibility score に寄与できます。

## User-defined Scenarios

user-defined scenario は、将来プロジェクトローカルの挙動を定義するための機能です。**ローカルテストの便宜** に限られ、**provider compatibility の主張ではありません**。runtime support の実装後はアプリ固有テスト向けにローカル応答や state を積み増せますが、単体では provider compatibility を証明せず、built-in adapter 契約のように文書化してはいけません。

### Built-in vs user-defined

| 観点 | built-in adapter scenario | user-defined scenario |
| --- | --- | --- |
| 出所 | Mockport が adapter spec と test で保守 | プロジェクトでローカル定義（将来の `scenarios:` ブロック） |
| 契約 | `docs/adapters/<adapter>.md` に公開される決定的な契約 | プロジェクトローカル挙動。public adapter 契約ではない |
| compatibility への影響 | test、docs、fixture evidence があれば compatibility scoring に寄与可能 | provider compatibility score や maturity を上げない |
| runtime 状態 | `adapters.<name>.scenario` と `X-Mockport-Scenario` で有効 | `scenarios:` ブロックはパースされるが runtime 未実装 |

### 最小構成例

built-in scenario の選択は adapter の `scenario:` フィールドを使います。docs、fixture、config の example は **fake local 値のみ**（mock secret、local URL、synthetic ID）を使ってください。本番 credential、顧客 payload、実 provider endpoint を scenario example に入れないでください。

```yaml
adapters:
  stripe:
    enabled: true
    base_path: /stripe
    scenario: payment_success
    fake_secret: mockport_stripe_secret
    webhook:
      target_url: http://app:3000/webhooks/stripe
      signing_secret: whsec_mockport
```

これは Stripe adapter spec の built-in `payment_success` 契約を選択します。user-defined scenario は定義しません。

完全な user-defined scenario システムが存在するまで、adapter は部分的な custom 挙動より明示的な built-in scenario を優先してください。user-defined scenario が後に昇格する場合は、built-in scenario 名、test、docs、sanitized fixture evidence が必要です。

ヘッダ override（`X-Mockport-Scenario`）はビルトインシナリオのみ対象です。ユーザー定義シナリオへのリクエスト単位の切り替えはスコープ外です。

> **現在の状況:** `mockport.yml` の `scenarios:` ブロックはパースされますが、ランタイムではまだ実装されていません。このブロックが存在する場合、Mockport は警告を出力します。詳細は [limitations](site/limitations.ja.md#未実装の設定ブロック) を参照してください。

## Compatibility Boundary

compatibility scoring は次を区別する必要があります。

- Mockport built-in scenario coverage
- SDK contract coverage
- workflow state coverage
- user-defined local behavior

最初の 3 つだけが provider compatibility maturity を上げられます。user-defined local behavior は報告できますが、unsupported provider behavior を隠してはいけません。

unsupported provider behavior は adapter spec、support matrix、compatibility report に **known gap** として残す必要があります。scenario の docs や example で coverage を過大に主張したり、user-defined override が文書化済み gap を埋めたかのように示したりしてはいけません。
