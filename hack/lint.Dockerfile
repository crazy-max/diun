# syntax=docker/dockerfile:1

ARG GO_VERSION="1.24"
ARG XX_VERSION="1.6.1"
ARG ALPINE_VERSION="3.22"
ARG GOLANGCI_LINT_VERSION="v2.1.6"
ARG GOLANGCI_FROM_SOURCE="true"

FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base
ENV GOFLAGS="-buildvcs=false"
RUN apk add --no-cache gcc linux-headers musl-dev
COPY --from=xx --link / /
WORKDIR /src

FROM base AS golangci-build
ARG GOLANGCI_LINT_VERSION
ADD "https://github.com/golangci/golangci-lint.git#${GOLANGCI_LINT_VERSION}" .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/ go mod download
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/ mkdir -p out && go build -o /out/golangci-lint ./cmd/golangci-lint

FROM --platform=$BUILDPLATFORM golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci-lint
FROM scratch AS golangci-binary-false
COPY --from=golangci-lint /usr/bin/golangci-lint golangci-lint
FROM scratch AS golangci-binary-true
COPY --from=golangci-build /out/golangci-lint golangci-lint
FROM golangci-binary-${GOLANGCI_FROM_SOURCE} AS golangci-binary

FROM base AS lint
ARG TARGETPLATFORM
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache,id=lint-cache-$TARGETPLATFORM \
    --mount=from=golangci-binary,source=/golangci-lint,target=/usr/bin/golangci-lint \
  xx-go --wrap && \
  golangci-lint run ./...
