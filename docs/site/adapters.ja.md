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
