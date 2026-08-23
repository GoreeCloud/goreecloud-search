#!/usr/bin/env python3
# SPDX-License-Identifier: AGPL-3.0-or-later

from __future__ import annotations

import hashlib
import json
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
CONTRACT_PATH = ROOT / "goreecloud" / "platform-integrations.json"
SETTINGS_PATH = ROOT / "goreecloud" / "settings.yml.example"
PRIVACY_ASSET = ROOT / "searx" / "static" / "themes" / "simple" / "img" / "privacy-shield.svg"
WARDVEIL_ASSET = ROOT / "searx" / "static" / "themes" / "simple" / "img" / "wardveil-security.svg"


def _sha256(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> None:
    contract = json.loads(CONTRACT_PATH.read_text(encoding="utf-8"))
    privacy = contract["privacy_shield"]
    wardveil = contract["wardveil_security"]

    assert contract["schema_version"] == 1
    assert contract["product"] == "GoreeCloud Search"

    assert privacy["authority_repository"] == "GoreeCloud/goreecloud-privacy-shield"
    assert privacy["authority_revision"] == "1edb768336591e0656226f555c57d32537de9274"
    assert privacy["adapter_state"] == "integrated"
    assert privacy["consumer_asset"] == "searx/static/themes/simple/img/privacy-shield.svg"
    assert _sha256(PRIVACY_ASSET) == privacy["consumer_asset_sha256"]

    assert wardveil["authority_repository"] == "GoreeCloud/goreecloud-wardveil-security"
    assert wardveil["authority_revision"] == "d044e04f35fc09d623dc2ee55810a0e1453b6c01"
    assert wardveil["adapter_state"] == "integrated"
    assert wardveil["runtime_status"] in {
        "protected",
        "attention",
        "degraded",
        "unknown",
        "not_applicable",
    }
    assert wardveil["runtime_status"] == "unknown"
    assert wardveil["protected_by_wardveil"] is False
    assert wardveil["consumer_asset"] == "searx/static/themes/simple/img/wardveil-security.svg"
    assert _sha256(WARDVEIL_ASSET) == wardveil["consumer_asset_sha256"]

    assert privacy["authority_repository"] != wardveil["authority_repository"]
    assert "Protected by Wardveil" not in privacy["boundary"]
    assert "source integration alone" in wardveil["boundary"]

    settings = SETTINGS_PATH.read_text(encoding="utf-8")
    for marker in (
        "enable_metrics: false",
        "image_proxy: true",
        "query_in_title: false",
        "bind_address: \"127.0.0.1\"",
        "X-Content-Type-Options: nosniff",
        "X-Frame-Options: DENY",
        "Referrer-Policy: no-referrer",
        "Permissions-Policy: \"camera=(), microphone=(), geolocation=()\"",
    ):
        assert marker in settings, marker


if __name__ == "__main__":
    main()
