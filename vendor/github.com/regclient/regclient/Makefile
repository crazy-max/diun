COMMANDS?=regctl regsync regbot
BINARIES?=$(addprefix bin/,$(COMMANDS))
IMAGES?=$(addprefix docker-,$(COMMANDS))
ARTIFACT_PLATFORMS?=linux-amd64 linux-arm64 linux-ppc64le linux-s390x linux-riscv64 darwin-amd64 darwin-arm64 windows-amd64.exe freebsd-amd64
ARTIFACTS?=$(foreach cmd,$(addprefix artifacts/,$(COMMANDS)),$(addprefix $(cmd)-,$(ARTIFACT_PLATFORMS)))
IMAGE_PLATFORMS?=linux/386,linux/amd64,linux/arm/v6,linux/arm/v7,linux/arm64,linux/ppc64le,linux/s390x,linux/riscv64
VCS_REPO?="https://github.com/regclient/regclient.git"
VCS_REF?=$(shell git rev-list -1 HEAD)
ifneq ($(shell git status --porcelain 2>/dev/null),)
  VCS_REF := $(VCS_REF)-dirty
endif
VCS_VERSION?=$(shell vcs_describe="$$(git describe --all)"; \
  vcs_version="(devel)"; \
  if [ "$${vcs_describe}" != "$${vcs_describe#tags/}" ]; then \
    vcs_version="$${vcs_describe#tags/}"; \
  elif [ "$${vcs_describe}" != "$${vcs_describe#heads/}" ]; then \
    vcs_version="$${vcs_describe#heads/}"; \
    if [ "main" = "$${vcs_version}" ]; then vcs_version=edge; fi; \
  fi; \
  echo "$${vcs_version}" | sed -r 's#/+#-#g')
VCS_TAG?=$(shell git describe --tags --abbrev=0 2>/dev/null || true)
VCS_SEC?=$(shell git log -1 --format=%ct)
VCS_DATE?=$(shell date -d "@$(VCS_SEC)" +%Y-%m-%dT%H:%M:%SZ --utc)
LD_FLAGS?=-s -w -extldflags -static -buildid= -X \"github.com/regclient/regclient/internal/version.vcsTag=$(VCS_TAG)\"
GO_BUILD_FLAGS?=-trimpath -ldflags "$(LD_FLAGS)"
DOCKERFILE_EXT?=$(shell if docker build --help 2>/dev/null | grep -q -- '--progress'; then echo ".buildkit"; fi)
DOCKER_ARGS?=--build-arg "VCS_REF=$(VCS_REF)" --build-arg "VCS_VERSION=$(VCS_VERSION)" --build-arg "SOURCE_DATE_EPOCH=$(VCS_SEC)"  --build-arg "BUILD_DATE=$(VCS_DATE)"
GOPATH?=$(shell go env GOPATH)
PWD:=$(shell pwd)
VER_BUMP?=$(shell command -v version-bump 2>/dev/null)
VER_BUMP_CONTAINER?=sudobmitch/version-bump:edge
ifeq "$(strip $(VER_BUMP))" ''
	VER_BUMP=docker run --rm \
		-v "$(shell pwd)/:$(shell pwd)/" -w "$(shell pwd)" \
		-u "$(shell id -u):$(shell id -g)" \
		$(VER_BUMP_CONTAINER)
endif
MARKDOWN_LINT_VER?=v0.21.0
GOFUMPT_VER?=v0.9.2
GOMAJOR_VER?=v0.15.0
GOSEC_VER?=v2.23.0
GO_VULNCHECK_VER?=v1.1.4
OSV_SCANNER_VER?=v2.3.3
SYFT?=$(shell command -v syft 2>/dev/null)
SYFT_CMD_VER:=$(shell [ -x "$(SYFT)" ] && echo "v$$($(SYFT) version | awk '/^Version: / {print $$2}')" || echo "0")
SYFT_VERSION?=v1.42.1
SYFT_CONTAINER?=anchore/syft:v1.42.1@sha256:392b65f29a410d2c1294d347bb3ad6f37608345ab6e7b43d2df03ea18bd6f5b0
ifneq "$(SYFT_CMD_VER)" "$(SYFT_VERSION)"
	SYFT=docker run --rm \
		-v "$(shell pwd)/:$(shell pwd)/" -w "$(shell pwd)" \
		-u "$(shell id -u):$(shell id -g)" \
		$(SYFT_CONTAINER)
