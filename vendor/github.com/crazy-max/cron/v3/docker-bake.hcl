variable "GO_VERSION" {
  default = "1.19"
}

target "_common" {
  args = {
    GO_VERSION = GO_VERSION
  }
}

group "default" {
  targets = ["test"]
}

target "test" {
  inherits = ["_common"]
  target = "test-coverage"
  output = ["."]
}
