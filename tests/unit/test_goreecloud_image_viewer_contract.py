# SPDX-License-Identifier: AGPL-3.0-or-later
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
BASE_TEMPLATE = ROOT / "searx" / "templates" / "simple" / "base.html"
IMAGE_STYLES = ROOT / "searx" / "static" / "themes" / "simple" / "goreecloud-image-results.css"


def test_image_result_styles_load_after_general_goreecloud_shell() -> None:
    base = BASE_TEMPLATE.read_text(encoding="utf-8")
    image_link = "goreecloud-image-results.css"
    platform_link = "goreecloud-platform-shell.css"

    assert image_link in base
    assert platform_link in base
    assert base.index(image_link) > base.index(platform_link)


def test_image_result_ancestor_cannot_contain_fixed_viewer() -> None:
    css = IMAGE_STYLES.read_text(encoding="utf-8")

    assert "article.result-images" in css
    assert "backdrop-filter: none !important" in css
    assert "-webkit-backdrop-filter: none !important" in css
    assert "transform: none !important" in css
    assert "padding: 0 !important" in css
    assert "margin: 0 !important" in css


def test_selected_image_detail_is_explicitly_viewport_fixed() -> None:
    css = IMAGE_STYLES.read_text(encoding="utf-8")
    selector = "#results.image-detail-open article.result-images[data-vim-selected] .detail"

    assert selector in css
    assert "position: fixed !important" in css
    assert "z-index: 4000 !important" in css
    assert "max-width: 100rem" in css
    assert "a.result-images-source img" in css
    assert "max-height: min(62vh, 46rem) !important" in css


def test_image_viewer_controls_meet_glaze_touch_target_floor() -> None:
    css = IMAGE_STYLES.read_text(encoding="utf-8")

    assert "var(--gc-target-comfortable)" in css
    assert "a.result-detail-close" in css
    assert "a.result-detail-previous" in css
    assert "a.result-detail-next" in css
