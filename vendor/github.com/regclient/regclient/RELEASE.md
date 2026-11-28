# Release Notes

## Release v0.11.7

Security:

- Upgrading Go fixes CVE-2026-111943 ([PR 1140][pr-1140])

Contributors:

- @sudo-bmitch

[pr-1140]: https://github.com/regclient/regclient/pull/1140

## Release v0.11.6

Features:

- image mod support for user, workdir, and author. ([PR 1101][pr-1101])
- Support RC releases. ([PR 1126][pr-1126])

Fixes:

- Add limits and controls to ReadAll. ([PR 1096][pr-1096])
- Skip local check on requests going through a proxy. ([PR 1099][pr-1099])
- URLs passed to ProxyFromEnvironment require a scheme. ([PR 1104][pr-1104])
- Avoid a panic if http.DefaultTransport is altered. ([PR 1109][pr-1109])
- Handle docker schema v1  with ocidir. ([PR 1111][pr-1111])
- Do not checkout main branch for releases. ([PR 1127][pr-1127])
- Pin Go version for regctl mod commands. ([PR 1130][pr-1130])
- Improve reproducibility with tags on main branch. ([PR 1131][pr-1131])
- Do not sign SBOMs. ([PR 1133][pr-1133])

Other changes:

- Add copyright headers. ([PR 1102][pr-1102])
- Improve error messages in ocidir scheme. ([PR 1106][pr-1106])
- Upgrade osv-scanner to v2. ([PR 1115][pr-1115])
- Remove go reportcard. ([PR 1119][pr-1119])

Contributors:

- @sudo-bmitch

[pr-1096]: https://github.com/regclient/regclient/pull/1096
[pr-1099]: https://github.com/regclient/regclient/pull/1099
[pr-1101]: https://github.com/regclient/regclient/pull/1101
[pr-1102]: https://github.com/regclient/regclient/pull/1102
[pr-1104]: https://github.com/regclient/regclient/pull/1104
[pr-1106]: https://github.com/regclient/regclient/pull/1106
[pr-1109]: https://github.com/regclient/regclient/pull/1109
[pr-1111]: https://github.com/regclient/regclient/pull/1111
[pr-1115]: https://github.com/regclient/regclient/pull/1115
[pr-1119]: https://github.com/regclient/regclient/pull/1119
[pr-1126]: https://github.com/regclient/regclient/pull/1126
[pr-1127]: https://github.com/regclient/regclient/pull/1127
[pr-1130]: https://github.com/regclient/regclient/pull/1130
[pr-1131]: https://github.com/regclient/regclient/pull/1131
[pr-1133]: https://github.com/regclient/regclient/pull/1133

## Release v0.11.5

Security:

- Prevent https to non-https downgrades and localhost redirects. ([PR 1093][pr-1093])
- Forbid sending auth on redirects. ([PR 1095][pr-1095])

Features:

- Add regbot `manifest.descriptor` to the sandbox. ([PR 1091][pr-1091])

Contributors:

- @GimmyDatBeeR
- @sudo-bmitch

[pr-1091]: https://github.com/regclient/regclient/pull/1091
[pr-1093]: https://github.com/regclient/regclient/pull/1093
[pr-1095]: https://github.com/regclient/regclient/pull/1095

## Release v0.11.4

Security:

- Validate server URL in token auth. ([PR 1075][pr-1075])
- Upgrading Go fixes CVE-2026-33814  and CVE-2026-39836, other vulnerabilities fixed in 1.26.3 were not called by this project. ([PR 1084][pr-1084])

Features:

- Support scanning OCI Layout for referrers. ([PR 1074][pr-1074])
- Add created timestamp in OCI Layout entries. ([PR 1081][pr-1081])
- `tag.ls` now accepts the same pagination parameters as `repo.ls`. ([PR 1086][pr-1086])

Fixes:

- Push tags for minor and major releases on Docker Hub. ([PR 1087][pr-1087])

Contributors:

- @ffried
- @sudo-bmitch

[pr-1074]: https://github.com/regclient/regclient/pull/1074
[pr-1075]: https://github.com/regclient/regclient/pull/1075
[pr-1081]: https://github.com/regclient/regclient/pull/1081
[pr-1084]: https://github.com/regclient/regclient/pull/1084
[pr-1086]: https://github.com/regclient/regclient/pull/1086
[pr-1087]: https://github.com/regclient/regclient/pull/1087

## Release v0.11.3

Security:

- Go 1.26.2 release fixes CVE-2026-32280 ([PR 1072][pr-1072])
- Go 1.26.2 release fixes CVE-2026-32281 ([PR 1072][pr-1072])
- Go 1.26.2 release fixes CVE-2026-32283 ([PR 1072][pr-1072])
- Go 1.26.2 release fixes CVE-2026-32288 ([PR 1072][pr-1072])
- Go 1.26.2 release fixes CVE-2026-33810 ([PR 1072][pr-1072])

Features:

- Add support for pushing digest with tags. ([PR 1062][pr-1062])
- Handle OCI-Tag headers with comma separators. ([PR 1070][pr-1070])

Contributors:

- @sudo-bmitch

[pr-1062]: https://github.com/regclient/regclient/pull/1062
[pr-1070]: https://github.com/regclient/regclient/pull/1070
[pr-1072]: https://github.com/regclient/regclient/pull/1072

## Release v0.11.2

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

## Release v0.11.1

Security:

- Go 1.25.5 fixes CVE-2025-61729 ([PR 1025][pr-1025])
- Go 1.25.5 fixes CVE-2025-61727 ([PR 1025][pr-1025])

Fixes:

- Correct selection of previous tag for releases. ([PR 1023][pr-1023])
- Make sure ContentLength is correctly set in the request. ([PR 1024][pr-1024])

Contributors:

- @sudo-bmitch

[pr-1023]: https://github.com/regclient/regclient/pull/1023
[pr-1024]: https://github.com/regclient/regclient/pull/1024
[pr-1025]: https://github.com/regclient/regclient/pull/1025

## Release v0.11.0

Features:

- Build artifacts for riscv64. ([PR 1011][pr-1011])
- Generate FreeBSD amd64 binaries. ([PR 1013][pr-1013])
- Add support for cosign v3 bundles. ([PR 1018][pr-1018])

Fixes:

- Fix ECR Helper version pin. ([PR 1017][pr-1017])
- Fix the cosign use-signing-config flag. ([PR 1019][pr-1019])
- Improve reproducibility in Dockerfiles. ([PR 1020][pr-1020])

Other Changes:

- Add a policy for LLM generated contributions. ([PR 1016][pr-1016])

Contributors:

- @ffgan
- @sudo-bmitch

[pr-1011]: https://github.com/regclient/regclient/pull/1011
[pr-1013]: https://github.com/regclient/regclient/pull/1013
[pr-1016]: https://github.com/regclient/regclient/pull/1016
[pr-1017]: https://github.com/regclient/regclient/pull/1017
[pr-1018]: https://github.com/regclient/regclient/pull/1018
[pr-1019]: https://github.com/regclient/regclient/pull/1019
[pr-1020]: https://github.com/regclient/regclient/pull/1020

## Release v0.10.0

Features:

- Feat: Support DOCKER_AUTH_CONFIG variable. ([PR 996][pr-996])
- Feat: Add regctl repo copy. ([PR 997][pr-997])
- Feat: regsync support for semantic versioning(semver) for matching tags ([PR 1005][pr-1005])
- Feat: Add `tagSets` to regsync. ([PR 1008][pr-1008])

Changes:

- Chore: Add go:fix lines to deprecated code. ([PR 994][pr-994])
- Chore: Add gofumpt to the build. ([PR 995][pr-995])
- Chore: Remove the unused bps field. ([PR 998][pr-998])
- Fix: Handle semver compare of numeric prerelease ([PR 1007][pr-1007])

Security:

