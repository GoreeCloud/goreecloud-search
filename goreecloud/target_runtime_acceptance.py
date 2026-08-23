# SPDX-License-Identifier: AGPL-3.0-or-later
"""Produce sanitized read-only target-runtime evidence for GoreeCloud Search.

The operator must stage the immutable candidate separately. This tool never starts, stops,
recreates, or modifies containers, routing, DNS, Caddy, NetBird, firewall, backups, or
persistent application data. It only inspects an already running loopback-only container
and writes a non-authorizing evidence artifact.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import pathlib
import re
import shutil
import subprocess
import sys
import urllib.request
from typing import Any


IMAGE_RE = re.compile(r"^ghcr\.io/goreecloud/goreecloud-search@sha256:[0-9a-f]{64}$")
SHA_RE = re.compile(r"^[0-9a-f]{40}$")
VERSION_RE = re.compile(r"^2026\.[0-9]{1,2}\.[0-9]{1,2}-[0-9a-f]{7,40}$")
LOOPBACK_URL_RE = re.compile(r"^https?://(?:127\.0\.0\.1|localhost|\[::1\])(?::[0-9]{1,5})?$")
REPOSITORY_URL = "https://github.com/GoreeCloud/goreecloud-search"


class AcceptanceError(ValueError):
    """Raised when the staged runtime does not satisfy the acceptance contract."""


def _nonempty(value: str, label: str) -> str:
    value = value.strip()
    if not value:
        raise AcceptanceError(f"{label} must not be empty")
    return value


def validate_inputs(base_url: str, image: str, source: str, version: str) -> None:
    """Validate the immutable identity and loopback-only target before any inspection."""
    if not LOOPBACK_URL_RE.fullmatch(base_url):
        raise AcceptanceError("base URL must identify loopback-only staging")
    if not IMAGE_RE.fullmatch(image):
        raise AcceptanceError("expected image must be an immutable GoreeCloud Search GHCR digest")
    if not SHA_RE.fullmatch(source):
        raise AcceptanceError("expected source must be a lowercase 40-character Git SHA")
    if not VERSION_RE.fullmatch(version):
        raise AcceptanceError("expected OCI version must use the GoreeCloud candidate version format")


def validate_ports(published: str) -> str:
    """Require at least one published mapping and reject every non-loopback mapping."""
    lines = [line.strip() for line in published.splitlines() if line.strip()]
    if not lines:
        raise AcceptanceError("candidate container must expose a loopback-only published port")
    for line in lines:
        if "127.0.0.1:" not in line and "[::1]:" not in line:
            raise AcceptanceError(f"non-loopback published port rejected: {line}")
    return "\n".join(lines)


def _run(arguments: list[str]) -> str:
    try:
        completed = subprocess.run(
            arguments,
            check=True,
            capture_output=True,
            text=True,
            encoding="utf-8",
        )
    except (OSError, subprocess.CalledProcessError) as exc:
        raise AcceptanceError(f"command failed: {' '.join(arguments)}") from exc
    return completed.stdout.strip()


def _docker_format(container: str, expression: str) -> str:
    return _nonempty(_run(["docker", "inspect", "-f", expression, container]), "docker inspect output")


def inspect_container(container: str, image: str, source: str, version: str) -> dict[str, Any]:
    """Read and validate the already staged container and immutable image metadata."""
    if shutil.which("docker") is None:
        raise AcceptanceError("Docker is required for immutable target-runtime acceptance")

    running = _docker_format(container, "{{.State.Running}}")
    health = _docker_format(container, "{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}")
    if running != "true":
        raise AcceptanceError("candidate container is not running")
    if health != "healthy":
        raise AcceptanceError("candidate container is not healthy")

    published = validate_ports(_run(["docker", "port", container]))
    observed_image = _docker_format(container, "{{.Config.Image}}")
    observed_id = _docker_format(container, "{{.Image}}")
    expected_id = _nonempty(
        _run(["docker", "image", "inspect", "-f", "{{.Id}}", image]),
        "expected image ID",
    )
    if observed_image != image:
        raise AcceptanceError("running container image reference does not match expected image")
    if observed_id != expected_id:
        raise AcceptanceError("running container image ID does not match expected digest image")

    labels = {
        "title": _image_label(image, "org.opencontainers.image.title"),
        "source": _image_label(image, "org.opencontainers.image.source"),
        "revision": _image_label(image, "org.opencontainers.image.revision"),
        "version": _image_label(image, "org.opencontainers.image.version"),
        "licenses": _image_label(image, "org.opencontainers.image.licenses"),
    }
    expected = {
        "title": "GoreeCloud Search",
        "source": REPOSITORY_URL,
        "revision": source,
        "version": version,
        "licenses": "AGPL-3.0-or-later",
    }
    for key, expected_value in expected.items():
        if labels[key] != expected_value:
            raise AcceptanceError(f"candidate OCI {key} does not match the expected identity")

    return {
        "status": "verified",
        "running": True,
        "health": "healthy",
        "published_ports": published,
        "identity_status": "verified",
        "expected_image": image,
        "expected_source_revision": source,
        "observed_image_reference": observed_image,
        "observed_image_id": observed_id,
        "oci": labels,
    }


def _image_label(image: str, label: str) -> str:
    expression = f'{{{{ index .Config.Labels "{label}" }}}}'
    return _nonempty(_run(["docker", "image", "inspect", "-f", expression, image]), f"OCI {label}")


def _fetch(url: str, method: str = "GET") -> tuple[str, dict[str, str]]:
    request = urllib.request.Request(url, method=method)
    try:
        with urllib.request.urlopen(request, timeout=20) as response:  # nosec B310 - loopback is validated
            body = response.read().decode("utf-8", errors="replace") if method != "HEAD" else ""
            headers = {key.lower(): value for key, value in response.headers.items()}
    except OSError as exc:
        raise AcceptanceError(f"HTTP acceptance request failed: {url}") from exc
    return body, headers


def inspect_http(base_url: str) -> dict[str, str]:
    """Verify GoreeCloud product identity, health, and privacy headers on loopback."""
    home, _ = _fetch(f"{base_url}/")
    preferences, _ = _fetch(f"{base_url}/preferences")
    about, _ = _fetch(f"{base_url}/about")
    health, _ = _fetch(f"{base_url}/healthz")
    _, headers = _fetch(f"{base_url}/", method="HEAD")

    if "<title>GoreeCloud Search</title>" not in home or "goreecloud.css" not in home:
        raise AcceptanceError("home page does not identify the GoreeCloud Search product shell")
    if "GoreeCloud Search" not in preferences:
        raise AcceptanceError("Preferences does not identify GoreeCloud Search")
    if "About GoreeCloud Search" not in about:
        raise AcceptanceError("About page does not identify GoreeCloud Search")
    if health.strip() != "OK":
        raise AcceptanceError("health endpoint did not return the expected OK response")

    required_headers = {
        "x-robots-tag": ("noindex", "nofollow"),
        "referrer-policy": ("no-referrer",),
        "x-frame-options": ("deny",),
        "permissions-policy": ("camera=()", "microphone=()", "geolocation=()"),
    }
    for key, fragments in required_headers.items():
        value = headers.get(key, "").lower()
        if not all(fragment in value for fragment in fragments):
            raise AcceptanceError(f"required privacy header is missing or incomplete: {key}")

    return {
        "home_identity": "passed",
        "preferences_identity": "passed",
        "about_identity": "passed",
        "health": "passed",
        "privacy_headers": "passed",
    }


def build_evidence(
    base_url: str,
    container: str,
    http_acceptance: dict[str, str],
    runtime: dict[str, Any],
) -> dict[str, Any]:
    """Build the sanitized schema-version 1 target-runtime artifact."""
    return {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "generated_at": dt.datetime.now(dt.timezone.utc).isoformat().replace("+00:00", "Z"),
        "target": {"base_url": base_url, "container": container},
        "http_acceptance": http_acceptance,
        "providers": "skipped",
        "container_runtime": runtime,
        "scope": {
            "target_runtime_identity_verified": True,
            "target_environment_configuration_rollback_tested": False,
            "target_environment_data_restore_tested": False,
            "backup_restore_tested": False,
            "production_cutover_authorized": False,
            "statement": (
                "This artifact records read-only loopback target-runtime identity acceptance. "
                "It does not prove provider, backup, restore, configuration rollback, data "
                "recovery, production cutover, or Stable promotion."
            ),
        },
    }


def parser() -> argparse.ArgumentParser:
    """Build the command line for operator-controlled target-runtime inspection."""
    root = argparse.ArgumentParser(description=__doc__)
    root.add_argument("--base-url", default="http://127.0.0.1:8888")
    root.add_argument("--container", required=True)
    root.add_argument("--expected-image", required=True)
    root.add_argument("--expected-source", required=True)
    root.add_argument("--expected-version", required=True)
    root.add_argument("--evidence-json", required=True)
    return root


def main() -> int:
    """Validate an already staged candidate and write sanitized evidence."""
    args = parser().parse_args()
    try:
        validate_inputs(args.base_url, args.expected_image, args.expected_source, args.expected_version)
        http_acceptance = inspect_http(args.base_url)
        runtime = inspect_container(
            args.container,
            args.expected_image,
            args.expected_source,
            args.expected_version,
        )
        evidence = build_evidence(args.base_url, args.container, http_acceptance, runtime)
        output = pathlib.Path(args.evidence_json)
        output.write_text(json.dumps(evidence, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    except (AcceptanceError, OSError) as exc:
        print(f"Target-runtime acceptance error: {exc}", file=sys.stderr)
        return 2

    print("GoreeCloud Search target-runtime acceptance passed.")
    print(f"Evidence: {args.evidence_json}")
    print("Production cutover authorized by this evidence: false")
    print("Stable promotion authorized by this evidence: false")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
