# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM ghcr.io/jdx/mise:2026.10.18-debian AS builder
LABEL maintainer="mingcheng<mingcheng@outlook.com>"

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=2.0.0
ARG BUILD_TIME=unknown
ARG BUILD_COMMIT=unknown

ENV CGO_ENABLED=0 GOTOOLCHAIN=local GOPATH=/go

WORKDIR /src

COPY mise.toml go.mod go.sum ./
RUN mise trust && mise install go
RUN --mount=type=cache,target=/go/pkg/mod mise exec go -- go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mise exec go -- go test ./... && \
    GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} mise exec go -- go build -trimpath \
      -ldflags="-s -w -X main.BuildVersion=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.BuildCommit=${BUILD_COMMIT}" \
      -o /bin/simplyddns ./cmd/simplyddns

# Stage2
FROM debian:13-slim

ENV TZ=Asia/Shanghai
# Keep security updates available within the pinned Debian release.
# hadolint ignore=DL3008
RUN echo "Asia/Shanghai" > /etc/timezone \
    && apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends ca-certificates tzdata dumb-init \
    && rm -rf /var/lib/apt/lists

COPY --from=builder /bin/simplyddns /bin/simplyddns

USER 65534:65534

ENTRYPOINT ["dumb-init", "--", "/bin/simplyddns"]
