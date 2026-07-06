# Quickstart

[日本語版](quickstart.ja.md)

## Install

### Option A: Docker (recommended, no install)

```bash
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/examples/stripe-checkout/mockport.yml:/etc/mockport/mockport.yml \
  ghcr.io/albert-einshutoin/mockport:0.1.0-alpha \
  run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

Use the explicit `0.1.0-alpha` image tag for preview installs. The `latest` tag follows the default branch image and is not the preview release contract.

### Option B: Release binary

Download `mockport_<version>_<os>_<arch>.tar.gz` and `checksums.txt` from [GitHub Releases](https://github.com/albert-einshutoin/mockport/releases), verify the checksum, extract, and run `./mockport version`. Full steps: [Distribution](distribution.md).

### Option C: From source

```bash
make build
./bin/mockport version
```

## Project Setup (Options B and C)

After you have the `mockport` binary on your PATH, generate project files and start Compose:

```bash
mockport init --adapter stripe
docker compose -f docker-compose.mockport.yml up
curl http://localhost:43101/health
mockport healthcheck
```

For multiple adapters:

```bash
mockport init --adapter stripe --adapter openai --adapter github-oauth --adapter slack --adapter line --adapter zoho-oauth
docker compose -f docker-compose.mockport.yml up
```

## Switching scenarios

Besides fixing a scenario in `mockport.yml`, you can switch per request using the `X-Mockport-Scenario` header — no server restart required.

```bash
# Test the Stripe failure path without restarting the server
curl -X POST http://localhost:43101/stripe/v1/checkout/sessions \
  -H "X-Mockport-Scenario: payment_failed" \
  -H "Authorization: Bearer $STRIPE_KEY" \
  -d "mode=payment&success_url=http://localhost/success&cancel_url=http://localhost/cancel"
```

See the [adapter reference](adapters.md) for the list of supported scenarios per adapter.

`mockport init` protects existing generated files by default. Use `--force` only when you intentionally want to replace `mockport.yml`, `.env.mockport`, and `docker-compose.mockport.yml`.

## Expected Output

`mockport init --adapter stripe` creates:

- `mockport.yml`
- `.env.mockport`
- `docker-compose.mockport.yml`

Health check after the server is running:

```bash
$ curl http://localhost:43101/health
{"status":"ok"}
```

Inspect the request and safety report:

```bash
mockport report
```

Or fetch the JSON report directly:

```bash
curl http://localhost:43101/_mockport/report
```

## Troubleshooting

**Docker is not running** — `docker compose up` fails with `Cannot connect to the Docker daemon` or similar. Start Docker Desktop (or your local Docker engine) and rerun `docker compose -f docker-compose.mockport.yml up`.

**Port 43101 is already in use** — Mockport fails to bind with `address already in use` on port `43101`. Stop the other process using that port, or change `server.port` in `mockport.yml` and the host port mapping in `docker-compose.mockport.yml` together so they stay aligned.
