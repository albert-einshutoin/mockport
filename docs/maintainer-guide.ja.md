# Maintainer Guide 日本語版

[English](maintainer-guide.md)

maintainer は、public preview の信頼性を保つために docs、tests、compatibility evidence、release artifacts を一貫して扱います。

## 運用方針

- adapter behavior を変更したら docs、fixtures、reports、tests を同時に確認します。
- release 前に public env safety と AI-safe warning を確認します。
- issue/PR では supported scope と known gap を明確にします。
