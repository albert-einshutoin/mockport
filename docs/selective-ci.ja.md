# 選択的CIテスト実行

Mockport の Pull Request CI は、失敗検出能力を落とさずフィードバックを短縮するため、
安全側の変更影響分析を使います。`main` と `release/**` への push、手動実行、日次の定期実行では、
引き続き全テストを実行します。

## 安全モデル

判定器は fail-closed を原則とします。

1. 現在のheadと対象ブランチ最新状態のmerge baseを解決する
2. Gitから追加・変更・削除・コピー・renameを取得する
3. `ci/config.json` に集約した完全検証ルールを適用する
4. Goアダプターが `go list -json ./...` のimportグラフから変更パッケージとリポジトリ内の全逆依存を選ぶ
5. 対応する結合・E2E検証と、常時実行するhealthスモークテストを追加する
6. 入力が不足・未対応・未分類の場合は全テストへフォールバックする

削除済みファイルは分類には利用しますが、ファイルの存在を前提とするツールには渡しません。
変更パスをshellコマンドへ展開せず、runnerは設定済みの信頼できるargv配列だけを実行します。

共通契約、設定、routing、security、state、build、dependency、test、container、CIの変更は、
意図的に全テスト対象とします。影響分析器自体の変更も全テスト対象です。

## 構成

- `ci/impact.py`: CIサービス非依存の差分取得、分類、計画、JSON出力、実行、計測、fallback制御
- `ci/adapters/go.py`: Go project検出、package graph解析、逆依存探索
- `ci/config.json`: 完全検証ルール、path mapping、smoke test、信頼済みcommand定義
- `ci/run-full.sh`: 判定器や設定を読み込めない場合にも使える独立した完全検証レーン

共通判定器にはGo固有の選択処理を含めていません。他言語は、repository pathを受け取り、
同じJSON契約でproject、影響module、unit test targetを返すadapterとして追加できます。

## ローカル検証

```sh
make ci-impact-test
python3 ci/impact.py plan --base origin/main --head HEAD --output ci-plan.json
python3 ci/impact.py run --plan ci-plan.json
```

merge後・release前と同じ完全検証を行う場合は、次を実行します。

```sh
make ci-full
```

`ci-plan.json` にはrevision、変更ファイル、検出・影響project、影響module、
unit/integration/E2E/smoke target、strategy、fallback理由を記録します。CIでは14日間artifactとして保存し、
選択判断を監査可能にします。

pathやproject形式を追加するときは、最初に判定器またはadapterの失敗テストを追加してください。
明示的に理解できないpathは、完全検証fallbackのままにします。
