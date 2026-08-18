# SPDX-License-Identifier: AGPL-3.0-or-later
"""Contract tests for the GoreeCloud Search query shell and category navigation."""

from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]


def _read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def test_search_results_header_has_explicit_glaze_structure():
    template = _read("searx/templates/simple/search.html")

    assert 'class="goreecloud-search-header"' in template
    assert 'class="goreecloud-search-primary"' in template
    assert 'class="goreecloud-search-query"' in template
    assert 'search_box goreecloud-search-box' in template
    assert 'search_filters goreecloud-search-filters' in template
    assert 'aria-label="Search query"' in template


def test_category_navigation_uses_semantic_icon_controls():
    template = _read("searx/templates/simple/categories.html")

    assert "'general': 'globe'" in template
    assert "'videos': 'film'" in template
    assert "'news': 'newspaper'" in template
    assert "'map': 'location'" in template
    assert "'music': 'musical-notes'" in template
    assert 'class="search_categories goreecloud-category-nav"' in template
    assert 'class="goreecloud-category-icon"' in template
    assert 'goreecloud-category-name' in template
    assert 'aria-pressed=' in template


def test_search_shell_stylesheet_loads_after_existing_glaze_layers():
    base = _read("searx/templates/simple/base.html")
    stylesheet = _read("searx/static/themes/simple/goreecloud-search-shell.css")

    mobile_position = base.index("goreecloud-mobile.css")
    shell_position = base.index("goreecloud-search-shell.css")
    assert shell_position > mobile_position

    required_contracts = (
        ".goreecloud-search-primary",
        ".goreecloud-search-box",
        ".goreecloud-category-nav",
        ".goreecloud-category-icon",
        "var(--gc-target-min)",
        "@media (max-width: 599px)",
        "prefers-reduced-motion",
        "prefers-reduced-transparency",
        "forced-colors",
    )
    for marker in required_contracts:
        assert marker in stylesheet
