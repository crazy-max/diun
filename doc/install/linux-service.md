# Run as service on Debian based distro

## Using systemd

> :warning: Make sure to follow the instructions to [install from binary](binary.md) before.

Run the below command in a terminal:

```
sudo vim /etc/systemd/system/diun.service
```

Copy the sample [diun.service](../../.res/systemd/diun.service).

Change the user, group, and other required startup values following your needs.

Enable and start Diun at boot:

```
sudo systemctl enable diun
sudo systemctl start diun
```

To view logs:

```
journalctl -fu diun.service
```
