# Installation from binary

## Download

Diun binaries are available on [releases]({{ config.repo_url }}releases) page.

Choose the archive matching the destination platform:

* [diun_{{ git.tag | trim('v') }}_darwin_arm64.tar.gz]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/diun_{{ git.tag | trim('v') }}_darwin_arm64.tar.gz)
* [diun_{{ git.tag | trim('v') }}_darwin_x86_64.tar.gz]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/diun_{{ git.tag | trim('v') }}_darwin_x86_64.tar.gz)
* [diun_{{ git.tag | trim('v') }}_linux_arm64.tar.gz]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/diun_{{ git.tag | trim('v') }}_linux_arm64.tar.gz)
* [diun_{{ git.tag | trim('v') }}_linux_armv5.tar.gz]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/diun_{{ git.tag | trim('v') }}_linux_armv5.tar.gz)
* [diun_{{ git.tag | trim('v') }}_linux_armv6.tar.gz]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/diun_{{ git.tag | trim('v') }}_linux_armv6.tar.gz)
* [diun_{{ git.tag | trim('v') }}_linux_armv7.tar.gz]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/diun_{{ git.tag | trim('v') }}_linux_armv7.tar.gz)
* [diun_{{ git.tag | trim('v') }}_linux_i386.tar.gz]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/diun_{{ git.tag | trim('v') }}_linux_i386.tar.gz)
* [diun_{{ git.tag | trim('v') }}_linux_ppc64le.tar.gz]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/diun_{{ git.tag | trim('v') }}_linux_ppc64le.tar.gz)
* [diun_{{ git.tag | trim('v') }}_linux_s390x.tar.gz]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/diun_{{ git.tag | trim('v') }}_linux_s390x.tar.gz)
* [diun_{{ git.tag | trim('v') }}_linux_x86_64.tar.gz]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/diun_{{ git.tag | trim('v') }}_linux_x86_64.tar.gz)
* [diun_{{ git.tag | trim('v') }}_windows_i386.zip]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/diun_{{ git.tag | trim('v') }}_windows_i386.zip)
* [diun_{{ git.tag | trim('v') }}_windows_x86_64.zip]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/diun_{{ git.tag | trim('v') }}_windows_x86_64.zip)

And extract diun:

```shell
wget -qO- {{ config.repo_url }}releases/download/v{{ git.tag | trim('v') }}/diun_{{ git.tag | trim('v') }}_linux_x86_64.tar.gz | tar -zxvf - diun
```

After getting the binary, it can be tested with [`./diun --help`](../usage/command-line.md#global-options) command
and moved to a permanent location.

## Server configuration

Steps below are the recommended server configuration.

### Prepare environment

Create user to run diun (ex. `diun`)

```shell
groupadd diun
useradd -s /bin/false -d /bin/null -g diun diun
```

### Create required directory structure

```shell
mkdir -p /var/lib/diun
chown diun:diun /var/lib/diun/
chmod -R 750 /var/lib/diun/
mkdir /etc/diun
chown diun:diun /etc/diun
chmod 770 /etc/diun
```

### Configuration

Create your first [configuration](../config/index.md) file in `/etc/diun/diun.yml` and type:

```shell
chown diun:diun /etc/diun/diun.yml
chmod 644 /etc/diun/diun.yml
```

!!! note
    Not required if you want to only rely on environment variables

### Copy binary to global location

```shell
cp diun /usr/local/bin/diun
```

## Running Diun

After the above steps, two options to run Diun:

### 1. Creating a service file (recommended)

See how to create [Linux service](linux-service.md) to start Diun automatically.

### 2. Running from terminal

```shell
DIUN_DB_PATH=/var/lib/diun/diun.db /usr/local/bin/diun serve --config /etc/diun/diun.yml
```

## Updating to a new version

You can update to a new version of Diun by stopping it, replacing the binary at `/usr/local/bin/diun` and restarting
the instance.

If you have carried out the installation steps as described above, the binary should have the generic name `diun`. Do
not change this, i.e. to include the version number.
