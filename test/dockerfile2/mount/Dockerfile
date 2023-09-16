# syntax=docker/dockerfile:1

# diun.platform=linux/amd64
FROM alpine:latest

# diun.platform=linux/amd64
# diun.metadata.foo=bar
RUN --mount=type=bind,target=.,rw \
  --mount=type=bind,from=crazymax/undock:0.5.0@sha256:736fdfde1268b93c2f733c53a7c45ece24e275318628fbb790bee7f89459961f,source=/usr/local/bin/undock,target=/usr/local/bin/undock \
  undock --version

# diun.platform=linux/amd64
# diun.metadata.foo=bar
RUN --mount=type=bind,target=.,rw \
  --mount=type=bind,from=crazymax/ddns-route53:foo@sha256:9cb3af44cdd00615266c87e60bc05cac534297be14c4596800b57322f9313615,source=/usr/local/bin/ddns-route53,target=/usr/local/bin/ddns-route53 \
  ddns-route53 --version
