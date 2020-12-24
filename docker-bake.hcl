// Go version
variable "GO_VERSION" {
  default = "1.15"
}

// GitHub reference as defined in GitHub Actions (eg. refs/head/master))
variable "GITHUB_REF" {
  default = ""
}

target "go-version" {
  args = {
    GO_VERSION = GO_VERSION
  }
}

target "ghaction-docker-meta" {
  tags = ["crazymax/diun:local"]
}

group "default" {
  targets = ["image"]
}

group "validate" {
  targets = ["lint", "vendor-validate"]
}

target "lint" {
  inherits = ["go-version"]
  dockerfile = "./hack/lint.Dockerfile"
  target = "lint"
}

target "vendor-validate" {
  inherits = ["go-version"]
  dockerfile = "./hack/vendor.Dockerfile"
  target = "validate"
}

target "vendor-update" {
  inherits = ["go-version"]
  dockerfile = "./hack/vendor.Dockerfile"
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
  args = {
    GIT_REF = GITHUB_REF
  }
  inherits = ["go-version"]
  target = "artifacts"
  output = ["./dist"]
}

target "artifact-all" {
  inherits = ["artifact"]
  platforms = ["linux/amd64", "linux/arm/v6", "linux/arm/v7", "linux/arm64", "linux/386", "linux/ppc64le", "linux/s390x", "windows/amd64", "windows/386", "darwin/amd64"]
}

target "image" {
  inherits = ["go-version", "ghaction-docker-meta"]
  cache-from = ["crazymax/diun:edge"]
}

target "image-all" {
  inherits = ["image"]
  platforms = ["linux/amd64", "linux/arm/v6", "linux/arm/v7", "linux/arm64", "linux/386", "linux/ppc64le"]
}
