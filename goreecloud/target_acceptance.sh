#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Read-only target-environment acceptance checks for GoreeCloud Search.

set -euo pipefail

base_url="${GOREECLOUD_SEARCH_BASE_URL:-http://127.0.0.1:8888}"
container_name="${GOREECLOUD_SEARCH_CONTAINER:-goreecloud-search}"
provider_suite="${GOREECLOUD_SEARCH_PROVIDER_SUITE:-1}"
expected_image="${GOREECLOUD_SEARCH_EXPECTED_IMAGE:-}"
expected_source="${GOREECLOUD_SEARCH_EXPECTED_SOURCE:-}"
evidence_json="${GOREECLOUD_SEARCH_EVIDENCE_JSON:-}"

usage() {
  cat <<'EOF'
Usage: goreecloud/target_acceptance.sh [OPTIONS]

Runs read-only acceptance checks against a staged or deployed GoreeCloud Search instance.
It does not modify Docker, Caddy, DNS, NetBird, firewall, backup, persistent application data,
or production routing. When --evidence-json is supplied, it writes only a sanitized local
acceptance artifact to the requested path.

Options:
  --base-url URL           Search instance to validate.
  --container NAME         Docker container name to inspect.
  --expected-image IMAGE   Exact immutable candidate image reference. Must be a
                           ghcr.io/goreecloud/goreecloud-search@sha256:... digest.
  --expected-source SHA    Exact 40-character Git source revision expected in OCI metadata.
                           Must be supplied together with --expected-image.
  --evidence-json PATH     Write sanitized machine-readable target acceptance evidence.
  --skip-providers         Skip representative external-provider acceptance.
  -h, --help               Show this help.

Environment equivalents:
  GOREECLOUD_SEARCH_BASE_URL
  GOREECLOUD_SEARCH_CONTAINER
  GOREECLOUD_SEARCH_PROVIDER_SUITE
  GOREECLOUD_SEARCH_EXPECTED_IMAGE
  GOREECLOUD_SEARCH_EXPECTED_SOURCE
  GOREECLOUD_SEARCH_EVIDENCE_JSON
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
    --expected-image)
      expected_image="${2:?missing immutable image reference}"
      shift 2
      ;;
    --expected-source)
      expected_source="${2:?missing source revision}"
      shift 2
      ;;
    --evidence-json)
      evidence_json="${2:?missing evidence path}"
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

fail() {
  echo "Target acceptance failed: $*" >&2
  exit 1
}

require curl
require python3

if [[ -n "$expected_image" || -n "$expected_source" ]]; then
  [[ -n "$expected_image" && -n "$expected_source" ]] \
    || fail "--expected-image and --expected-source must be supplied together"
  [[ "$expected_image" =~ ^ghcr\.io/goreecloud/goreecloud-search@sha256:[0-9a-f]{64}$ ]] \
    || fail "expected image must be the GoreeCloud Search GHCR image pinned by sha256 digest"
  [[ "$expected_source" =~ ^[0-9a-f]{40}$ ]] \
    || fail "expected source must be a lowercase 40-character Git SHA"
  require docker
fi

printf 'GoreeCloud Search target acceptance\n'
printf 'Base URL: %s\n' "$base_url"
printf 'Container: %s\n' "$container_name"
if [[ -n "$expected_image" ]]; then
  printf 'Expected immutable image: %s\n' "$expected_image"
  printf 'Expected source revision: %s\n' "$expected_source"
fi

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

container_runtime_status="not_available"
container_running="false"
container_health="not_checked"
loopback_ports="not_checked"
identity_status="not_requested"
observed_image_ref=""
observed_image_id=""
oci_title=""
oci_source=""
oci_revision=""
oci_version=""
oci_licenses=""

