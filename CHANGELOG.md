# Changelog

## 4.33.0 (2026/05/30)

* Diun now exposes a container `healthcheck` command by @crazy-max in #1758 #1759
* New containerd provider by @crazy-max in #1754
* Support multiple Nomad namespaces by @crazy-max in #1755
* Added proxy support for HTTP notifiers by @crazy-max in #1751
* Added Teams workflow webhook card support by @crazy-max in #1757
* Skip generated artifact tags during registry checks by @crazy-max in #1745
* Skip non-image artifacts during registry checks by @crazy-max in #1746
* Fix mail notifications with strict SMTP relays by @crazy-max in #1753
* Document global proxy configuration by @crazy-max in #1752
* Preserve logrus fields in zerolog output by @crazy-max in #1749

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.32.0...v4.33.0

## 4.32.0 (2026/05/28)

* Added an optional Prometheus metrics endpoint by @crazy-max in #1744
* Added string helpers to notification templates by @crazy-max in #1743
* Signal REST API notification now supports text mode by @sonntam in #1611
* Telegram notification now supports configurable API URL for proxy support by @GSemekhin in #1660
* ntfy now supports notification icon by @crazy-max in #1714
* ntfy now supports click action by @crazy-max in #1738
* Mail notification now uses default `HELO` when `localName` is unset by @crazy-max in #1735
* Discord notification now supports retry after rate limits by @crazy-max in #1723
* Slack notification now supports retry after rate limits by @crazy-max in #1724
* Telegram notification now supports retry after rate limits by @crazy-max in #1725
* Matrix notification now supports retry after rate limits by @crazy-max in #1726
* Teams notification now supports retry after rate limits by @crazy-max in #1731
* Webhook notification now fails on unexpected status by @crazy-max in #1722
* UUID as secret is now supported for healtchecks by @crazy-max in #1737
* Handle database manifest entries consistently by @crazy-max in #1715
* Log startup config as structured debug field by @crazy-max in #1648
* Simplify shutdown lifecycle by @crazy-max in #1704
* Simplify provider job iteration by @crazy-max in #1706
* Modernize config defaults and remove obsolete util helpers by @crazy-max in #1705
* Fix webhook notification where HTTP method would not be correctly propagated by @crazy-max in #1630
* Fix notification HTML paragraph trimming by @crazy-max in #1716
* Go 1.26 by @crazy-max in #1695
* MkDocs Materials 9.7.5 by @crazy-max in #1676
* Migrate github.com/docker/docker to github.com/moby/moby by @crazy-max in #1601
* Migrate github.com/go-gomail/gomail to github.com/wneessen/go-mail 0.7.3 by @crazy-max in #1732 #1733 #1734
* Bump filippo.io/edwards25519 to 1.1.1 in #1616
* Bump github.com/PaulSonOfLars/gotgbot/v2 to 2.0.0-rc.35 in #1618 #1742
* Bump github.com/alecthomas/kong to 1.15.0 in #1603 #1671
* Bump github.com/bmatcuk/doublestar/v4 to 4.10.0 in #1719
* Bump github.com/crazy-max/gohealthchecks to 0.6.0 in #1587
* Bump github.com/crazy-max/gonfig to 0.8.0 in #1668
* Bump github.com/docker/go-connections to 0.7.0 in #1685
* Bump github.com/dromara/carbon/v2 to 2.6.16 in #1595
* Bump github.com/go-playground/validator/v10 to 10.30.2 in #1585 #1675
* Bump github.com/hashicorp/nomad/api to 2.0.2 in #1717
* Bump github.com/jedib0t/go-pretty/v6 to 6.7.10 in #1669 #1691
* Bump github.com/moby/buildkit to 0.30.0 in #1596 #1672 #1707
* Bump github.com/panjf2000/ants/v2 to 2.12.1 in #1586 #1604 #1646 #1741
* Bump github.com/rabbitmq/amqp091-go to 1.11.0 in #1688
* Bump github.com/rs/zerolog to 1.35.1 in #1658 #1689
* Bump go.podman.io/image/v5 to 5.40.0 in #1600 #1670 #1601
* Bump golang.org/x/mod to 0.36.0 in #1606 #1635 #1673 #1712
* Bump golang.org/x/sys to 0.45.0 in #1591 #1602 #1666 #1699 #1711
* Bump google.golang.org/grpc to 1.81.1 in #1643 #1667 #1708
* Bump gopkg.in/yaml.v3 to 3.0.1 in #1720
* Bump k8s.io/client-go to 0.35.3 in #1607 #1631 #1644
* Bump maunium.net/go/mautrix to 0.28.0 in #1613 #1641 #1710

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.31.0...v4.32.0

## 4.31.0 (2025/12/24)

* Support for negating namespaces with Kubernetes provider by @crazy-max in #1582
* Support new Matrix servers for Matrix notifications by @artgpz, @sreinwald in #1529 #1551
* Use `X-Gotify-Key` header to send token for Gotify by @crazy-max in #1530
* Add RFC 5322 compliant Message-ID header to email notifications by @tkaufmann in #1557
* Add `renderEmbeds` option for Discord notifications by @crazy-max in #1580
* Go 1.25 by @crazy-max in #1573
* Alpine Linux 3.23 by @crazy-max in #1572
* MkDocs Material 9.6.20 by @crazy-max in #1509
* Bump github.com/alecthomas/kong to 1.13.0 in #1545
* Bump github.com/containerd/platforms to 1.0.0-rc.2 in #1533
* Bump github.com/docker/docker to 28.5.2+incompatible in #1517
* Bump github.com/dromara/carbon/v2 to 2.6.15 in #1526 #1546
* Bump github.com/eclipse/paho.mqtt.golang to 1.5.1 in #1504
* Bump github.com/go-playground/validator/v10 to 10.30.0 in #1513 #1576
* Bump github.com/hashicorp/nomad/api to 1.11.1 by @crazy-max in #1579
* Bump github.com/jedib0t/go-pretty/v6 to 6.7.8 in #1569 #1583
* Bump github.com/moby/buildkit to 0.25.3 in #1516 #1566
* Bump go.podman.io/image/v5 to 5.38.0 by @crazy-max in #1502 #1535
* Bump golang.org/x/mod to 0.30.0 in #1540
* Bump google.golang.org/grpc to 1.78.0 in #1565 #1584
* Bump google.golang.org/grpc/cmd/protoc-gen-go-grpc to 1.6.0 in #1577
* Bump google.golang.org/protobuf to 1.36.11 in #1564
* Bump k8s.io/client-go to 0.34.1 in #1501
* Bump maunium.net/go/mautrix to 0.26.1 in #1567

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.30.0...v4.31.0

