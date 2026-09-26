# Stripe Adapter 日本語版

[English](stripe.md)

共通ルール: [シナリオポリシー](../scenario-policy.ja.md)。

Stripe adapter は、payment integration の selected workflow を local で検証するための Stripe-like adapter です。

## 対応範囲

- checkout sessions、payment intents、customers、products、prices、subscriptions、invoices、refunds。
- `/stripe/v1/...` の全 API route は、`/stripe` prefix を除いた SDK-compatible な `/v1/...` alias でも同じ Stripe-like contract を返す。`/stripe/test/...` の test helper は他 adapter との衝突を避けるため `/test/...` では公開しない。
- fake signed webhook、validation error、stateful list/retrieve、idempotency replay。
- Checkoutアプリ向けに、loopback専用の `/stripe/test/webhook/send` は `{"session_id":"<作成済みSession ID>","event_id":"evt_order_1","event_type":"checkout.session.completed"}` を受け取り、保存済みSessionのIDと `client_reference_id` をイベントに含めます。選べるイベントは支払済みの `checkout.session.completed` と未払いの `checkout.session.async_payment_failed` です。未知のSessionは `404`、不正入力は `400`。同じevent IDを再配送でき、Mockportは重複排除しません。署名timestampは毎回現在時刻で作り直します。空本文は既存の固定デモイベントです。[アプリE2E例](../../examples/app-e2e/README.md)を参照してください。
- webhook send helper の outbound 配送は固定 `5s` timeout。timeout は `504` / `webhook_send_timeout`、target の non-2xx は `502` / `webhook_target_non_2xx`。
- `timeout` は即時の 504 レスポンス shape を返す。実レイテンシは server 全体の `X-Mockport-Delay`（`0`–`30000` ms、[Adapters](../site/adapters.ja.md#x-mockport-delay) 参照）で注入する。
- real payment processing、fraud、settlement、tax、disputes、Connect、full Billing lifecycle は対象外です。

`auth_required: true` を設定すると、`/stripe/v1/...` とSDK用 `/v1/...` aliasは `Authorization: Bearer <fake_secret>` を厳密に検証し、欠落・不正形式・誤キーに `401` / `invalid_api_key` を返します。未設定時は従来どおり検証しません。`fake_secret` 省略時は `FakeEnv` と同じ既定のfake keyを使います。`/stripe/test/...` は従来のloopback・送信先制限で保護します。`auth_error` scenarioは正しいキーでも失敗させる別の検証です。

## Scenarios

| Scenario | レスポンス shape | レイテンシ動作 |
| --- | --- | --- |
| `payment_success` | `200` / Stripe 風 success object | scenario による sleep なし。実レイテンシは `X-Mockport-Delay` で制御（`0`–`30000` ms、[Adapters](../site/adapters.ja.md#x-mockport-delay) 参照）。 |
| `payment_failed` | `402` / `card_declined` | scenario による sleep なし。実レイテンシは `X-Mockport-Delay` で制御（`0`–`30000` ms、[Adapters](../site/adapters.ja.md#x-mockport-delay) 参照）。 |
| `auth_error` | `401` / `invalid_api_key` | scenario による sleep なし。実レイテンシは `X-Mockport-Delay` で制御（`0`–`30000` ms、[Adapters](../site/adapters.ja.md#x-mockport-delay) 参照）。 |
| `rate_limited` | `429` / `rate_limited` | scenario による sleep なし。実レイテンシは `X-Mockport-Delay` で制御（`0`–`30000` ms、[Adapters](../site/adapters.ja.md#x-mockport-delay) 参照）。 |
| `timeout` | `504` / `mockport_timeout` | 即時の `504` レスポンス shape のみ。scenario は **sleep や遅延処理を行わない**。実レイテンシは `X-Mockport-Delay` で制御（`0`–`30000` ms、[Adapters](../site/adapters.ja.md#x-mockport-delay) 参照）。 |

`timeout` scenario はレスポンス shape を制御するもので、リクエスト処理時間は制御しない。実レイテンシを入れる場合は `X-Mockport-Delay` を使う。

詳細な endpoint、SDK contract、known gap は英語版を正とします。
