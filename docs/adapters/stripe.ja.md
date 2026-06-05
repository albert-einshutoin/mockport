# Stripe Adapter 日本語版

[English](stripe.md)

Stripe adapter は、payment integration の selected workflow を local で検証するための Stripe-like adapter です。

## 対応範囲

- checkout sessions、payment intents、customers、products、prices、subscriptions、invoices、refunds。
- fake signed webhook、validation error、stateful list/retrieve、idempotency replay。
- real payment processing、fraud、settlement、tax、disputes、Connect、full Billing lifecycle は対象外です。

詳細な endpoint、SDK contract、known gap は英語版を正とします。
