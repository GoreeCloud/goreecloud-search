#!/bin/sh
set -eu

purge=false
if [ "${1:-}" = "--purge" ]; then
    purge=true
elif [ "$#" -gt 0 ]; then
    printf '%s\n' "Usage: $0 [--purge]" >&2
    exit 2
fi

DATA_ROOT="$HOME/.local/share/goreecloud/search"
CONFIG_ROOT="$HOME/.config/goreecloud"
UNIT_ROOT="$HOME/.config/systemd/user"
ENV_FILE="$CONFIG_ROOT/search.env"
UNIT_FILE="$UNIT_ROOT/goreecloud-search.service"

if command -v systemctl >/dev/null 2>&1; then
    systemctl --user disable --now goreecloud-search.service >/dev/null 2>&1 || true
fi

rm -f "$UNIT_FILE"

if command -v systemctl >/dev/null 2>&1; then
    systemctl --user daemon-reload >/dev/null 2>&1 || true
    systemctl --user reset-failed goreecloud-search.service >/dev/null 2>&1 || true
fi

if [ "$purge" = true ]; then
    rm -rf "$DATA_ROOT"
    rm -f "$ENV_FILE"
fi

printf '%s\n'     "GoreeCloud Search user service removed."     "The existing SearXNG runtime was not changed."
if [ "$purge" = false ]; then
    printf '%s\n'         "Runtime data and protected configuration were retained."         "Run again with --purge only when permanent local removal is intended."
fi
