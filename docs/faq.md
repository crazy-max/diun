# FAQ

## Test notifications

Through the [command line](usage/cli.md) with:

```shell
$ diun --config ./diun.yml --test-notif
```

Or within a container:

```shell
$ docker run --rm -it -v "$(pwd)/diun.yml:/diun.yml" \
  crazymax/diun:latest --config /diun.yml --test-notif
```

## field docker|swarm uses unsupported type: invalid

If you have the error `failed to decode configuration from file: field docker uses unsupported type: invalid` that's because your `docker`, `swarm` or `kubernetes` provider is not initialized in your configuration:

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

If you encounter this kind of warning, you are probably using the [file provider](providers/file.md) containing an image with an erroneous or empty platform. If the platform is not filled in, it will be deduced automatically from the information of your operating system on which Diun is running.

In the example below, Diun is running (`diun_x.x.x_windows_i386.zip`) on Windows 10 and tries to analyze the `crazymax/cloudflared` image with the detected platform (`windows/386)`:

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
