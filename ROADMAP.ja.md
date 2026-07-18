# Roadmap 日本語版

[English](ROADMAP.md)

Mockport の roadmap は、Docker-first な local emulator から provider-compatible な selected workflow へ段階的に進めるための計画です。

## 主要な方向性

- public preview では、Stripe/OpenAI/GitHub OAuth/Slack/LINE/Zoho OAuth などの workflow-compatible adapter を安定させます。
- 互換性 track では manifest、SDK contract、fixture、known-gap report を追加します。
- distribution は Docker と GitHub release archive を主経路にし、Homebrew と npm は補助経路として扱います。
- AI-safe mode と public env safety を継続的に強化します。

将来候補の一覧、優先度 tier、adapter family 戦略、探索的な実装順は [Adapter Candidate Priorities（英語）](docs/planning/adapter-candidate-priorities.md) にまとめています。この資料は計画検討用であり、実装確約や現行サポート範囲を示すものではありません。

詳細な milestone と順序は英語版を正とします。
