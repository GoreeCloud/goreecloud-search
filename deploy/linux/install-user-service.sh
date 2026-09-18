#!/bin/sh
set -eu

umask 077

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)

DATA_ROOT="${XDG_DATA_HOME:-$HOME/.local/share}/goreecloud/search"
CONFIG_ROOT="${XDG_CONFIG_HOME:-$HOME/.config}/goreecloud"
UNIT_ROOT="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
VENV="$DATA_ROOT/venv"
ENV_FILE="$CONFIG_ROOT/search.env"
UNIT_FILE="$UNIT_ROOT/goreecloud-search.service"

require_command() {
    command -v "$1" >/dev/null 2>&1 || {
        printf '%s\n' "Required command not found: $1" >&2
        exit 1
    }
}

require_command python3
require_command systemctl

systemctl --user show-environment >/dev/null 2>&1 || {
    printf '%s\n'         "A working systemd user manager is required for this deployment profile." >&2
    exit 1
}

mkdir -p "$DATA_ROOT" "$CONFIG_ROOT" "$UNIT_ROOT"
chmod 700 "$DATA_ROOT" "$CONFIG_ROOT"

python3 -m venv "$VENV"
"$VENV/bin/python" -m pip install     --disable-pip-version-check     --no-deps     --no-build-isolation     --upgrade     "$REPO_ROOT"

install -m 0600 "$SCRIPT_DIR/goreecloud-search.service" "$UNIT_FILE"

if [ ! -e "$ENV_FILE" ]; then
    install -m 0600 "$SCRIPT_DIR/search.env.example" "$ENV_FILE"
else
    chmod 0600 "$ENV_FILE"
fi

systemctl --user daemon-reload

if ! grep -Eq '^BRAVE_SEARCH_API_KEY=.+$' "$ENV_FILE"; then
    printf '%s\n'         "GoreeCloud Search was installed but not started."         "Set BRAVE_SEARCH_API_KEY in $ENV_FILE, then run:"         "  systemctl --user enable --now goreecloud-search.service"         "The existing SearXNG runtime was not changed."
    exit 0
fi

systemctl --user enable --now goreecloud-search.service

"$VENV/bin/python" - <<'PY'
from urllib.request import urlopen
import json

with urlopen("http://127.0.0.1:8787/healthz", timeout=5) as response:
    payload = json.load(response)
if response.status != 200 or payload.get("status") != "ok":
    raise SystemExit("GoreeCloud Search health check failed")
PY

printf '%s\n'     "GoreeCloud Search is active on http://127.0.0.1:8787."     "The existing SearXNG runtime was not changed."
