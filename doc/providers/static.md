# Static provider

* [About](#about)
* [Quick start](#quick-start)

## About

The static provider is the most basic way to analyse Docker images. Nothing special to see here as everything is configured through the [providers field](../configuration.md#providers).

## Quick start

But let's take a look with a simple example:

```yml
db:
  path: diun.db

watch:
  workers: 20
  schedule: "* * * * *"

regopts:
  jfrog:
    username: foo
    password: bar

providers:
  static:
    - name: crazymax/cloudflared
      watch_repo: true
    - name: docker.bintray.io/jfrog/xray-mongo:3.2.6
      regopts_id: jfrog
```

Here we want to analyze all tags of `crazymax/cloudflared` and `docker.bintray.io/jfrog/xray-mongo:3.2.6` tag. Now let's start Diun:

```
$ diun --config diun.yml
Sat, 14 Dec 2019 15:32:23 UTC INF Starting Diun 2.0.0
Sat, 14 Dec 2019 15:32:23 UTC INF Found 2 static provider(s) to analyze...
Sat, 14 Dec 2019 15:32:25 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:latest provider=static
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.11.3 provider=static
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.11.0 provider=static
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.10.1 provider=static
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.9.0 provider=static
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.9.2 provider=static
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.10.2 provider=static
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.11.2 provider=static
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.9.1 provider=static
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=crazymax/cloudflared image=docker.io/crazymax/cloudflared:2019.10.4 provider=static
Sat, 14 Dec 2019 15:32:28 UTC INF New image found id=docker.bintray.io/jfrog/xray-mongo:3.2.6 image=docker.bintray.io/jfrog/xray-mongo:3.2.6 provider=static
Sat, 14 Dec 2019 15:32:28 UTC INF Cron initialized with schedule * * * * *
Sat, 14 Dec 2019 15:32:28 UTC INF Next run in 31 seconds (2019-12-14 15:33:00 +0000 UTC)
```
