# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS builder
LABEL maintainer="mingcheng<mingcheng@outlook.com>"

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=unknown
ARG BUILD_TIME=unknown
ARG BUILD_COMMIT=unknown

ENV CGO_ENABLED=0
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
      -ldflags="-s -w -X main.BuildVersion=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.BuildCommit=${BUILD_COMMIT} -extldflags -static" \
      -o /bin/simplyddns ./cmd/simplyddns

# Stage2
FROM debian:stable-slim

ENV TZ "Asia/Shanghai"
RUN echo "Asia/Shanghai" > /etc/timezone \
 	&& apt-get update \
 	&& apt-get install -y --no-install-recommends ca-certificates openssl tzdata curl netcat-openbsd dumb-init \
 	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/simplyddns /bin/simplyddns

USER nobody
HEALTHCHECK --interval=30s --timeout=3s CMD nc -w 3 -zv 114.114.114.114 53 || exit 1

ENTRYPOINT ["dumb-init", "/bin/simplyddns"]
