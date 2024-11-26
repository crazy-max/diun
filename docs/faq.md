# FAQ

## Timezone

By default, all interpretation and scheduling is done with your local timezone (`TZ` environment variable).

Cron schedule may also override the timezone to be interpreted in by providing an additional space-separated field
at the beginning of the cron spec, of the form `CRON_TZ=<timezone>`:

```yaml
watch:
  schedule: "CRON_TZ=Asia/Tokyo 0 */6 * * *"
```

## Test notifications

Through the [command line](usage/command-line.md#notif-test) with:

```shell
diun notif test
```

Or within a container:

```shell
docker compose exec diun diun notif test
```

While the test notification might work, it's important to note that, by default, Diun will only notify when the cronjob triggers and not on the first run. Check [.watch](config/watch.md) for `firstCheckNotif`.


## Customize the hostname

The hostname that appears in your notifications is the one associated with the
container if you use the Diun image with `docker run` or `docker compose up -d`.
By default, it's a random string like `d2219b854598`. To change it:

```console
$ docker run --hostname "diun" ...
```

Or if you use Docker Compose:

```yaml
services:
  diun:
    image: crazymax/diun:latest
    hostname: diun
```

## Notification template

The title and body of a notification message can be customized for each notifier through `templateTitle` and
`templateBody` fields except for those rendering _JSON_ or _Env_ like [Amqp](notif/amqp.md),
[MQTT](notif/mqtt.md), [Script](notif/script.md) and [Webhook](notif/webhook.md).

Templating is supported with the following fields:

| Key                             | Description                                                                           |
|---------------------------------|---------------------------------------------------------------------------------------|
| `.Meta.ID`                      | App ID: `diun`                                                                        |
| `.Meta.Name`                    | App Name: `Diun`                                                                      |
| `.Meta.Desc`                    | App description: `Docker image update notifier`                                       |
| `.Meta.URL`                     | App repo URL: `https://github.com/crazy-max/diun`                                     |
| `.Meta.Logo`                    | App logo URL: `https://raw.githubusercontent.com/crazy-max/diun/master/.res/diun.png` |
| `.Meta.Author`                  | App author: `CrazyMax`                                                                |
| `.Meta.Version`                 | App version: `v4.19.0`                                                                |
| `.Meta.UserAgent`               | App user-agent used to talk with registries: `diun/4.19.0 go/1.16 Linux`              |
| `.Meta.Hostname`                | Hostname                                                                              |
| `.Entry.Status`                 | Entry status. Can be `new`, `update`, `unchange`, `skip` or `error`                   |
| `.Entry.Provider`               | [Provider](config/providers.md) used                                                  |
| `.Entry.Image`                  | Docker image name. e.g. `docker.io/crazymax/diun:latest`                              |
| `.Entry.Image.Domain`           | Docker image domain. e.g. `docker.io`                                                 |
| `.Entry.Image.Path`             | Docker image path. e.g. `crazymax/diun`                                               |
| `.Entry.Image.Tag`              | Docker image tag. e.g. `latest`                                                       |
| `.Entry.Image.Digest`           | Docker image digest                                                                   |
| `.Entry.Image.HubLink`          | Docker image hub link (if available). e.g. `https://hub.docker.com/r/crazymax/diun`   |
| `.Entry.Manifest.Name`          | Manifest name. e.g. `docker.io/crazymax/diun`                                         |
| `.Entry.Manifest.Tag`           | Manifest tag. e.g. `latest`                                                           |
| `.Entry.Manifest.MIMEType`      | Manifest MIME type. e.g. `application/vnd.docker.distribution.manifest.list.v2+json`  |
| `.Entry.Manifest.Digest`        | Manifest digest                                                                       |
| `.Entry.Manifest.Created`       | Manifest created date. e.g. `2021-06-20T12:23:56Z`                                    |
| `.Entry.Manifest.DockerVersion` | Version of Docker that was used to build the image. e.g. `20.10.7`                    |
| `.Entry.Manifest.Labels`        | Image labels                                                                          |
| `.Entry.Manifest.Layers`        | Image layers                                                                          |
| `.Entry.Manifest.Platform`      | Platform that the image is runs on. e.g. `linux/amd64`                                |
| `.Entry.Metadata`               | Key-value pair of image metadata specific to each provider                            |

## Authentication against the registry

You can authenticate against the registry through the [`regopts` settings](config/regopts.md) or you can mount
your docker config file `$HOME/.docker/config.json` if you are already connected to the registry with `docker login`:

```yaml
name: diun

services:
  diun:
    image: crazymax/diun:latest
    container_name: diun
    command: serve
    volumes:
      - "./data:/data"
      - "/root/.docker/config.json:/root/.docker/config.json:ro"
      - "/var/run/docker.sock:/var/run/docker.sock"
    environment:
      - "TZ=Europe/Paris"
      - "DIUN_WATCH_SCHEDULE=0 */6 * * *"
      - "DIUN_PROVIDERS_DOCKER=true"
      - "DIUN_PROVIDERS_DOCKER_WATCHBYDEFAULT=true"
    restart: always
```

## field docker|swarm uses unsupported type: invalid

If you have the error `failed to decode configuration from file: field docker uses unsupported type: invalid` that's
because your `docker`, `swarm` or `kubernetes` provider is not initialized in your configuration:

!!! failure
    ```yaml
    providers:
      docker:
    ```

should be:

!!! success
    ```yaml
    providers:
      docker: {}
    ```

## No image found in manifest list for architecture, variant, OS

If you encounter this kind of warning, you are probably using the [file provider](providers/file.md) for an
image with an erroneous or empty platform. If the platform is not filled in, it will be deduced automatically from the
information of your operating system on which Diun is running.

In the example below, Diun is running (`diun_x.x.x_windows_i386.zip`) on Windows 10 and tries to analyze the
`crazymax/cloudflared` image with the detected platform (`windows/386)`:

```yaml
- name: crazymax/cloudflared:2020.2.1
```

But this platform is not supported by this image as you can see [on DockerHub](https://hub.docker.com/layers/crazymax/cloudflared/2020.2.1/images/sha256-137eea4e84ec4c6cb5ceb2017b9788dcd7b04f135d756e1f37e3e6673c0dd9d2?context=explore):

!!! warning
    `Fri, 27 Mar 2020 01:20:03 UTC WRN Cannot get remote manifest error="Cannot create image closer: Error choosing image instance: no image found in manifest list for architecture 386, variant \"\", OS windows" image=docker.io/image=crazymax/cloudflared:2020.2.1 provider=file`

You have to force the platform for this image if you are not on a supported platform:

```yaml
- name: crazymax/cloudflared:2020.2.1
  platform:
    os: linux
    arch: amd64
```

!!! success
    `Fri, 27 Mar 2020 01:24:33 UTC INF New image found image=docker.io/crazymax/cloudflared:2020.2.1 provider=file`

## Too many requests to registry

The error `Cannot create image closer: too many requests to registry` is returned when the HTTP status code returned
by the registry is 429.

This can happen on the DockerHub registry because of the [rate-limited anonymous pulls](https://docs.docker.com/docker-hub/download-rate-limit/).

To solve this you must first be authenticated against the registry through the [`regopts` settings](config/regopts.md): 

```yaml
regopts:
  - name: "docker.io"
    selector: image
    username: foo
    password: bar
```

If this is not enough, tweak the [`schedule` setting](config/watch.md#schedule) with something
like `0 */6 * * *` (every 6 hours).

## Docker Hub rate limits

Docker is now [enforcing Docker Hub pull rate limits](https://www.docker.com/increase-rate-limits). This means you can
make 100 pull image requests per six hours for anonymous usage, and 200 pull image requests per six hours
for free Docker accounts. But this rate limit is not necessarily an indicator on the number of times an image has
actually been downloaded. In fact, their _pulls_ counter/metric is actually a representation of the number of times a
manifest for a particular image has been retrieved.

As you probably know, Diun downloads the manifest of an image from its registry through a `GET` request to be able to
retrieve its inside metadata. Fortunately Diun doesn't perform a `GET` request at each scan but only when an image
has been updated or added on the registry. This allows us not to exceed this rate limit in our situation, but
it also **strongly depends on the number of images you scan**. To increase your pull rate limits you can upgrade
your account to a [Docker Pro or Team subscription](https://www.docker.com/pricing) and authenticate against the
registry through the [`regopts` settings](config/regopts.md): 

```yaml
regopts:
  - name: "docker.io"
    selector: image
    username: foo
    password: bar
```

Or you can tweak the [`schedule` setting](config/watch.md#schedule) with something like `0 */6 * * *` (every 6 hours).

!!! warning
    Also be careful with the `watch_repo` setting as it will fetch manifest for **ALL** tags available for the image.

## Tags sorting when using `watch_repo`

When you use the `watch_repo` setting, Diun will fetch all tags available for
the image. Depending on the registry, order of the tags list can change.

You can use the `sort_tags` setting available for each provider to use a
specific sorting method for the tags list.

* `default`: do not sort and use the expected tags list from the registry
* `reverse`: reverse order for the tags list from the registry
* `lexicographical`: sort the tags list lexicographically
* `semver`: sort the tags list using semantic versioning

Given the following list of tags received from the registry:

```json
[
  "0.1.0",
  "0.4.0",
  "3.0.0-beta.1",
  "3.0.0-beta.4",
  "4",
  "4.0.0",
  "4.0.0-beta.1",
  "4.1.0",
  "4.1.1",
  "4.10.0",
  "4.11.0",
  "4.20",
  "4.20.0",
  "4.20.1",
  "4.3.0",
  "4.3.1",
  "4.9.0",
  "edge",
  "latest"
]
```

Here is the result for `reverse`:

```json
[
  "latest",
  "edge",
  "4.9.0",
  "4.3.1",
  "4.3.0",
  "4.20.1",
  "4.20.0",
  "4.20",
  "4.11.0",
  "4.10.0",
  "4.1.1",
  "4.1.0",
  "4.0.0-beta.1",
  "4.0.0",
  "4",
  "3.0.0-beta.4",
  "3.0.0-beta.1",
  "0.4.0",
  "0.1.0"
]
```

And for `semver`:

```json
[
  "4.20.1",
  "4.20.0",
  "4.20",
  "4.11.0",
  "4.10.0",
  "4.9.0",
  "4.3.1",
  "4.3.0",
  "4.1.1",
  "4.1.0",
  "4.0.0",
  "4",
  "4.0.0-beta.1",
  "3.0.0-beta.4",
  "3.0.0-beta.1",
  "0.4.0",
  "0.1.0",
  "edge",
  "latest"
]
```

## Profiling

Diun provides a simple way to manage runtime/pprof profiling through the
[`--profiler-path` and `--profiler` flags with `serve` command](usage/command-line.md#serve):

```yaml
name: diun

services:
  diun:
    image: crazymax/diun:latest
    container_name: diun
    command: serve
    volumes:
      - "./data:/data"
      - "./profiler:/profiler"
      - "/var/run/docker.sock:/var/run/docker.sock"
    environment:
      - "TZ=Europe/Paris"
      - "LOG_LEVEL=info"
      - "PROFILER_PATH=/profiler"
      - "PROFILER=mem"
      - "DIUN_PROVIDERS_DOCKER=true"
    restart: always
```

The following profilers are available:

* `cpu` enables cpu profiling
* `mem` enables memory profiling
* `alloc` enables memory profiling and changes which type of memory to profile allocations
* `heap` enables memory profiling and changes which type of memory profiling to profile the heap
* `routines` enables goroutine profiling
* `mutex` enables mutex profiling
* `threads` enables thread creation profiling
* `block` enables block (contention) profiling

## Image with digest and `image:tag@digest` format

Analysis of an image with a digest but without tag will be done using `latest`
as a tag which could lead to false positives.

For example `crazymax/diun@sha256:fa80af32a7c61128ffda667344547805b3c5e7721ecbbafd70e35bb7bb7c989f`
is referring to `crazymax/diun:4.24.0` tag, so it's not correct to assume that
we want to analyze `crazymax/diun:latest`.

You can still pin an image to a specific digest and analyze the image if the
tag is specified using the `image:tag@digest` format. Taking the previous
example if we specify `crazymax/diun:4.24.0@sha256:fa80af32a7c61128ffda667344547805b3c5e7721ecbbafd70e35bb7bb7c989f`,
then `crazymax/diun:4.24.0` will be analyzed.
