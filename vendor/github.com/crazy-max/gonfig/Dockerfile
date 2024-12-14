# syntax=docker/dockerfile:1

ARG GO_VERSION="1.20"
ARG GOLANGCI_LINT_VERSION="v1.51"

FROM golang:${GO_VERSION}-alpine AS base
RUN apk add --no-cache gcc git linux-headers musl-dev
ENV CGO_ENABLED=0
ENV GO111MODULE=on
WORKDIR /src

FROM base AS vendored
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/go/pkg/mod \
  go mod tidy && go mod download && \
  mkdir /out && cp go.mod go.sum /out

FROM scratch AS vendor-update
COPY --from=vendored /out /

FROM vendored AS vendor-validate
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

FROM golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci-lint
FROM base AS lint
RUN --mount=type=bind,target=. \
  --mount=type=cache,target=/root/.cache \
  --mount=from=golangci-lint,source=/usr/bin/golangci-lint,target=/usr/bin/golangci-lint \
  golangci-lint run ./...

FROM vendored AS test
ENV CGO_ENABLED=1
RUN --mount=type=bind,target=. \
  --mount=type=cache,target=/root/.cache \
  --mount=type=cache,target=/go/pkg/mod \
  go test -v -coverprofile=/tmp/coverage.txt -covermode=atomic -race ./... && \
  go tool cover -func=/tmp/coverage.txt

FROM scratch AS test-coverage
COPY --from=test /tmp/coverage.txt /coverage.txt
