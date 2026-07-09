# Adapters

[日本語版](adapters.ja.md)

| Adapter | Base path | Maturity | Workflows |
| --- | --- | --- | --- |
| `stripe` | `/stripe` plus SDK-compatible `/v1` alias | `workflow-compatible` | checkout sessions, payment intents, customers, products, prices, subscriptions, invoices, refunds, fake signed webhooks |
| `openai` | `/openai` | `workflow-compatible` | models, chat completions, responses, embeddings, files, batches |
| `github-oauth` | `/github` | `workflow-compatible` | authorize redirect, token exchange, user profile, user emails, user orgs |
| `slack` | `/slack` | `workflow-compatible` | auth test, conversations list/history, message post/update/delete, Events API URL verification/message callback subset |
| `line` | `/line` | `workflow-compatible` | Messaging API send/content/signed webhook/rich menu/channel token workflows, LINE Login OAuth/profile, LIFF helpers, MINI App service messages, LINE Pay request/confirm, Mini Dapp wallet/payment helpers |
| `zoho-oauth` | `/zoho` | `workflow-compatible` | authorize redirect with state echo, access token exchange, user info via the `Zoho-oauthtoken` scheme |

Adapters are scenario-driven today and are moving toward provider-compatible local APIs for selected workflows. Use the [support matrix](support-matrix.md) and report behavior matrix to confirm supported paths.

`timeout` scenarios return an immediate 504-style response shape. To test client-side timeout behavior, add the server-wide `X-Mockport-Delay` header with a delay in milliseconds before Mockport handles the request.

| Header value | Behavior |
| --- | --- |
| Missing | No artificial delay; request proceeds immediately. |
| `0` | Accepted; no sleep before handling. |
| Positive (`1`–`30000`) | Sleep for the given milliseconds, then handle the request. |
| Empty or whitespace-only | Rejected with `400 Bad Request`; no sleep. |
| Non-integer | Rejected with `400 Bad Request`; no sleep. |
| Negative | Rejected with `400 Bad Request`; no sleep. |
| Above `30000` | Rejected with `400 Bad Request`; no sleep. |

Invalid values return:

```text
invalid X-Mockport-Delay: must be 0-30000 (milliseconds)
```

Example delayed request:

```bash
curl -H "X-Mockport-Delay: 250" http://localhost:43101/stripe/v1/customers
```

Mockport rejects request bodies larger than **1 MiB (1,048,576 bytes)** before adapter handlers run. This server-wide limit protects local and CI emulator runs from unbounded reads while staying high enough for current adapter workflows and fixtures. Oversized bodies return `413 Payload Too Large` with:

```text
request body too large
```

Adapter handlers may apply the same limit independently for provider-shaped error responses on bodies that pass the server check.

Detailed adapter specifications:

- [Stripe adapter](../adapters/stripe.md)
- [OpenAI adapter](../adapters/openai.md)
- [GitHub OAuth adapter](../adapters/github-oauth.md)
- [Slack adapter](../adapters/slack.md)
- [LINE adapter](../adapters/line.md)
- [Zoho OAuth adapter](../adapters/zoho-oauth.md)