## 4.30.0 (2025/08/31)

* Add TLS config options `tlsSkipVerify` and `tlsCaCertFiles` for all notifiers using an HTTP client by @crazy-max in #1489
* Apprise notifications support by @privacyfr3ak in #1457
* Elasticsearch notifications support by @robin-moser in #1452
* Add `disableNotification` option for Telegram by @imrebuild in #1354
* Switch from third-party module to HTTP client for Pushover by @crazy-max in #1490
* Align `chatIDs` and `chatIDsFile` format handling for Telegram by @crazy-max in #1316
* Switch from `github.com/hako/durafmt` to `github.com/dromara/carbon` module by @crazy-max in #1317
* Remove unneeded `openssl` package in the Docker image by @crazy-max in #1488
* Go 1.24 by @crazy-max in #1461
* Alpine Linux 3.22 by @crazy-max in #1462
* Bump dario.cat/mergo to 1.0.2 in #1436
* Bump github.com/PaulSonOfLars/gotgbot/v2 to 2.0.0-rc.33 in #1397 #1448
* Bump github.com/alecthomas/kong to 1.12.1 in #1324 #1456
* Bump github.com/containers/image/v5 to 5.36.1 in #1340 #1454 #1467
* Bump github.com/crazy-max/gohealthchecks to 0.5.0 in #1319
* Bump github.com/docker/docker to 28.3.3+incompatible in #1458
* Bump github.com/docker/go-connections to 0.6.0 in #1470
* Bump github.com/dromara/carbon/v2 to 2.6.11 in #1435 #1455
* Bump github.com/go-playground/validator/v10 to 10.27.0 in #1333 #1432 #1446
* Bump github.com/hashicorp/nomad/api to 1.10.4 by @crazy-max in #1487
* Bump github.com/jedib0t/go-pretty/v6 to 6.6.8 in #1430 #1466
* Bump github.com/moby/buildkit to 0.23.2 in #1445
* Bump github.com/opencontainers/image-spec to 1.1.1 in #1434
* Bump github.com/panjf2000/ants/v2 to 2.11.3 in #1331 #1433
* Bump github.com/rs/zerolog to 1.34.0 in #1431
* Bump github.com/stretchr/testify to 1.11.1 in #1482
* Bump go.etcd.io/bbolt to 1.4.3 in #1361 #1444 #1477
* Bump golang.org/x/crypto to 0.35.0 in #1398
* Bump golang.org/x/mod to 0.27.0 in #1377 #1450 #1469
* Bump golang.org/x/net to 0.38.0 in #1343 #1402
* Bump golang.org/x/sys to 0.35.0 in #1323 #1427 #1472
* Bump google.golang.org/grpc to 1.74.2 in #1407 #1465
* Bump google.golang.org/protobuf to 1.36.8 in #1389 #1471 #1479
* Bump k8s.io/client-go to 0.32.1 in #1338 #1453

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.29.0...v4.30.0

## 4.29.0 (2024/12/19)

:warning: See **Migration notes** in the documentation before upgrading.

* Topics support for Telegram notifications by @crazy-max in #1308
* Webhook url as secret support for Discord, Slack and Teams notifications by @crazy-max in #1302
* Enhance error message for JSON decode response issues for Gotify, ntfy and RocketChat by @crazy-max in #1309
* Fix TLS configuration handling for Nomad provider by @IamTheFij in #1178
* Go 1.23 by @crazy-max in #1286
* Alpine Linux 3.21 by @crazy-max in #1286
* Switch to github.com/containerd/platforms v0.2.1 by @crazy-max in #1287
* Switch to github.com/rabbitmq/amqp091-go v1.10.0 by @crazy-max in #1288
* Bump dario.cat/mergo to 1.0.1 by @crazy-max in #1301
* Bump github.com/PaulSonOfLars/gotgbot/v2 to 2.0.0-rc.30 in #1185 #1278
* Bump github.com/alecthomas/kong to 1.6.0 in #1298
* Bump github.com/containers/image/v5 to 5.33.0 by @crazy-max in #1274 #1284
* Bump github.com/distribution/reference to 0.6.0 in #1183
* Bump github.com/docker/docker to 27.3.1+incompatible by @crazy-max in #1181 #1275 #1291
* Bump github.com/eclipse/paho.mqtt.golang to 1.5.0 in #1299
* Bump github.com/go-playground/validator/v10 to 10.23.0 in #1179 #1191 #1297
* Bump github.com/gregdel/pushover to 1.3.1 in #1164
* Bump github.com/jedib0t/go-pretty/v6 to 6.6.5 in #1167 #1300
* Bump github.com/microcosm-cc/bluemonday to 1.0.27 in #1294
* Bump github.com/moby/buildkit to 0.17.3 by @crazy-max in #1160 #1312
* Bump github.com/panjf2000/ants/v2 to 2.10.0 in #1198
* Bump github.com/rs/zerolog to 1.33.0 in #1186
* Bump github.com/stretchr/testify to 1.10.0 in #1295
* Bump go.etcd.io/bbolt to 1.3.11 by @crazy-max in #1187 #1292
* Bump golang.org/x/crypto to 0.31.0 in #1271
* Bump golang.org/x/mod to 0.22.0 in #1188 #1296
* Bump golang.org/x/net to 0.23.0 in #1157
* Bump golang.org/x/sys to 0.25.0 in #1184 #1240
* Bump google.golang.org/grpc to 1.67.0 by @crazy-max in #1171 #1293
* Bump google.golang.org/grpc/cmd/protoc-gen-go-grpc to 1.5.1 in #1224
* Bump google.golang.org/protobuf to 1.35.2 in #1277
* Bump k8s.io/client-go to 0.32.0 in #1280

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.28.0...v4.29.0

