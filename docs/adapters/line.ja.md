# LINE Adapter 日本語版

[English](line.md)

LINE adapter は、Messaging API、LINE Login、LIFF helper、MINI App service message、LINE Pay、Mini Dapp helper の local workflow を扱います。本ドキュメントは LINE 公式プラットフォーム仕様の写しではなく、完全な LINE プラットフォーム互換を主張しません。

`line` adapter は **workflow-compatible**（アプリ統合テストに必要な workflow を local で再現可能）であり、**provider-compatible**（本番 LINE プラットフォームと同等の挙動・検証・運用境界）ではありません。

## 対応範囲

- message send、content、signed webhook、rich menu、channel token workflow。
- OAuth code/token/profile と local profile lookup。LINE Login flow では authorize と token exchange の `client_id` を必須とし、token exchange の値は code 発行時と一致する必要があります。
- LIFF local helper、MINI App service message、LINE Pay v3-like payment、Mini Dapp SDK-like wallet/payment helper。

endpoint の詳細、request/response shape、official reference map は英語版 [`line.md`](line.md) を正とします。

## 認証方式

公式 LINE Login overview が区別する auto login、email/password login、QR code login、SSO login といった UI レベルの認証方式は、Mockport では再現しません。local adapter は、ユーザー認証が成功した後に得られる authorization code flow を模擬します。

二要素認証も provider 側の policy 境界として扱います。Mockport は verification code 画面、trusted browser の有効期間、アカウント切り替え、channel console 設定をエミュレートしません。

## Scenarios

| Scenario | 挙動 |
| --- | --- |
| `line_success` | 既定の成功 local workflow。 |
| `auth_error` | token 保護 endpoint で認証失敗を返す。 |
| `rate_limited` | Messaging API-like send で rate limit 挙動を返す。 |
| `invalid_request` | request validation 風の失敗を返す。 |
| `pay_failed` | LINE Pay または Mini Dapp payment の失敗挙動を返す。 |

See also: [Scenario Policy 日本語版](../scenario-policy.ja.md).

## State

adapter は次の local deterministic state を保持します。

| Resource | 用途 |
| --- | --- |
| `oauth_code` | `/oauth2/v2.1/authorize` が発行する authorization code。 |
| `oauth_token` | `/oauth2/v2.1/token` が発行する access token。 |
| `message` | Messaging API-like の送信 message 記録。 |
| `rich_menu` | rich menu 定義と image upload 状態。 |
| `rich_menu_alias` | rich menu alias マッピング。 |
| `user_rich_menu` | ユーザー単位の rich menu link。 |
| `notification_token` | MINI App service message の notification token。 |
| `line_pay_payment` | LINE Pay-like の payment reservation と confirmation。 |
| `mini_dapp_payment` | Mini Dapp-like の local payment 記録。 |

ID はプロセス内で deterministic であり、Mockport プロセス再起動時にリセットされます。

## Known Gaps

`line` adapter は **workflow-compatible** であり、**provider-compatible** ではありません。

既知の gap:

- 実 LINE Login UI、QR code login、auto login、SSO login、二要素認証画面はない。
- 署名済みまたは provider で検証可能な ID token はない。
- Mockport が OpenID Connect discovery endpoint を公開しない。
- 実 LINE SDK contract harness は未対応。
- 実 LIFF browser runtime はない。
- provider 主導の webhook redelivery、retry scheduler、完全な webhook event catalog はない。local helper は要求に応じて signed webhook payload を送信できる。outbound 配送は固定 `5s` timeout で、timeout は `504` + `failure: "timeout"`、non-2xx target は `502` + `failure: "target_non_2xx"` と `target_status_code` を返す。
- monthly quota、free-message、rate-limit bucket、concurrent audience operation の enforcement は、deterministic scenario を超えてない。
- すべての message、Flex、template、action、audience、insight、coupon、membership、rich menu フィールドに対する完全な Messaging API schema validation はない。
- 実 media storage lifecycle はない。content と preview endpoint は local placeholder bytes を返す。
- LINE Developers Console の channel 設定や review workflow はない。
- regional policy enforcement はない。
- Mini Dapp endpoint は local SDK-style helper であり、Dapp Portal の完全 clone ではない。

## 検証

adapter と package test:

```bash
go test ./adapters/line ./internal/server ./internal/cli ./internal/config
```

public trust gate:

```bash
bash scripts/check-public-trust.sh
```

互換性ステータスと known gap は [`docs/site/support-matrix.md`](../site/support-matrix.md) を正とします。

SDK contract harness は LINE 未対応です。`bash scripts/run-sdk-contracts.sh line` は `unsupported provider: line` で終了します。LINE smoke test が [`contract/sdk/README.md`](../../contract/sdk/README.md) に追加されるまで、上記 adapter test を使用してください。
