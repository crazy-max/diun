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

Through the [command line](usage/cli.md) with:

```shell
$ diun --config ./diun.yml --test-notif
```

Or within a container:

```shell
$ docker-compose exec diun diun --test-notif
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

If you encounter this kind of warning, you are probably using the [file provider](providers/file.md) containing an
image with an erroneous or empty platform. If the platform is not filled in, it will be deduced automatically from the
information of your operating system on which Diun is running.

In the example below, Diun is running (`diun_x.x.x_windows_i386.zip`) on Windows 10 and tries to analyze the
`crazymax/cloudflared` image with the detected platform (`windows/386)`:

```yaml
- name: crazymax/cloudflared:2020.2.1
  watch_repo: true
```

But this platform is not supported by this image as you can see [on DockerHub](https://hub.docker.com/layers/crazymax/cloudflared/2020.2.1/images/sha256-137eea4e84ec4c6cb5ceb2017b9788dcd7b04f135d756e1f37e3e6673c0dd9d2?context=explore):

!!! warning
    `Fri, 27 Mar 2020 01:20:03 UTC WRN Cannot get remote manifest error="Cannot create image closer: Error choosing image instance: no image found in manifest list for architecture 386, variant \"\", OS windows" image=docker.io/image=crazymax/cloudflared:2020.2.1 provider=file`

You have to force the platform for this image if you are not on a supported platform:

```yaml
- name: crazymax/cloudflared:2020.2.1
  watch_repo: true
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