## 4.28.0 (2024/04/06)

* Add `tzdata` package to Docker image by @crazy-max in #1144
* Alpine Linux 3.19 by @crazy-max in #1140
* Bump github.com/jedib0t/go-pretty/v6 to 6.5.6 in #1137
* Bump github.com/panjf2000/ants/v2 to 2.9.1 in #1139
* Bump golang.org/x/mod to 0.17.0 in #1143
* Bump golang.org/x/sys to 0.19.0 in #1142
* Bump google.golang.org/grpc to 1.63.0 in #1141

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.27.0...v4.28.0

## 4.27.0 (2024/03/23)

* Sound option support for pushover by @szerencl in #996
* Fix NTFY markdown by @stecydube in #1025
* Fix global defaults for file provider by @adamantike in #1063
* Switch to `github.com/PaulSonOfLars/gotgbot/v2` for Telegram API client by @jon4hz in #1135
* Go 1.21 by @IamTheFij, @crazy-max in #1026 #1050 #1058
* Generate sbom and provenance by @crazy-max in #1116
* Bump github.com/alecthomas/kong to 0.9.0 in #1041 #1118
* Bump github.com/containerd/containerd to 1.7.14 in #1047 #1124
* Bump github.com/containers/image/v5 to 5.30.0 in #1029 #1112
* Bump github.com/docker/distribution to 2.8.3+incompatible in #991
* Bump github.com/docker/docker to 25.0.5+incompatible by @crazy-max in #1120 #1134
* Bump github.com/go-playground/validator/v10 to 10.19.0 in #1020 #1109
* Bump github.com/hashicorp/nomad/api to 1.7.2 by @IamTheFij in #1049
* Bump github.com/jedib0t/go-pretty/v6 to 6.5.5 in #1012 #1083 #1126
* Bump github.com/microcosm-cc/bluemonday to 1.0.26 in #1042
* Bump github.com/moby/buildkit to 0.13.1 by @crazy-max in #1043 #1111 #1117 #1128
* Bump github.com/opencontainers/image-spec to 1.1.0 in #1100
* Bump github.com/panjf2000/ants/v2 to 2.9.0 in #1046
* Bump github.com/rs/zerolog to 1.32.0 in #989 #1121
* Bump go.etcd.io/bbolt to 1.3.9 in #1044 #1106
* Bump golang.org/x/crypto to 0.17.0 in #1060
* Bump golang.org/x/mod to 0.16.0 in #1021 #1110
* Bump golang.org/x/net to 0.17.0 in #1002
* Bump golang.org/x/sys to 0.17.0 in #1035 #1092
* Bump google.golang.org/grpc to 1.62.1 in #1048 #1061 #1113
* Bump google.golang.org/protobuf to 1.33.0 in #1064 #1119
* Bump k8s.io/client-go to 0.29.3 in #1045 #1051 #1098 #1127

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.26.0...v4.27.0

## 4.26.0 (2023/09/23)

* Global `defaults` support for image configuration by @IamTheFij, @crazy-max in #887 #981 #982
* `image:tag@digest` format support by @crazy-max in #915
* Handle analysis of images with tag and digest by @crazy-max in #968
* Fix latest for `image list` command by @crazy-max in #983
* Fix dead link in reporting-issue docs by @IamTheFij in #963
* Alpine Linux 3.18 by @crazy-max in #914
* Bump github.com/AlecAivazis/survey/v2 to 2.3.7 in #900
* Bump github.com/alecthomas/kong to 0.8.0 in #905
* Bump github.com/containerd/containerd to 1.7.6 in #954
* Bump github.com/containers/image/v5 to 5.26.1 in #911
* Bump github.com/docker/docker to 24.0.6+incompatible in #947
* Bump github.com/eclipse/paho.mqtt.golang to 1.4.3 in #920
* Bump github.com/go-playground/validator/v10 to 10.15.4 in #972
* Bump github.com/gregdel/pushover to 1.3.0 in #975
* Bump github.com/jedib0t/go-pretty/v6 to 6.4.7 in #971
* Bump github.com/microcosm-cc/bluemonday to 1.0.25 in #927
* Bump github.com/moby/buildkit to 0.12.2 in #940
* Bump github.com/opencontainers/image-spec to 1.1.0-rc5 in #912 #974
* Bump github.com/panjf2000/ants/v2 to 2.8.2 in #913 #922 #978
* Bump github.com/rs/zerolog to 1.30.0 in #976
* Bump github.com/streadway/amqp to 1.1.0 in #904
* Bump golang.org/x/mod to 0.12.0 in #901 #917
* Bump golang.org/x/sys to 0.12.0 in #899 #945
* Bump google.golang.org/grpc to 1.58.2 in #906 #961 #980
* Bump google.golang.org/protobuf to 1.31.0 in #908
* Bump k8s.io/client-go to 0.28.2 in #960

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.25.0...v4.26.0

## 4.25.0 (2023/06/12)

