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


def test_provider_health_uses_accessible_glaze_status_structure():
    template = _read("searx/templates/simple/elements/engines_msg.html")
    stylesheet = _read("searx/static/themes/simple/goreecloud-states.css")

    required_template_contracts = (
        'class="goreecloud-provider-health"',
        'aria-labelledby="engines_msg-title"',
        'goreecloud-provider-health-summary',
        'goreecloud-provider-health-badge',
        '<caption>{{ _(\'Search provider diagnostics\') }}</caption>',
        '<th scope="col">{{ _(\'Provider\') }}</th>',
        '<th scope="col">{{ _(\'Status\') }}</th>',
        '<th scope="row" class="engine-name">',
        'data-provider-state="degraded"',
        'data-provider-state="available"',
        'goreecloud-provider-state',
        'View provider diagnostics',
    )
    for marker in required_template_contracts:
        assert marker in template

    assert 'Messages from the search engines' not in template
    assert 'aria-labelledby="{{engine_name}}_time"' not in template
    assert 'bar-chart-graph' not in template

    required_style_contracts = (
        '.goreecloud-provider-health-panel',
        '.goreecloud-provider-health-summary:focus-visible',
        '.goreecloud-provider-health-table-wrap',
        '.goreecloud-provider-health-table',
        '.goreecloud-provider-state[data-state="available"]',
        '.goreecloud-provider-state[data-state="degraded"]',
        'var(--gc-success)',
        'var(--gc-warning)',
        'overflow-x: auto',
        'prefers-reduced-transparency',
        'prefers-contrast: more',
        'forced-colors: active',
    )
    for marker in required_style_contracts:
        assert marker in stylesheet


def test_empty_state_stays_goreecloud_first_while_about_preserves_upstream_attribution():
    no_results = _read("searx/templates/simple/messages/no_results.html")
    about = _read("searx/infopage/en/about.md")

    assert 'GoreeCloud Search does not automatically send you to a separate public search service' in no_results
    assert 'another SearXNG instance' not in no_results
    assert 'built from the open-source [SearXNG] project' in about
    assert 'preserving SearXNG attribution and the terms of the upstream license' in about


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


def run_contract_checks() -> None:
    """Allow this contract to run in CI without adding a test-runner dependency."""
    test_search_results_header_has_explicit_glaze_structure()
    test_category_navigation_uses_semantic_icon_controls()
    test_provider_health_uses_accessible_glaze_status_structure()
    test_empty_state_stays_goreecloud_first_while_about_preserves_upstream_attribution()
    test_search_shell_stylesheet_loads_after_existing_glaze_layers()


if __name__ == "__main__":
    run_contract_checks()
    print("GoreeCloud Search shell contract passed.")
