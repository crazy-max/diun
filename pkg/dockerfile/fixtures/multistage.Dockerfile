FROM alpine:3.14 AS base
COPY ./config /etc/config

FROM base AS final
COPY --from=base /etc/config /etc/config
COPY --from=crazymax/yasu / /