* `runOnStartup` watch option by @crazy-max in #895
* ntfy notification support by @blueberryapple in #787
* Authentication support for ntfy by @crazy-max in #890
* Sorting for prefixed semver by @IamTheFij in #765
* Check Nomad group meta tags by @IamTheFij in #763
* Go 1.20 by @crazy-max in #858
* Bump github.com/containerd/containerd to 1.7.2 in #757 #792 #885
* Bump github.com/containers/image/v5 to 5.25.0 in #772 #791 #796 #855
* Bump github.com/crazy-max/gohealthchecks to 0.4.1 in #866
* Bump github.com/crazy-max/gonfig to 0.7.1 in #865
* Bump github.com/docker/distribution to 2.8.2+incompatible in #871
* Bump github.com/docker/docker 24.0.2+incompatible in #851 #883
* Bump github.com/go-playground/validator/v10 to 10.14.10 in #778 #852 #896
* Bump github.com/gregdel/pushover to 1.2.0 in #867
* Bump github.com/imdario/mergo to 0.3.16 in #830 #898
* Bump github.com/jedib0t/go-pretty/v6 to 6.4.4 in #760 #803
* Bump github.com/microcosm-cc/bluemonday to 1.0.24 in #780 #810 #876
* Bump github.com/moby/buildkit to 0.11.6 in #790 #809 #848
* Bump github.com/opencontainers/runc to 1.1.5 in #834
* Bump github.com/panjf2000/ants/v2 to 2.7.5 in #846 #889
* Bump github.com/rs/zerolog to 1.29.1 in #777 #854
* Bump github.com/stretchr/testify to 1.8.4 in #801 #897
* Bump go.etcd.io/bbolt to 1.3.7 in #781
* Bump golang.org/x/mod to 0.10.0 in #786 #808 #837
* Bump golang.org/x/net to 0.7.0 in #793
* Bump golang.org/x/sys to 0.8.0 in #784 #807 #857
* Bump google.golang.org/grpc to 1.52.0 in #762 #785 #826 #864
* Bump google.golang.org/grpc/cmd/protoc-gen-go-grpc to 1.3.0 in #806
* Bump google.golang.org/protobuf to 1.30.0 in #818

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.24.0...v4.25.0

## 4.24.0 (2022/12/29)

* Entry metadata field by @crazy-max in #749
* Jitter watch option by @crazy-max in #746
* Allow customizing Signal notification message by @crazy-max in #748

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.23.1...v4.24.0

## 4.23.1 (2022/12/28)

* Fix release file extension by @suzuki-shunsuke in #743

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.23.0...v4.23.1

## 4.23.0 (2022/12/28)

* Nomad provider by @IamTheFij, @crazy-max in #722 #739 #742
* Signal (REST API) notifications support by @MrRagga- in #650
* Fix email notification message template by @crazy-max in #740
* Fix panics when parsing notification templates by @crazy-max in #741
* Fix test notification typo by @vilm3r in #677
* docs: Fix `sort_tags` by @hatamiarash7 in #655
* docs: List valid log levels by @sliekens in #668
* docs: Fix the issues URL by @DougEdey in #697
* docs: New blog posts from the community by @crazy-max in #657
* Go 1.19 by @crazy-max in #701
* Alpine Linux 3.17 by @crazy-max in #735
* Fix proto gen by @crazy-max in #720
* Enhance workflow by @crazy-max in #706
* Use `GITHUB_REF` when tag pushed for versioning by @crazy-max in #707
* Bump github.com/AlecAivazis/survey/v2 to 2.3.6 in #686
* Bump github.com/alecthomas/kong to 0.7.1 in #718
* Bump github.com/containerd/containerd to 1.6.14 in #669 #719 #732
* Bump github.com/containers/image/v5 to 5.23.1 in #692 #716
* Bump github.com/crazy-max/gonfig to 0.6.0 in #651
* Bump github.com/docker/go-units to 0.5.0 in #678
* Bump github.com/eclipse/paho.mqtt.golang to 1.4.2 in #711
* Bump github.com/go-playground/validator/v10 to 10.11.1 in #699
* Bump github.com/jedib0t/go-pretty/v6 to 6.4.3 in #694 #715 #724
* Bump github.com/microcosm-cc/bluemonday to 1.0.21 in #695
* Bump github.com/moby/buildkit to 0.10.6 by @crazy-max in #738
* Bump github.com/opencontainers/image-spec to 1.1.0-rc2 in #702
* Bump github.com/panjf2000/ants/v2 to 2.7.1 in #709 #733
* Bump github.com/pkg/profile to 1.7.0 in #705
* Bump github.com/rs/zerolog to 1.28.0 in #676
* Bump github.com/stretchr/testify to 1.8.1 in #708
* Bump github.com/tidwall/pretty to 1.2.1 in #698
* Bump golang.org/x/mod to 0.7.0 by @crazy-max in #736
* Bump golang.org/x/sys to 0.3.0 by @crazy-max in #737
* Bump google.golang.org/grpc to 1.51.0 in #696 #721
* Bump google.golang.org/protobuf to 1.28.1 in #661
* Bump k8s.io/client-go to 0.25.4 in #689 #717

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.22.0...v4.23.0

## 4.22.0 (2022/07/17)

* Allow customizing the hub link by @crazy-max in #648
* Use OCI image url label to override hub link by @crazy-max in #646
* Tags sorting support by @crazy-max in #645
* Alpine Linux 3.16 by @crazy-max in #647
* Go 1.18 by @crazy-max in #592
* MkDocs Material 8.3.9 by @crazy-max in #644
* Explain roles required for rocketchat notification by @hofbi in #553
* Bump github.com/AlecAivazis/survey/v2 to 2.3.5 in #585 #625
* Bump github.com/alecthomas/kong to 0.6.1 in #549 #558 #576 #630
* Bump github.com/containerd/containerd to 1.6.0 in #557
* Bump github.com/containers/image/v5 to 5.21.1 in #552 #588 #603
* Bump github.com/docker/docker to 20.10.3-0.20220414164044-61404de7df1a in #575
* Bump github.com/eclipse/paho.mqtt.golang to 1.4.1 in #623
* Bump github.com/go-playground/validator/v10 to 10.11.0 in #568 #602
* Bump github.com/imdario/mergo to 0.3.13 in #617
* Bump github.com/jedib0t/go-pretty/v6 to 6.3.5 in #555 #584 #595 #642
* Bump github.com/microcosm-cc/bluemonday to 1.0.19 in #554 #636
* Bump github.com/moby/buildkit to 0.10.1-0.20220712094726-874eef9b70db by @crazy-max in #578 #590 #610 #643
* Bump github.com/panjf2000/ants/v2 to 2.5.0 in #563 #611
* Bump github.com/rs/zerolog to 1.27.0 in #626
* Bump github.com/stretchr/testify to 1.8.0 in #635
* Bump google.golang.org/grpc to 1.48.0 in #615 #639
* Bump google.golang.org/protobuf to 1.28.0 in #582
* Bump k8s.io/client-go to 0.24.3 in #561 #580 #604 #640

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.21.0...v4.22.0

