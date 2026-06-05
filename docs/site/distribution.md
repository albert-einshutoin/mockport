# Distribution

[日本語版](distribution.ja.md)

Primary distribution paths:

| Path | Status |
| --- | --- |
| Docker image | Preview via GHCR |
| Release binary archives | Preview via GitHub Releases |
| Homebrew | Not published; template only |
| npm | Not published; experimental wrapper only |

## Public Preview

Current preview version: `v0.1.0-alpha`.

Docker:

```bash
docker pull ghcr.io/albert-einshutoin/mockport:0.1.0-alpha
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/configs/mockport.example.yml:/etc/mockport/mockport.yml \
  ghcr.io/albert-einshutoin/mockport:0.1.0-alpha run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

Release archives:

```bash
curl -LO https://github.com/albert-einshutoin/mockport/releases/download/v0.1.0-alpha/mockport_0.1.0-alpha_darwin_arm64.tar.gz
curl -LO https://github.com/albert-einshutoin/mockport/releases/download/v0.1.0-alpha/checksums.txt
grep 'mockport_0.1.0-alpha_darwin_arm64.tar.gz' checksums.txt | sed 's# dist/# #' | shasum -a 256 -c -
tar -xzf mockport_0.1.0-alpha_darwin_arm64.tar.gz
./mockport_0.1.0-alpha_darwin_arm64/mockport version
```

Use the explicit `0.1.0-alpha` image tag for preview installs. The `latest` tag follows the default branch image and is not the preview release contract.

Local release archive check:

```bash
scripts/test-release-archives.sh
```

Published release verification:

```bash
tmpdir="$(mktemp -d)"
gh release download v0.1.0-alpha -D "$tmpdir"
scripts/verify-release-artifacts.sh 0.1.0-alpha "$tmpdir" ghcr.io/albert-einshutoin/mockport:0.1.0-alpha
```
