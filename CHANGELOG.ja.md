# Changelog 日本語版

[English](CHANGELOG.md)

この文書は Mockport の変更履歴を日本語で追うための入口です。詳細な release note、tag、差分の正準情報は英語版 `CHANGELOG.md` を参照してください。

## Unreleased

## v0.3.0-preview - 2026-09-27

- Stripe・OpenAIの通常リクエストに対して、設定時に偽キーを検証する機能を追加。未対応のOpenAI Responses `stream:true` は501を返します。
- 有効なinline明細でのStripe Checkout、作成時未払いSession、署名付き完了Webhook、注文の一回更新をアプリで確認します。
- OpenAIの通常応答・Chat streaming・有限の失敗制御と、無変更のLLM 0.36への技術接続を追加。模擬usageは固定値です。
- 署名付きSlack message event、公式SDKでの署名検証、同threadへの返信を確認します。
- Docker・binary・CLI init・実験的npm wrapperの版を揃えました。移行と制限は[日本語release notes](docs/releases/v0.3.0-preview.ja.md)を参照してください。

## v0.2.0-preview - 2026-07-19

### Compatibility release track

- Stripe、OpenAI、GitHub OAuth、Slack、公式 LINE SDK の contract check 向けに scheduled/manual compatibility CI を追加。
- compatibility score、provider API version、SDK/client evidence、known gap を含む生成済み compatibility report を追加。
- maturity label `experimental`、`sdk-compatible`、`workflow-compatible`、`provider-compatible` 向けの release check を追加。
- v0.1.0-alpha の scope と比べ、現行 mainline では6つの built-in adapter（Stripe、OpenAI、GitHub OAuth、Slack、LINE、Zoho OAuth）すべてを、文書化された selected workflow 向けの `workflow-compatible` として分類。`provider-compatible` parity を主張するものではない。
- deterministic state、reset、bounded request history、AI-safe summary、実行可能な Node SDK example、`llms.txt` を追加。

## v0.1.0-alpha - 2026-05-26

最初の公開 preview release。

### Included

- AI-safe な設定チェックを備えた Docker-first Mockport runtime。
- checkout session、payment intent、webhook 送信、一般的な error scenario 向けの Stripe-like payment adapter。
- Experimental な OpenAI-compatible、GitHub OAuth-like、Slack-like adapter。
- request history、scenario coverage、behavior matrix、safety findings 向けの `/_mockport/report`。
- Linux と macOS の amd64/arm64 向け GitHub Release archive。
- `ghcr.io/albert-einshutoin/mockport:0.1.0-alpha` として公開された GHCR image。

### Known Limits

- scenario-compatible であり、full provider-compatible ではない。
- provider SDK contract coverage は後続 Phase で開始。
- Homebrew と npm はまだ公開 distribution channel ではない。
