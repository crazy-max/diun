FROM alpine:3.14 AS base
COPY ./config /etc/config

FROM base AS final
COPY --from=crazymax/yasu / /
