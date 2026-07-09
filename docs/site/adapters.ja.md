# Adapters

[English](adapters.md)

Mockport の adapter は scenario-driven です。現時点では、選択された workflow をローカルおよび CI で検証できる `workflow-compatible` な surface に集中しています。

| Adapter | Base path | Maturity | Workflows |
| --- | --- | --- | --- |
| `stripe` | `/stripe` と SDK-compatible な `/v1` alias | `workflow-compatible` | checkout sessions, payment intents, customers, products, prices, subscriptions, invoices, refunds, fake signed webhooks |
| `openai` | `/openai` | `workflow-compatible` | models, chat completions, responses, embeddings, files, batches |
| `github-oauth` | `/github` | `workflow-compatible` | authorize redirect, token exchange, user profile, user emails, user orgs |
| `slack` | `/slack` | `workflow-compatible` | auth test, conversations list/history, message post/update/delete, Events API URL verification/message callback subset |
| `line` | `/line` | `workflow-compatible` | Messaging API send/content/signed webhook/rich menu/channel token workflows, LINE Login OAuth/profile, LIFF helpers, MINI App service messages, LINE Pay request/confirm, Mini Dapp wallet/payment helpers |
| `zoho-oauth` | `/zoho` | `workflow-compatible` | authorize redirect（state echo）, access token exchange, user info（`Zoho-oauthtoken` scheme） |

対応範囲を判断するときは、[support matrix](support-matrix.ja.md) と compatibility report を確認してください。Mockport は外部 provider の内部実装や未公開仕様を再現するものではなく、ローカル統合テストで必要になる成功、失敗、認証エラー、rate limit、timeout、webhook/callback などの検証に集中しています。

## シナリオの切り替え方

シナリオは2通りの方法で切り替えられます。

### 1. mockport.yml（設定ファイル）

```yaml
adapters:
  stripe:
    scenario: payment_failed
```

設定を変更するにはサーバーの再起動が必要です。

### 2. X-Mockport-Scenario ヘッダ（リクエスト単位）

リクエストに `X-Mockport-Scenario` ヘッダを付けることで、サーバーを再起動せずにリクエスト単位でシナリオを切り替えられます。

```bash
curl -X POST http://localhost:43101/stripe/v1/checkout/sessions \
  -H "X-Mockport-Scenario: payment_failed" \
  -H "Authorization: Bearer $STRIPE_KEY" \
  -d "mode=payment&success_url=http://localhost/success&cancel_url=http://localhost/cancel"
```

解決順序: **ヘッダ > config の scenario > アダプタのデフォルト**

- 未知のシナリオ名は拒否されます（黙って成功系にフォールバックしません）。エラー形式は各プロバイダ固有の形式に従い、多くのアダプタは HTTP 400 を返しますが、LINE Pay は実 API に合わせて HTTP 200 で `returnCode`/`returnMessage`（`unknown_mockport_scenario` を含む）を返します
- ヘッダによる切り替えはリクエスト単位なので並列テストでも干渉しません
- 対象はアダプタの `Metadata().Scenarios` に登録された組み込みシナリオのみです
- `/test/reset` などの管理用エンドポイントは状態リセット専用で、シナリオ検証の対象外です

## Request body 上限

Mockport は adapter handler より前に、**1 MiB（1,048,576 bytes）** を超える request body を拒否します。ローカルおよび CI の emulator 実行で unbounded read を避けるための server-wide 制限で、現行 adapter workflow と fixture には十分な上限です。超過時は `413 Payload Too Large` と次の本文を返します。

```text
request body too large
```

adapter handler 側でも同じ上限で provider 形式のエラーを返す場合があります。

詳細な adapter 仕様:

- [Stripe adapter](../adapters/stripe.ja.md)
- [OpenAI adapter](../adapters/openai.ja.md)
- [GitHub OAuth adapter](../adapters/github-oauth.ja.md)
- [Slack adapter](../adapters/slack.ja.md)
- [LINE adapter](../adapters/line.ja.md)
- [Zoho OAuth adapter](../adapters/zoho-oauth.ja.md)
