# Zoho OAuth Adapter 日本語版

[English](zoho-oauth.md)

共通ルール: [シナリオポリシー](../scenario-policy.ja.md)。

Zoho OAuth adapter は、Zoho OAuth2 の authorization-code flow を最小限・deterministic に再現する adapter です。アプリが `ZOHO_AUTH_BASE_URL` をこの Mockport の base path に向けるだけで、実際の Zoho に接続せずローカルでログインを完結できます。Zoho の全 API は対象外です。

## 対応範囲

Zoho OAuth client が実際に呼ぶ 3 endpoint のみを mock します。

- 認可 redirect（ログイン画面は出さず即 redirect）。
- authorization code の token exchange。
- Zoho 固有の auth scheme を使う user info 取得。

## Base Path

既定の base path（アプリが `ZOHO_AUTH_BASE_URL` に設定する値）:

```text
/zoho
```

設定例:

```yaml
adapters:
  zoho-oauth:
    enabled: true
    base_path: /zoho
    scenario: oauth_success
    fake_secret: mockport_zoho_secret
```

## Supported Endpoints

`base` は設定された base path（既定 `/zoho`）です。

| Method | Path | 用途 |
| --- | --- | --- |
| `GET` | `{base}/oauth/v2/auth` | fake authorization code を発行し、`redirect_uri` へ 302 redirect（`code` 付与・`state` echo）。 |
| `POST` | `{base}/oauth/v2/token` | authorization code を access token に交換。 |
| `GET` | `{base}/oauth/user/info` | deterministic な user info を返す。`Zoho-oauthtoken` scheme が必須。 |
| `POST` | `{base}/test/reset` | test isolation 用に local OAuth state を消去（loopback からのみ）。 |

挙動の要点:

- **authorize**: `client_id` と loopback の `redirect_uri` が必須。ログイン画面は出さず、`redirect_uri` へ `302` し、生成した `code` と request の `state` を付与します。
- **token**（`application/x-www-form-urlencoded`）: 成功時は `200` で `{"access_token":"<token>"}`。失敗時（`grant_type` 不正、`code` が不正・未知）は `{"error":"<reason>"}`。Zoho の挙動に合わせ token 交換失敗も HTTP `200` で返し、client は status code ではなく `error` field を見ます。authorization code は one-time use かつ client-bound です。
- **user info**: `Authorization: Zoho-oauthtoken <access_token>`（`Bearer` ではない）が必須。成功時は `200` で `{"Email":"<email>","Display_Name":"<name>"}`（大文字始まりの key）。token 欠落・未知 token・`Zoho-oauthtoken` 以外の scheme は `401`。

## 返す User の設定

返す `Email` / `Display_Name` は差し替え可能です。

- 既定値は環境変数 `ZOHO_USER_EMAIL` / `ZOHO_USER_NAME`（未設定時は `mockport@example.test` / `Mockport User`）。
- authorize の query parameter `mock_email` / `mock_name` で、1 回の flow だけ上書きできます。上書きは発行した code に紐づき、後段の user info に反映されます。

## Scenarios

| Scenario | 挙動 |
| --- | --- |
| `oauth_success` | 既定の成功 workflow。 |
| `invalid_code` | token 交換失敗（不正・未知 code）を強制。 |
| `invalid_token` | user info の認証失敗（`401`）を強制。 |

詳細な known gap は英語版を正とします。
