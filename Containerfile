# syntax=docker/dockerfile:1.6
# Multi-stage: golang builder -> gcr.io/distroless/static (smallest possible).
FROM golang:1.25-alpine AS build
WORKDIR /src

# Cache deps separately
COPY go.mod ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy and build (pure-Go, statically linked)
COPY . .
ARG VERSION="dev"
ARG COMMIT="unknown"
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath \
      -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
      -o /out/platform-demo .

# Distroless static = ~2 MiB base + our ~7 MiB binary
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/platform-demo /platform-demo
EXPOSE 9898
USER nonroot
ENTRYPOINT ["/platform-demo"]
