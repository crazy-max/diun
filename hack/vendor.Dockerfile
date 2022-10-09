# syntax=docker/dockerfile:1

ARG GO_VERSION="1.19"
ARG GOMOD_OUTDATED_VERSION="v0.8.0"

FROM golang:${GO_VERSION}-alpine AS base
RUN apk add --no-cache git linux-headers musl-dev
WORKDIR /src

FROM base AS vendored
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/go/pkg/mod <<EOT
set -e
go mod tidy
go mod download
mkdir /out
cp go.mod go.sum /out
EOT

FROM scratch AS update
COPY --from=vendored /out /

FROM vendored AS validate
RUN --mount=type=bind,target=.,rw <<EOT
set -e
git add -A
cp -rf /out/* .
diff=$(git status --porcelain -- go.mod go.sum)
if [ -n "$diff" ]; then
  echo >&2 'ERROR: Vendor result differs. Please vendor your package with "docker buildx bake vendor"'
  echo "$diff"
  exit 1
fi
EOT

FROM psampaz/go-mod-outdated:${GOMOD_OUTDATED_VERSION} AS go-mod-outdated
FROM base AS outdated
RUN --mount=type=bind,target=. \
  --mount=type=cache,target=/go/pkg/mod \
  --mount=from=go-mod-outdated,source=/home/go-mod-outdated,target=/usr/bin/go-mod-outdated \
  go list -mod=readonly -u -m -json all | go-mod-outdated -update -direct
