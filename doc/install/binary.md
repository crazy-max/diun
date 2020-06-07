# Installation from binary

## Download

Diun binaries are available in [releases](https://github.com/crazy-max/diun/releases) page.

Choose the archive matching the destination platform and extract diun:

```
wget -qO- https://github.com/crazy-max/diun/releases/download/v3.0.0/diun_3.0.0_linux_x86_64.tar.gz | tar -zxvf - diun
```

After getting the binary, it can be tested with [`./diun --help`](../getting-started.md#diun-cli) command and moved to a permanent location.

## Server configuration

Steps below are the recommended server configuration.

### Prepare environment

Create user to run diun (ex. `diun`)

```
groupadd diun
useradd -s /bin/false -d /bin/null -g diun diun
```

### Create required directory structure

```
mkdir -p /var/lib/diun
chown diun:diun /var/lib/diun/
chmod -R 750 /var/lib/diun/
mkdir /etc/diun
chown diun:diun /etc/diun
chmod 770 /etc/diun
```

### Configuration

Create your first [configuration](../configuration.md) file in `/etc/diun/diun.yml` and type:

```
chown diun:diun /etc/diun/diun.yml
chmod 644 /etc/diun/diun.yml
```

> 💡 Not required if you want to only rely on environment variables

### Copy binary to global location

```
cp diun /usr/local/bin/diun
```

## Running Diun

After the above steps, two options to run Diun:

### 1. Creating a service file (recommended)

See how to create [Linux service](linux-service.md) to start Diun automatically.

### 2. Running from command-line/terminal

```
DIUN_DB_PATH=/var/lib/diun/diun.db /usr/local/bin/diun --config /etc/diun/diun.yml
```

## Updating to a new version

You can update to a new version of Diun by stopping it, replacing the binary at `/usr/local/bin/diun` and restarting the instance.

If you have carried out the installation steps as described above, the binary should have the generic name `diun`. Do not change this, i.e. to include the version number.