## 4.21.0 (2022/01/26)

* Add `image prune` command by @crazy-max in #519
* Fix matrix login scheme by @crazy-max in #487
* Move `syscall` to `golang.org/x/sys` by @crazy-max in #525
* Move from `io/ioutil` to `os` package by @crazy-max in #524
* Fix notif template in docs
* Enhance dockerfiles by @crazy-max in #523
* Add binary bake target by @crazy-max in #517
* MkDocs Material 8.1.8 by @crazy-max in #520 #548
* Alpine Linux 3.15 by @crazy-max in #527
* goreleaser-xx 1.2.5 by @crazy-max in #539
* Bump github.com/alecthomas/kong to 0.3.0 in #507 #537
* Bump github.com/containerd/containerd to 1.5.8 in #494 #496 #509
* Bump github.com/containers/image/v5 to 5.19.0 in #498 #536 #546
* Bump github.com/docker/docker to 20.10.12+incompatible in #500 #510 #531
* Bump github.com/go-playground/validator/v10 to 10.10.0 in #538
* Bump github.com/jedib0t/go-pretty/v6 to 6.2.5 in #543
* Bump github.com/microcosm-cc/bluemonday to 1.0.17 in #499 #535
* Bump github.com/moby/buildkit to 0.9.3 in #495 #506 #512
* Bump github.com/opencontainers/image-spec to v1.0.2-0.20211117181255-693428a734f5 by @crazy-max in #513
* Bump github.com/panjf2000/ants/v2 to 2.4.7 in #532
* Bump github.com/rs/zerolog to 1.26.1 in #485 #502 #534
* Bump google.golang.org/grpc to 1.44.0 in #492 #505 #529 #545
* Bump google.golang.org/grpc/cmd/protoc-gen-go-grpc to 1.2.0 in #533
* Bump k8s.io/client-go to 0.22.4 in #490 #511

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.20.1...v4.21.0

## 4.20.1 (2021/09/06)

* Fix notification title by @crazy-max in #483

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.20.0...v4.20.1

## 4.20.0 (2021/09/05)

* Option to render fields by @crazy-max in #480
* Allow choosing status to be notified by @crazy-max in #475
* Enhance notif wording by @crazy-max in #467
* Wrong remaining time displayed by @crazy-max in #469
* Allow multi recipients for email notifier by @crazy-max in #463
* Provide mutable tags for Diun image by @crazy-max in #462
* Fix Dockerfile parser and add tests by @crazy-max in #459
* Add e2e tests by @crazy-max in #471
* Use args in kubernetes documentation example by @paddatrapper in #424
* Fix j2 variable in docs by @crazy-max in #422
* Note to customize the hostname by @crazy-max in #465
* Go 1.17 by @crazy-max in #458
* Add `windows/arm64` artifact by @crazy-max in #472
* Add `linux/riscv64` artifact by @crazy-max in #427
* Alpine Linux 3.14 by @crazy-max in #426
* MkDocs Material 7.2.6 by @crazy-max in #428 #482
* Protoc 3.17.3 by @crazy-max in #461
* Bump codecov/codecov-action to 2
* Bump github.com/containerd/containerd to 1.5.5 in #433 #440 #447
* Bump github.com/containers/image/v5 to 5.16.0 in #460 #476
* Bump github.com/crazy-max/gonfig to 0.5.0 in #474
* Bump github.com/docker/docker to 20.10.8 in #451
* Bump github.com/go-playground/validator/v10 to 10.9.0 in #429 #445 #455
* Bump github.com/gregdel/pushover to 1.1.0 by @crazy-max in #470
* Bump github.com/jedib0t/go-pretty/v6 to 6.2.4 in #432
* Bump github.com/microcosm-cc/bluemonday to 1.0.15 in #430
* Bump github.com/moby/buildkit to 0.9.0 in #437
* Bump github.com/rs/zerolog to 1.24.0 in #477
* Bump github.com/streadway/amqp to 1.0.0 by @crazy-max in #470
* Bump google.golang.org/grpc to 1.40.0 in #421 #453 #456
* Bump google.golang.org/protobuf to 1.27.1 in #420
* Bump k8s.io/client-go to 0.22.1 in #466

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.19.0...v4.20.0

## 4.19.0 (2021/06/26)

* Allow customizing notification message by @crazy-max in #415
* Bump github.com/containers/image/v5 to 5.13.2 in #412
* Bump github.com/microcosm-cc/bluemonday to 1.0.14 in #413
* Bump github.com/panjf2000/ants/v2 to 2.4.6 in #416
* Bump k8s.io/client-go to 0.21.2 in #414

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.18.0...v4.19.0

## 4.18.0 (2021/06/18)

* Handle registry auth config by @crazy-max in #411
* Bump github.com/alecthomas/kong to 0.2.17 in #401
* Bump github.com/containers/image/v5 to 5.13.1 in #409
* Bump github.com/eclipse/paho.mqtt.golang to 1.3.5 in #400
* Bump github.com/jedib0t/go-pretty/v6 to 6.2.2 in #388
* Bump github.com/microcosm-cc/bluemonday to 1.0.13 in #403 #410
* Bump github.com/rs/zerolog to 1.23.0 in #405
* Bump github.com/tidwall/pretty to 1.2.0 in #390 #402
* Bump go.etcd.io/bbolt to 1.3.6 in #394
* Bump google.golang.org/grpc to 1.38.0 in #389
* Bump k8s.io/client-go to 0.21.1 in #381
* Avoid notification for unupdated image by @crazy-max in #406
* Use `openssl` pkg by @crazy-max in #407
* Bumps github.com/docker/docker to 20.10.7+incompatible in #397
* Set `cacheonly` output for validators by @crazy-max in #395
* Define serve command by @crazy-max in #393
* Save raw manifest in db by @crazy-max in #391

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.17.0...v4.18.0

