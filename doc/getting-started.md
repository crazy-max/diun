# Getting started

## Launch Diun with the Docker provider

Create a `docker-compose.yml` file that uses the official Diun image:

```yaml
version: "3.5"

services:
  diun:
    image: crazymax/diun:latest
    volumes:
      - "./data:/data"
      - "/var/run/docker.sock:/var/run/docker.sock"
    environment:
      - "TZ=Europe/Paris"
      - "LOG_LEVEL=info"
      - "LOG_JSON=false"
      - "DIUN_WATCH_WORKERS=20"
      - "DIUN_WATCH_SCHEDULE=*/30 * * * *"
      - "DIUN_PROVIDERS_DOCKER=true"
      - "DIUN_PROVIDERS_DOCKER_WATCHBYDEFAULT=true"
    restart: always
```

Here we use a minimal configuration to analyze **all running containers** (watch by default enable) of your **local Docker** instance **every 30 minutes**.

That's it. Now you can launch Diun!
