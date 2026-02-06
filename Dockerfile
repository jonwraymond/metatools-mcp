# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.25

FROM golang:${GO_VERSION} AS build
WORKDIR /src

ARG TARGETOS
ARG TARGETARCH

# Bitwarden's `sdk-go` uses cgo to link a vendored static library (`libbitwarden_c.a`).
# We need a C toolchain in the build image.
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

# This Dockerfile is intended to be built from the ApertureStack root as context,
# so local `replace ../toolops` and `replace ../toolops-integrations` directives work.
COPY metatools-mcp/go.mod metatools-mcp/go.sum ./metatools-mcp/
COPY toolops/go.mod toolops/go.sum ./toolops/
COPY toolops-integrations/go.mod toolops-integrations/go.sum ./toolops-integrations/

WORKDIR /src/metatools-mcp
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY metatools-mcp /src/metatools-mcp
COPY toolops /src/toolops
COPY toolops-integrations /src/toolops-integrations

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 CGO_LDFLAGS="-lm" GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -mod=mod -o /out/metatools ./cmd/metatools

# We link a cgo-based Bitwarden SDK; runtime needs libgcc_s and other C runtime libs.
FROM gcr.io/distroless/cc-debian12
WORKDIR /app
COPY --from=build /out/metatools /app/metatools
COPY --from=build /src/metatools-mcp/examples /app/examples

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/metatools"]
CMD ["serve", "--transport=streamable", "--host=0.0.0.0", "--port=8080", "--config=/app/examples/metatools.yaml"]
