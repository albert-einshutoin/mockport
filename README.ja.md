# Mockport

AI ネイティブ開発のための、シークレット不要なサービスエミュレーション。

[English](README.md)

Mockport は、外部サービス連携をローカル Docker 環境で再現するためのエミュレーターです。ローカル開発、CI、AI コーディングワークフローで、実サービスの API キーや Webhook secret を使わずに統合パスを検証できます。

公開プレビューの範囲、対応状況、サンプル、制限は [Mockport Docs 日本語版](docs/site/index.ja.md) を参照してください。

プロジェクト計画とメンテナンス:

- [Roadmap](ROADMAP.ja.md)
- [Maintainer Guide](docs/maintainer-guide.ja.md)
- [Support Policy](docs/public-support-policy.ja.md)

## Quickstart

ローカルインストールなしで試せます。

```bash
docker build -t mockport:local -f docker/Dockerfile .
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/configs/mockport.example.yml:/etc/mockport/mockport.yml \
  mockport:local run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

ローカル API を確認します。

```bash
curl http://localhost:43101/health
curl -X POST http://localhost:43101/stripe/v1/checkout/sessions
curl http://localhost:43101/_mockport/report
```

ビルド済みバイナリからの CLI ワークフロー:

```bash
mockport init --adapter stripe
docker compose -f docker-compose.mockport.yml up
```

`mockport init` は、既存の生成ファイルをデフォルトで保護します。`mockport.yml`、`.env.mockport`、`docker-compose.mockport.yml` を意図的に置き換える場合だけ `--force` を使ってください。

アプリケーション側の `.env` 例:

```env
STRIPE_API_URL=http://localhost:43101/stripe
STRIPE_SECRET_KEY=mockport_stripe_secret
STRIPE_WEBHOOK_SECRET=whsec_mockport
```

生成された fake 値を変更しない限り、この Mockport 用 env は commit しても安全です。詳細は [Public Env Safety](docs/public-env-safety.ja.md) を参照してください。

Stripe 風 checkout session を試します。

```bash
curl -X POST http://localhost:43101/stripe/v1/checkout/sessions
```

リクエストと安全性レポートを確認します。

```bash
mockport report
```

## Source Smoke Test

このリポジトリの checkout から実行します。

```bash
bash scripts/smoke-empty-dir.sh
```

この smoke test は、ローカル Docker image をビルドし、一時的な空ディレクトリで `mockport init --adapter stripe` を実行し、Docker Compose を起動して `/health`、Stripe 風 checkout request、`mockport report` を確認します。

## Docker

```bash
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/configs/mockport.example.yml:/etc/mockport/mockport.yml \
  mockport:local run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

## Install And Distribution

Mockport は Docker-first です。最初の public preview は `v0.1.0-alpha` です。

Docker preview image:

```bash
docker pull ghcr.io/albert-einshutoin/mockport:0.1.0-alpha
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/configs/mockport.example.yml:/etc/mockport/mockport.yml \
  ghcr.io/albert-einshutoin/mockport:0.1.0-alpha run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

Release archives:

```bash
curl -LO https://github.com/albert-einshutoin/mockport/releases/download/v0.1.0-alpha/mockport_0.1.0-alpha_darwin_arm64.tar.gz
curl -LO https://github.com/albert-einshutoin/mockport/releases/download/v0.1.0-alpha/checksums.txt
grep 'mockport_0.1.0-alpha_darwin_arm64.tar.gz' checksums.txt | sed 's# dist/# #' | shasum -a 256 -c -
tar -xzf mockport_0.1.0-alpha_darwin_arm64.tar.gz
./mockport_0.1.0-alpha_darwin_arm64/mockport version
```

| Channel | Status | Notes |
| --- | --- | --- |
| Docker / GHCR | Preview | `ghcr.io/albert-einshutoin/mockport:0.1.0-alpha`; `latest` は default branch に追従し、preview release contract ではありません |
| GitHub release archives | Preview | `mockport_<version>_<os>_<arch>.tar.gz` と `checksums.txt` |
| Homebrew | Not published | Formula template は `packaging/homebrew/` 配下 |
| npm | Not published | npm wrapper は experimental。Go binary と Docker が主経路です |

Docs site source は `docs/site/` 配下です。

## Services

対応済み:

| Service | Adapter | Base path | Maturity | Supported workflows |
| --- | --- | --- | --- | --- |
| Stripe-like payments | `stripe` | `/stripe` と SDK-compatible な `/v1` alias | `workflow-compatible` | checkout sessions, payment intents, customers, products, prices, subscriptions, invoices, refunds, fake signed webhooks, SDK contract, state, validation, idempotency |
| OpenAI-compatible API | `openai` | `/openai` | `workflow-compatible` | models, chat completions, responses, streaming, embeddings, files, batches, SDK contract, state, validation |
| GitHub OAuth-like API | `github-oauth` | `/github` | `workflow-compatible` | authorize redirect, access token exchange, user profile, user emails, user orgs, client contract, state, scope validation |
| Slack-like messaging API | `slack` | `/slack` | `workflow-compatible` | auth test, conversations list/history, message post/update/delete, Events API URL verification/message callback subset, client contract, state, Slack-style errors |
| LINE-like platform APIs | `line` | `/line` | `workflow-compatible` | Messaging API send/content/signed webhook/rich menu/channel token workflows, LINE Login OAuth/profile, LIFF helpers, MINI App service messages, LINE Pay v3 request/confirm, Mini Dapp wallet/payment helpers |

計画中:

| Service | Planned adapter | Target workflows | Status |
| --- | --- | --- | --- |
| SendGrid-like email API | `sendgrid` | email send success/failure, auth error, rate limit, webhook events | Later candidate |

## AI-safe By Default

Mockport は、実在の credential らしい値や本番外部サービス URL を検出すると警告します。`strict` mode では unsafe configuration があると startup に失敗します。

サーバーを起動せずに config を確認できます。

```bash
mockport run --config examples/unsafe-config/mockport.yml --check
```

警告カテゴリ、strict mode、redaction の挙動は [AI-safe Development](docs/ai-safe-development.ja.md) を参照してください。

Mockport は外部サービスの完全な clone ではありません。success、failure、auth error、rate limit、timeout、webhook/callback など、ローカル統合テストに必要な scenario に集中しています。
