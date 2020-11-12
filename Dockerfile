ARG GO_VERSION=1.15
ARG VERSION=dev

FROM --platform=${BUILDPLATFORM:-linux/amd64} tonistiigi/xx:golang AS xgo

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:${GO_VERSION}-alpine AS base
RUN apk add --no-cache curl gcc git musl-dev
COPY --from=xgo / /
WORKDIR /src

FROM base AS gomod
COPY . .
RUN go mod download

FROM gomod AS build
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.io,direct
RUN go build -ldflags "-w -s -X 'main.version=${VERSION}'" -v -o /opt/diun cmd/main.go

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:latest
LABEL maintainer="CrazyMax"

RUN apk --update --no-cache add \
    ca-certificates \
    libressl \
  && rm -rf /tmp/* /var/cache/apk/*

COPY --from=build /opt/diun /usr/local/bin/diun
RUN diun --version

ENV DIUN_DB_PATH="/data/diun.db"

VOLUME [ "/data" ]
ENTRYPOINT [ "diun" ]
