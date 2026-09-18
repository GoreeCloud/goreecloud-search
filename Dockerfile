FROM python:3.12.14-slim-bookworm@sha256:782412e85d0f0984994c290652577d4018aff08145c85b262bb63dc0c7522254

ARG GORECLOUD_SEARCH_VERSION=development
ARG GORECLOUD_SEARCH_REVISION=unknown

LABEL org.opencontainers.image.title="GoreeCloud Search" \
      org.opencontainers.image.description="Native GoreeCloud Search service" \
      org.opencontainers.image.source="https://github.com/GoreeCloud/goreecloud-search" \
      org.opencontainers.image.version="${GORECLOUD_SEARCH_VERSION}" \
      org.opencontainers.image.revision="${GORECLOUD_SEARCH_REVISION}" \
      org.opencontainers.image.licenses="AGPL-3.0-or-later"

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONPATH=/opt/goreecloud-search

WORKDIR /opt/goreecloud-search

COPY --chown=10001:10001 src/goreecloud_search ./goreecloud_search
COPY --chown=10001:10001 LICENSE NOTICE VERSION /usr/share/doc/goreecloud-search/

USER 10001:10001

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["python", "-c", "import json,urllib.request; r=urllib.request.urlopen('http://127.0.0.1:8080/healthz', timeout=2); p=json.load(r); raise SystemExit(0 if r.status == 200 and p.get('status') == 'ok' else 1)"]

ENTRYPOINT ["python", "-m", "goreecloud_search.cli"]
CMD ["serve-container", "--port", "8080", "--service-hostname", "search.goreecloud.com"]
