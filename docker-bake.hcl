variable "GO_VERSION" {
  default = "1.17"
}

target "_common" {
  args = {
    GO_VERSION = GO_VERSION
    BUILDKIT_CONTEXT_KEEP_GIT_DIR = 1
  }
}

// Special target: https://github.com/docker/metadata-action#bake-definition
target "docker-metadata-action" {
  tags = ["diun:local"]
}

group "default" {
  targets = ["image-local"]
}

target "binary" {
  inherits = ["_common"]
  target = "binary"
  output = ["./bin"]
}

target "artifact" {
  inherits = ["_common"]
  target = "artifact"
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
    "windows/amd64",
    "windows/arm64"
  ]
}

target "image" {
  inherits = ["_common", "docker-metadata-action"]
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

target "test" {
  inherits = ["_common"]
  target = "test-coverage"
  output = ["."]
}

target "vendor" {
  inherits = ["_common"]
  dockerfile = "./hack/vendor.Dockerfile"
  target = "update"
  output = ["."]
}

target "gen" {
  inherits = ["_common"]
  dockerfile = "./hack/gen.Dockerfile"
  target = "update"
  output = ["."]
}

target "docs" {
  dockerfile = "./hack/docs.Dockerfile"
  target = "release"
  output = ["./site"]
}

target "gomod-outdated" {
  inherits = ["_common"]
  dockerfile = "./hack/vendor.Dockerfile"
  target = "outdated"
  output = ["type=cacheonly"]
}

group "validate" {
  targets = ["lint", "vendor-validate", "gen-validate"]
}

target "lint" {
  inherits = ["_common"]
  dockerfile = "./hack/lint.Dockerfile"
  target = "lint"
  output = ["type=cacheonly"]
}

target "vendor-validate" {
  inherits = ["_common"]
  dockerfile = "./hack/vendor.Dockerfile"
  target = "validate"
  output = ["type=cacheonly"]
}

target "gen-validate" {
  inherits = ["_common"]
  dockerfile = "./hack/gen.Dockerfile"
  target = "validate"
  output = ["type=cacheonly"]
}
