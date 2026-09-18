FROM python:3.12-slim-bookworm

ARG VCS_REF=unknown
ARG VERSION=development

LABEL org.opencontainers.image.title="GoreeCloud Search" \
      org.opencontainers.image.description="Privacy-focused GoreeCloud Search service" \
      org.opencontainers.image.source="https://github.com/GoreeCloud/goreecloud-search" \
      org.opencontainers.image.revision="$VCS_REF" \
      org.opencontainers.image.version="$VERSION" \
      org.opencontainers.image.licenses="AGPL-3.0-or-later"

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONPATH=/opt/goreecloud-search

RUN useradd --system --uid 10001 --home /nonexistent --shell /usr/sbin/nologin goreecloud

WORKDIR /opt/goreecloud-search
COPY --chown=10001:10001 src/goreecloud_search ./goreecloud_search

USER 10001:10001
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["python","-c","from urllib.request import urlopen; r=urlopen('http://127.0.0.1:8080/healthz', timeout=2); raise SystemExit(0 if r.status == 200 else 1)"]

CMD ["python","-m","goreecloud_search.cli","serve","--host","0.0.0.0","--port","8080"]