if command -v docker >/dev/null 2>&1; then
  container_runtime_status="checked"
  docker inspect "$container_name" >/dev/null
  running="$(docker inspect -f '{{.State.Running}}' "$container_name")"
  health="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$container_name")"
  test "$running" = "true"
  test "$health" = "healthy"
  container_running="true"
  container_health="$health"

  published="$(docker port "$container_name" 2>/dev/null || true)"
  if [[ -n "$published" ]] && grep -Evq '127\.0\.0\.1|\[::1\]' <<<"$published"; then
    echo "Unexpected non-loopback published container port:" >&2
    echo "$published" >&2
    exit 1
  fi
  loopback_ports="verified"

  if [[ -n "$expected_image" ]]; then
    observed_image_ref="$(docker inspect -f '{{.Config.Image}}' "$container_name")"
    observed_image_id="$(docker inspect -f '{{.Image}}' "$container_name")"
    expected_image_id="$(docker image inspect -f '{{.Id}}' "$expected_image")"

    [[ "$observed_image_ref" = "$expected_image" ]] \
      || fail "running container image reference does not equal the immutable expected image"
    [[ "$observed_image_id" = "$expected_image_id" ]] \
      || fail "running container image ID does not match the expected digest image"

    oci_title="$(docker image inspect -f '{{ index .Config.Labels "org.opencontainers.image.title" }}' "$expected_image")"
    oci_source="$(docker image inspect -f '{{ index .Config.Labels "org.opencontainers.image.source" }}' "$expected_image")"
    oci_revision="$(docker image inspect -f '{{ index .Config.Labels "org.opencontainers.image.revision" }}' "$expected_image")"
    oci_version="$(docker image inspect -f '{{ index .Config.Labels "org.opencontainers.image.version" }}' "$expected_image")"
    oci_licenses="$(docker image inspect -f '{{ index .Config.Labels "org.opencontainers.image.licenses" }}' "$expected_image")"

    [[ "$oci_title" = "GoreeCloud Search" ]] \
      || fail "candidate OCI title is not GoreeCloud Search"
    [[ "$oci_source" = "https://github.com/GoreeCloud/goreecloud-search" ]] \
      || fail "candidate OCI source does not identify the GoreeCloud Search repository"
    [[ "$oci_revision" = "$expected_source" ]] \
      || fail "candidate OCI revision does not match the expected source revision"
    [[ -n "$oci_version" ]] \
      || fail "candidate OCI version is empty"
    [[ "$oci_licenses" = "AGPL-3.0-or-later" ]] \
      || fail "candidate OCI license metadata is not AGPL-3.0-or-later"
    identity_status="verified"
  fi
elif [[ -n "$expected_image" ]]; then
  fail "Docker is required when immutable candidate identity is requested"
fi

provider_status="skipped"
if [[ "$provider_suite" = "1" ]]; then
  python3 goreecloud/provider_acceptance.py \
    --base-url "$base_url" \
    --suite
  provider_status="passed"
fi

if [[ -n "$evidence_json" ]]; then
  GOREECLOUD_EVIDENCE_BASE_URL="$base_url" \
  GOREECLOUD_EVIDENCE_CONTAINER="$container_name" \
  GOREECLOUD_EVIDENCE_PROVIDER_STATUS="$provider_status" \
  GOREECLOUD_EVIDENCE_RUNTIME_STATUS="$container_runtime_status" \
  GOREECLOUD_EVIDENCE_CONTAINER_RUNNING="$container_running" \
  GOREECLOUD_EVIDENCE_CONTAINER_HEALTH="$container_health" \
  GOREECLOUD_EVIDENCE_LOOPBACK_PORTS="$loopback_ports" \
  GOREECLOUD_EVIDENCE_IDENTITY_STATUS="$identity_status" \
  GOREECLOUD_EVIDENCE_EXPECTED_IMAGE="$expected_image" \
  GOREECLOUD_EVIDENCE_EXPECTED_SOURCE="$expected_source" \
  GOREECLOUD_EVIDENCE_OBSERVED_IMAGE_REF="$observed_image_ref" \
  GOREECLOUD_EVIDENCE_OBSERVED_IMAGE_ID="$observed_image_id" \
  GOREECLOUD_EVIDENCE_OCI_TITLE="$oci_title" \
  GOREECLOUD_EVIDENCE_OCI_SOURCE="$oci_source" \
  GOREECLOUD_EVIDENCE_OCI_REVISION="$oci_revision" \
  GOREECLOUD_EVIDENCE_OCI_VERSION="$oci_version" \
  GOREECLOUD_EVIDENCE_OCI_LICENSES="$oci_licenses" \
  python3 - "$evidence_json" <<'PY'
