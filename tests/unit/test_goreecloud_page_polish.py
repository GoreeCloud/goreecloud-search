# SPDX-License-Identifier: AGPL-3.0-or-later
"""Contract tests for GoreeCloud Search landing and Preferences presentation."""

from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]


def _read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def test_landing_page_has_single_search_shell_and_product_principles():
    index = _read("searx/templates/simple/index.html")
    search = _read("searx/templates/simple/simple_search.html")

    assert 'index goreecloud-search-landing' in index
    assert 'goreecloud-search-principles' in index
    assert 'Private by design' in index
    assert 'Multi-provider' in index
    assert 'Open source' in index
    assert 'placeholder="Search privately across the web"' in search
    assert 'aria-label="Search query"' in search


def test_preferences_surface_uses_goreecloud_language_and_actions():
    preferences = _read("searx/templates/simple/preferences.html")
    footer = _read("searx/templates/simple/preferences/footer.html")
    theme = _read("searx/templates/simple/preferences/theme.html")
    hotkeys = _read("searx/templates/simple/preferences/hotkeys.html")
    alignment = _read("searx/templates/simple/preferences/center_alignment.html")

    assert 'goreecloud-preferences-heading' in preferences
    assert 'goreecloud-preferences-form' in preferences
    assert 'Search settings' in preferences
    assert 'goreecloud-preferences-footer' in footer
    assert 'goreecloud-preferences-actions' in footer
    assert 'goreecloud-preferences-save' in footer
    assert 'goreecloud-preferences-reset' in footer
    assert 'Glaze UI' in theme
    assert '>SearXNG<' not in hotkeys
    assert 'Oscar layout' not in alignment


def test_page_polish_stylesheet_is_last_goreecloud_visual_layer():
    base = _read("searx/templates/simple/base.html")
    stylesheet = _read("searx/static/themes/simple/goreecloud-page-polish.css")

    result_polish_position = base.index("goreecloud-result-polish.css")
    page_polish_position = base.index("goreecloud-page-polish.css")
    assert page_polish_position > result_polish_position

    required_contracts = (
        ".goreecloud-search-landing",
        ".goreecloud-search-principles",
        ".goreecloud-preferences-form",
        ".goreecloud-preferences-footer",
        "input.checkbox-onoff:not(.reversed-checkbox)",
        "@media (max-width: 599px)",
        "prefers-reduced-motion",
        "prefers-reduced-transparency",
        "prefers-contrast: more",
        "forced-colors: active",
    )
    for marker in required_contracts:
        assert marker in stylesheet


def run_contract_checks() -> None:
    """Allow this contract to run without adding a test-runner dependency."""
    test_landing_page_has_single_search_shell_and_product_principles()
    test_preferences_surface_uses_goreecloud_language_and_actions()
    test_page_polish_stylesheet_is_last_goreecloud_visual_layer()


if __name__ == "__main__":
    run_contract_checks()
    print("GoreeCloud Search landing/preferences contract passed.")
