FROM --platform=${BUILDPLATFORM:-linux/amd64} tonistiigi/xx:golang AS xgo
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.13-alpine AS builder

ARG BUILD_DATE
ARG VCS_REF
ARG VERSION=dev

ENV CGO_ENABLED 0
ENV GO111MODULE on
ENV GOPROXY https://goproxy.io
COPY --from=xgo / /

ARG TARGETPLATFORM
RUN go env

RUN apk --update --no-cache add \
    build-base \
    gcc \
    git \
  && rm -rf /tmp/* /var/cache/apk/*

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ./
RUN go build -ldflags "-w -s -X 'main.version=${VERSION}'" -v -o diun cmd/main.go

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:latest

ARG BUILD_DATE
ARG VCS_REF
ARG VERSION

LABEL maintainer="CrazyMax" \
  org.opencontainers.image.created=$BUILD_DATE \
  org.opencontainers.image.url="https://github.com/crazy-max/diun" \
  org.opencontainers.image.source="https://github.com/crazy-max/diun" \
  org.opencontainers.image.version=$VERSION \
  org.opencontainers.image.revision=$VCS_REF \
  org.opencontainers.image.vendor="CrazyMax" \
  org.opencontainers.image.title="Diun" \
  org.opencontainers.image.description="Docker image update notifier" \
  org.opencontainers.image.licenses="MIT"

RUN apk --update --no-cache add \
    ca-certificates \
    libressl \
    tzdata \
  && rm -rf /tmp/* /var/cache/apk/*

COPY --from=builder /app/diun /usr/local/bin/diun
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
RUN diun --version

ENV DIUN_DB="/data/diun.db"

VOLUME [ "/data" ]

ENTRYPOINT [ "diun" ]
CMD [ "--config", "/diun.yml" ]
