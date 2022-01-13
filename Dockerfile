# syntax=docker/dockerfile:1-labs

ARG GO_VERSION="1.17"
ARG GORELEASER_XX_VERSION="1.2.5"

FROM --platform=$BUILDPLATFORM crazymax/goreleaser-xx:${GORELEASER_XX_VERSION} AS goreleaser-xx
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS base
ENV CGO_ENABLED=0
COPY --from=goreleaser-xx / /
RUN apk add --no-cache file git
WORKDIR /src

FROM base AS vendored
RUN --mount=type=bind,source=.,target=/src,rw \
  --mount=type=cache,target=/go/pkg/mod \
  go mod tidy && go mod download

FROM vendored AS test
ENV CGO_ENABLED=1
RUN apk add --no-cache gcc linux-headers musl-dev
RUN --mount=type=bind,target=. \
  --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build <<EOT
go test -v -coverprofile=/tmp/coverage.txt -covermode=atomic -race ./...
go tool cover -func=/tmp/coverage.txt
EOT

FROM scratch AS test-coverage
COPY --from=test /tmp/coverage.txt /coverage.txt

FROM vendored AS build
ARG TARGETPLATFORM
RUN --mount=type=bind,target=. \
  --mount=type=cache,target=/root/.cache \
  --mount=target=/go/pkg/mod,type=cache \
  goreleaser-xx --debug \
    --name "diun" \
    --dist "/out" \
    --main="./cmd" \
    --flags="-trimpath" \
    --ldflags="-s -w -X 'main.version={{.Version}}'" \
    --files="CHANGELOG.md" \
    --files="LICENSE" \
    --files="README.md"

FROM scratch AS artifact
COPY --from=build /out/*.tar.gz /
COPY --from=build /out/*.zip /

FROM scratch AS binary
COPY --from=build /usr/local/bin/diun* /

FROM alpine:3.15
RUN apk --update --no-cache add ca-certificates openssl
COPY --from=build /usr/local/bin/diun /usr/local/bin/diun
ENV PROFILER_PATH="/profiler" \
  DIUN_DB_PATH="/data/diun.db"
VOLUME [ "/data" ]
ENTRYPOINT [ "diun" ]
CMD [ "serve" ]
