# syntax=docker/dockerfile:1.2

FROM squidfunk/mkdocs-material:7.1.3 AS base
RUN apk add --no-cache \
    git \
    git-fast-import \
    openssh \
  && apk add --no-cache --virtual .build gcc musl-dev \
  && pip install --no-cache-dir \
    'markdown-include' \
    'mkdocs-awesome-pages-plugin' \
    'mkdocs-exclude' \
    'mkdocs-git-revision-date-localized-plugin' \
    'mkdocs-macros-plugin' \
  && apk del .build gcc musl-dev \
  && rm -rf /tmp/*

FROM base AS generate
RUN --mount=type=bind,target=. \
  mkdocs build --strict --site-dir /tmp/site

FROM scratch AS release
COPY --from=generate /tmp/site/ /
