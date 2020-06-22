# Golang testing image with some tools already installed
FROM golang:1.13-alpine

LABEL maintainer="Tom Whiston <tom.whiston@gmail.com>"

ENV SEMAPHORE_VERSION="development" SEMAPHORE_ARCH="linux_amd64" \
    SEMAPHORE_CONFIG_PATH="${SEMAPHORE_CONFIG_PATH:-/etc/semaphore}" \
    APP_ROOT="/go/src/github.com/CodyGuo/semaphore/" \
    GO111MODULE="on"

# hadolint ignore=DL3013
RUN set -ex \
    && apk add --no-cache --virtual mysql-client \
    && apk add --no-cache --virtual openssl openssl-dev ca-certificates curl curl-dev openssh-client tini nodejs nodejs-npm bash rsync \
    && apk add --no-cache --virtual python py-pip py-openssl python-dev \
    && apk --update add --virtual build-dependencies libffi-dev build-base \
    && pip install --upgrade pip cffi   \
    && pip install ansible              \
    && apk del build-dependencies \
    && rm -rf /var/cache/apk/*    \
    && adduser -D -u 1002 -g 0 semaphore \
    && mkdir -p /go/src/github.com/CodyGuo/semaphore \
    && mkdir -p /tmp/semaphore \
    && mkdir -p /etc/semaphore \
    && chown -R semaphore:0 /go \
    && chown -R semaphore:0 /tmp/semaphore \
    && chown -R semaphore:0 /etc/semaphore \
    && ssh-keygen -t rsa -q -f "/root/.ssh/id_rsa" -N "" \
    && ssh-keyscan -H github.com > /root/.ssh/known_hosts \
    && cd $(go env GOPATH) && git clone https://github.com/go-task/task \
    && cd task && go install -v ./cmd/task

# Copy in app source
WORKDIR ${APP_ROOT}
COPY . ${APP_ROOT}
RUN deployment/docker/ci/bin/install

USER semaphore
EXPOSE 3000
ENTRYPOINT ["/usr/local/bin/semaphore-wrapper"]
CMD ["./bin/semaphore", "--config", "/etc/semaphore/config.json"]