import datetime as dt
import json
import os
import pathlib
import sys

output = pathlib.Path(sys.argv[1])

def optional(name: str):
    value = os.environ.get(name, "")
    return value if value else None

payload = {
    "schema_version": 1,
    "product": "GoreeCloud Search",
    "generated_at": dt.datetime.now(dt.timezone.utc).isoformat().replace("+00:00", "Z"),
    "target": {
        "base_url": os.environ["GOREECLOUD_EVIDENCE_BASE_URL"],
        "container": os.environ["GOREECLOUD_EVIDENCE_CONTAINER"],
    },
    "http_acceptance": {
        "home_identity": "passed",
        "preferences_identity": "passed",
        "about_identity": "passed",
        "health": "passed",
        "privacy_headers": "passed",
    },
    "providers": os.environ["GOREECLOUD_EVIDENCE_PROVIDER_STATUS"],
    "container_runtime": {
        "status": os.environ["GOREECLOUD_EVIDENCE_RUNTIME_STATUS"],
        "running": os.environ["GOREECLOUD_EVIDENCE_CONTAINER_RUNNING"] == "true",
        "health": os.environ["GOREECLOUD_EVIDENCE_CONTAINER_HEALTH"],
        "published_ports": os.environ["GOREECLOUD_EVIDENCE_LOOPBACK_PORTS"],
        "identity_status": os.environ["GOREECLOUD_EVIDENCE_IDENTITY_STATUS"],
        "expected_image": optional("GOREECLOUD_EVIDENCE_EXPECTED_IMAGE"),
        "expected_source_revision": optional("GOREECLOUD_EVIDENCE_EXPECTED_SOURCE"),
        "observed_image_reference": optional("GOREECLOUD_EVIDENCE_OBSERVED_IMAGE_REF"),
        "observed_image_id": optional("GOREECLOUD_EVIDENCE_OBSERVED_IMAGE_ID"),
        "oci": {
            "title": optional("GOREECLOUD_EVIDENCE_OCI_TITLE"),
            "source": optional("GOREECLOUD_EVIDENCE_OCI_SOURCE"),
            "revision": optional("GOREECLOUD_EVIDENCE_OCI_REVISION"),
            "version": optional("GOREECLOUD_EVIDENCE_OCI_VERSION"),
            "licenses": optional("GOREECLOUD_EVIDENCE_OCI_LICENSES"),
        },
    },
    "scope": {
        "target_runtime_identity_verified": os.environ["GOREECLOUD_EVIDENCE_IDENTITY_STATUS"] == "verified",
        "target_environment_configuration_rollback_tested": False,
        "target_environment_data_restore_tested": False,
        "backup_restore_tested": False,
        "production_cutover_authorized": False,
        "statement": (
            "This artifact records read-only Search runtime acceptance and optional immutable "
            "candidate identity verification. It does not prove backup restoration, persistent-data "
            "recovery, configuration rollback, or authorize production cutover."
        ),
    },
}
output.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n", encoding="utf-8")
PY
  printf 'Sanitized target evidence: %s\n' "$evidence_json"
fi

printf '\nTarget acceptance checks passed for %s\n' "$base_url"
if [[ "$identity_status" = "verified" ]]; then
  printf 'Immutable candidate runtime identity verified.\n'
fi
printf 'This result does not by itself authorize production cutover.\n'
