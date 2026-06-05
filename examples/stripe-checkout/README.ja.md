# Stripe Checkout Example 日本語版

[English](README.md)

この example は、Stripe-like checkout session を local Mockport に送るための最小構成です。

## 確認すること

- `STRIPE_API_URL` を local Mockport に向けること。
- `mockport_stripe_secret` などの fake credential を使うこと。
- checkout session、idempotency、fake webhook/report の流れ。
