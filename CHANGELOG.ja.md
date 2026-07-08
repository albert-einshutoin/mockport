# Changelog 日本語版

[English](CHANGELOG.md)

この文書は Mockport の変更履歴を日本語で追うための入口です。詳細な release note、tag、差分の正準情報は英語版 `CHANGELOG.md` を参照してください。

## Unreleased

### Compatibility release track

- Stripe、OpenAI、GitHub OAuth、Slack の contract check 向けに scheduled/manual compatibility CI を追加。
- compatibility score、provider API version、SDK/client evidence、known gap を含む生成済み compatibility report を追加。
- maturity label `experimental`、`sdk-compatible`、`workflow-compatible`、`provider-compatible` 向けの release check を追加。

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
