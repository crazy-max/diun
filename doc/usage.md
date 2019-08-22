# Usage

## Command line

`diun --config=CONFIG [<flags>]`

* `--help` : Show help text and exit. _Optional_.
* `--version` : Show version and exit. _Optional_.
* `--config <path>` : Diun YAML configuration file. **Required**. (example: `diun.yml`).
* `--timezone <timezone>` : Timezone assigned to Diun. _Optional_. (default: `UTC`).
* `--log-level <level>` : Log level output. _Optional_. (default: `info`).
* `--log-json` : Enable JSON logging output. _Optional_. (default: `false`).
* `--log-caller` : Enable to add file:line of the caller. _Optional_. (default: `false`).
