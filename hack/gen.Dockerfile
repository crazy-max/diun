# syntax=docker/dockerfile:1.2
ARG GO_VERSION
ARG PROTOC_VERSION
ARG GLIBC_VERSION=2.33-r0

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

FROM base AS gomod
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/go/pkg/mod \
  go mod tidy && go mod download && go install -v $(sed -n -e 's|^\s*_\s*"\(.*\)".*$|\1| p' tools.go)

FROM gomod AS generate
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/go/pkg/mod \
  go generate ./... && mkdir /out && cp -Rf pb /out

FROM scratch AS update
COPY --from=generate /out /

FROM generate AS validate
RUN --mount=type=bind,target=.,rw \
  git add -A && cp -rf /out/* .; \
  if [ -n "$(git status --porcelain -- pb)" ]; then \
    echo >&2 'ERROR: Generate result differs. Please update with "docker buildx bake gen-update"'; \
    git status --porcelain -- pb; \
    exit 1; \
  fi
