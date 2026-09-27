# Quickstart

[English](quickstart.md)

Stripe 風 adapter を生成して、ローカルで起動します。

Dockerでの初回起動はDockerとGitを用意し、[配布案内](distribution.ja.md)の公開タグ`v0.3.0-preview`から設定を取得します。以下の`mockport init`は公開archiveから取得したbinaryをPATHに置いた後の手順です。

```bash
mockport init --adapter stripe
docker compose -f docker-compose.mockport.yml up
curl http://localhost:43101/health
```

`mockport init` が生成する `docker-compose.mockport.yml` も同じ意図です。ホスト側では `127.0.0.1:43101` のみにポートを公開し、コンテナ内のプロセスは `--host 0.0.0.0` で全インターフェースを listen します。どちらか一方ではなく、意図的に組み合わせた設定です。

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
  -H "Authorization: Bearer mockport_stripe_secret" \
  -d 'mode=payment' -d 'success_url=http://localhost/success' -d 'cancel_url=http://localhost/cancel' \
  -d 'line_items[0][price_data][currency]=usd' -d 'line_items[0][price_data][unit_amount]=1200' \
  -d 'line_items[0][price_data][product_data][name]=Mockport item' -d 'line_items[0][quantity]=1'
```

各アダプタの対応シナリオ一覧は [アダプタリファレンス](adapters.ja.md) を参照してください。
