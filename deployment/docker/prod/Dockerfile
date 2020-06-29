# Golang testing image with some tools already installed
FROM golang:1.13-alpine as builder

LABEL maintainer="CodyGuo <codyguo@aliyun.com>"

ENV APP_ROOT="/go/src/github.com/CodyGuo/semaphore/" \
    GO111MODULE="on"

# hadolint ignore=DL3013
RUN apk add --no-cache git curl openssh-client && \
    adduser -D -u 1001 -G root semaphore && \
    mkdir -p /tmp/semaphore && \
    mkdir -p /etc/semaphore && \
    chown -R semaphore:0 /tmp/semaphore && \
    chown -R semaphore:0 /etc/semaphore

COPY ./ ${APP_ROOT}
WORKDIR ${APP_ROOT}

RUN apk add --no-cache -U libc-dev nodejs nodejs-npm bash && \
  ./deployment/docker/prod/bin/install

FROM frolvlad/alpine-glibc:alpine-3.7

LABEL maintainer="Tom Whiston <tom.whiston@gmail.com>"

RUN apk add --no-cache git ansible mysql-client curl openssh-client tini && \
    adduser -D -u 1001 -G root semaphore && \
    mkdir -p /tmp/semaphore && \
    mkdir -p /etc/semaphore && \
    chown -R semaphore:0 /tmp/semaphore && \
    chown -R semaphore:0 /etc/semaphore

COPY --from=builder /usr/local/bin/semaphore-wrapper /usr/local/bin/semaphore-wrapper
COPY --from=builder /usr/local/bin/semaphore /usr/local/bin/semaphore

RUN chown 1001:0 /usr/local/bin/semaphore && \
    chown 1001:0 /usr/local/bin/semaphore-wrapper

WORKDIR /home/semaphore
USER 1001

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/usr/local/bin/semaphore-wrapper", "/usr/local/bin/semaphore", "--config", "/etc/semaphore/config.json"]
