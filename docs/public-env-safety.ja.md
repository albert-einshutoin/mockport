# Public Env Safety 日本語版

[English](public-env-safety.md)

Mockport が生成する env file は、fake value のままであれば commit 可能にすることを目的にしています。

## Safe values

- `mockport_` prefix の fake credential。
- `whsec_mockport` の fake webhook signing secret。
- `http://localhost:43101` など local-only URL。

## Unsafe values

- 実 provider API key、token、webhook secret。
- production provider URL。
- 顧客 payload や captured response。
- 将来 real value に置き換わりそうな曖昧 placeholder。

## Check

確認には `bash scripts/check-public-env.sh` を使います。この check は public docs、`.ja.md` 翻訳、GitHub template/workflow、example config、packaging docs、contract docs を対象に、real-looking provider credential、production provider URL、曖昧 placeholder を検出します。

Detector reference docs と intentionally unsafe な warning fixture だけは、狭い `mockport-public-safety` allow block で危険例を囲みます。

```md
<!-- mockport-public-safety: allow-begin detector-reference -->
unsafe detector example only
<!-- mockport-public-safety: allow-end -->
```

通常の setup 手順、generated example、quickstart には allow block を使いません。
