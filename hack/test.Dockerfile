# syntax=docker/dockerfile:1.2
ARG GO_VERSION

FROM golang:${GO_VERSION}-alpine AS base
RUN apk add --no-cache gcc linux-headers musl-dev
WORKDIR /src

FROM base AS gomod
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/go/pkg/mod \
  go mod tidy && go mod download

FROM gomod AS test
RUN --mount=type=bind,target=. \
  --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go test -v -coverprofile=/tmp/coverage.txt -covermode=atomic -race ./... && \
  go tool cover -func=/tmp/coverage.txt

FROM scratch AS test-coverage
COPY --from=test /tmp/coverage.txt /coverage.txt
