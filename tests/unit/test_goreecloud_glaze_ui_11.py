# SPDX-License-Identifier: AGPL-3.0-or-later
"""Fail-closed source contract for GoreeCloud Search Glaze UI 1.1 adoption."""

from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
CANONICAL_VERSION = "1.1.0"
CANONICAL_REVISION = "5c8320de4f770614a3e2bcf9de2a27f7fcfd920c"


def _read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def test_runtime_records_exact_glaze_ui_version_and_viewport_contract():
    base = _read("searx/templates/simple/base.html")

    assert f'<meta name="goreecloud-glaze-ui" content="{CANONICAL_VERSION}">' in base
    assert "viewport-fit=cover" in base
    assert "goreecloud-glaze-1.1.css" in base
    assert base.index("goreecloud-glaze-1.1.css") > base.index("goreecloud-page-compact.css")


def test_glaze_ui_11_semantic_layer_is_bound_to_canonical_release():
    stylesheet = _read("searx/static/themes/simple/goreecloud-glaze-1.1.css")

    assert f"Canonical Glaze UI version: {CANONICAL_VERSION}" in stylesheet
    assert f"Canonical Glaze UI source: {CANONICAL_REVISION}" in stylesheet

    required_tokens = (
        "--gc-on-accent:",
        "--gc-info:",
        "--gc-scrim:",
        "--gc-state-hover:",
        "--gc-state-pressed:",
        "--gc-state-focus:",
        "--gc-state-selected:",
        "--gc-radius-2xl:",
        "--gc-radius-pill:",
        "--gc-icon-sm:",
        "--gc-icon-md:",
        "--gc-density-control-compact:",
        "--gc-blur-overlay:",
        "--gc-opacity-disabled:",
        "--gc-shadow-overlay:",
        "--gc-content-max:",
        "--gc-reading-max:",
        "--gc-gutter-compact:",
        "--gc-gutter-wide:",
        "--gc-z-navigation:",
        "--gc-z-overlay:",
        "--gc-safe-left:",
        "--gc-safe-right:",
        "--gc-safe-bottom:",
    )
    for marker in required_tokens:
        assert marker in stylesheet


def test_glaze_ui_11_interaction_accessibility_and_resilience_contracts_exist():
    stylesheet = _read("searx/static/themes/simple/goreecloud-glaze-1.1.css")

    required_behaviors = (
        "var(--gc-state-hover)",
        "var(--gc-state-pressed)",
        "var(--gc-state-selected)",
        "var(--gc-on-accent)",
        "var(--gc-opacity-disabled)",
        ":disabled",
        '[aria-disabled="true"]',
        ":focus-visible",
        "@media (max-width: 599px)",
        "@media (min-width: 600px) and (max-width: 1023px)",
        "@media (min-width: 1024px) and (max-width: 1439px)",
        "@media (min-width: 1440px)",
        "prefers-reduced-transparency",
        "forced-colors: active",
        "env(safe-area-inset-left",
        "env(safe-area-inset-right",
        "env(safe-area-inset-bottom",
    )
    for marker in required_behaviors:
        assert marker in stylesheet


def test_glaze_ui_11_layer_has_no_remote_ui_dependency():
    stylesheet = _read("searx/static/themes/simple/goreecloud-glaze-1.1.css").lower()

    assert "@import" not in stylesheet
    assert "http://" not in stylesheet
    assert "https://" not in stylesheet
    assert "url(" not in stylesheet


def test_conformance_record_is_version_specific_and_does_not_overclaim_visual_acceptance():
    record = _read("docs/goreecloud/GLAZE-UI-CONFORMANCE.md")

    assert f"Glaze UI {CANONICAL_VERSION}" in record
    assert CANONICAL_REVISION in record
    assert "Source-conformance status:" in record
    assert "Stable conformance status: Pending visual acceptance" in record
    assert "Compact" in record
    assert "Expanded" in record
    assert "light" in record.lower()
    assert "dark" in record.lower()
    assert "Glaze UI 1.1 conformant" in record
    assert "only when" in record


def test_first_stable_gate_requires_glaze_11_visuals_and_files_provider_acceptance():
    stable = _read("docs/goreecloud/STABLE-CUTOVER.md")

    assert "Glaze UI 1.1 acceptance" in stable
    assert "Compact and Expanded" in stable
    assert "light and dark" in stable
    assert "General, Images, Videos, News, and Files" in stable


def run_contract_checks() -> None:
    """Allow this contract to run without adding a test-runner dependency."""
    test_runtime_records_exact_glaze_ui_version_and_viewport_contract()
    test_glaze_ui_11_semantic_layer_is_bound_to_canonical_release()
    test_glaze_ui_11_interaction_accessibility_and_resilience_contracts_exist()
    test_glaze_ui_11_layer_has_no_remote_ui_dependency()
    test_conformance_record_is_version_specific_and_does_not_overclaim_visual_acceptance()
    test_first_stable_gate_requires_glaze_11_visuals_and_files_provider_acceptance()


if __name__ == "__main__":
    run_contract_checks()
    print("GoreeCloud Search Glaze UI 1.1 source contract passed.")
