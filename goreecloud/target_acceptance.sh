#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Read-only target-environment acceptance checks for GoreeCloud Search.

set -euo pipefail

base_url="${GOREECLOUD_SEARCH_BASE_URL:-http://127.0.0.1:8888}"
container_name="${GOREECLOUD_SEARCH_CONTAINER:-goreecloud-search}"
provider_suite="${GOREECLOUD_SEARCH_PROVIDER_SUITE:-1}"

usage() {
  cat <<'EOF'
Usage: goreecloud/target_acceptance.sh [--base-url URL] [--container NAME] [--skip-providers]

Runs read-only acceptance checks against a staged or deployed GoreeCloud Search instance.
It does not modify Docker, Caddy, DNS, NetBird, firewall, backup, or production state.
EOF
}

while (($#)); do
  case "$1" in
    --base-url)
      base_url="${2:?missing URL}"
      shift 2
      ;;
    --container)
      container_name="${2:?missing container name}"
      shift 2
      ;;
    --skip-providers)
      provider_suite=0
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

require() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Required command not found: $1" >&2
    exit 1
  }
}

require curl
require python3

printf 'GoreeCloud Search target acceptance\n'
printf 'Base URL: %s\n' "$base_url"
printf 'Container: %s\n' "$container_name"

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

curl --fail --silent --show-error --location \
  "$base_url/" -o "$workdir/index.html"
curl --fail --silent --show-error --location \
  "$base_url/preferences" -o "$workdir/preferences.html"
curl --fail --silent --show-error --location \
  "$base_url/about" -o "$workdir/about.html"
curl --fail --silent --show-error \
  "$base_url/healthz" -o "$workdir/healthz.txt"
curl --fail --silent --show-error --head \
  "$base_url/" -o "$workdir/headers.txt"

grep -q '<title>GoreeCloud Search</title>' "$workdir/index.html"
grep -q 'goreecloud.css' "$workdir/index.html"
grep -q 'GoreeCloud Search' "$workdir/preferences.html"
grep -q 'About GoreeCloud Search' "$workdir/about.html"
grep -qi '^X-Robots-Tag: noindex, nofollow' "$workdir/headers.txt"
grep -qi '^Referrer-Policy: no-referrer' "$workdir/headers.txt"
grep -qi '^X-Frame-Options: DENY' "$workdir/headers.txt"
grep -qi '^Permissions-Policy:.*camera=().*microphone=().*geolocation=()' "$workdir/headers.txt"

if command -v docker >/dev/null 2>&1; then
  docker inspect "$container_name" >/dev/null
  running="$(docker inspect -f '{{.State.Running}}' "$container_name")"
  health="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$container_name")"
  test "$running" = "true"
  test "$health" = "healthy"

  published="$(docker port "$container_name" 2>/dev/null || true)"
  if [[ -n "$published" ]]; then
    if grep -Evq '127\.0\.0\.1|\[::1\]' <<<"$published"; then
      echo "Unexpected non-loopback published container port:" >&2
      echo "$published" >&2
      exit 1
    fi
  fi
fi

if [[ "$provider_suite" = "1" ]]; then
  python3 goreecloud/provider_acceptance.py \
    --base-url "$base_url" \
    --suite
fi

printf '\nTarget acceptance checks passed for %s\n' "$base_url"
printf 'This result does not by itself authorize production cutover.\n'