endif
STATICCHECK_VER?=v0.7.0
CI_DISTRIBUTION_VER?=3.0.0
CI_ZOT_VER?=v2.1.14

.PHONY: .FORCE
.FORCE:

.PHONY: all
all: fmt gofumpt gofix goimports vet test lint binaries ## Full build of Go binaries (including fmt, vet, test, and lint)

.PHONY: fmt
fmt: ## go fmt
	go fmt ./...

.PHONY: gofumpt
gofumpt: $(GOPATH)/bin/gofumpt ## gofumpt is a stricter alternative to go fmt
	gofumpt -l -w .

.PHONY: gofix
gofix: ## go fix
	go fix ./...

goimports: $(GOPATH)/bin/goimports
	$(GOPATH)/bin/goimports -w -format-only -local github.com/regclient .

.PHONY: vet
vet: ## go vet
	go vet ./...

.PHONY: test
test: ## go test
	go test -cover -race ./...

.PHONY: lint
lint: lint-go lint-goimports lint-md lint-gosec ## Run all linting

.PHONY: lint-go
lint-go: $(GOPATH)/bin/gofumpt $(GOPATH)/bin/staticcheck .FORCE ## Run linting for Go
	$(GOPATH)/bin/staticcheck -checks all ./...
	$(GOPATH)/bin/gofumpt -l -d .
	errors=$$(go fix -diff ./...); if [ "$${errors}" != "" ]; then echo "$${errors}"; exit 1; fi

lint-goimports: $(GOPATH)/bin/goimports
	@if [ -n "$$($(GOPATH)/bin/goimports -l -format-only -local github.com/regclient .)" ]; then \
		echo $(GOPATH)/bin/goimports -d -format-only -local github.com/regclient .; \
		$(GOPATH)/bin/goimports -d -format-only -local github.com/regclient .; \
		exit 1; \
	fi

# excluding types/platform pending resultion to https://github.com/securego/gosec/issues/1116
.PHONY: lint-gosec
lint-gosec: $(GOPATH)/bin/gosec .FORCE ## Run gosec
	$(GOPATH)/bin/gosec -terse -exclude-dir types/platform ./...

.PHONY: lint-md
lint-md: .FORCE ## Run linting for markdown
	docker run --rm -v "$(PWD):/workdir:ro" davidanson/markdownlint-cli2:$(MARKDOWN_LINT_VER) \
	  "**/*.md" "#vendor"

.PHONY: vulnerability-scan
vulnerability-scan: osv-scanner vulncheck-go ## Run all vulnerability scanners

.PHONY: osv-scanner
osv-scanner: $(GOPATH)/bin/osv-scanner .FORCE ## Run OSV Scanner
	$(GOPATH)/bin/osv-scanner scan --config .osv-scanner.toml -r --licenses="Apache-2.0,BSD-3-Clause,MIT,CC-BY-SA-4.0,UNKNOWN" .

.PHONY: vulncheck-go
vulncheck-go: $(GOPATH)/bin/govulncheck .FORCE ## Run govulncheck
	$(GOPATH)/bin/govulncheck ./...

.PHONY: vendor
vendor: ## Vendor Go modules
	go mod vendor

.PHONY: binaries
binaries: $(BINARIES) ## Build Go binaries

bin/%: .FORCE
	CGO_ENABLED=0 go build ${GO_BUILD_FLAGS} -o bin/$* ./cmd/$*

.PHONY: docker
docker: $(IMAGES) ## Build Docker images

docker-%: .FORCE
	docker build -t regclient/$* -f build/Dockerfile.$*$(DOCKERFILE_EXT) $(DOCKER_ARGS) .
	docker build -t regclient/$*:alpine -f build/Dockerfile.$*$(DOCKERFILE_EXT) --target release-alpine $(DOCKER_ARGS) .

