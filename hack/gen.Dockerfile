# syntax=docker/dockerfile:1

ARG GO_VERSION="1.20"
ARG PROTOC_VERSION="3.17.3"

# protoc is dynamically linked to glibc so can't use alpine base
FROM golang:${GO_VERSION}-bullseye AS base
RUN apt-get update && apt-get --no-install-recommends install -y git unzip
ARG PROTOC_VERSION
ARG TARGETOS
ARG TARGETARCH
RUN <<EOT
  set -e
  arch=$(echo $TARGETARCH | sed -e s/amd64/x86_64/ -e s/arm64/aarch_64/)
  wget -q https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-${TARGETOS}-${arch}.zip
  unzip protoc-${PROTOC_VERSION}-${TARGETOS}-${arch}.zip -d /usr/local
EOT
WORKDIR /src

FROM base AS vendored
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download

FROM vendored AS tools
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/go/pkg/mod \
  go install \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc \
    google.golang.org/protobuf/cmd/protoc-gen-go

FROM tools AS generate
RUN --mount=type=bind,target=.,rw \
    --mount=type=cache,target=/go/pkg/mod <<EOT
  set -e
  go generate -v ./...
  mkdir /out
  cp -Rf pb /out
EOT

FROM scratch AS update
COPY --from=generate /out /

FROM generate AS validate
RUN --mount=type=bind,target=.,rw <<EOT
  set -e
  git add -A
  cp -rf /out/* .
  diff=$(git status --porcelain -- pb)
  if [ -n "$diff" ]; then
    echo >&2 'ERROR: Vendor result differs. Please vendor your package with "docker buildx bake gen"'
    echo "$diff"
    exit 1
  fi
EOT
