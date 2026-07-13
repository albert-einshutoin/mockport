# Stripe Adapter 日本語版

[English](stripe.md)

Stripe adapter は、payment integration の selected workflow を local で検証するための Stripe-like adapter です。

## 対応範囲

- checkout sessions、payment intents、customers、products、prices、subscriptions、invoices、refunds。
- fake signed webhook、validation error、stateful list/retrieve、idempotency replay。
- webhook send helper の outbound 配送は固定 `5s` timeout。timeout は `504` / `webhook_send_timeout`、target の non-2xx は `502` / `webhook_target_non_2xx`。
- `timeout` は即時の 504 レスポンス shape を返す。実レイテンシは server 全体の `X-Mockport-Delay`（`0`–`30000` ms、[Adapters](../site/adapters.md) 参照）で注入する。
- real payment processing、fraud、settlement、tax、disputes、Connect、full Billing lifecycle は対象外です。
- list レスポンスは `object`、`data`、`has_more`（常に `false`）、`url`（リクエスト path）を返す。cursor/limit ページングは未対応。

## Scenarios

| Scenario | レスポンス shape | レイテンシ動作 |
| --- | --- | --- |
| `payment_success` | `200` / Stripe 風 success object | scenario による sleep なし。実レイテンシは `X-Mockport-Delay` で制御（`0`–`30000` ms、[Adapters](../site/adapters.md) 参照）。 |
| `payment_failed` | `402` / `card_declined` | scenario による sleep なし。実レイテンシは `X-Mockport-Delay` で制御（`0`–`30000` ms、[Adapters](../site/adapters.md) 参照）。 |
| `auth_error` | `401` / `invalid_api_key` | scenario による sleep なし。実レイテンシは `X-Mockport-Delay` で制御（`0`–`30000` ms、[Adapters](../site/adapters.md) 参照）。 |
| `rate_limited` | `429` / `rate_limited` | scenario による sleep なし。実レイテンシは `X-Mockport-Delay` で制御（`0`–`30000` ms、[Adapters](../site/adapters.md) 参照）。 |
| `timeout` | `504` / `mockport_timeout` | 即時の `504` レスポンス shape のみ。scenario は **sleep や遅延処理を行わない**。実レイテンシは `X-Mockport-Delay` で制御（`0`–`30000` ms、[Adapters](../site/adapters.md) 参照）。 |

`timeout` scenario はレスポンス shape を制御するもので、リクエスト処理時間は制御しない。実レイテンシを入れる場合は `X-Mockport-Delay` を使う。

See also: [Scenario Policy 日本語版](../scenario-policy.ja.md).

詳細な endpoint、SDK contract、known gap は英語版を正とします。
