# SPDX-License-Identifier: AGPL-3.0-or-later
"""Contract checks for GoreeCloud Search filter, file, and video result polish."""

from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]


def _read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def test_filter_toolbar_has_explicit_glaze_controls():
    template = _read("searx/templates/simple/search.html")

    assert "goreecloud-filter-control goreecloud-filter-language" in template
    assert "goreecloud-filter-control goreecloud-filter-time" in template
    assert "goreecloud-filter-control goreecloud-filter-safe" in template
    assert template.count("goreecloud-filter-label") == 3


def test_file_result_uses_structured_metadata_and_action():
    template = _read("searx/templates/simple/result_templates/file.html")

    assert "goreecloud-file-summary" in template
    assert "goreecloud-file-attributes" in template
    assert "<dl" in template
    assert "<dt>" in template
    assert "<dd>" in template
    assert "goreecloud-file-actions" in template
    assert "icon_small('download')" in template


def test_video_result_uses_media_card_hooks():
    template = _read("searx/templates/simple/result_templates/videos.html")

    assert "goreecloud-video-action-row" in template
    assert "goreecloud-video-description" in template
    assert "goreecloud-embedded-media" in template
    assert "icon_small('film')" in template


def test_polish_stylesheet_is_last_goreecloud_layer_and_resilient():
    base = _read("searx/templates/simple/base.html")
    stylesheet = _read("searx/static/themes/simple/goreecloud-result-polish.css")

    shell_position = base.index("goreecloud-search-shell.css")
    polish_position = base.index("goreecloud-result-polish.css")
    assert polish_position > shell_position

    required_contracts = (
        ".goreecloud-filter-control",
        ".goreecloud-result-action",
        ".result-file",
        ".result-videos",
        "grid-template-areas",
        "@media (max-width: 760px)",
        "prefers-reduced-transparency",
        "prefers-contrast: more",
        "forced-colors",
    )
    for marker in required_contracts:
        assert marker in stylesheet


def run_contract_checks() -> None:
    test_filter_toolbar_has_explicit_glaze_controls()
    test_file_result_uses_structured_metadata_and_action()
    test_video_result_uses_media_card_hooks()
    test_polish_stylesheet_is_last_goreecloud_layer_and_resilient()


if __name__ == "__main__":
    run_contract_checks()
    print("GoreeCloud result polish contract passed.")