.PHONY: oci-image
oci-image: $(addprefix oci-image-,$(COMMANDS)) ## Build reproducible images to an OCI Layout

oci-image-%: bin/regctl .FORCE
	PATH="$(PWD)/bin:$(PATH)" build/oci-image.sh -r scratch -i "$*" -p "$(IMAGE_PLATFORMS)"
	PATH="$(PWD)/bin:$(PATH)" build/oci-image.sh -r alpine  -i "$*" -p "$(IMAGE_PLATFORMS)" -b "alpine:3"

.PHONY: test-docker
test-docker: $(addprefix test-docker-,$(COMMANDS)) ## Build multi-platform docker images (but do not tag)

test-docker-%:
	docker buildx build --platform="$(IMAGE_PLATFORMS)" -f build/Dockerfile.$*.buildkit .
	docker buildx build --platform="$(IMAGE_PLATFORMS)" -f build/Dockerfile.$*.buildkit --target release-alpine .

.PHONY: ci
ci: ci-distribution ci-zot ## Run CI tests against self hosted registries

.PHONY: ci-distribution
ci-distribution:
	docker run --rm -d -p 5000 \
		--label regclient-ci=true --name regclient-ci-distribution \
		-e "REGISTRY_STORAGE_DELETE_ENABLED=true" \
		docker.io/library/registry:$(CI_DISTRIBUTION_VER)
	./build/ci-test.sh -t localhost:$$(docker port regclient-ci-distribution 5000 | head -1 | cut -f2 -d:)/test-ci
	docker stop regclient-ci-distribution

.PHONY: ci-zot
ci-zot:
	docker run --rm -d -p 5000 \
		--label regclient-ci=true --name regclient-ci-zot \
		-v "$$(pwd)/build/zot-config.json:/etc/zot/config.json:ro" \
		ghcr.io/project-zot/zot-linux-amd64:$(CI_ZOT_VER)
	./build/ci-test.sh -t localhost:$$(docker port regclient-ci-zot 5000 | head -1 | cut -f2 -d:)/test-ci
	docker stop regclient-ci-zot

.PHONY: artifacts
artifacts: $(ARTIFACTS) ## Generate artifacts

.PHONY: artifact-pre
artifact-pre:
	mkdir -p artifacts

artifacts/%: artifact-pre .FORCE
	@set -e; \
	target="$*"; \
	command="$${target%%-*}"; \
	platform_ext="$${target#*-}"; \
	platform="$${platform_ext%.*}"; \
	export GOOS="$${platform%%-*}"; \
	export GOARCH="$${platform#*-}"; \
	echo export GOOS=$${GOOS}; \
	echo export GOARCH=$${GOARCH}; \
	echo go build ${GO_BUILD_FLAGS} -o "$@" ./cmd/$${command}/; \
	CGO_ENABLED=0 go build ${GO_BUILD_FLAGS} -o "$@" ./cmd/$${command}/; \
	$(SYFT) scan -q "file:$@" --source-name "$${command}" -o cyclonedx-json >"artifacts/$${command}-$${platform}.cyclonedx.json"; \
	$(SYFT) scan -q "file:$@" --source-name "$${command}" -o spdx-json >"artifacts/$${command}-$${platform}.spdx.json"

.PHONY: clean
clean: ## delete generated content
	[ ! -d artifacts ] || rm -r artifacts
	[ ! -d bin ] || rm -r bin
	[ ! -d output ] || rm -r output
	[ ! -d vendor ] || rm -r vendor

.PHONY: plugin-user
plugin-user:
	mkdir -p ${HOME}/.docker/cli-plugins/
	cp docker-plugin/docker-regclient ${HOME}/.docker/cli-plugins/docker-regctl

.PHONY: plugin-host
plugin-host:
	sudo cp docker-plugin/docker-regclient /usr/libexec/docker/cli-plugins/docker-regctl

.PHONY: util-golang-major
util-golang-major: $(GOPATH)/bin/gomajor ## check for major dependency updates
	$(GOPATH)/bin/gomajor list