## 4.17.0 (2021/05/26)

:warning: See **Migration notes** in the documentation before upgrading.

* Add CLI to interact with Diun through gRPC by @crazy-max in #382
    * Create `image` and `notif` proto services
    * Implement proto definitions
    * New commands `serve`, `image` and `notif`
    * Refactor command line usage doc
    * Better CLI error handling
    * Tools build constraint to manage tools deps through go modules
    * Compile and validate protos through a dedicated Dockerfile and a bake target    
    * Merge validate and build workflow
    * Add upgrade notes
* Bump github.com/containerd/containerd to 1.5.2 by @crazy-max in #353 #361 #362 #383
* Bump github.com/containers/image/v5 to 5.12.0 in #360
* Bump github.com/eclipse/paho.mqtt.golang to 1.3.4 in #359
* Bump github.com/go-playground/validator/v10 to 10.6.1 in #377
* Bump github.com/moby/buildkit to 0.8.3 in #354
* Bump github.com/panjf2000/ants/v2 to 2.4.5 in #380
* Bump github.com/pkg/profile to 1.6.0 in #363
* Bump github.com/rs/zerolog to 1.22.0 in #379
* MkDocs Materials 7.1.5 by @crazy-max in #386
* Add `NO_COLOR` support by @crazy-max in #384
* Move to docker/metadata-action by @crazy-max in #366
* Add blog posts by @crazy-max in #355 #385

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.16.1...v4.17.0

## 4.16.1 (2021/04/30)

* Fix Swarm Provider by @crazy-max in #351

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.16.0...v4.16.1

## 4.16.0 (2021/04/29)

* Dockerfile provider by @crazy-max in #329
* Note about `watch_repo` setting by @crazy-max in #348
* Contribute to doc by @crazy-max in #347
* Update docs for Podman support by @signed-log in #345
* Optional profiler volume by @crazy-max in #344

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.15.2...v4.16.0

## 4.15.2 (2021/04/25)

* Make profiler optional by @crazy-max in #341

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.15.1...v4.15.2

## 4.15.1 (2021/04/25)

* Fix profiler path by @crazy-max in #339

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.15.0...v4.15.1

## 4.15.0 (2021/04/25)

* Add `darwin/arm64` artifact by @crazy-max in #338
* MkDocs Materials 7.1.3 by @crazy-max in #337
* Add profiler flag by @crazy-max in #336
* Handle digest based image reference by @crazy-max in #335
* Bump github.com/containers/image/v5 to 5.11.1 in #323 #330
* Bump github.com/docker/docker to 20.10.6+incompatible in #324
* Bump github.com/eclipse/paho.mqtt.golang to 1.3.3 in #316
* Bump github.com/go-playground/validator/v10 to 10.5.0 in #319
* Bump github.com/microcosm-cc/bluemonday to 1.0.9 in #311 #321 #325 #333
* Bump github.com/panjf2000/ants/v2 to 2.4.4 in #312
* Bump github.com/rs/zerolog to 1.21.0 in #309
* Deploy docs on workflow dispatch or tag by @crazy-max in #305

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.14.0...v4.15.0

## 4.14.0 (2021/03/15)

