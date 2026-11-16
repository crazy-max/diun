# syntax=docker/dockerfile:1

ARG GO_VERSION="1.27"
ARG XX_VERSION="1.10.0"
ARG ALPINE_VERSION="3.24"
ARG GOLANGCI_LINT_VERSION="v2.15.0"
ARG GOLANGCI_FROM_SOURCE="true"
ARG GOPLS_ANALYZERS_VERSION="v0.3.0"

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
ARG TARGETNAME
ARG TARGETPLATFORM
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache,id=lint-cache-${TARGETNAME}-${TARGETPLATFORM} \
    --mount=from=golangci-binary,source=/golangci-lint,target=/usr/bin/golangci-lint \
  xx-go --wrap && \
  golangci-lint run ./...

FROM base AS gopls-analyzers
ARG GOPLS_ANALYZERS_VERSION
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/ \
  GOBIN=/out go install github.com/crazy-max/gopls-analyzers/cmd/gopls-analyzers@${GOPLS_ANALYZERS_VERSION}

FROM base AS gopls-analyze
ENV GOFLAGS="-buildvcs=false -mod=vendor -tags=containers_image_openpgp,exclude_graphdriver_btrfs,exclude_graphdriver_devicemapper"
ARG TARGETNAME
ARG TARGETPLATFORM
RUN --mount=target=. \
  --mount=target=/root/.cache,type=cache,id=gopls-cache-${TARGETNAME}-${TARGETPLATFORM} \
  --mount=target=/usr/bin/gopls-analyzers,from=gopls-analyzers,source=/out/gopls-analyzers <<EOF
  set -ex
  xx-go --wrap
  gopls-analyzers -list-analyzers
  go vet -vettool=/usr/bin/gopls-analyzers -exclude-generated ./...
EOF