- CVE-2025-58187: Fixed with Go upgrade (<https://osv.dev/GO-2025-4007>).
- CVE-2025-58189: Fixed with Go upgrade (<https://osv.dev/GO-2025-4008>).
- CVE-2025-61723: Fixed with Go upgrade (<https://osv.dev/GO-2025-4009>).
- CVE-2025-47912: Fixed with Go upgrade (<https://osv.dev/GO-2025-4010>).
- CVE-2025-58185: Fixed with Go upgrade (<https://osv.dev/GO-2025-4011>).
- CVE-2025-58186: Fixed with Go upgrade (<https://osv.dev/GO-2025-4012>).
- CVE-2025-58188: Fixed with Go upgrade (<https://osv.dev/GO-2025-4013>).
- CVE-2025-58183: Fixed with Go upgrade (<https://osv.dev/GO-2025-4014>).
- CVE-2025-9230: Fixed with Alpine image upgrade.
- CVE-2025-9230: Fixed with Alpine image upgrade.
- CVE-2025-9232: Fixed with Alpine image upgrade.
- CVE-2025-9232: Fixed with Alpine image upgrade.
- CVE-2025-9231: Fixed with Alpine image upgrade.
- CVE-2025-9231: Fixed with Alpine image upgrade.

Contributors:

- @daimoniac
- @sudo-bmitch

[pr-994]: https://github.com/regclient/regclient/pull/994
[pr-995]: https://github.com/regclient/regclient/pull/995
[pr-996]: https://github.com/regclient/regclient/pull/996
[pr-997]: https://github.com/regclient/regclient/pull/997
[pr-998]: https://github.com/regclient/regclient/pull/998
[pr-1005]: https://github.com/regclient/regclient/pull/1005
[pr-1007]: https://github.com/regclient/regclient/pull/1007
[pr-1008]: https://github.com/regclient/regclient/pull/1008

## Release v0.9.2

Security:

- xz upgrade fixes CVE-2025-58058 ([PR 989][pr-989])

Miscellaneous:

- Fix CLI lint errors. ([PR 983][pr-983])
- Cleanup version output. ([PR 985][pr-985])
- Dockerfile cleanup. ([PR 986][pr-986])

Contributors:

- @sudo-bmitch

[pr-983]: https://github.com/regclient/regclient/pull/983
[pr-985]: https://github.com/regclient/regclient/pull/985
[pr-986]: https://github.com/regclient/regclient/pull/986
[pr-989]: https://github.com/regclient/regclient/pull/989

## Release v0.9.1

Features:

- Allow relative urls in bearer auth. ([PR 963][pr-963])
- Add "ns" query param to registry mirror requests. ([PR 976][pr-976])

Miscellaneous:

- Update to SLSA v1 provenance. ([PR 968][pr-968])
- Add a "make clean" command. ([PR 969][pr-969])

Contributors:

- @sudo-bmitch
- @wjordan

[pr-963]: https://github.com/regclient/regclient/pull/963
[pr-968]: https://github.com/regclient/regclient/pull/968
[pr-969]: https://github.com/regclient/regclient/pull/969
[pr-976]: https://github.com/regclient/regclient/pull/976

## Release v0.9.0

Breaking:

- Drop support for 3rd Go release because of upstream forced upgrades (see <https://github.com/golang/go/issues/69095>). ([PR 948][pr-948])

Features:

- Add a script to reproduce regclient images. ([PR 940][pr-940])
- Support IPv6 hosts. ([PR 956][pr-956])

Fixes:

- Convert  docker attestations built with `oci-artifact=true`. ([PR 949][pr-949])
- Allow duplicate keys in yaml config. ([PR 952][pr-952])

Miscellaneous:

- Migrate yaml library. ([PR 947][pr-947])
- Convert the build to use OCI style attestations. ([PR 950][pr-950])

Contributors:

- @JimmyMa
- @sudo-bmitch

[pr-940]: https://github.com/regclient/regclient/pull/940
[pr-947]: https://github.com/regclient/regclient/pull/947
[pr-948]: https://github.com/regclient/regclient/pull/948
[pr-949]: https://github.com/regclient/regclient/pull/949
[pr-950]: https://github.com/regclient/regclient/pull/950
[pr-952]: https://github.com/regclient/regclient/pull/952
[pr-956]: https://github.com/regclient/regclient/pull/956

## Release v0.8.3

Features:

- Add `ref.AddDigest` method that does not unset the tag. ([PR 910][pr-910])
- Adding a `regctl registry whoami` command. ([PR 912][pr-912])
- Improve `regctl image check-base` output. ([PR 917][pr-917])
- regsync option to abort on errors. ([PR 924][pr-924])
- Improve fallback tag handling. ([PR 925][pr-925])
- regctl flag to ignore missing images on delete. ([PR 930][pr-930])

Fixes:

- Validate registry names. ([PR 911][pr-911])
- Escape regexp example. ([PR 920][pr-920])
- Auth header parsing. ([PR 936][pr-936])

Changes:

- Update supported Go releases to 1.22, 1.23, and 1.24. ([PR 909][pr-909])
- Modernize Go to the 1.22 specs. ([PR 910][pr-910])
- Refactor cobra commands. ([PR 915][pr-915])
- Include Docker Hub repository documentation. ([PR 918][pr-918])
- Move documentation pointers to the website. ([PR 939][pr-939])

Contributors:

- @sudo-bmitch

[pr-909]: https://github.com/regclient/regclient/pull/909
[pr-910]: https://github.com/regclient/regclient/pull/910
[pr-911]: https://github.com/regclient/regclient/pull/911
[pr-912]: https://github.com/regclient/regclient/pull/912
[pr-915]: https://github.com/regclient/regclient/pull/915
[pr-917]: https://github.com/regclient/regclient/pull/917
[pr-918]: https://github.com/regclient/regclient/pull/918
[pr-920]: https://github.com/regclient/regclient/pull/920
[pr-924]: https://github.com/regclient/regclient/pull/924
[pr-925]: https://github.com/regclient/regclient/pull/925
[pr-930]: https://github.com/regclient/regclient/pull/930
[pr-936]: https://github.com/regclient/regclient/pull/936
[pr-939]: https://github.com/regclient/regclient/pull/939

## Release v0.8.2

This fixes a regression in v0.8.1 for users authenticating using a refresh token.

Fixes:

- Allow authentication with a token. ([PR 908][pr-908])

Contributors:

- @sudo-bmitch

[pr-908]: https://github.com/regclient/regclient/pull/908

## Release v0.8.1

Security:

- Go v1.23.6 fixes CVE-2025-22866. ([PR 906][pr-906])

Features:

- Improve regctl arg completion. ([PR 895][pr-895])
- Add cobra command for documentation. ([PR 900][pr-900])

Fixes:

- Do not request offline refresh token. ([PR 893][pr-893])
- Ignore unsupported entries in docker config. ([PR 894][pr-894])
- Align log levels with slog. ([PR 901][pr-901])
- Interval overrides a default schedule in regsync and regbot. ([PR 904][pr-904])

Miscellaneous:

- Adding a logo. ([PR 889][pr-889])

Contributors:

- @obaibula
- @sudo-bmitch

[pr-889]: https://github.com/regclient/regclient/pull/889
[pr-893]: https://github.com/regclient/regclient/pull/893
[pr-894]: https://github.com/regclient/regclient/pull/894
[pr-895]: https://github.com/regclient/regclient/pull/895
[pr-900]: https://github.com/regclient/regclient/pull/900
[pr-901]: https://github.com/regclient/regclient/pull/901
[pr-904]: https://github.com/regclient/regclient/pull/904
[pr-906]: https://github.com/regclient/regclient/pull/906

## Release v0.8.0

## Highlights

There are three headline changes in this release: slog support, external referrers, and deprecating legacy packages.

This release switches from logrus to slog.
Migration methods are included to minimize the impact on existing users.
Anyone parsing the logging output from regctl, regsync, and regbot will notice the format has changed.

External referrers allow referrers to be pushed and pulled from a separate repository from the subject image.
This feature requires users to provide the external repository themselves since a registry has no way to communicate this to the user.
An example use case of this feature are third parties, like security scanners, providing attestations of images they do not control.

Legacy packages have been disabled by default and will eventually be removed.
To continue using legacy packages until their removal, you may compile with `-tags legacy`.

## Breaking

- Breaking: Warning handlers switched from `logrus` to `slog` which will only impact those with a custom warning handler. ([PR 847][pr-847])
- Breaking: Disable legacy packages by default. ([PR 852][pr-852])

## Features

- Feat: Refactor logging to use log/slog. ([PR 847][pr-847])
- Feat: Switch regbot to slog. ([PR 849][pr-849])
- Feat: Switch regctl to slog. ([PR 850][pr-850])
- Feat: Switch regsync to slog. ([PR 851][pr-851])
- Feat: Move logrus calls into files excluded by wasm. ([PR 853][pr-853])
- Feat: Allow plus in ocidir path. ([PR 856][pr-856])
- Feat: Support referrers in an external repository. ([PR 866][pr-866])
- Feat: Image mod environment variables. ([PR 867][pr-867])
- Feat: Include source in referrers response. ([PR 870][pr-870])
- Feat: Add external flag to regctl artifact put. ([PR 873][pr-873])
- Feat: Copy image with external referrers. ([PR 874][pr-874])
- Feat: Document community maintained packages. ([PR 878][pr-878])
- Feat: Support external referrers in regsync. ([PR 881][pr-881])
- Feat: Support incomplete subject descriptor. ([PR 885][pr-885])

## Fixes

- Fix: Inject release notes by file. ([PR 854][pr-854])
- Fix: Platform test for darwin/macos should not add variant. ([PR 879][pr-879])
- Fix: Handle repeated digest in copy with external referrers. ([PR 882][pr-882])

## Chores

- Chore: Improve error message when inspecting artifacts. ([PR 862][pr-862])
- Chore: Remove unused short arg parameters. ([PR 877][pr-877])

## Contributors

- @sudo-bmitch

[pr-847]: https://github.com/regclient/regclient/pull/847
[pr-849]: https://github.com/regclient/regclient/pull/849
[pr-850]: https://github.com/regclient/regclient/pull/850
[pr-851]: https://github.com/regclient/regclient/pull/851
[pr-852]: https://github.com/regclient/regclient/pull/852
[pr-853]: https://github.com/regclient/regclient/pull/853
[pr-854]: https://github.com/regclient/regclient/pull/854
[pr-856]: https://github.com/regclient/regclient/pull/856
[pr-862]: https://github.com/regclient/regclient/pull/862
[pr-866]: https://github.com/regclient/regclient/pull/866
[pr-867]: https://github.com/regclient/regclient/pull/867
[pr-870]: https://github.com/regclient/regclient/pull/870
[pr-873]: https://github.com/regclient/regclient/pull/873
[pr-874]: https://github.com/regclient/regclient/pull/874
[pr-877]: https://github.com/regclient/regclient/pull/877
[pr-878]: https://github.com/regclient/regclient/pull/878
[pr-879]: https://github.com/regclient/regclient/pull/879
[pr-881]: https://github.com/regclient/regclient/pull/881
[pr-882]: https://github.com/regclient/regclient/pull/882
[pr-885]: https://github.com/regclient/regclient/pull/885

## Release v0.7.2

Breaking Changes:

The breaking changes are to internal methods and undocumented features that should not be encountered by users.

- Update scheme to use pqueue instead of throttle. ([PR 803][pr-803])
- Removes an undocumented API for deleting images from Hub. ([PR 803][pr-803])
- `config.Host.Throttle()` has been removed. Use `scheme.Throttler` instead. ([PR 813][pr-813])

Features:

- Significant refactor of http APIs to speed up image copies. ([PR 803][pr-803])
- Add a priority queue for network requests. ([PR 803][pr-803])
- Move logging into transport and rework backoff. ([PR 803][pr-803])
- Remove default rate limit. ([PR 803][pr-803])
- Add priority queue algorithm and reorder image copy steps. ([PR 803][pr-803])
- Consolidate warnings. ([PR 810][pr-810])
- Limit number of retries for a request. ([PR 812][pr-812])
- Add default host config. ([PR 821][pr-821])

Fixes:

- Update GHA output generating steps. ([PR 800][pr-800])
- Lookup referrers when registry does not give digest with head. ([PR 801][pr-801])
- Support auth on redirect. ([PR 805][pr-805])
- Prevent data race when reading blob and seeking. ([PR 814][pr-814])
- Detect integer overflows on type conversion. ([PR 830][pr-830])
- Add a warning if syft is not installed. ([PR 841][pr-841])
- Race condition in the pqueue tests. ([PR 843][pr-843])
- Dedup warnings on image mod. ([PR 846][pr-846])

Chores:

- Update staticcheck and fix linter warnings for Go 1.23. ([PR 804][pr-804])
- Remove digest calculation from reghttp. ([PR 803][pr-803])
- Remove `ReqPerSec` in tests. ([PR 806][pr-806])
- Move throttle from `config` to `reghttp`. ([PR 813][pr-813])
- Refactoring to remove globals in regsync. ([PR 815][pr-815])
- Refactor to remove globals in regbot. ([PR 816][pr-816])
- Remove throttle package. ([PR 817][pr-817])
- Update version-bump config for processors. ([PR 828][pr-828])
- Update config to use yaml anchors and aliases ([PR 829][pr-829])
- Do not automatically assign myself to GitHub issues. ([PR 831][pr-831])
- Remove OpenSSF scorecard and best practices. ([PR 832][pr-832])
- Update docker image base filesystem. ([PR 837][pr-837])

Contributors:

- @sudo-bmitch

[pr-800]: https://github.com/regclient/regclient/pull/800
[pr-801]: https://github.com/regclient/regclient/pull/801
[pr-804]: https://github.com/regclient/regclient/pull/804
[pr-803]: https://github.com/regclient/regclient/pull/803
[pr-805]: https://github.com/regclient/regclient/pull/805
[pr-806]: https://github.com/regclient/regclient/pull/806
[pr-810]: https://github.com/regclient/regclient/pull/810
[pr-812]: https://github.com/regclient/regclient/pull/812
[pr-813]: https://github.com/regclient/regclient/pull/813
[pr-814]: https://github.com/regclient/regclient/pull/814
[pr-815]: https://github.com/regclient/regclient/pull/815
[pr-816]: https://github.com/regclient/regclient/pull/816
[pr-817]: https://github.com/regclient/regclient/pull/817
[pr-821]: https://github.com/regclient/regclient/pull/821
[pr-828]: https://github.com/regclient/regclient/pull/828
[pr-829]: https://github.com/regclient/regclient/pull/829
[pr-830]: https://github.com/regclient/regclient/pull/830
[pr-831]: https://github.com/regclient/regclient/pull/831
[pr-832]: https://github.com/regclient/regclient/pull/832
[pr-837]: https://github.com/regclient/regclient/pull/837
[pr-841]: https://github.com/regclient/regclient/pull/841
[pr-843]: https://github.com/regclient/regclient/pull/843
[pr-846]: https://github.com/regclient/regclient/pull/846

## Release v0.7.1

[PR 798][pr-798] fixes an issue where a malicious registry could return a pinned manifest different from the request.
Commands like `regctl manifest get $image@$digest` will now verify the digest of the returned manifest matches the request rather than the registry headers.

Security updates:

- Validate the digest of the ref when provided. ([PR 798][pr-798])

Features:

- Add a `WithDockerCredsFile() regclient.Opt`. ([PR 784][pr-784])
- Add `regctl artifact get --config` option to only return the config. ([PR 795][pr-795])

Fixes:

- Detect `amd64` variants for `--platform local`. ([PR 782][pr-782])
- Mod tracking of changed manifests. ([PR 783][pr-783])
- Tar path separator should always be a `/`. ([PR 788][pr-788])

Other Changes:

- Remove docker build cache in GHA. ([PR 780][pr-780])

Contributors:

- @mmonaco
- @stormyyd
- @sudo-bmitch

[pr-780]: https://github.com/regclient/regclient/pull/780
[pr-782]: https://github.com/regclient/regclient/pull/782
[pr-783]: https://github.com/regclient/regclient/pull/783
[pr-784]: https://github.com/regclient/regclient/pull/784
[pr-788]: https://github.com/regclient/regclient/pull/788
[pr-795]: https://github.com/regclient/regclient/pull/795
[pr-798]: https://github.com/regclient/regclient/pull/798

## Release v0.7.0

CVEs:

- CVE-2024-24790 fix included with Go 1.22.4 upgrade. ([PR 762][pr-762])
- CVE-2024-24791 fix included with Go 1.22.5 upgrade. ([PR 777][pr-777])

Breaking:

- `regctl registry set` and `regctl registry login` will return a non-zero if the ping fails. ([PR 751][pr-751])
- Removed `WithFS` which required access to an internal interface to use. ([PR 772][pr-772])

Features:

- Add an experimental `regctl ref` command. ([PR 765][pr-765])
- Support digest algorithms beyond sha256. ([PR 776][pr-776])
- Support modifying the digest algorithm on an image. ([PR 776][pr-776])
- Experimental support for pushing tagged manifests with different digest algorithms. ([PR 778][pr-778])

Fixes:

- Prevent panic on interrupted image mod. ([PR 746][pr-746])
- Enable deletion on olareg tests. ([PR 758][pr-758])
- Allow `~` (tilde) in ocidir reference paths. ([PR 763][pr-763])
- Allow well known architectures as a platform. ([PR 771][pr-771])
- Validate digests before calling methods that could panic. ([PR 776][pr-776])

Other changes:

- Refactor pulling manifests by platform. ([PR 768][pr-768])
- Cleanup Dockerfile linter warnings. ([PR 770][pr-770])
- Enable docker caching of GHA builds. ([PR 773][pr-773])
- Include a contributor list in the readme. ([PR 774][pr-774])

Contributors:

- @sudo-bmitch
- @thesayyn

[pr-746]: https://github.com/regclient/regclient/pull/746
[pr-751]: https://github.com/regclient/regclient/pull/751
[pr-758]: https://github.com/regclient/regclient/pull/758
[pr-762]: https://github.com/regclient/regclient/pull/762
[pr-763]: https://github.com/regclient/regclient/pull/763
[pr-765]: https://github.com/regclient/regclient/pull/765
[pr-768]: https://github.com/regclient/regclient/pull/768
[pr-770]: https://github.com/regclient/regclient/pull/770
[pr-771]: https://github.com/regclient/regclient/pull/771
[pr-772]: https://github.com/regclient/regclient/pull/772
[pr-773]: https://github.com/regclient/regclient/pull/773
[pr-774]: https://github.com/regclient/regclient/pull/774
[pr-776]: https://github.com/regclient/regclient/pull/776
[pr-777]: https://github.com/regclient/regclient/pull/777
[pr-778]: https://github.com/regclient/regclient/pull/778

## Release v0.6.1

CVEs:

- Go update fixes CVE-2024-24788. ([PR 739][pr-739])

Breaking:

- pkg/archive.Compress no longer decompresses the input. ([PR 732][pr-732])

Features:

- Add the `regclient.ImageConfig` method. ([PR 706][pr-706])
- Add ability to modify the layer compression. ([PR 730][pr-730])
- Add support for zstd compressed layers. ([PR 732][pr-732])
- Add image mod ability to append layers to an image. ([PR 736][pr-736])
- `regctl image mod` add layer from directory. ([PR 740][pr-740])

Fixes:

- Override the Go version used by the OSV Scanner. ([PR 691][pr-691])
- Validate media types on `regctl artifact put`. ([PR 707][pr-707])
- Use the provided descriptor in the BlobGet/Head to a registry. ([PR 724][pr-724])
- Replace "whitelist" with "known list" for inclusivity. ([PR 725][pr-725])
- Handle nil pointer when config file is a directory. ([PR 738][pr-738])

Chores:

- Limit token permission on the coverage action. ([PR 705][pr-705])
- Clarify `regctl manifest head --platform` will trigger a get request. ([PR 713][pr-713])
- Reenable OSV Scanner weekly check in GitHub Actions. ([PR 715][pr-715])
- Add fuzzing tests for compression. ([PR 741][pr-741])

Contributors:

- @sudo-bmitch

[pr-691]: https://github.com/regclient/regclient/pull/691
[pr-705]: https://github.com/regclient/regclient/pull/705
[pr-706]: https://github.com/regclient/regclient/pull/706
[pr-707]: https://github.com/regclient/regclient/pull/707
[pr-713]: https://github.com/regclient/regclient/pull/713
[pr-715]: https://github.com/regclient/regclient/pull/715
[pr-724]: https://github.com/regclient/regclient/pull/724
[pr-725]: https://github.com/regclient/regclient/pull/725
[pr-730]: https://github.com/regclient/regclient/pull/730
[pr-732]: https://github.com/regclient/regclient/pull/732
[pr-736]: https://github.com/regclient/regclient/pull/736
[pr-738]: https://github.com/regclient/regclient/pull/738
[pr-739]: https://github.com/regclient/regclient/pull/739
[pr-740]: https://github.com/regclient/regclient/pull/740
[pr-741]: https://github.com/regclient/regclient/pull/741

## Release v0.6.0

Breaking:

- `regctl artifact put` no longer includes the filename annotation by default. Use `--file-title` to include. ([PR 659][pr-659])
- Dropping Go 1.19 support ([PR 656][pr-656])
- The platform string for windows images no longer includes the non-standard OS Version value. ([PR 685][pr-685])

Fixes:

- Allow pushing artifacts without an artifactType value. ([PR 658][pr-658])
- Image mod where created image is in a different repository ([PR 662][pr-662])
- Improve returned errors from `regclient.ImageCopy`. ([PR 663][pr-663])
- Cancel blob uploads on failures. ([PR 666][pr-666])
- Allow ctrl-c on `regctl registry login` ([PR 671][pr-671])
- Promoting annotations should ignore child manifests that have been removed from the tree. ([PR 675][pr-675])
- Pin base image digest in build scripts to match Dockerfile pins. ([PR 678][pr-678])
- Error wrapping fixed in several locations. ([PR 682][pr-682])
- Platform selection now finds the best match rather than the first compatible match. ([PR 685][pr-685])
- Update registry versions in CI tests. ([PR 687][pr-687])
- Missing lines from diff context. ([PR 688][pr-688])
- Replace `syft packages` with `syft scan`. ([PR 695][pr-695])
- Image mod can manage the data file on the config descriptor of artifacts. ([PR 697][pr-697])

Features:

- Adding Go 1.22 support ([PR 656][pr-656])
- Add `BlobDelete` support for ocidir references. ([PR 669][pr-669])
- Add `regctl blob delete` command. ([PR 669][pr-669])
- Support formatting output on `regctl registry config`. ([PR 673][pr-673])
- Add image mod ability to promote common annotations in the child images to the index. ([PR 674][pr-674])
- Specifying windows OS Version now uses a comma separated syntax in the platform string. ([PR 685][pr-685])
- Detect AMD64 variant when looking up local platform. ([PR 692][pr-692])
- Add ability to set the config platform setting with `regctl image mod`. ([PR 693][pr-693])
- Image mod support for setting the entrypoint and cmd. ([PR 694][pr-694])

Deprecations:

- Errors in `types` are moved to the `errs` package. ([PR 686][pr-686])
- MediaTypes in `types` are moved to the `mediatype` package. ([PR 686][pr-686])
- Descriptor and associated variables in `types` are moved to the `descriptor` package. ([PR 686][pr-686])
- `github.com/regclient/regclient/regclient` (3 levels of regclient) deprecations are now identified by the standard comment to trigger linters. ([PR 686][pr-686])

Other changes:

- Update OSV scanner to monitor for unapproved licenses. ([PR 655][pr-655])
- Include an API example in the Go docs. ([PR 657][pr-657])
- Add examples to regctl help messages. ([PR 660][pr-660])
- Include the Go Report Card badge. ([PR 664][pr-664])
- Document the availability of the GitHub Actions installer for `regctl`. ([PR 665][pr-665])
- Add examples to regctl help messages. ([PR 672][pr-672])
- Redesign how annotations are added to the regclient images. ([PR 676][pr-676])
- Remove uuid dependency from test code, replace with a random string generator. ([PR 677][pr-677])
- Manage base image annotation with version-bump. ([PR 679][pr-679])
- Use `t.Fatal` where appropriate. ([PR 680][pr-680])
- Remove wraperr package. ([PR 681][pr-681])
- Add links to the GHA workflow badges. ([PR 683][pr-683])
- Include a download count badge. ([PR 684][pr-684])
- Refactoring `types` package to avoid circular dependency issues. ([PR 686][pr-686])
- Cleanup unused parameters on private functions. ([PR 698][pr-698])
- Resume push of SBOMs to Docker Hub. ([PR 701][pr-701])

Contributors:

- @sudo-bmitch

[pr-655]: https://github.com/regclient/regclient/pull/655
[pr-656]: https://github.com/regclient/regclient/pull/656
[pr-657]: https://github.com/regclient/regclient/pull/657
[pr-658]: https://github.com/regclient/regclient/pull/658
[pr-659]: https://github.com/regclient/regclient/pull/659
[pr-660]: https://github.com/regclient/regclient/pull/660
[pr-662]: https://github.com/regclient/regclient/pull/662
[pr-663]: https://github.com/regclient/regclient/pull/663
[pr-664]: https://github.com/regclient/regclient/pull/664
[pr-665]: https://github.com/regclient/regclient/pull/665
[pr-666]: https://github.com/regclient/regclient/pull/666
[pr-669]: https://github.com/regclient/regclient/pull/669
[pr-671]: https://github.com/regclient/regclient/pull/671
[pr-672]: https://github.com/regclient/regclient/pull/672
[pr-673]: https://github.com/regclient/regclient/pull/673
[pr-674]: https://github.com/regclient/regclient/pull/674
[pr-675]: https://github.com/regclient/regclient/pull/675
[pr-676]: https://github.com/regclient/regclient/pull/676
[pr-677]: https://github.com/regclient/regclient/pull/677
[pr-678]: https://github.com/regclient/regclient/pull/678
[pr-679]: https://github.com/regclient/regclient/pull/679
[pr-680]: https://github.com/regclient/regclient/pull/680
[pr-681]: https://github.com/regclient/regclient/pull/681
[pr-682]: https://github.com/regclient/regclient/pull/682
[pr-683]: https://github.com/regclient/regclient/pull/683
[pr-684]: https://github.com/regclient/regclient/pull/684
[pr-686]: https://github.com/regclient/regclient/pull/686
[pr-685]: https://github.com/regclient/regclient/pull/685
[pr-687]: https://github.com/regclient/regclient/pull/687
[pr-688]: https://github.com/regclient/regclient/pull/688
[pr-692]: https://github.com/regclient/regclient/pull/692
[pr-693]: https://github.com/regclient/regclient/pull/693
[pr-694]: https://github.com/regclient/regclient/pull/694
[pr-695]: https://github.com/regclient/regclient/pull/695
[pr-697]: https://github.com/regclient/regclient/pull/697
[pr-698]: https://github.com/regclient/regclient/pull/698
[pr-701]: https://github.com/regclient/regclient/pull/701

## Release v0.5.7

Changes:

- Add a `--skip-check` option to `regctl registry set` and `regctl registry login`. ([PR 646][pr-646])

Fixes:

- Improve error handling on blob put retries when source is not an io.Seeker. ([PR 622][pr-622])
- Preserve descriptor contents on chunked blob push. (@edigaryev) ([PR 637][pr-637])
- Validate descriptor contents on chunked blob push. ([PR 637][pr-637])

Chores:

- Improve testing to detect race conditions in registry operations. ([PR 634][pr-634])
- Update `ImageCopy` test to not depend on `ImageCopy` for setup. ([PR 635][pr-635])
- leverage `t.Setenv` in tests of environment variables. ([PR 636][pr-636])
- reduce logging of context canceled messages in an image copy failure. ([PR 639][pr-639])
- Add tests for TagList and TagDelete. ([PR 640][pr-640])
- Upgrade olareg testing harness to latest version. ([PR 648][pr-648])
- Update OSV scanner to use new syntax. ([PR 652][pr-652])

Contributors:

- @sudo-bmitch

[pr-634]: https://github.com/regclient/regclient/pull/634
[pr-635]: https://github.com/regclient/regclient/pull/635
[pr-636]: https://github.com/regclient/regclient/pull/636
[pr-622]: https://github.com/regclient/regclient/pull/622
[pr-637]: https://github.com/regclient/regclient/pull/637
[pr-639]: https://github.com/regclient/regclient/pull/639
[pr-640]: https://github.com/regclient/regclient/pull/640
[pr-646]: https://github.com/regclient/regclient/pull/646
[pr-648]: https://github.com/regclient/regclient/pull/648
[pr-652]: https://github.com/regclient/regclient/pull/652

## Release v0.5.6

Changes:

- Chore: go.mod version is now set to oldest supported release. ([PR 623][pr-623])
- Chore: Make vendoring optional. ([PR 632][pr-632])

Contributors:

- @sudo-bmitch

[pr-623]: https://github.com/regclient/regclient/pull/623
[pr-632]: https://github.com/regclient/regclient/pull/632

## Release v0.5.5

New Features:

- Add OpenSSF Best Practices Badge. ([PR 607][pr-607])
- Adding OpenSSF Scorecard badge and GHA workflow. ([PR 609][pr-609])

Fixes:

- Validate references in regclient methods. ([PR 595][pr-595])
- Data race in the reghttp fallback timeout handling. ([PR 599][pr-599])
- HTTP proxy using environment variables. ([PR 615][pr-615])

Chores:

- Reorder descriptor fields. ([PR 594][pr-594])
- Add test for ocidir throttle race. ([PR 601][pr-601])
- Add gomajor utility to Makefile. ([PR 602][pr-602])
- Add commands to Makefile for managing releases. ([PR 604][pr-604])
- Pin GitHub actions. ([PR 605][pr-605])
- Use full semver on dependencies where available. ([PR 605][pr-605])
- Adjust token permissions on GitHub actions. ([PR 606][pr-606])
- Include disclosure timeline in security policy. ([PR 608][pr-608])
- Improve contributor guidelines. ([PR 612][pr-612])
- Improve BlobPut tests. ([PR 613][pr-613])

Contributors:

- @peusebiu
- @sudo-bmitch

[pr-594]: https://github.com/regclient/regclient/pull/594
[pr-595]: https://github.com/regclient/regclient/pull/595
[pr-599]: https://github.com/regclient/regclient/pull/599
[pr-601]: https://github.com/regclient/regclient/pull/601
[pr-602]: https://github.com/regclient/regclient/pull/602
[pr-604]: https://github.com/regclient/regclient/pull/604
[pr-605]: https://github.com/regclient/regclient/pull/605
[pr-606]: https://github.com/regclient/regclient/pull/606
[pr-607]: https://github.com/regclient/regclient/pull/607
[pr-608]: https://github.com/regclient/regclient/pull/608
[pr-609]: https://github.com/regclient/regclient/pull/609
[pr-612]: https://github.com/regclient/regclient/pull/612
[pr-613]: https://github.com/regclient/regclient/pull/613
[pr-615]: https://github.com/regclient/regclient/pull/615

## Release v0.5.4

New Features:

- Add `regctl --host` flag to configure registries for a single command. ([PR 572][pr-572])
- Configure HTTP client timeouts. ([PR 584][pr-584])
- Add `regclient.Ping` method. ([PR 590][pr-590])
- regctl: warn on failed logins or bad registry configuration changes. ([PR 590][pr-590])

Fixes:

- Fix handling of invalid hostname in `regclient.RepoList`. ([PR 577][pr-577])
- Fix bug in regsync tag filtering when running as a server. ([PR 579][pr-579])
- Enable parallel builds of the make "binaries" target. ([PR 588][pr-588])

Chores:

- Update Go docs on blob APIs and the config. ([PR 573][pr-573])
- Refactor the Ref package. ([PR 587][pr-587])

Contributors:

- @andyli
- @Juneezee
- @sudo-bmitch

[pr-572]: https://github.com/regclient/regclient/pull/572
[pr-573]: https://github.com/regclient/regclient/pull/573
[pr-577]: https://github.com/regclient/regclient/pull/577
[pr-579]: https://github.com/regclient/regclient/pull/579
[pr-584]: https://github.com/regclient/regclient/pull/584
[pr-587]: https://github.com/regclient/regclient/pull/587
[pr-588]: https://github.com/regclient/regclient/pull/588
[pr-590]: https://github.com/regclient/regclient/pull/590

## Release v0.5.3

Fixes:

- Fix formatting variables in `regctl image inspect`. ([PR 554][pr-554])

New Features:

- Add a `GetSize` method to image manifests (OCI and Docker2 manifests). ([PR 565][pr-565])

Chores:

- Refactoring CLIs to remove global state. ([PR 550][pr-550])
- Set GOTOOLCHAIN=local in CI ([PR 556][pr-556])
- Reorder Go imports to move local packages last. ([PR 557][pr-557])
- Remove duplicated tests from ci-registry action. ([PR 559][pr-559])
- Run tests using t.Parallel where possible. ([PR 564][pr-564])
- Update install guidance for quarantined binaries on MacOS. ([PR 569][pr-569])
- Release notes now include contributors. ([PR 570][pr-570])

Contributors:

- @felipecrs
- @sorenisanerd
- @sudo-bmitch

[pr-550]: https://github.com/regclient/regclient/pull/550
[pr-554]: https://github.com/regclient/regclient/pull/554
[pr-556]: https://github.com/regclient/regclient/pull/556
[pr-557]: https://github.com/regclient/regclient/pull/557
[pr-559]: https://github.com/regclient/regclient/pull/559
[pr-564]: https://github.com/regclient/regclient/pull/564
[pr-565]: https://github.com/regclient/regclient/pull/565
[pr-569]: https://github.com/regclient/regclient/pull/569
[pr-570]: https://github.com/regclient/regclient/pull/570

## Release v0.5.2

Breaking Changes:

- A few interfaces in the blob package were converted to pointers to structs. ([PR 547][pr-547])

Features:

- Expose the underlying OCI Layout Index to `regctl tag ls --format` ([PR 518][pr-518])
- Support compression on image export and import. ([PR 522][pr-522])
- Image mod of timestamps is now aware of base images. ([PR 524][pr-524])
- Add reproducible method / option to image mod. ([PR 525][pr-525])
- Support setting labels on specific platforms with image mod. ([PR 528][pr-528])
- Add `WithFileTarTime` method and `regctl image mod --file-tar-time` option to edit timestamps inside tar files. ([PR 530][pr-530])
- Support digest-tags in artifact list and tree output ([PR 531][pr-531])
- Add support for decompressing xz layers ([PR 534][pr-534])
- Support getting an artifact from an index of artifacts ([PR 536][pr-536])
- Add repo filters to regsync when copying registries ([PR 538][pr-538])
- Add gosec security linter ([PR 541][pr-541])
- Refactor type/blob package. ([PR 547][pr-547])
- Support pushing artifact to an index entry with `regctl artifact put --index` ([PR 548][pr-548])

Fixes:

- Add size limits on manifests ([PR 512][pr-512])
- Always set `artifactType` with `regctl artifact put` ([PR 513][pr-513])
- Manifest delete should not fail when referrer file is missing ([PR 515][pr-515])
- Artifact put of referrer should not add a manifest reference in ocidir ([PR 515][pr-515])
- Reproducible image creation scripts should prune stale referrers ([PR 515][pr-515])
- Fail faster on image copy when target registry is unreachable ([PR 517][pr-517])
- Avoid changing docker build attestations when converting to referrers ([PR 527][pr-527])

Chores:

- Cleanup docs on regclient package. ([PR 543][pr-543])
- Upgrade yaml package to v3. ([PR 544][pr-544])

New Contributors:

- @fanthos

[pr-512]: https://github.com/regclient/regclient/pull/512
[pr-513]: https://github.com/regclient/regclient/pull/513
[pr-515]: https://github.com/regclient/regclient/pull/515
[pr-517]: https://github.com/regclient/regclient/pull/517
[pr-518]: https://github.com/regclient/regclient/pull/518
[pr-522]: https://github.com/regclient/regclient/pull/522
[pr-524]: https://github.com/regclient/regclient/pull/524
[pr-525]: https://github.com/regclient/regclient/pull/525
[pr-527]: https://github.com/regclient/regclient/pull/527
[pr-528]: https://github.com/regclient/regclient/pull/528
[pr-530]: https://github.com/regclient/regclient/pull/530
[pr-531]: https://github.com/regclient/regclient/pull/531
[pr-534]: https://github.com/regclient/regclient/pull/534
[pr-536]: https://github.com/regclient/regclient/pull/536
[pr-538]: https://github.com/regclient/regclient/pull/538
[pr-541]: https://github.com/regclient/regclient/pull/541
[pr-543]: https://github.com/regclient/regclient/pull/543
[pr-544]: https://github.com/regclient/regclient/pull/544
[pr-547]: https://github.com/regclient/regclient/pull/547
[pr-548]: https://github.com/regclient/regclient/pull/548

## Release v0.5.1

Features:

- Add options to `regctl index create` for `artifactType` and `subject` ([PR 490][pr-490])
- Add `--latest` flag to `regctl artifact get` and `regctl artifact list` ([PR 507][pr-507])
- Add in memory caching support. ([PR 510][pr-510])

Fixes:

- Fix typo in OCI annotations ([PR 485][pr-485])
- Support multiple dashes in repo names. ([PR 489][pr-489])
- Fix auth failures, do not trigger a backoff ([PR 492][pr-492])
- Update Go dependencies from all subdirectories and used in tests ([PR 495][pr-495])
- Warning header log message is fixed. ([PR 497][pr-497])
- Include version annotation in regclient images ([PR 499][pr-499])
- Fix digest check when running `regctl image get-file` ([PR 503][pr-503])
- Switch `org.opencontainers.artifact.*` to `org.opencontainers.image.*` annotations in regclient images. ([PR 506][pr-506])

Chores:

- Sync OCI types to align with upstream sources. ([PR 488][pr-488])
- Add OSV vulnerability scanner ([PR 498][pr-498])

[pr-485]: https://github.com/regclient/regclient/pull/485
[pr-488]: https://github.com/regclient/regclient/pull/488
[pr-489]: https://github.com/regclient/regclient/pull/489
[pr-490]: https://github.com/regclient/regclient/pull/490
[pr-492]: https://github.com/regclient/regclient/pull/492
[pr-495]: https://github.com/regclient/regclient/pull/495
[pr-497]: https://github.com/regclient/regclient/pull/497
[pr-498]: https://github.com/regclient/regclient/pull/498
[pr-499]: https://github.com/regclient/regclient/pull/499
[pr-503]: https://github.com/regclient/regclient/pull/503
[pr-506]: https://github.com/regclient/regclient/pull/506
[pr-507]: https://github.com/regclient/regclient/pull/507
[pr-510]: https://github.com/regclient/regclient/pull/510

## Release v0.5.0

The two key features are:

- Updating the image copy to copy layers concurrently and with an improved UI.
- Update support for OCI with the Referrers changes coming in their 1.1 releases.

Image Copy:

- Add progress display to `regctl image copy` ([PR 413][pr-413])
- Image copy is now run with concurrency. ([PR 419][pr-419])
- Fix `regctl image copy` output on narrow terminals. ([PR 440][pr-440])
- Add a fast check option for copying images with referrers and digest tags. ([PR 441][pr-441])
- Update `regctl image copy` for tty displays. ([PR 447][pr-447])

OCI Support:

- Add support for `artifactType` in image manifest ([PR 400][pr-400])
- Accept manifests with OCI artifact media type (experimental). ([PR 418][pr-418])
- Handle the OCI-Subject header to detect referrer support. ([PR 446][pr-446])
- Embed the `Platform` field directly in the `ImageConfig` ([PR 456][pr-456])
- Switch from scratch to empty JSON media type ([PR 463][pr-463])
- Support artifactType and subject fields on OCI Index ([PR 476][pr-476])

Other Features:

- Image mod pushes directly to the target ref without an extra copy step ([PR 438][pr-438])
- Performance improvements for regsync ([PR 449][pr-449])
- Support client certs and keys for mTLS registry auth. ([PR 454][pr-454])
- Support updating annotations on platform specific manifests in a manifest list. ([PR 457][pr-457])
- Add ability to sort referrers by annotation. ([PR 467][pr-467])
- Use `SOURCE_DATE_EPOCH` build arg support in buildkit. ([PR 472][pr-472])
- Add regctl tag list filtering ([PR 477][pr-477])
- Add option to import a specific image or tag from an export of multiple images ([PR 482][pr-482])

Fixes:

- Invalid references are detected before querying the registry ([PR 414][pr-414])
- Fix handling of content-type headers. ([PR 418][pr-418])
- Fix race when creating ocidir ([PR 420][pr-420])
- Avoid an internal race condition when managing the referrers fallback tag. ([PR 427][pr-427])
- Fix: close reader when converting a blob to an OCI config ([PR 434][pr-434])
- Support manifests missing a mediaType field. ([PR 436][pr-436])
- Fix GitHub badges. ([PR 437][pr-437])
- Handle symlinks in the tar file with `regctl image import` ([PR 452][pr-452])
- Fix GCR credential helper to work on Artifact Registry ([PR 455][pr-455])
- Fix deadlock when referrers or digest tags refer to a parent manifest ([PR 464][pr-464])
- Improve error handling of `regctl artifact tree`. ([PR 470][pr-470])
- Fix copy when both referrers and digest-tags are included. ([PR 471][pr-471])
- Fix handling of copy with looping referrers or digest-tags to validating registries ([PR 475][pr-475])

[pr-400]: https://github.com/regclient/regclient/pull/400
[pr-413]: https://github.com/regclient/regclient/pull/413
[pr-414]: https://github.com/regclient/regclient/pull/414
[pr-418]: https://github.com/regclient/regclient/pull/418
[pr-419]: https://github.com/regclient/regclient/pull/419
[pr-420]: https://github.com/regclient/regclient/pull/420
[pr-427]: https://github.com/regclient/regclient/pull/427
[pr-434]: https://github.com/regclient/regclient/pull/434
[pr-436]: https://github.com/regclient/regclient/pull/436
[pr-437]: https://github.com/regclient/regclient/pull/437
[pr-438]: https://github.com/regclient/regclient/pull/438
[pr-440]: https://github.com/regclient/regclient/pull/440
[pr-441]: https://github.com/regclient/regclient/pull/441
[pr-446]: https://github.com/regclient/regclient/pull/446
[pr-447]: https://github.com/regclient/regclient/pull/447
[pr-449]: https://github.com/regclient/regclient/pull/449
[pr-452]: https://github.com/regclient/regclient/pull/452
[pr-454]: https://github.com/regclient/regclient/pull/454
[pr-455]: https://github.com/regclient/regclient/pull/455
[pr-456]: https://github.com/regclient/regclient/pull/456
[pr-457]: https://github.com/regclient/regclient/pull/457
[pr-463]: https://github.com/regclient/regclient/pull/463
[pr-464]: https://github.com/regclient/regclient/pull/464
[pr-467]: https://github.com/regclient/regclient/pull/467
[pr-470]: https://github.com/regclient/regclient/pull/470
[pr-471]: https://github.com/regclient/regclient/pull/471
[pr-472]: https://github.com/regclient/regclient/pull/472
[pr-475]: https://github.com/regclient/regclient/pull/475
[pr-476]: https://github.com/regclient/regclient/pull/476
[pr-477]: https://github.com/regclient/regclient/pull/477
[pr-482]: https://github.com/regclient/regclient/pull/482

## Release v0.4.8

Breaking Changes:

- Deprecated: `regclient.WithConfigHosts` is replaced by a variadic on `regclient.WithConfigHost` ([PR 409][pr-409])
- Deprecated: `regclient.WithBlobLimit`, `regcleint.WithBlobSize`, `regclient.WithCertDir`, `regclient.WithRetryDelay`, and `regclient.WithRetryLimit` are replaced by `regclient.WithRegOpts` ([PR 409][pr-409])

New Features:

- Add `--platform` option to `regctl image copy/export` ([PR 379][pr-379])
- Add option to override name in `regctl image export` ([PR 380][pr-380])
- Add platforms option to `regctl index add/create` ([PR 381][pr-381])
- Add `--referrers` and `--digest-tags` options to `regctl index add/create` ([PR 382][pr-382])
- Add `regctl blob copy` command ([PR 385][pr-385])
- Adding `regctl image mod --to-docker` to convert manifests from OCI to Docker schema2 ([PR 388][pr-388])
- Support `OCI-Chunk-Min-Length` header ([PR 394][pr-394])
- Add support for registry warning headers ([PR 396][pr-396])
- Add `regclient.WithRegOpts` ([PR 408][pr-408])

Bug Fixes:

- Improve handling of the referrers API with Harbor ([PR 389][pr-389])
- Fix an issue on `regctl tag rm` to support registries that require a layer ([PR 395][pr-395])
- Image mod only converts `config.mediaType` between known values ([PR 399][pr-399])
- Ignore anonymous blob mount failures ([PR 401][pr-401])
- Fix handling of docker registry logins with `credStore` ([PR 405][pr-405])
- Fix regsync handling of the paginated repo listing when syncing registries ([PR 406][pr-406])

Other Changes:

- Recursively sign manifest list and platform specific images with cosign ([PR 378][pr-378])
- Include tag in the version output ([PR 392][pr-392])

[pr-378]: https://github.com/regclient/regclient/pull/378
[pr-379]: https://github.com/regclient/regclient/pull/379
[pr-380]: https://github.com/regclient/regclient/pull/380
[pr-381]: https://github.com/regclient/regclient/pull/381
[pr-382]: https://github.com/regclient/regclient/pull/382
[pr-385]: https://github.com/regclient/regclient/pull/385
[pr-388]: https://github.com/regclient/regclient/pull/388
[pr-389]: https://github.com/regclient/regclient/pull/389
[pr-392]: https://github.com/regclient/regclient/pull/392
[pr-394]: https://github.com/regclient/regclient/pull/394
[pr-395]: https://github.com/regclient/regclient/pull/395
[pr-396]: https://github.com/regclient/regclient/pull/396
[pr-399]: https://github.com/regclient/regclient/pull/399
[pr-401]: https://github.com/regclient/regclient/pull/401
[pr-405]: https://github.com/regclient/regclient/pull/405
[pr-406]: https://github.com/regclient/regclient/pull/406
[pr-408]: https://github.com/regclient/regclient/pull/408
[pr-409]: https://github.com/regclient/regclient/pull/409

## Release v0.4.7

This is a repeat of v0.4.6 with fixes to GitHub Actions performing the release.

## Release v0.4.6

Breaking Changes:

- Tag details are no longer available in `version` commands ([PR 305][pr-305])
- `regclient.VCSRef` and `VCSTag` variables removed ([PR 305][pr-305])
- Detailed version support removed for Go 1.17 and earlier ([PR 305][pr-305])
- Artifact manifest is experimental ([PR 372][pr-372])

New Features:

- Make image builds reproducible ([PR 309][pr-309])
- Improve handling of network failures on blob push ([PR 324][pr-324])
- Add configurable limits on concurrency and requests per second ([PR 330][pr-330])
- regbot: adding options to `repo.ls` to support paginated responses ([PR 332][pr-332])
- Improve retries of blob copies between registries ([PR 333][pr-333])
- `regctl registry login` supports reading the password from stdin ([PR 348][pr-348])
- Add `mod.WithManifestToOCIReferrers()` to convert Docker to OCI referrers ([PR 349][pr-349])
- Add `regctl image mod --to-oci-referrers` ([PR 349][pr-349])
- Add SBOMs to artifacts to images ([PR 351][pr-351])
- Add `regctl manifest head` and `regctl blob head` commands ([PR 358][pr-358])
- Adding platform option to `regctl artifact` commands ([PR 359][pr-359])
- Adding platform option to the `ReferrerList` API ([PR 359][pr-359])
- Add `regctl artifact tree` command ([PR 361][pr-361])
- Binary artifacts are signed with cosign keyless signing ([PR 365][pr-365])
- Add referrers support to regsync ([PR 366][pr-366])

Bug Fixes:

- Fix error handling on manifests created by `ManifestHead` ([PR 343][pr-343])
- Fix `regctl image mod` for non-tar layers and non-image configs ([PR 347][pr-347])
- Converting to OCI now handles OCI manifests with mixed media types inside manifest ([PR 353][pr-353])
- Fix issue importing images created with tar ([PR 355][pr-355])
- Fallback to `GET` when `HEAD` request is missing digest header ([PR 363][pr-363])
- Fix handling of data field ([PR 368][pr-368])

Other Changes:

- Add help text to Makefile ([PR 308][pr-308])
- Make `artifactType` field optional for artifact manifests ([PR 327][pr-327])
- Fix reproducible builds with docker provenance ([PR 350][pr-350])
- Update Go support to 1.20 ([PR 356][pr-356])
- Upgrade Go to 1.20.1 ([PR 362][pr-362])
- Pin the version of staticcheck ([PR 364][pr-364])

[pr-305]: https://github.com/regclient/regclient/pull/305
[pr-308]: https://github.com/regclient/regclient/pull/308
[pr-309]: https://github.com/regclient/regclient/pull/309
[pr-324]: https://github.com/regclient/regclient/pull/324
[pr-327]: https://github.com/regclient/regclient/pull/327
[pr-330]: https://github.com/regclient/regclient/pull/330
[pr-332]: https://github.com/regclient/regclient/pull/332
[pr-333]: https://github.com/regclient/regclient/pull/333
[pr-343]: https://github.com/regclient/regclient/pull/343
[pr-347]: https://github.com/regclient/regclient/pull/347
[pr-348]: https://github.com/regclient/regclient/pull/348
[pr-349]: https://github.com/regclient/regclient/pull/349
[pr-350]: https://github.com/regclient/regclient/pull/350
[pr-351]: https://github.com/regclient/regclient/pull/351
[pr-353]: https://github.com/regclient/regclient/pull/353
[pr-355]: https://github.com/regclient/regclient/pull/355
[pr-356]: https://github.com/regclient/regclient/pull/356
[pr-358]: https://github.com/regclient/regclient/pull/358
[pr-359]: https://github.com/regclient/regclient/pull/359
[pr-361]: https://github.com/regclient/regclient/pull/361
[pr-362]: https://github.com/regclient/regclient/pull/362
[pr-363]: https://github.com/regclient/regclient/pull/363
[pr-364]: https://github.com/regclient/regclient/pull/364
[pr-365]: https://github.com/regclient/regclient/pull/365
[pr-366]: https://github.com/regclient/regclient/pull/366
[pr-368]: https://github.com/regclient/regclient/pull/368
[pr-372]: https://github.com/regclient/regclient/pull/372

## Release v0.4.5

Breaking Changes:

- `regclient.ManifestWithDesc` has been renamed to `regclient.WithManifestDesc` ([PR 283][pr-283])
- `manifest.Referrer` interface has been renamed to `manifest.Subjecter` ([PR 283][pr-283], [PR 302][pr-302])
- `ReferrerPut` method has been removed, use `ManifestPut` instead ([PR 283][pr-283])
- `regclient.ManifestPut` options are now in `regclient` rather than scheme ([PR 283][pr-283])

New Features:

- Adding diff commands to regctl for manifests, configs, and layers ([PR 269][pr-269])
- Supporting OCI subject/referrers ([PR 271][pr-271], [PR 302][pr-302])
- Adding support for client side filters on artifact list ([PR 273][pr-273])
- Support Link header in tag pagination seen on quay.io ([PR 276][pr-276])
- Update referrers to support deletes ([PR 283][pr-283])
- Add `regctl image check-base` ([PR 288][pr-288])
- Adding experimental support for rebasing images ([PR 291][pr-291])
- regclient adds a blob TarReader.ReadFile method ([PR 296][pr-296])
- Adding `regctl blob get-file` to fetch a file from a layer ([PR 296][pr-296])
- Adding `regctl image get-file` to fetch a file from the image layers ([PR 296][pr-296])
- Fallback to using linux platforms on mac and windows when querying manifest lists ([PR 303][pr-303])

Bug Fixes:

- Fix reference for ocidir exports for importing into docker ([PR 264][pr-264])
- Fix support for authentication parsing of token characters without quotes ([PR 266][pr-266])
- Fix `regctl artifact put` with both a refers and a tag ([PR 290][pr-290])
- Fixing regbot tests to run on other platforms ([PR 300][pr-300])

Other Changes:

- Upgrade to Go 1.19 ([PR 278][pr-278], [PR 286][pr-286])
- Add tips to common errors in regctl ([PR 285][pr-285])
- Descriptor comparison now allows for Docker to OCI conversion ([PR 289][pr-289])

[pr-264]: https://github.com/regclient/regclient/pull/264
[pr-266]: https://github.com/regclient/regclient/pull/266
[pr-269]: https://github.com/regclient/regclient/pull/269
[pr-271]: https://github.com/regclient/regclient/pull/271
[pr-273]: https://github.com/regclient/regclient/pull/273
[pr-276]: https://github.com/regclient/regclient/pull/276
[pr-278]: https://github.com/regclient/regclient/pull/278
[pr-283]: https://github.com/regclient/regclient/pull/283
[pr-285]: https://github.com/regclient/regclient/pull/285
[pr-286]: https://github.com/regclient/regclient/pull/286
[pr-288]: https://github.com/regclient/regclient/pull/288
[pr-289]: https://github.com/regclient/regclient/pull/289
[pr-290]: https://github.com/regclient/regclient/pull/290
[pr-291]: https://github.com/regclient/regclient/pull/291
[pr-296]: https://github.com/regclient/regclient/pull/296
[pr-300]: https://github.com/regclient/regclient/pull/300
[pr-302]: https://github.com/regclient/regclient/pull/302
[pr-303]: https://github.com/regclient/regclient/pull/303

## Release v0.4.4

Breaking Changes:

- Redundant manifest methods have been deprecated ([PR 257][pr-257])
- Manifest methods specific to Images and Indexes/ManifestLists have a different interface ([PR 258][pr-258])

New Features:

- `regctl blob put` supports a format string ([PR 255][pr-255])
- `regsync` immediately syncs any tags missing on the target without waiting for first sync schedule ([PR 256][pr-256])
- Adding `regctl index` command to create and mutate an Index or Manifest List ([PR 260][pr-260])

Bug Fixes:

- Credential helper support for Docker Hub ([PR 246][pr-246])
- Fix handling of `DOCKER_CONFIG` variable ([PR 249][pr-249])
- Fix handling of custom TLS settings on a registry with authentication ([PR 253][pr-253])

Other Changes:

- Normalize parsing of registry names in various components ([PR 247][pr-247])
- Exclude tags from referrers when copying digest tags ([PR 250][pr-250])

[pr-246]: https://github.com/regclient/regclient/pull/246
[pr-247]: https://github.com/regclient/regclient/pull/247
[pr-249]: https://github.com/regclient/regclient/pull/249
[pr-250]: https://github.com/regclient/regclient/pull/250
[pr-253]: https://github.com/regclient/regclient/pull/253
[pr-255]: https://github.com/regclient/regclient/pull/255
[pr-256]: https://github.com/regclient/regclient/pull/256
[pr-257]: https://github.com/regclient/regclient/pull/257
[pr-258]: https://github.com/regclient/regclient/pull/258
[pr-260]: https://github.com/regclient/regclient/pull/260

## Release v0.4.3

Breaking Changes:

- Media type variable for docker image layers has been renamed ([PR 220][pr-220])

New Features:

- Improve handling of external URLs ([PR 220][pr-220])
- Call credential helpers only when needed ([PR 234][pr-234])
- Support credential helpers directly in regsync and regbot ([PR 238][pr-238])
- Allow pushing artifacts and manifest by digest ([PR 242][pr-242])

Bug Fixes:

- regctl image mod args for expose and volume rm actually rm ([PR 216][pr-216])
- manifest head request now set the descriptor size correctly ([PR 222][pr-222])
- Fix for chunked uploads ([PR 228][pr-228])
- Fix image import from a multi-platform image to an ocidir:// layout ([PR 235][pr-235])

Other Changes:

- Adding experimental support for OCI referrers ([PR 225][pr-225])
- Adds experimental referrer support to regctl artifact commands ([PR 226][pr-226])
- Experimental: adding support for OCI artifact media type ([PR 229][pr-229])
- Switch to internal processing of docker config.json ([PR 234][pr-234])

[pr-216]: https://github.com/regclient/regclient/pull/216
[pr-220]: https://github.com/regclient/regclient/pull/220
[pr-222]: https://github.com/regclient/regclient/pull/222
[pr-225]: https://github.com/regclient/regclient/pull/225
[pr-226]: https://github.com/regclient/regclient/pull/226
[pr-228]: https://github.com/regclient/regclient/pull/228
[pr-229]: https://github.com/regclient/regclient/pull/229
[pr-234]: https://github.com/regclient/regclient/pull/234
[pr-235]: https://github.com/regclient/regclient/pull/235
[pr-238]: https://github.com/regclient/regclient/pull/238
[pr-242]: https://github.com/regclient/regclient/pull/242

## Release v0.4.2

Breaking Changes:

- Format templates no longer support `title`.
  ([PR 194][pr-194])

New Features:

- Add the ability to remove buildarg from image history.
  ([PR 207][pr-207])
- regsync now supports the `.Sync` template variable on source and target.
  ([PR 208][pr-208])
- Add support for annotations on docker2 schemas.
  ([PR 209][pr-209])
- Adding image signing with cosign.
  ([PR 212][pr-212])

Bug Fixes:

- Fix handling of http_proxy and https_proxy.
  ([PR 197][pr-197])
- manifest head on an ocidir with a digest no longer succeeds if manifest is missing.
  ([PR 199][pr-199])
- image export of manifest list to ocidir should only define parent manifest in `index.json`.
  ([PR 200][pr-200])
- Add `/tmp` directory to scratch images.
  ([PR 205][pr-205])
- Handle multiple tags pointing to the same digest in ocidir.
  ([PR 211][pr-211])

Other Changes:

- Upgrade to Go 1.17.
  ([PR 193][pr-193])
- Go build now runs with `-trimpath`.
  ([PR 196][pr-196])

[pr-193]: https://github.com/regclient/regclient/pull/193
[pr-194]: https://github.com/regclient/regclient/pull/194
[pr-196]: https://github.com/regclient/regclient/pull/196
[pr-197]: https://github.com/regclient/regclient/pull/197
[pr-199]: https://github.com/regclient/regclient/pull/199
[pr-200]: https://github.com/regclient/regclient/pull/200
[pr-205]: https://github.com/regclient/regclient/pull/205
[pr-207]: https://github.com/regclient/regclient/pull/207
[pr-208]: https://github.com/regclient/regclient/pull/208
[pr-209]: https://github.com/regclient/regclient/pull/209
[pr-211]: https://github.com/regclient/regclient/pull/211
[pr-212]: https://github.com/regclient/regclient/pull/212

## Release v0.4.1

Breaking Changes:

- Blob methods have been updated to use a descriptor instead of the digest and size.
  ([PR 173][pr-173])

New Features:

- Allow the user-agent to be overridden.
  ([PR 172][pr-172])
- Support the Data field in the descriptor on manifest and blob gets.
  ([PR 174][pr-174])
- Added image modification functionality.
  ([PR 182][pr-182])

Bug Fixes:

- Fix an issue with dangling references in the `ocidir` `index.json`.
  ([PR 176][pr-176])
- Fix handling of relative paths in `ocidir`.
  ([PR 177][pr-177])
- Fix handling of `/etc/docker/certs.d`.
  ([PR 179][pr-179])
- Fix handling of registry CA configuration.
  ([PR 180][pr-180])

[pr-172]: https://github.com/regclient/regclient/pull/172
[pr-173]: https://github.com/regclient/regclient/pull/173
[pr-174]: https://github.com/regclient/regclient/pull/174
[pr-176]: https://github.com/regclient/regclient/pull/176
[pr-177]: https://github.com/regclient/regclient/pull/177
[pr-179]: https://github.com/regclient/regclient/pull/179
[pr-180]: https://github.com/regclient/regclient/pull/180
[pr-182]: https://github.com/regclient/regclient/pull/182

## Release v0.4.0

Breaking Changes:

- Repository has been restructured to remove the `regclient` sub-directory.
  Backwards compatible aliases and stub functions have been added to minimize the impact.
  ([PR 130][pr-130])
- Refactored the manifest type. This results in breaking changes to exported methods.
  ([PR 157][pr-157])
- Blob methods have been refactored with breaking changes to exported APIs.
  ([PR 158][pr-158])
- External dependencies have been minimized, particularly for struct definitions.
  This impacts variable types used by some APIs.
  ([PR 160][pr-160])
- Default to displaying manifest lists in regctl.
  ([PR 166][pr-166])
- Image manifests now display with a pretty printer by default.
  ([PR 168][pr-168])

New Features:

- APIs have been updated to support the `ocidir://` scheme.
  You can now `regctl image copy alpine:latest ocidir://alpine:latest` to copy an image to a directory as you would another registry.
  ([PR 138][pr-138], [PR 146][pr-146])
- Added the ability to modify a manifest type.
  ([PR 157][pr-157])
- Blob SetConfig method added to allow modifying an OCI config.
  ([PR 158][pr-158])
- Added repoAuth option for gcr.io support.
  ([PR 159][pr-159])

Bug Fixes:

- Lots of linting and testing has been added.
  This will result in a change to some error messages.
  ([PR 146][pr-146], [PR 147][pr-147], [PR 151][pr-151])
- Fix handling of docker logins where scheme is included in the hostname.
  ([PR 137][pr-137])
- Fix for tag rm with authentication when the blob upload location is a different host.
  ([PR 144][pr-144])

[pr-130]: https://github.com/regclient/regclient/pull/130
[pr-137]: https://github.com/regclient/regclient/pull/137
[pr-138]: https://github.com/regclient/regclient/pull/138
[pr-144]: https://github.com/regclient/regclient/pull/144
[pr-146]: https://github.com/regclient/regclient/pull/146
[pr-147]: https://github.com/regclient/regclient/pull/147
[pr-151]: https://github.com/regclient/regclient/pull/151
[pr-157]: https://github.com/regclient/regclient/pull/157
[pr-158]: https://github.com/regclient/regclient/pull/158
[pr-159]: https://github.com/regclient/regclient/pull/159
[pr-160]: https://github.com/regclient/regclient/pull/160
[pr-166]: https://github.com/regclient/regclient/pull/166
[pr-168]: https://github.com/regclient/regclient/pull/168

## Release v0.3.10

Bug fixes:

- Insecure TLS should handle unknown TLS keys
- regsync backup errors are now a warning only, allowing automated recovery from a corrupt registry
- Fix for deleting tags pointing to an OCI manifest

Features and Changes:

- Verifying media type and digest on manifests
- Adding option to delete manifests by tag instead of digest
- Image copy: support added for digest tags used by tools like projectsigstore/cosign
- Image copy: support added to force a recursive copy when manifest already exists at target
- Handle registries without HEAD support (older versions of Nexus)

## Release v0.3.9

Features and Changes:

- Binaries are built for darwin-arm64 for Mac M1 users
- GitHub actions are fixed to not push on forks
- `regctl artifact` command has been added to get and put OCI artifacts

## Release v0.3.8

Breaking changes:

- `regclient.ImageImport` API uses an `io.ReadSeeker` instead of a filename
- OCI export/import is corrected to use annotations, previous exports of an OCI image will not import with this change

Features and Changes:

- Disable default blob max to support registries that don't handle chunked uploads
- Update to Go 1.16 and dependencies have been bumped
- Images now push to GHCR in addition to Docker Hub
- GHA fixes caching for faster docker image builds
- Allow a 201 http status for chunked blob uploads

## Release v0.3.7

Features and Changes:

- Fixes a bug that breaks chunked blob pushes
- Adds registry configurable chunk sizes and threshold to switch to chunked pushes
- Docker improves caching of credential helpers between builds

## Release v0.3.6

Features and Changes:

- add retry capability on blob push/pulls
- buildkit cache images can be copied and imported/exported
- ci: binaries for s390x and ppc64le are now built

## Release v0.3.5

Features and Changes:

- regctl: bash completion with colons in arg is now supported
- regbot and regsync: run "once" commands single threaded when parallel set to 0
- regsync: support for copying entire registry
- regclient: Improve http status code handling to better follow docker registry spec
- regclient: OCI distribution-spec 1.0 support

## Release v0.3.4

Features and Changes:

- regctl: adding command completion
- regclient: Include messages from registry in errors
- regclient: Adding OCI export (Docker export was previously available)
- regclient: Adding OCI and Docker import from tar
- regctl: adding image import/export support
- regbot: adding image import/export support

## Release v0.3.3

Features and Changes:

- regbot: Support for push API calls to modify images
- regclient: Refactoring blobs into a separate package
- Update auth token shortly before expiration to avoid race conditions
- Adding support for docker credential helpers
- Include ECR and GCR helpers in alpine images

## Release v0.3.2

Fixes:

- Fixing broken build of regctl on Windows

## Release v0.3.1

Changes:

- regclient: moved manifest into separate package
- regclient: chunked upload of a blob is now supported
- regctl: adding manifest and blob put commands
- regsync: registry field can now be templated
- GitHub actions: nightly edge build has been added

Fixes:

- Update interactive password call for windows

Experimental:

- Adding Docker Hub tag delete

## Release v0.3.0

Deprecated:

- Multiple regclient APIs have received breaking changes.
- Configuration file handling has been removed from regclient. Configuration
  should be injected by options when creating the regclient instance.
- DNS and Schema configuration options have been removed. Schema is now handled
  using the TLS setting, and DNS has been separated into Hostname and Mirrors.

Command Changes:

- Docker Hub is now handled as `docker.io` in various commands and is configured
  as the default when no registry is provided.
- Redesign of the mirror handling allows each mirror to have different
  credentials and a path prefix option.
- `raw`, `raw-headers`, and `raw-body` formats added for viewing the original
  response from the registry server.
- `version` sub-commands added to output the tag and git commit.
- `printPretty` format added to output tags and manifest list in a more user
  friendly format.
- regbot creates a new reference instead of reusing an existing one.
- regbot reference adds a digest get/set operation.
- Documentation has been updated.
- `regctl image digest` on a manifest list only performs one GET request instead
  of two.
- Docker schema v1 is supported to handle older images.

regclient API Changes:

- User-Agent in headers now reports which regclient command and commit.
- Retryable code now maintains state across multiple requests to avoid retrying
  a down mirror for each layer of an image.
- API handling to the registry has been abstracted to hopefully allow registry
  specific API requests.
- Error handling updated to support `errors.Is()`.
- Image config moved behind the blob interface.

## Release v0.2.1

- `regctl registry` commands `login`, `logout`, and `set` previously required
  the nonintuitive DNS name "registry-1.docker.io" for Hub. They now accept
  "docker.io" and default to Hub when no registry is provided.

## Release v0.2.0

- Adding regbot command and image to support Lua scripts with the regclient api.
  The regbot scripts allow complex mirroring logic and enable registry cleanup
  policies to be automated.
- Support registry auth without service or scope in the challenge header.

## Release v0.1.0

- Adding regsync command and image to copy images between registries.
- Locking added to regclient to handle concurrent requests.
- Improved formatting / templating functionality.
- Handling images with external layers (copying Windows images).
- Fixing permissions of `/home/appuser` in the `regclient/regctl` and
  `regclient/regsync` images.
- Bug fix for passing multiple host configs to regclient.

## Release v0.0.5

- Adding `regctl image ratelimit` command

## Release v0.0.4

- CI fix to build latest image

## Release v0.0.3

- Add alpine image variant for CI pipelines
- Add support for `repo ls` command and _catalog API
- Add pagination support on tags and repo ls
- Add support for rate limit headers in manifest API

## Release v0.0.2

- Add support for json log output
- Prompt for user and password on registry login if not set with flag
- Adding wrapper for docker CLI plugin
- Using manifest HEAD requests for digests
- Check manifest HEAD before pulling manifest in copy

## Release v0.0.1

- regclient API is usable, but not considered stable yet, there's a potential for refactoring and breaking changes.
- regctl CLI is ready for testing, please report any issues you encounter.
