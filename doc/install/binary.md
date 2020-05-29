# Installation from binary

## Download

Diun binaries are available in [releases](https://github.com/crazy-max/diun/releases) page.

Choose the archive matching the destination platform and extract diun:

```
wget -qO- https://github.com/crazy-max/diun/releases/download/v2.6.1/diun_2.6.1_linux_x86_64.tar.gz | tar -zxvf - diun
```

## Test

After getting the binary, it can be tested with `./diun --help` or moved to a permanent location.

```
$ ./diun --help
Usage: diun --config=STRING

Docker image update notifier. More info: https://github.com/crazy-max/diun

Flags:
  --help                Show context-sensitive help.
  --version
  --config=STRING       Diun configuration file ($CONFIG).
  --timezone="UTC"      Timezone assigned to Diun ($TZ).
  --log-level="info"    Set log level ($LOG_LEVEL).
  --log-json            Enable JSON logging output ($LOG_JSON).
  --log-caller          Add file:line of the caller to log output ($LOG_CALLER).
  --test-notif          Test notification settings ($TEST_NOTIF).
```

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

You must create your first [configuration](../configuration.md) file in `/etc/diun/diun.yml` and type:

```
chown diun:diun /etc/diun/diun.yml
chmod 644 /etc/diun/diun.yml
```

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
DIUN_DB=/var/lib/diun/diun.db /usr/local/bin/diun --config /etc/diun/diun.yml
```

## Updating to a new version

You can update to a new version of Diun by stopping it, replacing the binary at `/usr/local/bin/diun` and restarting the instance.

If you have carried out the installation steps as described above, the binary should have the generic name `diun`. Do not change this, i.e. to include the version number.
