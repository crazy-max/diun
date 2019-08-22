# Configuration

Here is a YAML structure example:

```yml
db:
  path: diun.db

watch:
  workers: 10
  schedule: "0 0 * * * *"

notif:
  mail:
    enable: false
    host: localhost
    port: 25
    ssl: false
    insecure_skip_verify: false
    username:
    password:
    from:
    to:
  webhook:
    enable: false
    endpoint: http://webhook.foo.com/sd54qad89azd5a
    method: GET
    headers:
      Content-Type: application/json
      Authorization: Token123456
    timeout: 10

regopts:
  someregistryoptions:
    username: foo
    password: bar
    timeout: 20
  onemore:
    username: foo2
    password: bar2
    insecure_tls: true

image:
  # Watch latest tag of crazymax/nextcloud image on docker.io (DockerHub) with registry ID 'someregistryoptions'.
  - name: docker.io/crazymax/nextcloud:latest
    regopts_id: someregistryoptions
  # Watch 4.0.0 tag of jfrog/artifactory-oss image on frog-docker-reg2.bintray.io (Bintray) with registry ID 'onemore'.
  - name: jfrog-docker-reg2.bintray.io/jfrog/artifactory-oss:4.0.0
    regopts_id: onemore
  # Watch coreos/hyperkube image on quay.io (Quay) and assume latest tag.
  - name: quay.io/coreos/hyperkube
  # Watch crazymax/swarm-cronjob image and assume docker.io registry and latest tag.
  # Only include tags matching regexp ^1\.2\..*
  - name: crazymax/swarm-cronjob
    watch_repo: true
    include_tags:
      - ^1\.2\..*
  # Watch portainer/portainer image on docker.io (DockerHub) and assume latest tag
  # Only watch latest 10 tags and include tags matching regexp ^(0|[1-9]\d*)\..*
  - name: docker.io/portainer/portainer
    watch_repo: true
    max_tags: 10
    include_tags:
      - ^(0|[1-9]\d*)\..*
  # Watch alpine image (library) and assume docker.io registry and latest tag.
  # Only check linux/arm64v8 image
  - name: alpine
    watch_repo: true
    os: linux
    arch: arm64v8
```

## db

* `db`
  * `path`: Path to Bolt database file where images manifests are stored (default: `diun.db`). Environment var `DIUN_DB` override this value.

## watch

* `watch`
  * `workers`: Maximum number of workers that will execute tasks concurrently. _Optional_. (default: `10`).
  * `schedule`: [CRON expression](https://godoc.org/github.com/crazy-max/cron#hdr-CRON_Expression_Format) to schedule Diun watcher. _Optional_. (default: `0 0 * * * *`).

## notif

* `notif`
  * `mail`
    * `enable`: Enable email reports (default: `false`).
    * `host`: SMTP server host (default: `localhost`). **required**
    * `port`: SMTP server port (default: `25`). **required**
    * `ssl`: SSL defines whether an SSL connection is used. Should be false in most cases since the auth mechanism should use STARTTLS (default: `false`).
    * `insecure_skip_verify`: Controls whether a client verifies the server's certificate chain and host name (default: `false`).
    * `username`: SMTP username.
    * `password`: SMTP password.
    * `from`: Sender email address. **required**
    * `to`: Recipient email address. **required**
  * `webhook`
    * `enable`: Enable webhook notification (default: `false`).
    * `endpoint`: URL of the HTTP request. **required**
    * `method`: HTTP method (default: `GET`). **required**
    * `headers`: Map of additional headers to be sent.
    * `timeout`: Timeout specifies a time limit for the request to be made. (default: `10`).

## regopts

* `regopts`: Map of registry options to use with images. Key is the ID and value is a struct with the following fields:
  * `username`: Registry username.
  * `password`: Registry password.
  * `timeout`: Timeout is the maximum amount of time for the TCP connection to establish. 0 means no timeout (default: `10`).
  * `insecure_tls`: Allow contacting docker registry over HTTP, or HTTPS with failed TLS verification (default: `false`).

## image

* `image`: Slice of image to watch with the following fields:
  * `name`: Docker image name to watch using `registry/path:tag` format. If registry is omitted, `docker.io` will be used and if tag is omitted, `latest` will be used. **required**
  * `os`: OS to use. _Optional_. (default: `linux`).
  * `arch`: Architecture to use. _Optional_. (default: `amd64`).
  * `regopts_id`: Registry options ID from `regopts` to use.
  * `watch_repo`: Watch all tags of this `image` repository (default: `false`).
  * `max_tags`: Maximum number of tags to watch if `watch_repo` enabled. 0 means all of them (default: `0`).
  * `include_tags`: List of regular expressions to include tags. Can be useful if you enable `watch_repo`.
  * `exclude_tags`: List of regular expressions to exclude tags. Can be useful if you enable `watch_repo`.
