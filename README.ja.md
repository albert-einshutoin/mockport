# Mockport

AI ネイティブ開発のための、シークレット不要なサービスエミュレーション。

[![CI](https://github.com/albert-einshutoin/mockport/actions/workflows/ci.yml/badge.svg)](https://github.com/albert-einshutoin/mockport/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/albert-einshutoin/mockport)](https://github.com/albert-einshutoin/mockport/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

[English](README.md)

## Why Mockport

Stripe、OpenAI、Slack、GitHub OAuth、LINE、Zoho OAuth の統合コードを、AI コーディングエージェントや CI に本物の API キーを渡さずにローカルで動かせます。SDK や HTTP クライアントの向き先を `localhost` に変えるだけです。

[stripe-mock](https://github.com/stripe/stripe-mock) など単一サービス向け mock とは異なり、Mockport は複数 SaaS adapter を 1 つの Docker プロセスで動かし、fake state、組み込み error scenario、webhook helper、secret-safe な既定値を提供します。機能比較は [Comparison](docs/site/comparison.ja.md) を参照してください。

## 30 秒クイックスタート

ローカルインストールなしで試せます。

```bash
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/examples/stripe-checkout/mockport.yml:/etc/mockport/mockport.yml \
  ghcr.io/albert-einshutoin/mockport:0.1.0-alpha \
  run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

別ターミナルで:

```bash
curl http://localhost:43101/health
curl -X POST http://localhost:43101/stripe/v1/checkout/sessions
curl http://localhost:43101/_mockport/report
```

### ソースから

リポジトリ checkout から binary と image をビルドします。

```bash
make build
docker build -t mockport:local -f docker/Dockerfile .
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/examples/stripe-checkout/mockport.yml:/etc/mockport/mockport.yml \
  mockport:local run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

空ディレクトリ smoke test:

```bash
bash scripts/smoke-empty-dir.sh
```

## 動作イメージ

ヘルスチェック:

```bash
$ curl http://localhost:43101/health
{"status":"ok"}
```

Stripe 風 checkout session:

```bash
$ curl -X POST http://localhost:43101/stripe/v1/checkout/sessions
{"id":"stripe_checkout_session_000001","object":"checkout.session","payment_status":"paid"}
```

CLI がインストール済みの場合、同じリクエストと安全性レポートを整形表示できます。

```bash
$ mockport report --url http://localhost:43101/_mockport/report
Mockport Report

Mode: ai-safe
Safety: safe=true real-looking-secrets=0 external-urls=0
Public env safe-to-commit: true

Adapters:
- stripe enabled at /stripe maturity=workflow-compatible

Requests:
- #1 POST /stripe/v1/checkout/sessions -> 200
```

## 対応サービス

| Service | Adapter | Base path | Maturity | Supported workflows |
| --- | --- | --- | --- | --- |
| Stripe-like payments | `stripe` | `/stripe` と SDK-compatible な `/v1` alias | `workflow-compatible` | checkout sessions, payment intents, fake signed webhooks, …（[spec](docs/adapters/stripe.ja.md)） |
| OpenAI-compatible API | `openai` | `/openai` | `workflow-compatible` | models, chat completions, streaming, embeddings, …（[spec](docs/adapters/openai.ja.md)） |
| GitHub OAuth-like API | `github-oauth` | `/github` | `workflow-compatible` | authorize redirect, token exchange, user profile, …（[spec](docs/adapters/github-oauth.ja.md)） |
| Slack-like messaging API | `slack` | `/slack` | `workflow-compatible` | auth test, conversations, message post/update/delete, …（[spec](docs/adapters/slack.ja.md)） |
| LINE-like platform APIs | `line` | `/line` | `workflow-compatible` | Messaging API, LINE Login, LINE Pay, …（[spec](docs/adapters/line.ja.md)） |
| Zoho OAuth-like API | `zoho-oauth` | `/zoho` | `workflow-compatible` | authorize redirect, token exchange, user info, …（[spec](docs/adapters/zoho-oauth.ja.md)） |

計画中:

| Service | Planned adapter | Target workflows | Status |
| --- | --- | --- | --- |
| SendGrid-like email API | `sendgrid` | email send success/failure, auth error, rate limit, webhook events | Later candidate |

## SDK 接続

公式 SDK の向き先を provider から Mockport に差し替えます。[Examples](docs/site/examples.ja.md) と [OpenAI Chat example](examples/openai-chat/README.ja.md) を参照してください。

```javascript
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: "mockport_openai_key",
  baseURL: "http://localhost:43101/openai/v1",
});
```

Stripe 用アプリケーション env（fake 値を変更しない限り commit しても安全）:

```env
STRIPE_API_URL=http://localhost:43101/stripe
STRIPE_SECRET_KEY=mockport_stripe_secret
STRIPE_WEBHOOK_SECRET=whsec_mockport
```

詳細は [Public Env Safety](docs/public-env-safety.ja.md) を参照してください。

## AI-safe By Default

Mockport は、実在の credential らしい値や本番外部サービス URL を検出すると警告します。`strict` mode では unsafe configuration があると startup に失敗します。

```bash
mockport run --config examples/unsafe-config/mockport.yml --check
```

警告カテゴリ、strict mode、redaction の挙動は [AI-safe Development](docs/ai-safe-development.ja.md) を参照してください。

## レポートと互換性

各 run は `/_mockport/report` と `mockport report` で request history、scenario coverage、behavior matrix、safety summary を公開します。

[Reports](docs/site/reports.ja.md) と [Support matrix](docs/site/support-matrix.ja.md) を参照してください。

## ドキュメントと配布

docs、install 経路、release verification は [docs/site/](docs/site/index.ja.md) 配下にあります。現在の preview は `v0.1.0-alpha`（[Docker / GHCR](docs/site/distribution.ja.md)、[GitHub release archives](docs/site/distribution.ja.md)）。npm wrapper は experimental。Go binary と Docker が主経路です。

> **⚠️ アーカイブ**: 実装開始前(2026-05)の設計ドキュメントは [docs/archive/design/](docs/archive/design/README.ja.md) に保存されています。内容は保守されておらず、実装と乖離している可能性があります。

## コントリビュート

開発は spec-first TDD に従います。[CONTRIBUTING.ja.md](CONTRIBUTING.ja.md)、[Adapter onboarding guide](docs/adding-an-adapter.md)、[Maintainer Guide](docs/maintainer-guide.ja.md)、[Roadmap](ROADMAP.ja.md)、[Support Policy](docs/public-support-policy.ja.md) を参照してください。

Mockport は外部サービスの完全な clone ではありません。success、failure、auth error、rate limit、timeout、webhook/callback など、ローカル統合テストに必要な scenario に集中しています。

---

[Quickstart](docs/site/quickstart.ja.md) · [Docs](docs/site/index.ja.md) · [Examples](docs/site/examples.ja.md) · [Roadmap](ROADMAP.ja.md) · [English](README.md)
