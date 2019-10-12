# Labels

Docker containers can be configured to be watched using [labels](https://docs.docker.com/config/labels-custom-metadata/):

| Label | Description |
| ----- | ----------- |
| `diun`<br>`diun.enable=true` | Enable Diun |
| `diun=1.2.*` | Shorthand for `diun.enable=true` and `diun.include_tags=1.2.*` |
| `diun.enable=false` | Disable Diun, even if `unlabeled-containers` is `true` |
| `diun.os=...`<br>`diun.arch=...`<br>`diun.watch_repo=true|false`<br>`diun.max_tags=0`<br>`diun.regopts_id=...` | Set the corresponding [configuration option](configuration.md#image) |
| `diun.include_tags=...`<br>`diun.include_tags.1=...`<br>`diun.include_tags.xyz=...` | Add an expression to the `include_tags` option, and set `watch_repo` to `true` if missing |
| `diun.exclude_tags=...`<br>`diun.exclude_tags.1=...`<br>`diun.exclude_tags.xyz=...` | Add an expression to the `exclude_tags` option |
