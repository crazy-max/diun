# Release v0.11.2

Features:

- Add support for regctl config in XDG and APPDATA. ([PR 1038][pr-1038])
- Add `ImageWithBlobReaderHook` for callbacks per layer when copying an image. ([PR 1046][pr-1046])

Fixes:

- Do not sign released images multiple times. ([PR 1027][pr-1027])
- regctl/action update for path fix. ([PR 1031][pr-1031])
- Remove default values from regctl config. ([PR 1039][pr-1039])
- Apply Go modernizations with `go fix` from 1.26.0. ([PR 1053][pr-1053])
- Adjust test repo names to avoid races. ([PR 1054][pr-1054])
- Automatically upgrade goimports and gorelease. ([PR 1056][pr-1056])

Other Changes:

- Add `REGCTL_CONFIG` to `regctl` help messages. ([PR 1037][pr-1037])
- Go upgrade fixes CVE-2025-68121, govulncheck indicates this project is not vulnerable. ([PR 1047][pr-1047])

Contributors:

- @sudo-bmitch
- @vrajashkr

[pr-1027]: https://github.com/regclient/regclient/pull/1027
[pr-1031]: https://github.com/regclient/regclient/pull/1031
[pr-1037]: https://github.com/regclient/regclient/pull/1037
[pr-1038]: https://github.com/regclient/regclient/pull/1038
[pr-1039]: https://github.com/regclient/regclient/pull/1039
[pr-1047]: https://github.com/regclient/regclient/pull/1047
[pr-1046]: https://github.com/regclient/regclient/pull/1046
[pr-1053]: https://github.com/regclient/regclient/pull/1053
[pr-1054]: https://github.com/regclient/regclient/pull/1054
[pr-1056]: https://github.com/regclient/regclient/pull/1056
