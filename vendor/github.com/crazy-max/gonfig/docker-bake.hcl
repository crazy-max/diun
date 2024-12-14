variable "GO_VERSION" {
  default = null
}

target "_common" {
  args = {
    GO_VERSION = GO_VERSION
  }
}

group "default" {
  targets = ["test"]
}

group "validate" {
  targets = ["lint", "vendor-validate"]
}

target "lint" {
  inherits = ["_common"]
  target = "lint"
  output = ["type=cacheonly"]
}

target "vendor-validate" {
  inherits = ["_common"]
  target = "vendor-validate"
  output = ["type=cacheonly"]
}

target "vendor-update" {
  inherits = ["_common"]
  target = "vendor-update"
  output = ["."]
}

target "test" {
  inherits = ["_common"]
  target = "test-coverage"
  output = ["."]
}
