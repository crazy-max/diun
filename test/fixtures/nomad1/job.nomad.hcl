job "diun-e2e" {
  type = "service"

  group "app" {
    shutdown_delay = "5s"

    service {
      name = "diun-e2e"
      provider = "nomad"
      tags = [
        "diun.enable=true",
        "diun.metadata.fixture=nomad1",
      ]
    }

    task "busybox" {
      driver = "docker"

      config {
        image = "busybox:1.36.1"
      }
    }
  }
}
