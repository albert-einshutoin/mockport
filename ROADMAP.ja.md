# Roadmap 日本語版

[English](ROADMAP.md)

Mockport の roadmap は、Docker-first な local emulator から provider-compatible な selected workflow へ段階的に進めるための計画です。

現在の公開版は [`v0.3.0-preview`](docs/releases/v0.3.0-preview.ja.md) です。6 adapter の `workflow-compatible` の範囲を維持し、Stripe・OpenAI・Slack の選択したアプリフローと無変更の LLM 0.36 技術接続を含みます。

## 主要な方向性

- public preview では、Stripe/OpenAI/GitHub OAuth/Slack/LINE/Zoho OAuth などの workflow-compatible adapter を安定させます。
- 互換性 track では manifest、SDK contract、fixture、known-gap report を追加します。
- distribution は Docker と GitHub release archive を主経路にし、Homebrew と npm は補助経路として扱います。
- AI-safe mode と public env safety を継続的に強化します。

## 次の優先順位（2026-09-27時点）

1. P0はmainへ反映済みです。checkout imageの取り違え [#315](https://github.com/albert-einshutoin/mockport/issues/315) / [PR #384](https://github.com/albert-einshutoin/mockport/pull/384)、通常HTTPのfake-key認証 [#82](https://github.com/albert-einshutoin/mockport/issues/82) / [PR #385](https://github.com/albert-einshutoin/mockport/pull/385)、Responses `stream:true` の未対応明示 [#386](https://github.com/albert-einshutoin/mockport/issues/386) / [PR #387](https://github.com/albert-einshutoin/mockport/pull/387) を含みます。
2. P1は `v0.3.0-preview` に含まれます。[#86](https://github.com/albert-einshutoin/mockport/issues/86) の [Stripe PR #389](https://github.com/albert-einshutoin/mockport/pull/389) と [OpenAI PR #390](https://github.com/albert-einshutoin/mockport/pull/390) は、Checkout→署名付きWebhook→業務状態更新と、Python appのHTTP入口→公式SDK→応答・有限再試行・timeoutを同じMockportプロセスで再現します。
3. [#84](https://github.com/albert-einshutoin/mockport/issues/84) のSlack署名付きmessage event→同threadへのSDK返信と、[#395](https://github.com/albert-einshutoin/mockport/issues/395) の無変更LLM 0.36接続も同版に含まれます。[公開image試用](docs/site/app-trial.ja.md)は版付きimageとサンプルsourceを記録します。初見利用者の導入・継続利用は未検証です。

Block Kit、interactions、LINE追加機能、OpenAI追加APIは利用者フローを示してから優先します。Responses SSE本体は利用アプリと最小event契約を定めた後に判断します。SendGridは具体的需要、限定フロー、公式SDK等の検証方法、保守見通しが揃うまで着手しません。初見利用者の試行は、今回のローカル・CI再現後の製品検証です。

将来候補の一覧、優先度 tier、adapter family 戦略、探索的な実装順は [Adapter Candidate Priorities（英語）](docs/planning/adapter-candidate-priorities.md) にまとめています。この資料は計画検討用であり、実装確約や現行サポート範囲を示すものではありません。

詳細な milestone と順序は英語版を正とします。
