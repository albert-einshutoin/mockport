# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

# Resolve dependencies in a cacheable layer. This layer is only rebuilt when
# go.mod or go.sum change, and the module cache is reused across builds.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Build the static binary. BuildKit cache mounts keep the module and build
# caches warm between builds; -trimpath strips local paths for reproducibility
# and -s -w drops the symbol table and DWARF debug info to shrink the binary.
COPY . .
ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath \
      -ldflags="-s -w -X github.com/albert-einshutoin/mockport/internal/cli.Version=${VERSION}" \
      -o /mockport ./cmd/mockport

# Distroless static + nonroot: no shell, no package manager, runs as UID 65532.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /mockport /mockport

EXPOSE 43101

USER nonroot:nonroot

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD ["/mockport", "healthcheck"]

ENTRYPOINT ["/mockport"]
CMD ["run", "--config", "/etc/mockport/mockport.yml"]
