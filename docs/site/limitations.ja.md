# Limitations 日本語版

[English](limitations.md)

Mockport は selected workflow の local integration test を目的にしており、provider の内部実装や未公開 behavior は再現しません。

## 現在の Preview スコープ

mainline adapter は selected local / CI integration path 向けに workflow-compatible です。provider sandbox や production validation の代替ではありません。

## つまずきやすい具体例

adapter spec、compatibility report、runtime 挙動と突合した症状ベースの制限です。

- **Stripe: 3DS / SCA（`requires_action`）フローは返しません** — PaymentIntent / Checkout Session は built-in シナリオの成功または decline 形状のみ。カード認証 UI の分岐や `requires_action` 処理は local では試せません。
- **Stripe: 課金ネットワークの計算はしません** — 金額・通貨はリクエスト値をそのまま返します。tax、按分、settlement、disputes、Connect、Billing lifecycle 全体は再現しません。
- **OpenAI: `/v1/responses` のストリーミングは未対応** — `/v1/chat/completions` は `stream: true` または `stream_success` シナリオで SSE 対応しますが、Responses API は常に JSON を返します（`stream_success` 設定時も同様）。Responses API の provider-compatible な SSE streaming は未実装です。

  ```bash
  curl -i -X POST http://localhost:43101/openai/v1/responses \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -H 'Content-Type: application/json' \
    -d '{"model":"gpt-mockport","input":"hello","stream":true}'
  ```

  Mockport は `Content-Type: application/json` と deterministic な response object を返します。`text/event-stream`、`data:` chunk、`response.output_text.delta` のような named event は返しません。
- **OpenAI: 実 inference 品質は再現しません** — 応答は deterministic な placeholder。model quality、tokenization parity、hosted tools、vector stores、provider scheduling は対象外。
- **Slack: 実メッセージ配送や Events API 全体は対象外** — local message state と URL verification / message callback の subset のみ。実 workspace 配送、Block Kit validation、files、app scopes、enterprise directory は未対応。
- **LINE: 実 Login UI や LIFF browser はありません** — OAuth code/token/profile は local で動作。QR login、LIFF runtime、署名付き ID token、provider 側 webhook 再配送、quota enforcement（シナリオ以外）は未対応。
- **全般: `mockport.yml` の `scenarios:` ブロックは未実装** — パースされますが runtime には適用されません。存在時は起動時・`--check`・`/_mockport/report` で警告します。
- **全般: 状態はメモリ内のみ** — コンテナまたはプロセス再起動で消えます。永続化レイヤーはありません。

adapter ごとの gap 一覧は [support matrix](support-matrix.ja.md)、`docs/adapters/`、 [compatibility reports](../compatibility-reports/latest.ja.md) を参照してください。

## 対象外

- 実 payment processing、fraud、settlement、billing network。
- 実 AI inference、tokenization parity、private scheduling。
- GitHub / Slack / LINE の enterprise policy や full directory state。
- 未公開 provider behavior。

## 未実装の設定ブロック

`mockport.yml` の `scenarios:` ブロックはパースされますが**未実装**です。ランタイムでは無視されます。
このブロックが存在する場合、起動時（`--check` 出力および `/_mockport/report`）に警告が出力されます。

レスポンスの切り替えやエラーケースのシミュレーションには以下を使用してください：

- アダプタの `scenario:` フィールドによる built-in シナリオ
- `X-Mockport-Scenario` リクエストヘッダ（issue #80 参照）

ユーザー定義シナリオの将来の計画については [scenario-policy.md](../scenario-policy.md) を参照してください。

## 運用上の注意

### ポート衝突（デフォルト `43101`）

Mockport のデフォルトポートは `43101` です。既に使用中だと起動に失敗し、次のようなエラーになります。

```text
listen on 127.0.0.1:43101: address already in use; choose another port or stop the existing process
```

`mockport.yml` でポートを変更できます。

```yaml
server:
  port: 43102
```

アプリの env も合わせて更新してください（例: `STRIPE_API_URL=http://localhost:43102/stripe`）。

### Docker Compose ネットワーク

同一 Compose ネットワーク上の別コンテナからは、サービス名で指します（`localhost` ではありません）。

```env
STRIPE_API_URL=http://mockport:43101/stripe
```

アプリコンテナ内の `localhost` は Mockport コンテナを指しません。

## サポート確認方法

- [Support matrix](support-matrix.ja.md)
- `/_mockport/report`
- `mockport report --json`（`--format json` も引き続き利用可能）
- Adapter examples
- Public env safety checks

採用前に unsupported endpoint と approximation を確認してください。
