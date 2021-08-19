// Go version
variable "GO_VERSION" {
  default = "1.17"
}

target "go-version" {
  args = {
    GO_VERSION = GO_VERSION
  }
}

// protoc version
variable "PROTOC_VERSION" {
  default = "3.17.3"
}

target "protoc-version" {
  args = {
    PROTOC_VERSION = PROTOC_VERSION
  }
}

// GitHub reference as defined in GitHub Actions (eg. refs/head/master))
variable "GITHUB_REF" {
  default = ""
}

target "git-ref" {
  args = {
    GIT_REF = GITHUB_REF
  }
}

// Special target: https://github.com/docker/metadata-action#bake-definition
target "docker-metadata-action" {
  tags = ["crazymax/diun:local"]
}

group "default" {
  targets = ["image-local"]
}

group "validate" {
  targets = ["lint", "vendor-validate", "gen-validate"]
}

target "lint" {
  inherits = ["go-version"]
  dockerfile = "./hack/lint.Dockerfile"
  target = "lint"
  output = ["type=cacheonly"]
}

target "vendor-validate" {
  inherits = ["go-version"]
  dockerfile = "./hack/vendor.Dockerfile"
  target = "validate"
  output = ["type=cacheonly"]
}

target "vendor-update" {
  inherits = ["go-version"]
  dockerfile = "./hack/vendor.Dockerfile"
  target = "update"
  output = ["."]
}

target "gen-validate" {
  inherits = ["go-version", "protoc-version"]
  dockerfile = "./hack/gen.Dockerfile"
  target = "validate"
  output = ["type=cacheonly"]
}

target "gen-update" {
  inherits = ["go-version", "protoc-version"]
  dockerfile = "./hack/gen.Dockerfile"
  target = "update"
  output = ["."]
}

target "test" {
  inherits = ["go-version"]
  dockerfile = "./hack/test.Dockerfile"
  target = "test-coverage"
  output = ["."]
}

target "docs" {
  dockerfile = "./hack/docs.Dockerfile"
  target = "release"
  output = ["./site"]
}

target "artifact" {
  inherits = ["go-version", "git-ref"]
  target = "artifacts"
  output = ["./dist"]
}

target "artifact-all" {
  inherits = ["artifact"]
  platforms = [
    "darwin/amd64",
    "darwin/arm64",
    "linux/386",
    "linux/amd64",
    "linux/arm/v5",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le",
    "linux/riscv64",
    "linux/s390x",
    "windows/386",
    "windows/amd64"
  ]
}

target "image" {
  inherits = ["go-version", "git-ref", "docker-metadata-action"]
}

target "image-local" {
  inherits = ["image"]
  output = ["type=docker"]
}

target "image-all" {
  inherits = ["image"]
  platforms = [
    "linux/386",
    "linux/amd64",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le"
  ]
}
