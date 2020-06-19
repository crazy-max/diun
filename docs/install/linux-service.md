# Run as service on Debian based distro

## Using systemd

!!! warning
    Make sure to follow the instructions to [install from binary](binary.md) before.

To create a new service, paste this content in `/etc/systemd/system/diun.service`:

```
[Unit]
Description=Diun
Documentation={{ config.site_url }}
After=syslog.target
After=network.target

[Service]
RestartSec=2s
Type=simple
User=diun
Group=diun
ExecStart=/usr/local/bin/diun --config /etc/diun/diun.yml --log-level info
Restart=always
Environment=DIUN_DB_PATH=/var/lib/diun/diun.db

[Install]
WantedBy=multi-user.target
```

Change the user, group, and other required startup values following your needs.

Enable and start Diun at boot:

```shell
$ sudo systemctl enable diun
$ sudo systemctl start diun
```

To view logs:

```shell
$ journalctl -fu diun.service
```
