FROM golang:1.24 AS builder
LABEL maintainer="mingcheng<mingcheng@outlook.com>"

ENV PACKAGE github.com/mingcheng/simplyddns
ENV BUILD_DIR ${GOPATH}/src/${PACKAGE}
ENV GOPROXY https://goproxy.cn,direct

# Build
COPY . ${BUILD_DIR}
WORKDIR ${BUILD_DIR}
RUN make build && cp ./simplyddns /bin/simplyddns

# Stage2
FROM debian:stable

ENV TZ "Asia/Shanghai"
RUN echo "Asia/Shanghai" > /etc/timezone \
 	&& apt -y update && apt -y install ca-certificates openssl tzdata curl netcat-openbsd dumb-init \
 	&& apt -y autoremove

COPY --from=builder /bin/simplyddns /bin/simplyddns

USER nobody
HEALTHCHECK --interval=30s --timeout=3s CMD nc -w 3 -zv 114.114.114.114 53 || exit 1

ENTRYPOINT ["dumb-init", "/bin/simplyddns"]