* Bump github.com/alecthomas/kong to 0.2.16 in #295
* Bump github.com/containers/image/v5 to 5.10.5 in #290
* Bump github.com/crazy-max/gohealthchecks to 0.3.0 in #296
* Bump github.com/imdario/mergo to 0.3.12 in #298
* Bump k8s.io/client-go to 0.20.4 in #280
* Docker client 20.10.5 by @crazy-max in #303
* Allow telegram chat IDs as file by @crazy-max in #301
* Go 1.16 by @crazy-max in #302
* Handle git ref for artifact target
* Allow configuring scheme for MQTT broker by @fblackburn1 in #292
* Switch to [goreleaser-xx](https://github.com/crazy-max/goreleaser-xx) by @crazy-max in #291

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.13.0...v4.14.0

## 4.13.0 (2021/03/01)

* Missing token as secret setting for some notifiers by @crazy-max in #289
* Allow disabling log color output by @crazy-max in #288
* Bump github.com/containers/image/v5 to 5.10.4 in #271 #282 #284
* Cleanup workflows by @crazy-max in #281 #287
* Do not check recipient details for Pushover by @crazy-max in #277
* MkDocs Materials 6.2.8 by @crazy-max in #276
* Fix markdown renderer by @crazy-max in #275
* Add message client for notifiers by @crazy-max in #273

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.12.0...v4.13.0

## 4.12.0 (2021/02/09)

* Use digest as comparison footprint by @crazy-max in #269
* Bump github.com/alecthomas/kong to 0.2.15 in #270
* Bump github.com/containers/image/v5 to 5.10.1 in #265
* Bump github.com/eclipse/paho.mqtt.golang to 1.3.2 in #268
* Move to [docker/bake-action](https://github.com/docker/bake-action) by @crazy-max in #266
* Typo in documentation by @TheCatLady in #258
* Log image validation

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.11.0...v4.12.0

## 4.11.0 (2021/01/04)

* Fix DB migration by @crazy-max in #255
* Add Pushover notification by @crazy-max in #254
* Avoid duplicated notifications with Kubernetes DaemonSet by @crazy-max in #252
* Make scheduler optional by @crazy-max in #251
* Bump github.com/eclipse/paho.mqtt.golang to 1.3.1 in #249
* Handle exclusions as a distinct status by @crazy-max in #248

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.10.0...v4.11.0

## 4.10.0 (2020/12/26)

* Refactor CI and dev workflow with buildx bake by @crazy-max in #247
    * Upload artifacts
    * Add `image-local` target
    * Single job for artifacts and image
    * Add `armv5` artifact
* MQTT Reconnection Log Spam by @aschoelzhorn in #241
* Add Docker + File providers user guide by @crazy-max in #239
* Bump github.com/alecthomas/kong to 0.2.12 in #231
* Bump github.com/containers/image/v5 to 5.8.1 in #226
* Bump github.com/containers/image/v5 to 5.9.0 in #236
* Bump github.com/eclipse/paho.mqtt.golang to 1.3.0 in #235
* Bump gopkg.in/yaml.v2 to 2.4.0 in #228

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.9.0...v4.10.0

## 4.9.0 (2020/11/16)

* Fix duplicated notifications
* Remove support for `freebsd/*` (moby/moby#38818)
* Add support for `linux/ppc64le` and `linux/s390x` (binary)
* Bump github.com/containers/image/v5 to 5.8.0
* Bump k8s.io/client-go to 0.19.4 in #224

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.8.1...v4.9.0

## 4.8.1 (2020/11/14)

* Fix registry timeout context in #221
* Image closer not required while fetching tags

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.8.0...v4.8.1

## 4.8.0 (2020/11/13)

* Go 1.15 by @crazy-max in #218
* Remove `linux/s390x` platform support (for now)
* Check digest from HEAD request by @crazy-max in #217
* Add FAQ note about Docker Hub rate limits
* Compare digest as watch setting
* Optimize build time
* Add hub link for GitHub Container Registry by @crazy-max in #211
* Update deps

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.7.0...v4.8.0

## 4.7.0 (2020/11/02)

* Add MQTT notification by @aschoelzhorn in #192
* Docker image also available on [GitHub Container Registry](https://github.com/users/crazy-max/packages/container/package/diun)
* Use zoneinfo from Go in #202
* Remove `--timezone` flag
* Use Docker meta action to handle tags and labels
* Update deps

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.6.1...v4.7.0

## 4.6.1 (2020/10/22)

* Typos in documentation
* Bump codecov/codecov-action to v1.0.14 in #195
* Bump docker/login-action to v1.6.0 in #198
* Bump github.com/go-playground/validator/v10 to 10.4.1 in #197
* Bump github.com/panjf2000/ants/v2 to 2.4.3 in #196
* Bump k8s.io/client-go to 0.19.3 in #199

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.6.0...v4.6.1

## 4.6.0 (2020/10/13)

* Add support for [Healthchecks](https://healthchecks.io/) to monitor Diun watcher in #78
* Add option to mention specific users or roles for Discord notifier in #188
* Update docker install documentation
* Add "Too many requests to registry" section in FAQ in #168
* Update deps
* Switch to [Docker actions](https://github.com/docker/build-push-action)

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.5.0...v4.6.0

## 4.5.0 (2020/08/29)

* Allow setting the hostname sent to the SMTP server with the HELO command for mail notification in #165
* Fix Telegram notification error in #162

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.4.1...v4.5.0

## 4.4.1 (2020/08/20)

* Allow using `--test-notif` without providers and DB connection in #157 #150
* Update deps

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.4.0...v4.4.1

## 4.4.0 (2020/08/08)

* Allow customizing message type for Matrix notifications in #143

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.3.1...v4.4.0

## 4.3.1 (2020/07/30)

* Hostname not taken into account for Matrix notifications

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.3.0...v4.3.1

## 4.3.0 (2020/07/29)

* Add Matrix notification in #124

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.2.0...v4.3.0

## 4.2.0 (2020/07/16)

* Seek configuration file from default places in #107
* Switch to [gonfig](https://github.com/crazy-max/gonfig)
* Update deps

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.1.1...v4.2.0

## 4.1.1 (2020/06/26)

* Small typo

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.1.0...v4.1.1

## 4.1.0 (2020/06/26)

* Discord notifications by @crazy-max in #110 #111
* Update migration notes in #107
* Logging when configuration file or `DIUN_` env vars not found in #107
* Bump github.com/containers/image/v5 to 5.5.1 in #96

**Full Changelog**: https://github.com/crazy-max/diun/compare/v4.0.0...v4.1.0

## 4.0.0 (2020/06/22)

:warning: See **Migration notes** in the documentation for breaking changes.

* Display hostname in notifications in #102
* Automatically determine registry options based on image name by @crazy-max in #103
* Docs website with mkdocs in #99
* Skip dangling images in #98
* More explicit message if manifest not found in #94
* Add swarm example
* Update doc for file and Swarm providers
* Add Kubernetes provider in #25
* Update Teams notification screenshot by @margaale in #93
* Send message as markdown for Gotify and Telegram notifiers
* Add link to respective hub in #40
* Configuration transposed into environment variables by @crazy-max in #82
* Configuration file not required anymore
* `DIUN_DB` env var renamed `DIUN_DB_PATH`
* Only accept duration as timeout value (`10` becomes `10s`)
* Enhanced documentation in #83
* Add note about test notifications by @Tooa in #79
* Improve configuration validation
* Fix telegram init
* All fields in configuration are now _camelCased_
* Docker API version negotiation in #29
* Add Mattermost compatibility via Slack webhooks by @Twilek-de in #80
* Update deps

**Full Changelog**: https://github.com/crazy-max/diun/compare/v3.0.0...v4.0.0

## 3.0.0 (2020/05/27)

:warning: See **Migration notes** in the documentation for breaking changes.

* Add script notification in #53
* Add Teams notification by @margaale in #72
* Add `--test-notif` flag in #23
* Allow only one Docker and Swarm provider
* Remove "enable" setting for notifiers
* Logging when no image is found
* Add Amqp notification client by @margaale in #63
* Fix default log level
* Move static to file provider by @crazy-max in #71
* Reload config on change for file provider in #16
* Switch to kong command-line parser by @crazy-max in #66
* Enhanced Dockerfile
* Review of platform detection in #57
* Leave default image platform empty for file provider (see FAQ doc)
* Handle platform variant
* Add database migration process
* Switch to Open Container Specification labels as label-schema.org ones are deprecated
* Remove unneeded `diun.os` and `diun.arch` docker labels
* Add upgrade notes
* Update deps

**Full Changelog**: https://github.com/crazy-max/diun/compare/v2.6.1...v3.0.0

## 2.6.1 (2020/03/26)

* Downgrade containers/image to 5.2.1 in #54

**Full Changelog**: https://github.com/crazy-max/diun/compare/v2.6.0...v2.6.1

## 2.6.0 (2020/03/26)

* Fix service image inspection in #52
* Docker client v19.03.8
* Update deps

**Full Changelog**: https://github.com/crazy-max/diun/compare/v2.5.0...v2.6.0

## 2.5.0 (2020/03/01)

* Add Rocket.Chat notifier by @crazy-max in #44

**Full Changelog**: https://github.com/crazy-max/diun/compare/v2.4.0...v2.5.0

## 2.4.0 (2020/02/17)

* Add Gotify notification client by @crazy-max in #36
* Bump containers/image v5 by @crazy-max in #35

**Full Changelog**: https://github.com/crazy-max/diun/compare/v2.3.0...v2.4.0

## 2.3.0 (2020/01/28)

* Add Telegram notifier by @DanNixon in #30
* Docker client struct options
* Move registry client to a dedicated package

**Full Changelog**: https://github.com/crazy-max/diun/compare/v2.2.1...v2.3.0

## 2.2.1 (2020/01/07)

* Set user agent for Docker registry client
* Update deps

**Full Changelog**: https://github.com/crazy-max/diun/compare/v2.2.0...v2.2.1

## 2.2.0 (2019/12/22)

* Add option to skip notification at the very first analysis of an image in #10
* Skip analysis of locally built image

**Full Changelog**: https://github.com/crazy-max/diun/compare/v2.1.0...v2.2.0

## 2.1.0 (2019/12/17)

* Add Slack notifier in #8

**Full Changelog**: https://github.com/crazy-max/diun/compare/v2.0.0...v2.1.0

## 2.0.0 (2019/12/14)

:warning: See **Migration notes** in the documentation for breaking changes.

* Include provider in notifications
* Add providers documentation
* Move image validation and improve job execution
* Add Swarm provider
* Add fields to load sensitive values from file in #7
* Add Docker provider in #3
* Docker client v19.03.5
* Move `image` field to providers layer and rename it `static`
* Update deps
* Go 1.13.5
* Seconds field optional for schedule

**Full Changelog**: https://github.com/crazy-max/diun/compare/v1.4.1...v2.0.0

## 1.4.1 (2019/10/20)

* Update deps
* Fix Docker labels

**Full Changelog**: https://github.com/crazy-max/diun/compare/v1.4.0...v1.4.1

## 1.4.0 (2019/10/01)

* Multi-platform Docker image
* Switch to GitHub Actions
* Stop publishing Docker image on Quay
* Go 1.12.10
* Use GOPROXY

**Full Changelog**: https://github.com/crazy-max/diun/compare/v1.3.0...v1.4.0

## 1.3.0 (2019/08/22)

* Add Linux service doc and sample
* Move documentation
* Fix go mod
* Remove `--docker` flag
* Allow overriding database path through `DIUN_DB` env var

**Full Changelog**: https://github.com/crazy-max/diun/compare/v1.2.0...v1.3.0

## 1.2.0 (2019/08/18)

* Update deps
* Display containers/image logs
* Fix registry options not setted in #5

**Full Changelog**: https://github.com/crazy-max/diun/compare/v1.1.0...v1.2.0

## 1.1.0 (2019/07/24)

* Update deps

**Full Changelog**: https://github.com/crazy-max/diun/compare/v1.0.2...v1.1.0

## 1.0.2 (2019/07/01)

* Worker pool can be full while retrieving tags

**Full Changelog**: https://github.com/crazy-max/diun/compare/v1.0.1...v1.0.2

## 1.0.1 (2019/07/01)

* Fix runtime error

**Full Changelog**: https://github.com/crazy-max/diun/compare/v1.0.0...v1.0.1

## 1.0.0 (2019/07/01)

:warning: See **Migration notes** in the documentation for breaking changes.

* Always run on startup. Flag `--run-startup` removed.
* Display next execution time
* Use v3 robfig/cron
* Move `Os` and `Arch` filters to image
* Retrieve all tags by default
* Review config file structure
* Improve worker pool

**Full Changelog**: https://github.com/crazy-max/diun/compare/v0.5.0...v1.0.0

## 0.5.0 (2019/06/09)

* Add worker pool to parallelize analyses

**Full Changelog**: https://github.com/crazy-max/diun/compare/v0.4.1...v0.5.0

## 0.4.1 (2019/06/08)

* Filter tags before return them

**Full Changelog**: https://github.com/crazy-max/diun/compare/v0.4.0...v0.4.1

## 0.4.0 (2019/06/08)

* Add option to set the maximum number of tags to watch for an item if `watch_repo` is enabled

**Full Changelog**: https://github.com/crazy-max/diun/compare/v0.3.2...v0.4.0

## 0.3.2 (2019/06/08)

* Fix registry client context

**Full Changelog**: https://github.com/crazy-max/diun/compare/v0.3.1...v0.3.2

## 0.3.1 (2019/06/08)

* Fix email template
* Add flag to log caller

**Full Changelog**: https://github.com/crazy-max/diun/compare/v0.3.0...v0.3.1

## 0.3.0 (2019/06/08)

* Allow overriding os and architecture when watching
* Move `insecure_tls` and `timeout` options to registry option
* Rename Bolt bucket
* Change default schedule
* Review registry client

**Full Changelog**: https://github.com/crazy-max/diun/compare/v0.2.0...v0.3.0

## 0.2.0 (2019/06/05)

* Don't skip repo analysis if default tag not found
* Docker engine 18.09.6

**Full Changelog**: https://github.com/crazy-max/diun/compare/v0.1.1...v0.2.0

## 0.1.1 (2019/06/04)

* Increase default timeout
* Fix `data` volume mount

**Full Changelog**: https://github.com/crazy-max/diun/compare/v0.1.0...v0.1.1

## 0.1.0 (2019/06/04)

* Initial version
