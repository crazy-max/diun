variable "GO_VERSION" {
  default = null
}

variable "DESTDIR" {
  default = "./bin"
}

group "default" {
  targets = ["test"]
}

target "test" {
  target = "test-coverage"
  output = ["${DESTDIR}/coverage"]
  args = {
    GO_VERSION = GO_VERSION
  }
}

target "lint" {
  target = "lint"
  output = ["type=cacheonly"]
}
