# syntax=docker/dockerfile:1

ARG ALPINE_VERSION="3.24"

FROM alpine:${ALPINE_VERSION}
RUN apk add --no-cache containerd containerd-ctr runc
COPY test/fixtures/containerd1/start.sh /start.sh
CMD ["sh", "/start.sh"]
