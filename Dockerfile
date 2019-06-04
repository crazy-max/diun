FROM golang:1.12.4 as builder

ARG BUILD_DATE
ARG VCS_REF
ARG VERSION

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go version
RUN go mod download
COPY . ./
RUN cp /usr/local/go/lib/time/zoneinfo.zip ./ \
  && CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-w -s -X 'main.version=${VERSION}'" \
    -v -o diun cmd/main.go

FROM alpine:latest

ARG BUILD_DATE
ARG VCS_REF
ARG VERSION

LABEL maintainer="CrazyMax" \
  org.label-schema.build-date=$BUILD_DATE \
  org.label-schema.name="Diun" \
  org.label-schema.description="Docker image update notifier" \
  org.label-schema.version=$VERSION \
  org.label-schema.url="https://github.com/crazy-max/diun" \
  org.label-schema.vcs-ref=$VCS_REF \
  org.label-schema.vcs-url="https://github.com/crazy-max/diun" \
  org.label-schema.vendor="CrazyMax" \
  org.label-schema.schema-version="1.0"

RUN apk --update --no-cache add \
    ca-certificates \
    libressl \
    tzdata \
  && rm -rf /tmp/* /var/cache/apk/*

COPY --from=builder /app/diun /usr/local/bin/diun
COPY --from=builder /app/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip

VOLUME [ "/data" ]

CMD [ "diun", "--config", "/diun.yml", "--docker" ]
