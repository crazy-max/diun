# syntax=docker/dockerfile:1

ARG GO_VERSION="1.23"
ARG ALPINE_VERSION="3.21"
ARG GOLANGCI_LINT_VERSION="v1.62"

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base
ENV GOFLAGS="-buildvcs=false"
RUN apk add --no-cache gcc linux-headers musl-dev
WORKDIR /src

FROM golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci-lint
FROM base AS lint
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache \
    --mount=from=golangci-lint,source=/usr/bin/golangci-lint,target=/usr/bin/golangci-lint \
  golangci-lint run ./...
