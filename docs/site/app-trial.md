# Digest-pinned app trial

[日本語版](app-trial.ja.md)

This trial runs three sample apps against a published Mockport image. A separate, first-time user should substitute one flow in their own app; the bundled sample is preparation evidence, not an external-app trial. No real provider token, Slack workspace, payment, or AI inference is used.

## Verified combinations

| Flow set | Mockport source commit | Registry image | Sample commit | Platform | Result |
| --- | --- | --- | --- | --- | --- |
| P0/P1 Stripe + OpenAI | `67a81591e48526c88b96381cd56c54386d93341a` | `ghcr.io/albert-einshutoin/mockport@sha256:eae6ba56f01cc9969038da36982e6699732379028f27313ba86dba5d1e43b8fb` | `f6fe7d573ee37b282767aa112d30615be1941298` | `linux/arm64` | Fresh-directory Compose runner exit 0; Stripe and OpenAI success |
| #84 Stripe + OpenAI + Slack | `01ab1614e40e0d92122123639eba799633e3e6c1` | `ghcr.io/albert-einshutoin/mockport@sha256:534cfaa092373cf1a307a9a140415a015b414cad8301a24e0b546b9ed46dbb04` | `f6fe7d573ee37b282767aa112d30615be1941298` | `linux/arm64` | Fresh-directory Compose runner exit 0; all three success |

The versioned `0.2.0-preview` is an older release and does not contain these app flows. `latest` alone is not a reproducible image selector. The sample app Dockerfiles download their pinned SDK packages during build. A Docker image digest fixes Mockport only; the sample commit and package locks fix this example's source/dependency versions.

The P0/P1 row was exercised from a new temporary directory with a local Git clone of the pinned commit, the public digest pulled from GHCR, and the published Compose file. The running container image ID matched the digest, and `down --remove-orphans` removed its containers. The local clone took 0.76 seconds and Compose took 7.02 seconds with already cached base images and package layers. These are warm measurements; prerequisites and dependency downloads were not measured as a fresh installation.

The #84 row was exercised from a new directory with a GitHub clone of the pinned sample commit. Git clone took 0.94 seconds, a separate `docker compose build --no-cache` of the sample app images took 8.52 seconds, the published Mockport digest pull took 3.21 seconds, and `up --no-build` with all three flows took 5.43 seconds. A second new-directory run of the exact `up --build` command below took 6.77 seconds after dependencies had been cached. Docker, Git, base images, OS caches, and registry caches were already available; these figures are phase measurements, not a complete cold-machine installation time. The running Mockport image ID matched the pinned digest and Compose cleanup removed its containers.

## Run from a new directory

Prerequisites: Docker with Compose, Git, access to GHCR and package registries, and four free local ports 43101/33001/33002/33003. Start timing before installing any missing prerequisite. From a new working directory, use the exact sample commit and image digest from the matching table row:

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

The published Compose file has an `image:` selector for Mockport and no Mockport `build:`. The runner prints the Stripe order, OpenAI streamed text, and the Slack SDK reply. A nonzero runner exit means the selected flow failed. Record `git rev-parse HEAD`, the pulled digest, `docker image inspect` platform/image ID, the command and result together. Clean up with the `down` command even after an unsuccessful run. To exercise only the earlier P0/P1 image, use its digest from the first table row and set `PUBLISHED_FLOWS=p0p1`; that image does not contain #84.

## First-time external app trial card

Choose one ordinary app flow and wire only its provider URL and fake credentials to Mockport. For Slack, configure distinct fake `SLACK_BOT_TOKEN` and `SLACK_SIGNING_SECRET`, then deliver the selected message event. Expected success is one accepted event and one SDK reply to the event channel/thread. As a representative failure, send a changed body after signing or an old timestamp and check that the app rejects it without another reply. For Stripe or OpenAI, use their [app examples](../../examples/app-e2e/README.md) to select a matching failure case. Avoid real credentials.

| Record | Observation |
| --- | --- |
| Connected app and selected flow |  |
| Preparation start → first app success; prerequisite install / dependency fetch / execution split |  |
| Code changes beyond endpoint and fake credential configuration |  |
| Steps completed from docs alone; help or rework needed |  |
| Expected failure noticed and root cause identified |  |

Record the environment preparation start and each phase boundary with wall-clock times. A cached run measures only a warm rerun. A Codex-run bundled sample does not count as a first-time user's separate app. If the trial exposes a setup blocker, prioritize that repair before expanding the API surface.
