# SPDX-License-Identifier: AGPL-3.0-or-later
"""Contract tests for GoreeCloud Search result-card presentation."""

from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]


def _read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def test_file_results_expose_structured_glaze_metadata():
    template = _read("searx/templates/simple/result_templates/file.html")

    assert 'goreecloud-file-attributes' in template
    assert 'goreecloud-file-attribute-label' in template
    assert 'goreecloud-file-attribute-value' in template
    assert 'aria-label="File details"' in template
    assert 'goreecloud-result-action' in template
    assert "icon_small('download')" in template


def test_video_results_use_shared_result_actions_and_snippets():
    template = _read("searx/templates/simple/result_templates/videos.html")

    assert 'goreecloud-result-action' in template
    assert 'goreecloud-video-snippet' in template
    assert 'goreecloud-embedded-media' in template
    assert "icon_small('film')" in template
    assert '\n</p>\n{{- result_sub_footer' not in template


def test_result_styles_cover_media_file_sidebar_and_accessibility_states():
    stylesheet = _read("searx/static/themes/simple/goreecloud-results.css")

    required_contracts = (
        ".result-file",
        ".goreecloud-file-attributes",
        ".result-videos",
        ".goreecloud-result-action",
        ".goreecloud-embedded-media",
        ".goreecloud-sidebar-panel",
        "@media (max-width: 599px)",
        "prefers-reduced-motion",
        "prefers-reduced-transparency",
        "prefers-contrast: more",
        "forced-colors",
    )
    for marker in required_contracts:
        assert marker in stylesheet


def run_contract_checks() -> None:
    """Allow this contract to run directly in lightweight CI jobs."""
    test_file_results_expose_structured_glaze_metadata()
    test_video_results_use_shared_result_actions_and_snippets()
    test_result_styles_cover_media_file_sidebar_and_accessibility_states()


if __name__ == "__main__":
    run_contract_checks()
    print("GoreeCloud Search result-card contract passed.")
