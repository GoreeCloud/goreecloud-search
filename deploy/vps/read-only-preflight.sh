#!/usr/bin/env bash

set -euo pipefail

SEARCH_COMPOSE="${SEARCH_COMPOSE:-/srv/docker/stacks/searxng/docker-compose.yml}"
SEARCH_CONTAINER="${SEARCH_CONTAINER:-searxng-core}"
LEGACY_VALKEY_CONTAINER="${LEGACY_VALKEY_CONTAINER:-searxng-valkey}"
SEARCH_HOST="${SEARCH_HOST:-search.goreecloud.com}"
PROXY_NETWORK="${PROXY_NETWORK:-proxy}"
CADDY_ROOT="${CADDY_ROOT:-/srv/docker/stacks/caddy}"
SEARCH_SECRET_FILE="${SEARCH_SECRET_FILE:-/srv/docker/secrets/searxng/brave-search-api-key}"

fail() {
    printf 'ERROR=%s\n' "$*" >&2
    exit 1
}

section() {
    printf '\n=== %s ===\n' "$1"
}

kv() {
    printf '%s=%s\n' "$1" "$2"
}

require_command() {
    command -v "$1" >/dev/null 2>&1 || fail "required command not available: $1"
}

absolute_path() {
    case "$1" in
        /*) return 0 ;;
        *) return 1 ;;
    esac
}

for path in "$SEARCH_COMPOSE" "$CADDY_ROOT" "$SEARCH_SECRET_FILE"; do
    absolute_path "$path" || fail "preflight paths must be absolute"
done

for command_name in docker hostname uname stat sha256sum grep curl getent awk sort sed wc paste; do
    require_command "$command_name"
done

DOCKER=(docker)
if docker info >/dev/null 2>&1; then
    :
elif command -v sudo >/dev/null 2>&1 && sudo -n docker info >/dev/null 2>&1; then
    DOCKER=(sudo -n docker)
else
    fail "Docker read access is unavailable; run with approved Docker read permissions"
fi

sudo_read_available() {
    command -v sudo >/dev/null 2>&1 && sudo -n true >/dev/null 2>&1
}

file_exists() {
    local path="$1"
    if test -e "$path"; then
        return 0
    fi
    if sudo_read_available && sudo -n test -e "$path"; then
        return 0
    fi
    return 1
}

file_readable() {
    local path="$1"
    if test -r "$path"; then
        return 0
    fi
    if sudo_read_available && sudo -n test -r "$path"; then
        return 0
    fi
    return 1
}

file_stat() {
    local path="${@: -1}"
    if test -r "$path"; then
        stat "$@"
        return
    fi
    sudo_read_available || fail "path requires approved read privilege: $path"
    sudo -n stat "$@"
}

file_sha256() {
    local path="$1"
    if test -r "$path"; then
        sha256sum "$path"
        return
    fi
    sudo_read_available || fail "path requires approved read privilege: $path"
    sudo -n sha256sum "$path"
}

grep_paths() {
    local pattern="$1"
    local root="$2"
    if test -r "$root"; then
        grep -RIl --binary-files=without-match --fixed-strings -- "$pattern" "$root" 2>/dev/null || true
        return
    fi
    sudo_read_available || fail "Caddy root requires approved read privilege: $root"
    sudo -n grep -RIl --binary-files=without-match --fixed-strings -- "$pattern" "$root" 2>/dev/null || true
}

container_exists() {
    "${DOCKER[@]}" container inspect "$1" >/dev/null 2>&1
}

network_exists() {
    "${DOCKER[@]}" network inspect "$1" >/dev/null 2>&1
}

network_has_container() {
    local network="$1"
    local container="$2"
    "${DOCKER[@]}" network inspect "$network" \
        --format '{{range .Containers}}{{println .Name}}{{end}}' \
        | grep -Fxq -- "$container"
}

printf 'GoreeCloud Search VPS read-only preflight\n'
kv mode "read-only"
kv mutation_permitted "no"
kv secret_values_printed "no"
kv live_provider_query_performed "no"
kv expected_search_host "$SEARCH_HOST"

section "Host"
kv hostname "$(hostname)"
kv architecture "$(uname -m)"
kv docker_server_version "$("${DOCKER[@]}" version --format '{{.Server.Version}}')"
kv docker_compose_version "$("${DOCKER[@]}" compose version --short)"

section "Authoritative Search Compose"
file_exists "$SEARCH_COMPOSE" || fail "authoritative Search Compose file is absent: $SEARCH_COMPOSE"
file_readable "$SEARCH_COMPOSE" || fail "authoritative Search Compose file is not readable: $SEARCH_COMPOSE"
kv compose_path "$SEARCH_COMPOSE"
kv compose_sha256 "$(file_sha256 "$SEARCH_COMPOSE" | awk '{print $1}')"
kv compose_mode_owner "$(file_stat -c '%a %U:%G' "$SEARCH_COMPOSE")"
printf '%s\n' "compose_services_begin"
"${DOCKER[@]}" compose -f "$SEARCH_COMPOSE" config --services
printf '%s\n' "compose_services_end"
printf '%s\n' "compose_images_begin"
"${DOCKER[@]}" compose -f "$SEARCH_COMPOSE" config --images
printf '%s\n' "compose_images_end"

section "Search Container"
if ! container_exists "$SEARCH_CONTAINER"; then
    kv search_container_exists "no"
    fail "expected active Search container is absent: $SEARCH_CONTAINER"
fi
kv search_container_exists "yes"
kv search_container_name "$SEARCH_CONTAINER"
kv search_container_image_ref "$("${DOCKER[@]}" container inspect "$SEARCH_CONTAINER" --format '{{.Config.Image}}')"
SEARCH_IMAGE_ID="$("${DOCKER[@]}" container inspect "$SEARCH_CONTAINER" --format '{{.Image}}')"
kv search_container_image_id "$SEARCH_IMAGE_ID"
kv search_container_state "$("${DOCKER[@]}" container inspect "$SEARCH_CONTAINER" --format '{{.State.Status}}')"
kv search_container_health "$("${DOCKER[@]}" container inspect "$SEARCH_CONTAINER" --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}')"
kv search_container_user "$("${DOCKER[@]}" container inspect "$SEARCH_CONTAINER" --format '{{.Config.User}}')"
kv search_container_read_only_root "$("${DOCKER[@]}" container inspect "$SEARCH_CONTAINER" --format '{{.HostConfig.ReadonlyRootfs}}')"
kv search_container_init "$("${DOCKER[@]}" container inspect "$SEARCH_CONTAINER" --format '{{.HostConfig.Init}}')"
kv search_container_cap_drop "$("${DOCKER[@]}" container inspect "$SEARCH_CONTAINER" --format '{{json .HostConfig.CapDrop}}')"
kv search_container_security_opt "$("${DOCKER[@]}" container inspect "$SEARCH_CONTAINER" --format '{{json .HostConfig.SecurityOpt}}')"
kv search_container_restart "$("${DOCKER[@]}" container inspect "$SEARCH_CONTAINER" --format '{{.HostConfig.RestartPolicy.Name}}')"
kv search_container_host_port_bindings "$("${DOCKER[@]}" container inspect "$SEARCH_CONTAINER" --format '{{json .HostConfig.PortBindings}}')"
kv search_container_networks "$("${DOCKER[@]}" container inspect "$SEARCH_CONTAINER" --format '{{range $name, $_ := .NetworkSettings.Networks}}{{printf "%s " $name}}{{end}}' | sed 's/[[:space:]]*$//')"
printf '%s\n' "search_container_mounts_begin"
"${DOCKER[@]}" container inspect "$SEARCH_CONTAINER" \
    --format '{{range .Mounts}}{{printf "%s -> %s (rw=%t)\n" .Source .Destination .RW}}{{end}}'
printf '%s\n' "search_container_mounts_end"

section "Search Image"
kv search_image_repo_digests "$("${DOCKER[@]}" image inspect "$SEARCH_IMAGE_ID" --format '{{json .RepoDigests}}')"
kv search_image_oci_version "$("${DOCKER[@]}" image inspect "$SEARCH_IMAGE_ID" --format '{{index .Config.Labels "org.opencontainers.image.version"}}')"
kv search_image_oci_revision "$("${DOCKER[@]}" image inspect "$SEARCH_IMAGE_ID" --format '{{index .Config.Labels "org.opencontainers.image.revision"}}')"

section "Docker Network Boundaries"
if network_exists "$PROXY_NETWORK"; then
    kv proxy_network_exists "yes"
    if network_has_container "$PROXY_NETWORK" "$SEARCH_CONTAINER"; then
        kv search_on_proxy_network "yes"
    else
        kv search_on_proxy_network "no"
    fi
else
    kv proxy_network_exists "no"
    kv search_on_proxy_network "no"
fi

if container_exists "$LEGACY_VALKEY_CONTAINER"; then
    kv legacy_valkey_exists "yes"
    kv legacy_valkey_state "$("${DOCKER[@]}" container inspect "$LEGACY_VALKEY_CONTAINER" --format '{{.State.Status}}')"
    kv legacy_valkey_networks "$("${DOCKER[@]}" container inspect "$LEGACY_VALKEY_CONTAINER" --format '{{range $name, $_ := .NetworkSettings.Networks}}{{printf "%s " $name}}{{end}}' | sed 's/[[:space:]]*$//')"
else
    kv legacy_valkey_exists "no"
fi

section "Protected Runtime Credential Metadata"
if file_exists "$SEARCH_SECRET_FILE"; then
    kv provider_secret_file_exists "yes"
    kv provider_secret_file_mode_owner "$(file_stat -c '%a %U:%G' "$SEARCH_SECRET_FILE")"
else
    kv provider_secret_file_exists "no"
fi
kv provider_secret_value "not-read"

section "Caddy Search Route References"
if file_exists "$CADDY_ROOT"; then
    kv caddy_root_exists "yes"
    HOST_MATCHES="$(grep_paths "$SEARCH_HOST" "$CADDY_ROOT")"
    BACKEND_MATCHES="$(grep_paths "$SEARCH_CONTAINER:8080" "$CADDY_ROOT")"
    kv caddy_search_hostname_match_count "$(printf '%s\n' "$HOST_MATCHES" | sed '/^$/d' | wc -l | awk '{print $1}')"
    kv caddy_search_backend_match_count "$(printf '%s\n' "$BACKEND_MATCHES" | sed '/^$/d' | wc -l | awk '{print $1}')"
    printf '%s\n' "caddy_search_hostname_match_files_begin"
    printf '%s\n' "$HOST_MATCHES" | sed '/^$/d'
    printf '%s\n' "caddy_search_hostname_match_files_end"
    printf '%s\n' "caddy_search_backend_match_files_begin"
    printf '%s\n' "$BACKEND_MATCHES" | sed '/^$/d'
    printf '%s\n' "caddy_search_backend_match_files_end"
else
    kv caddy_root_exists "no"
fi

section "Private DNS And HTTPS"
RESOLVED_ADDRESSES="$(getent ahosts "$SEARCH_HOST" | awk '{print $1}' | sort -u | paste -sd, -)"
if [ -n "$RESOLVED_ADDRESSES" ]; then
    kv private_dns_resolves "yes"
    kv resolved_addresses "$RESOLVED_ADDRESSES"
else
    kv private_dns_resolves "no"
fi

printf '%s\n' "https_probe_begin"
curl --fail --silent --show-error --output /dev/null --max-time 10 \
    --write-out 'https_status=%{http_code}\nhttps_ssl_verify_result=%{ssl_verify_result}\nhttps_content_type=%{content_type}\nhttps_time_total=%{time_total}\n' \
    "https://$SEARCH_HOST/"
printf '%s\n' "https_probe_end"

section "Preflight Boundary"
kv compose_raw_contents_printed "no"
kv caddy_raw_contents_printed "no"
kv container_environment_printed "no"
kv container_logs_printed "no"
kv files_written "no"
kv docker_state_changed "no"
kv caddy_state_changed "no"
kv result "read-only evidence collected; deployment authorization is separate"
