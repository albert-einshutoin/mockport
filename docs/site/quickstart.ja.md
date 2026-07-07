# Quickstart

[English](quickstart.md)

Stripe 風 adapter を生成して、ローカルで起動します。

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

### 既知の UX ギャップ

この Quickstart では `mockport up` ではなく、明示的な `docker compose -f docker-compose.mockport.yml up` を使っています。`mockport up` は利用可能ですが、Docker Compose まわりの UX 改善（エラー表示の明確化や `--detach` / `--build` の扱い）は [public preview follow-up #8](https://github.com/albert-einshutoin/mockport/issues/8) で別途追跡しています。一覧は [ロードマップ](../../ROADMAP.md#public-preview-follow-up) を参照してください。

起動後は、`/_mockport/report` または `mockport report` で、実行された scenario と safety summary を確認できます。
