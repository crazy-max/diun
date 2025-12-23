# syntax=docker/dockerfile:1

ARG GO_VERSION="1.24"
ARG ALPINE_VERSION="3.23"

ARG GOVULNCHECK_VERSION="v1.1.4"
ARG FORMAT="text"

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base
WORKDIR /go/src/github.com/docker/buildx
RUN apk add --no-cache moreutils
ARG GOVULNCHECK_VERSION
RUN --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    go install golang.org/x/vuln/cmd/govulncheck@$GOVULNCHECK_VERSION

FROM base AS run
ARG FORMAT
RUN --mount=type=bind,target=. <<EOT
  set -ex
  mkdir /out
  govulncheck -format ${FORMAT} ./... | tee /out/govulncheck.out
EOT

FROM scratch AS output
COPY --from=run /out /
