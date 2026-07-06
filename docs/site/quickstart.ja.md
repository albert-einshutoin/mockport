# Quickstart

[English](quickstart.md)

Stripe 風 adapter を生成して、ローカルで起動します。

```bash
mockport init --adapter stripe
docker compose -f docker-compose.mockport.yml up
curl http://localhost:43101/health
```

複数 adapter をまとめて生成する場合:

```bash
mockport init --adapter stripe --adapter openai --adapter github-oauth --adapter slack --adapter line
docker compose -f docker-compose.mockport.yml up
```

`mockport init` は既存の生成ファイルを保護します。既存の `mockport.yml`、`.env.mockport`、`docker-compose.mockport.yml` を置き換える必要がある場合だけ `--force` を指定してください。

起動後は、`/_mockport/report` または `mockport report` で、実行された scenario と safety summary を確認できます。

## シナリオの切り替え

`mockport.yml` でシナリオを固定するほかに、リクエストごとに `X-Mockport-Scenario` ヘッダで切り替えられます（サーバー再起動不要）。

```bash
# Stripe の失敗系をテストする（サーバー再起動不要）
curl -X POST http://localhost:43101/stripe/v1/checkout/sessions \
  -H "X-Mockport-Scenario: payment_failed" \
  -H "Authorization: Bearer $STRIPE_KEY" \
  -d "mode=payment&success_url=http://localhost/success&cancel_url=http://localhost/cancel"
```

各アダプタの対応シナリオ一覧は [アダプタリファレンス](adapters.ja.md) を参照してください。
