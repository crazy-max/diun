# syntax=docker/dockerfile:1.2

# diun.platform=linux/amd64
FROM alpine:latest

# diun.watch_repo=true
# diun.max_tags=10
# diun.platform=linux/amd64
COPY --from=crazymax/yasu / /

# diun.watch_repo=true
# diun.include_tags=^\d+\.\d+\.\d+$
# diun.platform=linux/amd64
RUN --mount=type=bind,target=.,rw \
  --mount=type=bind,from=docker/buildx-bin:0.6.0,source=/buildx,target=/usr/bin/buildx \
  yasu --version
