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
