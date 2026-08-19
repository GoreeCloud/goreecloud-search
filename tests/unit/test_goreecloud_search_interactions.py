# SPDX-License-Identifier: AGPL-3.0-or-later
"""Contracts for GoreeCloud Search query-control interaction states."""

from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]


def _read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def test_empty_query_does_not_expose_clear_control():
    search = _read("searx/templates/simple/simple_search.html")
    interactions = _read("searx/static/themes/simple/goreecloud-search-interactions.css")

    assert 'id="q"' in search
    assert 'placeholder="Search privately across the web"' in search
    assert 'id="clear_search" type="reset"' in search
    assert '#q:placeholder-shown + #clear_search' in interactions
    assert 'display: none !important;' in interactions
    assert '#send_search' not in interactions


def test_interaction_layer_is_loaded_after_search_shell():
    base = _read("searx/templates/simple/base.html")

    shell_position = base.index("goreecloud-search-shell.css")
    interaction_position = base.index("goreecloud-search-interactions.css")
    result_position = base.index("goreecloud-result-polish.css")

    assert shell_position < interaction_position < result_position


if __name__ == "__main__":
    test_empty_query_does_not_expose_clear_control()
    test_interaction_layer_is_loaded_after_search_shell()
    print("GoreeCloud Search interaction contract passed.")
