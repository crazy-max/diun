# syntax=docker/dockerfile:1

ARG GO_VERSION=1.19

FROM golang:${GO_VERSION}-alpine AS base
RUN apk add --no-cache gcc git linux-headers musl-dev
WORKDIR /src

FROM base AS test
RUN --mount=type=bind,target=. \
  --mount=type=cache,target=/root/.cache \
  go test -v -coverprofile=/tmp/coverage.txt -covermode=atomic -race ./... && \
  go tool cover -func=/tmp/coverage.txt

FROM scratch AS test-coverage
COPY --from=test /tmp/coverage.txt /coverage.txt
