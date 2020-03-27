# Usage

## Command line

`diun --config=CONFIG [<flags>]`

* `--help` : Show help text and exit.
* `--version` : Show version and exit.
* `--config <path>` : Diun YAML configuration file. **Required**. (example: `diun.yml`).
* `--timezone <timezone>` : Timezone assigned to Diun. (default: `UTC`).
* `--log-level <level>` : Log level output. (default: `info`).
* `--log-json` : Enable JSON logging output. (default: `false`).
* `--log-caller` : Enable to add file:line of the caller. (default: `false`).