.PHONY: util-golang-update
util-golang-update: ## update go module versions
	go get -u -t ./...
	go mod tidy
	[ ! -d vendor ] || go mod vendor

.PHONY: util-release-preview
util-release-preview: $(GOPATH)/bin/gorelease ## preview changes for next release
	git checkout main
	./.github/release.sh -d
	gorelease

.PHONY: util-release-run
util-release-run: ## generate a new release
	git checkout main
	./.github/release.sh

.PHONY: util-version-check
util-version-check: ## check all dependencies for updates
	$(VER_BUMP) check

.PHONY: util-version-update
util-version-update: ## update versions on all dependencies
	$(VER_BUMP) update

$(GOPATH)/bin/gofumpt: .FORCE
	@[ -f "$(GOPATH)/bin/gofumpt" ] \
	&& [ "$$($(GOPATH)/bin/gofumpt -version | cut -f 1 -d ' ')" = "$(GOFUMPT_VER)" ] \
	|| go install mvdan.cc/gofumpt@$(GOFUMPT_VER)

$(GOPATH)/bin/gomajor: .FORCE
	@[ -f "$(GOPATH)/bin/gomajor" ] \
	&& [ "$$($(GOPATH)/bin/gomajor version | grep '^version' | cut -f 2 -d ' ')" = "$(GOMAJOR_VER)" ] \
	|| go install github.com/icholy/gomajor@$(GOMAJOR_VER)

$(GOPATH)/bin/goimports: .FORCE
	@[ -f "$(GOPATH)/bin/goimports" ] && [ "$$(go version | cut -f3 -d' ')" = "$$(go version $(GOPATH)/bin/goimports | cut -f2 -d' ')" ] \
	||	go install golang.org/x/tools/cmd/goimports@latest

$(GOPATH)/bin/gorelease: .FORCE
	@[ -f "$(GOPATH)/bin/gorelease" ] && [ "$$(go version | cut -f3 -d' ')" = "$$(go version $(GOPATH)/bin/gorelease | cut -f2 -d' ')" ] \
	|| go install golang.org/x/exp/cmd/gorelease@latest

$(GOPATH)/bin/gosec: .FORCE
	@[ -f $(GOPATH)/bin/gosec ] \
	&& [ "$$($(GOPATH)/bin/gosec -version | grep '^Version' | cut -f 2 -d ' ')" = "$(GOSEC_VER)" ] \
	|| go install -ldflags '-X main.Version=$(GOSEC_VER) -X main.GitTag=$(GOSEC_VER)' \
	    github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VER)

$(GOPATH)/bin/staticcheck: .FORCE
	@[ -f $(GOPATH)/bin/staticcheck ] \
	&& [ "$$($(GOPATH)/bin/staticcheck -version | cut -f 3 -d ' ' | tr -d '()')" = "$(STATICCHECK_VER)" ] \
	|| go install "honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VER)"

$(GOPATH)/bin/govulncheck: .FORCE
	@[ -f $(GOPATH)/bin/govulncheck ] \
	&& [ $$(go version -m $(GOPATH)/bin/govulncheck | \
		awk -F ' ' '{ if ($$1 == "mod" && $$2 == "golang.org/x/vuln") { printf "%s\n", $$3 } }') = "$(GO_VULNCHECK_VER)" ] \
	|| CGO_ENABLED=0 go install "golang.org/x/vuln/cmd/govulncheck@$(GO_VULNCHECK_VER)"

$(GOPATH)/bin/osv-scanner: .FORCE
	@[ -f $(GOPATH)/bin/osv-scanner ] \
	&& [ "$$(osv-scanner --version | awk -F ': ' '{ if ($$1 == "osv-scanner version") { printf "%s\n", $$2 } }')" = "$(OSV_SCANNER_VER)" ] \
	|| CGO_ENABLED=0 go install "github.com/google/osv-scanner/v2/cmd/osv-scanner@$(OSV_SCANNER_VER)"

.PHONY: help
help: # Display help
	@awk -F ':|##' '/^[^\t].+?:.*?##/ { printf "\033[36m%-30s\033[0m %s\n", $$1, $$NF }' $(MAKEFILE_LIST)
