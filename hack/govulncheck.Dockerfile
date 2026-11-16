# syntax=docker/dockerfile:1

ARG GO_VERSION="1.27"
# renovate: datasource=docker depName=alpine versioning=docker
ARG ALPINE_VERSION="3.24"

# renovate: datasource=go depName=golang.org/x/vuln
ARG GOVULNCHECK_VERSION="v1.9.0"
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
  set -ex -o pipefail
  mkdir /out
  govulncheck -format ${FORMAT} ./... | tee /out/govulncheck.out
EOT

FROM scratch AS output
COPY --from=run /out /
