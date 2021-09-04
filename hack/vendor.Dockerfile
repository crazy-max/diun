# syntax=docker/dockerfile:1.2
ARG GO_VERSION

FROM golang:${GO_VERSION}-alpine AS base
RUN apk add --no-cache git linux-headers musl-dev
WORKDIR /src

FROM base AS vendored
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/go/pkg/mod \
  go mod tidy && go mod download && \
  mkdir /out && cp go.mod go.sum /out

FROM scratch AS update
COPY --from=vendored /out /

FROM vendored AS validate
RUN --mount=type=bind,target=.,rw \
  git add -A && cp -rf /out/* .; \
  if [ -n "$(git status --porcelain -- go.mod go.sum)" ]; then \
    echo >&2 'ERROR: Vendor result differs. Please vendor your package with "docker buildx bake vendor-update"'; \
    git status --porcelain -- go.mod go.sum; \
    exit 1; \
  fi

FROM psampaz/go-mod-outdated:v0.8.0 AS go-mod-outdated
FROM base AS outdated
RUN --mount=type=bind,target=.,ro \
  --mount=type=cache,target=/go/pkg/mod \
  --mount=from=go-mod-outdated,source=/home/go-mod-outdated,target=/usr/bin/go-mod-outdated \
  go list -mod=readonly -u -m -json all | go-mod-outdated -update -direct
