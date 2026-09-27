# Slack Adapter 日本語版

[English](slack.md)

共通ルール: [シナリオポリシー](../scenario-policy.ja.md)。

Slack adapter は、messaging workflow と Events API の selected subset を local で検証するための adapter です。

## 対応範囲

- `auth.test`、conversation list/history、message post/update/delete。
- URL verification、message callback subset、固定 message event のローカル署名付き配送。
- `chat.postMessage` の `thread_ts` を応答と保存状態に保持し、通常の channel history から thread 返信を除外します。
- 実 workspace 配送、親 message の取得、thread 全体、`conversations.replies`、Block Kit validation、files、enterprise policy、workspace directory は対象外です。

## 選択フローと設定移行

`fake_secret` は Web API token、`webhook.signing_secret` は Events API の署名鍵です。従来 `fake_secret` を署名鍵としていた設定では、別の偽値を `webhook.signing_secret` に明示してください。署名鍵がない `/slack/events` は `503 missing_signing_secret`、送信 helper は `missing_signing_secret` で失敗します。暗黙の fallback はありません。

loopback からの `POST /slack/test/event/send` は、設定済みの安全なローカル `webhook.target_url` に [`events_message_callback.json`](../../compat/fixtures/slack/events_message_callback.json) の event を送ります。送信時の Unix 秒と raw body で Slack v0 署名を作ります。対象は `message_success` の通常 message 一種です。非 2xx は配送先 status 付き 502、timeout は 504 です。

[Python app 例](../../examples/app-e2e/slack-app/server.py) は公式 `SignatureVerifier` で raw body と headers を検証し、固定文を `slack-sdk==3.44.1` の `WebClient.chat_postMessage` から event の channel と ts の thread へ返信します。同期 SDK 呼び出しは 1 秒 timeout です。E2E は SDK 応答と report の 1 要求を、adapter 結合テストは handler 経由で保存した返信を確認します。Slack の受信 2xx は 3 秒以内の選択正常ケースで確認しますが、高負荷時の配送保証は扱いません。

公式根拠: [Events API](https://docs.slack.dev/apis/events-api/)、[署名](https://docs.slack.dev/authentication/verifying-requests-from-slack/)、[SignatureVerifier](https://docs.slack.dev/tools/python-slack-sdk/reference/signature/index.html)、[chat.postMessage](https://docs.slack.dev/reference/methods/chat.postMessage/)。

詳細な endpoint と error model は英語版を正とします。
