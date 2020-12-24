#!/usr/bin/env sh

: ${TARGETPLATFORM=}
: ${TARGETOS=}
: ${TARGETARCH=}
: ${TARGETVARIANT=}
: ${CGO_ENABLED=}
: ${GOARCH=}
: ${GOOS=}
: ${GOARM=}
: ${GOBIN=}
: ${GIT_REF=}

set -eu

if [ ! -z "$TARGETPLATFORM" ]; then
  os="$(echo $TARGETPLATFORM | cut -d"/" -f1)"
  arch="$(echo $TARGETPLATFORM | cut -d"/" -f2)"
  if [ ! -z "$os" ] && [ ! -z "$arch" ]; then
    export GOOS="$os"
    export GOARCH="$arch"
    if [ "$arch" = "arm" ]; then
      case "$(echo $TARGETPLATFORM | cut -d"/" -f3)" in
      "v5")
        export GOARM="5"
        ;;
      "v6")
        export GOARM="6"
        ;;
      *)
        export GOARM="7"
        ;;
      esac
    fi
  fi
fi

if [ ! -z "$TARGETOS" ]; then
  export GOOS="$TARGETOS"
fi

if [ ! -z "$TARGETARCH" ]; then
  export GOARCH="$TARGETARCH"
fi

if [ "$TARGETARCH" = "arm" ]; then
  if [ ! -z "$TARGETVARIANT" ]; then
    case "$TARGETVARIANT" in
    "v5")
      export GOARM="5"
      ;;
    "v6")
      export GOARM="6"
      ;;
    *)
      export GOARM="7"
      ;;
    esac
  else
    export GOARM="7"
  fi
fi

if [ "$CGO_ENABLED" = "1" ]; then
  case "$GOARCH" in
  "amd64")
    export CC="x86_64-linux-gnu-gcc"
    ;;
  "ppc64le")
    export CC="powerpc64le-linux-gnu-gcc"
    ;;
  "s390x")
    export CC="s390x-linux-gnu-gcc"
    ;;
  "arm64")
    export CC="aarch64-linux-gnu-gcc"
    ;;
  "arm")
    case "$GOARM" in
    "5")
      export CC="arm-linux-gnueabi-gcc"
      ;;
    *)
      export CC="arm-linux-gnueabihf-gcc"
      ;;
    esac
    ;;
  esac
fi

if [ "$GOOS" = "wasi" ]; then
  export GOOS="js"
fi

if [ -z "$GOBIN" ] && [ -n "$GOPATH" ] && [ -n "$GOARCH" ] && [ -n "$GOOS" ]; then
  export PATH=${GOPATH}/bin/${GOOS}_${GOARCH}:${PATH}
fi

cat > ./.goreleaser.yml <<EOL
dist: /out

builds:
  -
    main: ./cmd/main.go
    ldflags:
      - -s -w -X "main.version={{ .Version }}" -X "main.commit={{ .ShortCommit }}"
    env:
      - CGO_ENABLED=0
    goos:
      - ${GOOS}
    goarch:
      - ${GOARCH}
    goarm:
      - ${GOARM}
    hooks:
      post:
        - cp "{{ .Path }}" /usr/local/bin/diun

archives:
  -
    replacements:
      386: i386
      amd64: x86_64
    format_overrides:
      - goos: windows
        format: zip
    files:
      - CHANGELOG.md
      - LICENSE
      - README.md

release:
  disable: true
EOL

gitTag=""
case "$GIT_REF" in
  refs/tags/v*)
    gitTag="${GIT_REF#refs/tags/v}"
    export GORELEASER_CURRENT_TAG=$gitTag
    ;;
  *)
    if gitTag=$(git tag --points-at HEAD --sort -version:creatordate | head -n 1); then
      if [ -z "$gitTag" ]; then
        gitTag=$(git describe --tags --abbrev=0)
      fi
    fi
    ;;
esac
echo "git tag found: ${gitTag}"

gitDirty="true"
if git describe --exact-match --tags --match "$gitTag" >/dev/null 2>&1; then
  gitDirty="false"
fi
echo "git dirty: ${gitDirty}"

flags=""
if [ "$gitDirty" = "true" ]; then
  flags="--snapshot"
fi

set -x
/usr/local/bin/goreleaser release $flags
