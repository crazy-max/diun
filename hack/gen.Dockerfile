# syntax=docker/dockerfile:1

ARG GO_VERSION="1.19"
ARG PROTOC_VERSION="3.17.3"
ARG GLIBC_VERSION="2.33-r0"

FROM golang:${GO_VERSION}-alpine AS base
ARG GLIBC_VERSION
RUN apk add --no-cache curl file git unzip \
  && curl -sSL "https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub" -o "/etc/apk/keys/sgerrand.rsa.pub" \
  && curl -sSL "https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VERSION}/glibc-${GLIBC_VERSION}.apk" -o "glibc.apk" \
  && apk add glibc.apk \
  && rm /etc/apk/keys/sgerrand.rsa.pub glibc.apk
ARG PROTOC_VERSION
RUN curl -sSL "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip" -o "protoc.zip" \
  && unzip "protoc.zip" -d "/usr/local" \
  && protoc --version \
  && rm "protoc.zip"
WORKDIR /src

FROM base AS vendored
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/go/pkg/mod \
  go mod tidy && go mod download

FROM vendored AS tools
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/go/pkg/mod \
  go install -v $(sed -n -e 's|^\s*_\s*"\(.*\)".*$|\1| p' tools.go)

FROM tools AS generate
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/go/pkg/mod <<EOT
set -e
go generate ./...
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
