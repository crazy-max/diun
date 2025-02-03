# syntax=docker/dockerfile:1

ARG GO_VERSION="1.23"
ARG XX_VERSION="1.6.1"
ARG ALPINE_VERSION="3.21"
ARG GOLANGCI_LINT_VERSION="v1.62"

FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base
ENV GOFLAGS="-buildvcs=false"
RUN apk add --no-cache gcc linux-headers musl-dev
COPY --from=xx --link / /
WORKDIR /src

FROM --platform=$BUILDPLATFORM golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci-lint
FROM base AS lint
ARG TARGETPLATFORM
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache,id=lint-cache-$TARGETPLATFORM \
    --mount=from=golangci-lint,source=/usr/bin/golangci-lint,target=/usr/bin/golangci-lint \
  xx-go --wrap && \
  golangci-lint run ./...
