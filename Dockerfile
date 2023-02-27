FROM golang:1.19 AS builder
LABEL maintainer="mingcheng<mingcheng@outlook.com>"

ARG GITEA_TOKEN
ENV GITEA_TOKEN ${GITEA_TOKEN}

ENV PACKAGE github.com/mingcheng/simplyddns
ENV BUILD_DIR ${GOPATH}/src/${PACKAGE}
ENV GOPROXY https://goproxy.cn,direct

# Build
COPY . ${BUILD_DIR}
WORKDIR ${BUILD_DIR}
RUN git config --global url."https://${GITEA_TOKEN}@repo.wooramel.cn/".insteadOf "https://repo.wooramel.cn/" \
 	&& make clean build \
	&& cp ./simplyddns /bin/simplyddns \
	&& cp ./example/basic.yml /etc/simplyddns.yml

# Stage2
FROM debian:bullseye

ENV TZ "Asia/Shanghai"
RUN sed -i 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list \
	&& sed -i 's/security.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list \
	&& echo "Asia/Shanghai" > /etc/timezone \
	&& apt -y update \
	&& apt -y upgrade \
	&& apt -y install ca-certificates openssl tzdata curl netcat dumb-init \
	&& apt -y autoremove

COPY --from=builder /bin/simplyddns /bin/simplyddns
COPY --from=builder /etc/simplyddns.yml /etc/simplyddns.yml

USER nobody
HEALTHCHECK --interval=30s --timeout=3s CMD nc -w 3 -zv 114.114.114.114 53 || exit 1

ENTRYPOINT ["dumb-init", "/bin/simplyddns"]
