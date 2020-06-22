FROM golang:1.13-alpine as Builder

ARG APP_HOME=/go/src/github.com/CodyGuo/semaphore

ENV GO111MODULE="on" \
    GOPATH=/go

WORKDIR ${APP_HOME}

COPY . ${APP_HOME}

RUN set -ex \
    && apk add --no-cache git curl bash \
    && cd $(go env GOPATH) \
    && git clone https://github.com/go-task/task \
    && cd task \
    && go install -v ./cmd/task \
    && cd ${APP_HOME} \
    && task deps:tools \
    && task deps:be \
    && task compile:be \
    && task compile:api:hooks \
    && go get -u -v github.com/snikch/goodman/cmd/goodman

FROM apiaryio/dredd:13.0.0

ARG GOPATH=/go

ENV GOPATH=${GOPATH} \
    PATH=${GOPATH}/bin:/usr/local/bin:${PATH} \
    SEMAPHORE_SERVICE=semaphore_ci \
    SEMAPHORE_PORT=3000 \
    MYSQL_SERVICE=mysql \
    MYSQL_PORT=3306

WORKDIR /go/src/github.com/CodyGuo/semaphore

COPY --from=Builder /go/src /go/src
COPY --from=Builder /go/bin /go/bin
COPY deployment/docker/ci/dredd/entrypoint /usr/local/bin

ENTRYPOINT ["/usr/local/bin/entrypoint"]
