# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.25

FROM golang:${GO_VERSION} AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=mod -o /out/metatools ./cmd/metatools

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/metatools /app/metatools
COPY --from=build /src/examples /app/examples

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/metatools"]
CMD ["serve", "--transport=streamable", "--host=0.0.0.0", "--port=8080", "--config=/app/examples/metatools.yaml"]
