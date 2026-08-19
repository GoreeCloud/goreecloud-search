# SPDX-License-Identifier: AGPL-3.0-or-later
"""Public provider-diagnostics privacy and Glaze UI contract."""

from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
STATS = ROOT / "searx" / "templates" / "simple" / "stats.html"
BASE = ROOT / "searx" / "templates" / "simple" / "base.html"
STYLE = ROOT / "searx" / "static" / "themes" / "simple" / "goreecloud-provider-diagnostics.css"


def test_public_provider_diagnostics_are_sanitized():
    stats = STATS.read_text(encoding="utf-8")

    assert "Provider diagnostics" in stats
    assert "Sanitized provider performance metrics" in stats
    assert "Internal exception messages" in stats
    assert "exception_classname" not in stats
    assert "error.log_message" not in stats
    assert "error.log_parameters" not in stats
    assert "error.filename" not in stats
    assert "error.function" not in stats
    assert "error.code" not in stats
    assert "log_parameters" not in stats
    assert "engine-tooltip" not in stats


def test_provider_diagnostics_keep_accessible_metrics():
    stats = STATS.read_text(encoding="utf-8")

    assert '<thead>' in stats
    assert '<tbody>' in stats
    assert 'scope="col"' in stats
    assert 'scope="row"' in stats
    assert 'aria-label="Provider performance table"' in stats
    assert 'tabindex="0"' in stats
    assert "Provider name" in stats
    assert "Result count" in stats
    assert "Response time" in stats
    assert "Reliability" in stats


def test_provider_diagnostics_use_glaze_surface_and_product_language():
    base = BASE.read_text(encoding="utf-8")
    style = STYLE.read_text(encoding="utf-8")

    assert "goreecloud-provider-diagnostics.css" in base
    assert "Provider diagnostics" in base
    assert "Engine stats" not in base
    assert ".goreecloud-provider-diagnostics" in style
    assert "var(--gc-border)" in style
    assert "var(--gc-surface)" in style
    assert "var(--gc-accent)" in style
    assert "prefers-reduced-transparency" in style
    assert "prefers-contrast: more" in style
    assert "forced-colors: active" in style


if __name__ == "__main__":
    test_public_provider_diagnostics_are_sanitized()
    test_provider_diagnostics_keep_accessible_metrics()
    test_provider_diagnostics_use_glaze_surface_and_product_language()
    print("GoreeCloud provider diagnostics contract passed.")
