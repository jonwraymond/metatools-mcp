# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.25

FROM golang:${GO_VERSION} AS build
WORKDIR /src

ARG TARGETOS
ARG TARGETARCH

# This Dockerfile is intended to be built from the ApertureStack root as context,
# so local `replace ../toolops` and `replace ../toolops-integrations` directives work.
COPY metatools-mcp/go.mod metatools-mcp/go.sum ./metatools-mcp/
COPY toolops/go.mod toolops/go.sum ./toolops/
COPY toolops-integrations/go.mod toolops-integrations/go.sum ./toolops-integrations/

WORKDIR /src/metatools-mcp
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY metatools-mcp ./metatools-mcp
COPY toolops ./toolops
COPY toolops-integrations ./toolops-integrations

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -mod=mod -o /out/metatools ./cmd/metatools

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/metatools /app/metatools
COPY --from=build /src/metatools-mcp/examples /app/examples

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/metatools"]
CMD ["serve", "--transport=streamable", "--host=0.0.0.0", "--port=8080", "--config=/app/examples/metatools.yaml"]
