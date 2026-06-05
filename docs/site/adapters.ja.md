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

対応範囲を判断するときは、[support matrix](support-matrix.ja.md) と compatibility report を確認してください。Mockport は外部 provider の内部実装や未公開仕様を再現するものではなく、ローカル統合テストで必要になる成功、失敗、認証エラー、rate limit、timeout、webhook/callback などの検証に集中しています。

詳細な adapter 仕様:

- [Stripe adapter](../adapters/stripe.ja.md)
- [OpenAI adapter](../adapters/openai.ja.md)
- [GitHub OAuth adapter](../adapters/github-oauth.ja.md)
- [Slack adapter](../adapters/slack.ja.md)
- [LINE adapter](../adapters/line.ja.md)
