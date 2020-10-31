FROM --platform=${BUILDPLATFORM:-linux/amd64} tonistiigi/xx:golang AS xgo
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.13-alpine AS builder

ARG VERSION=dev

ENV CGO_ENABLED 0
ENV GO111MODULE on
ENV GOPROXY https://goproxy.io,direct
COPY --from=xgo / /

RUN apk --update --no-cache add \
    build-base \
    gcc \
    git \
  && rm -rf /tmp/* /var/cache/apk/*

WORKDIR /app

COPY . ./
RUN go mod download

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
RUN go env
RUN go build -ldflags "-w -s -X 'main.version=${VERSION}'" -v -o diun cmd/main.go

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:latest
LABEL maintainer="CrazyMax"

RUN apk --update --no-cache add \
    ca-certificates \
    libressl \
  && rm -rf /tmp/* /var/cache/apk/*

COPY --from=builder /app/diun /usr/local/bin/diun
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
RUN diun --version

ENV DIUN_DB_PATH="/data/diun.db"

VOLUME [ "/data" ]
ENTRYPOINT [ "diun" ]
