# digest 固定のアプリ試用

[English](app-trial.md)

公開 Mockport image に対してサンプルアプリを動かす導入手順です。初見利用者による別アプリ試用では、協力者自身のアプリの 1 フローに置き換えます。実 token・Slack workspace・決済・AI 推論は不要です。

## 検証済みの組み合わせ

| フロー | Mockport source commit | registry image | サンプル commit | platform | 結果 |
| --- | --- | --- | --- | --- | --- |
| P0/P1 Stripe + OpenAI | `67a81591e48526c88b96381cd56c54386d93341a` | `ghcr.io/albert-einshutoin/mockport@sha256:eae6ba56f01cc9969038da36982e6699732379028f27313ba86dba5d1e43b8fb` | `f6fe7d573ee37b282767aa112d30615be1941298` | `linux/arm64` | 新規ディレクトリの Compose runner exit 0、2フロー成功 |
| #84 Stripe + OpenAI + Slack | `01ab1614e40e0d92122123639eba799633e3e6c1` | `ghcr.io/albert-einshutoin/mockport@sha256:534cfaa092373cf1a307a9a140415a015b414cad8301a24e0b546b9ed46dbb04` | `f6fe7d573ee37b282767aa112d30615be1941298` | `linux/arm64` | 新規ディレクトリの Compose runner exit 0、3フロー成功 |

版付き `0.2.0-preview` は従来版です。可変の `latest` だけでは版を再現できません。サンプル Dockerfile は build 時に固定した SDK 依存を取得するため、Mockport の digest に加えてサンプル commit と lock file を記録します。

P0/P1 は新規一時ディレクトリに固定 commit をローカル clone し、GHCR から公開 digest を pull して配布用 Compose で実行しました。実行 container の image ID は digest と一致し、`down --remove-orphans` 後に対象 container は残りませんでした。ローカル clone 0.76 秒、Compose 7.02 秒ですが、base image と依存 layer はキャッシュ済みです。前提導入・新規依存取得を含む cold 時間として扱いません。

#84 は新規ディレクトリへ固定 sample commit を GitHub から clone しました。clone 0.94 秒、sample app image の `docker compose build --no-cache` 8.52 秒、公開 Mockport digest の pull 3.21 秒、3フローの `up --no-build` 5.43 秒でした。別の新規ディレクトリで下記と同じ `up --build` も実行し、依存 cache 済みで 6.77 秒でした。Docker・Git・base image・OS・registry 側 cache は既存のため、完全な新規マシンの cold 導入時間ではありません。実行 Mockport image ID は固定 digest と一致し、対象 container は後始末済みです。

## 新規作業ディレクトリから実行

前提は Docker Compose、Git、GHCR と依存 package registry への接続、空いているローカル port 43101/33001/33002/33003 です。不足する前提の導入前から時間を計測します。表の行に対応する commit と digest を使用します。

```sh
git clone https://github.com/albert-einshutoin/mockport.git
cd mockport
git checkout f6fe7d573ee37b282767aa112d30615be1941298
export MOCKPORT_IMAGE='ghcr.io/albert-einshutoin/mockport@sha256:534cfaa092373cf1a307a9a140415a015b414cad8301a24e0b546b9ed46dbb04'
export PUBLISHED_FLOWS=all
docker pull "$MOCKPORT_IMAGE"
docker image inspect "$MOCKPORT_IMAGE" --format '{{.Id}} {{.Os}}/{{.Architecture}}'
docker compose -f examples/app-e2e/compose.published.yml -p mockport-published up --build --abort-on-container-exit --exit-code-from runner
docker compose -f examples/app-e2e/compose.published.yml -p mockport-published down --remove-orphans
```

配布用 Compose には Mockport の `build:` がなく、`MOCKPORT_IMAGE` の digest を使います。runner の正常終了時は Stripe の注文、OpenAI の stream、Slack の SDK 返信が出ます。終了 code が非 0 なら成功と扱いません。`git rev-parse HEAD`、digest、`docker image inspect` の image ID/platform、実行 command と結果を記録し、失敗時も `down` で片付けます。従来の P0/P1 image だけを試す場合は表の1行目の digest と `PUBLISHED_FLOWS=p0p1` を使用します。その image に #84 は含まれません。

## 初見利用者・別アプリ試用の記録票

利用者自身の通常フローを一つ選び、接続先と偽 credential を Mockport に向けます。Slack なら token と署名鍵を別の偽値にして、選択 message event を送り、受信の成功と同じ channel/thread への SDK 返信を確認します。代表的な失敗として署名後の本文改ざんか古い timestamp を送り、返信が増えず拒否されることを確かめます。Stripe/OpenAI の失敗例は [app E2E](../../examples/app-e2e/README.md) を参照します。

| 記録項目 | 観測結果 |
| --- | --- |
| 接続した別アプリと選択フロー |  |
| 環境準備開始→最初のアプリ成功、前提導入・依存取得・実行の内訳 |  |
| 接続先と偽 credential 以外のコード変更 |  |
| 文書だけで進めた箇所、手助け・手戻りが必要だった箇所 |  |
| 期待した失敗を認識し、原因を特定できたか |  |

各段階の開始・終了を壁時計で記録します。キャッシュ後の再実行は cold 導入時間ではありません。Codex がサンプルを別ディレクトリで動かしても初見利用者・別アプリの成功とは数えません。導入の阻害要因が見つかれば次の API 拡張より先に修正します。
