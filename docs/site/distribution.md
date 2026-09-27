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

Current preview version: `v0.3.0-preview`.

Prerequisites for the container path: Docker and Git. The repository tag supplies the configuration file; the GHCR image supplies the executable. No local Go or Python installation is needed for this basic run.

Docker:

```bash
git clone --branch v0.3.0-preview --depth 1 https://github.com/albert-einshutoin/mockport.git
cd mockport
docker pull ghcr.io/albert-einshutoin/mockport:0.3.0-preview
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/configs/mockport.example.yml:/etc/mockport/mockport.yml \
  ghcr.io/albert-einshutoin/mockport:0.3.0-preview run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

Release archives:

The example below is for macOS arm64. The Release also contains `darwin_amd64`, `linux_amd64`, and `linux_arm64`; choose the archive for the host. `curl`, `shasum`, and `tar` are needed for this path.

```bash
curl -LO https://github.com/albert-einshutoin/mockport/releases/download/v0.3.0-preview/mockport_0.3.0-preview_darwin_arm64.tar.gz
curl -LO https://github.com/albert-einshutoin/mockport/releases/download/v0.3.0-preview/checksums.txt
grep 'mockport_0.3.0-preview_darwin_arm64.tar.gz' checksums.txt | sed 's# dist/# #' | shasum -a 256 -c -
tar -xzf mockport_0.3.0-preview_darwin_arm64.tar.gz
./mockport_0.3.0-preview_darwin_arm64/mockport version
```

Use the explicit `0.3.0-preview` image tag for preview installs. The `latest` tag follows the default branch image and is not the preview release contract.

The [release record](../releases/v0.3.0-preview.md) and [app trial](app-trial.md) record the registry digest after publication. Use the digest rather than `latest` for repeated trials. A registry digest and the local platform image ID are different identifiers; record both.

Local release archive check:

```bash
scripts/test-release-archives.sh
```

Published release verification:

From the tagged checkout, install the GitHub CLI (`gh`) and Go 1.26.8 if using the existing verification script, then run:

```bash
tmpdir="$(mktemp -d)"
gh release download v0.3.0-preview -D "$tmpdir"
scripts/verify-release-artifacts.sh 0.3.0-preview "$tmpdir" ghcr.io/albert-einshutoin/mockport:0.3.0-preview
```
