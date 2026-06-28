> **⚠️ Archive notice — Not maintained, may diverge, do not cite as authoritative.**
>
> Pre-implementation design archive. This is **not** the authoritative source for current implementation.
> For current specs see [docs/site/](../../../site/index.md).

# 09. Docker Runtime

[日本語版](09_docker_runtime.ja.md)

## Docker Engine version

Recommended Docker Engine: `29.5.2` or later in the 29.x stable line.

## Runtime strategy

Mockport should be Docker-first.

Primary usage:

```bash
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/mockport.yml:/etc/mockport/mockport.yml \
  ghcr.io/albert-einshutoin/mockport:0.1.0-alpha run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

## Docker Compose usage

```yaml
services:
  mockport:
    image: ghcr.io/albert-einshutoin/mockport:0.1.0-alpha
    command: ["run", "--config", "/etc/mockport/mockport.yml", "--host", "0.0.0.0"]
    ports:
      - "127.0.0.1:43101:43101"
    volumes:
      - ./mockport.yml:/etc/mockport/mockport.yml
```

## Application env switching

Production:

```env
STRIPE_API_URL=<provider Stripe API URL>
STRIPE_SECRET_KEY=<redacted real provider key>
```

Local/CI/AI:

```env
STRIPE_API_URL=http://mockport:43101/stripe
STRIPE_SECRET_KEY=mockport_stripe_secret
```

## Initial image model

Use one all-in-one image:

```txt
ghcr.io/albert-einshutoin/mockport:0.1.0-alpha
```

Use explicit release image tags for reproducible preview installs. The `latest` tag follows the default branch image and is not the preview release contract.

All MVP adapters are compiled in, but disabled unless config enables them.

## Future image model

Later:

```txt
ghcr.io/mockport/mockport:latest
ghcr.io/mockport/mockport-stripe:latest
ghcr.io/mockport/mockport-openai:latest
```

Do not start with adapter-specific images unless image size becomes a problem.

## Dockerfile

Recommended:

```dockerfile
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /mockport ./cmd/mockport

FROM gcr.io/distroless/static-debian12

COPY --from=builder /mockport /mockport

EXPOSE 43101

ENTRYPOINT ["/mockport"]
CMD ["run", "--config", "/etc/mockport/mockport.yml"]
```

## Why distroless

- smaller runtime surface
- no shell by default
- better security posture
- suitable for static Go binaries

## Build command

```bash
docker build -t mockport:local -f docker/Dockerfile .
```

## Run command

```bash
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/mockport.yml:/etc/mockport/mockport.yml \
  mockport:local run --config /etc/mockport/mockport.yml --host 0.0.0.0
```
