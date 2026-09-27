# Distribution 日本語版

[English](distribution.md)

Mockport は Docker-first で配布します。public preview では Docker image と GitHub release archive が主経路です。

現在の preview は `v0.3.0-preview` です。Docker image は
`ghcr.io/albert-einshutoin/mockport:0.3.0-preview` を使用してください。

Docker と Git を用意し、公開タグから設定ファイルを取得します。基本起動にはローカルのGo・Pythonは不要です。

```bash
git clone --branch v0.3.0-preview --depth 1 https://github.com/albert-einshutoin/mockport.git
cd mockport
docker pull ghcr.io/albert-einshutoin/mockport:0.3.0-preview
docker run --rm -p 127.0.0.1:43101:43101 \
  -v "$(pwd)/configs/mockport.example.yml:/etc/mockport/mockport.yml:ro" \
  ghcr.io/albert-einshutoin/mockport:0.3.0-preview run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

別ターミナルで `curl http://localhost:43101/health` と `curl http://localhost:43101/_mockport/report` を実行できます。macOS arm64のarchive取得・checksum確認例と4種類の配布名は[英語版](distribution.md)にあります。checksumとbinary版の一括確認には、公開タグのcheckoutでGitHub CLI・Go 1.26.8を用意し、`gh release download v0.3.0-preview -D "$tmpdir"` の後に `scripts/verify-release-artifacts.sh 0.3.0-preview "$tmpdir"` を実行します。`tmpdir="$(mktemp -d)"` を先に設定してください。

`latest`はmainに追従して変わります。再現には版付きtagと公開registry digest `ghcr.io/albert-einshutoin/mockport@sha256:497f6ec7dc6f1fea8afd292f9480872983bc940425a090aa8ad8580350d8e9a5`を使用してください。source、image ID、実行platformの対応は[公開記録](../releases/v0.3.0-preview.ja.md)、3アプリとLLMの手順は[公開アプリ試用](app-trial.ja.md)を参照してください。

## 経路

- Docker/GHCR: preview image を取得してすぐに起動できます。
- GitHub release archives: OS/arch 別 binary と checksum を確認できます。
- Homebrew/npm: template や wrapper はありますが、主経路ではありません。
