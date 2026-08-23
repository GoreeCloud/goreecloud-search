# SPDX-License-Identifier: AGPL-3.0-or-later
"""Source contracts for the GoreeCloud Search homepage/Preferences rebuild."""

from pathlib import Path
import re
import unittest


ROOT = Path(__file__).resolve().parents[2]


def _read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def _css_variables(css: str, selector: str) -> dict[str, str]:
    match = re.search(rf"{selector}\s*\{{(?P<body>.*?)\}}", css, flags=re.DOTALL)
    if match is None:
        raise AssertionError(f"CSS variable scope not found: {selector}")
    return dict(re.findall(r"(--[\w-]+):\s*(#[0-9a-fA-F]{6})", match.group("body")))


def _contrast_ratio(first: str, second: str) -> float:
    def luminance(value: str) -> float:
        channels = [int(value[index : index + 2], 16) / 255 for index in (1, 3, 5)]
        linear = [channel / 12.92 if channel <= 0.04045 else ((channel + 0.055) / 1.055) ** 2.4 for channel in channels]
        return 0.2126 * linear[0] + 0.7152 * linear[1] + 0.0722 * linear[2]

    lighter, darker = sorted((luminance(first), luminance(second)), reverse=True)
    return (lighter + 0.05) / (darker + 0.05)


class GoreeCloudHomePreferencesV2Test(unittest.TestCase):
    """Keep the rebuilt product surfaces and accessibility contract intact."""

    def test_layer_order(self):
        base = _read("searx/templates/simple/base.html")
        foundation = base.index("goreecloud.css")
        states = base.index("goreecloud-states.css")
        rebuild = base.index("goreecloud-home-preferences-v2.css")
        containment = base.index("goreecloud-home-preferences-v2-containment.css")
        self.assertLess(foundation, states)
        self.assertLess(states, rebuild)
        self.assertLess(rebuild, containment)

    def test_home_markup(self):
        index = _read("searx/templates/simple/index.html")
        self.assertIn("goreecloud-index-masthead", index)
        self.assertIn("Private search.", index)
        self.assertIn("Total control.", index)
        self.assertIn("goreecloud-index-principles", index)
        self.assertIn("simple/simple_search.html", index)

    def test_search_form(self):
        search = _read("searx/templates/simple/simple_search.html")
        self.assertIn("aria-label=\"{{ _('Search query') }}\"", search)
        self.assertIn("if endpoint == 'index'", search)
        self.assertIn("simple/categories.html", search)
        self.assertIn("{% else %}", search)
        self.assertIn("name=\"category_{{ category }}\"", search)
        self.assertIn("goreecloud-index-privacy-note", search)

    def test_pref_markup(self):
        preferences = _read("searx/templates/simple/preferences.html")
        for tab_id in (
            "general",
            "ui",
            "privacy",
            "engines",
            "query",
            "cookies",
        ):
            self.assertIn(f"tab_header('maintab', '{tab_id}'", preferences)
        self.assertIn('role="tab"', preferences)
        self.assertIn('role="tabpanel"', preferences)
        self.assertIn('name="{{ name }}"', preferences)

    def test_pref_keyboard(self):
        css = _read("searx/static/themes/simple/goreecloud-home-preferences-v2.css")
        radio = '#main_preferences #search_form > .tabs > input[type="radio"]'
        self.assertIn(radio, css)
        self.assertIn("display: block;", css)
        self.assertIn('input[type="radio"]:focus-visible + label', css)
        self.assertIn("#main_preferences #tab-label-general", css)
        self.assertIn("#main_preferences #tab-label-cookies", css)

    def test_pref_actions(self):
        footer = _read("searx/templates/simple/preferences/footer.html")
        self.assertIn("goreecloud-preferences-actions", footer)
        self.assertIn("goreecloud-preferences-save", footer)
        self.assertIn("url_for('index')", footer)
        self.assertIn("url_for('clear_cookies')", footer)
        self.assertIn("type=\"submit\"", footer)

    def test_adaptive_css(self):
        css = _read("searx/static/themes/simple/goreecloud-home-preferences-v2.css")
        containment = _read("searx/static/themes/simple/goreecloud-home-preferences-v2-containment.css")
        self.assertIn("@media (min-width: 1024px) and (max-width: 1439px)", css)
        self.assertIn("@media (min-width: 1440px)", css)
        self.assertIn("@media (min-width: 600px) and (max-width: 1023px)", css)
        self.assertIn("@media (max-width: 599px)", css)
        self.assertIn("@media (prefers-reduced-motion: reduce)", css)
        self.assertIn("@media (prefers-reduced-transparency: reduce)", css)
        self.assertIn("@media (prefers-contrast: more)", css)
        self.assertIn("@media (forced-colors: active)", css)
        self.assertIn("overflow-x: clip", containment)
        self.assertIn("width: 44px", containment)
        self.assertIn("overflow-x: auto", containment)
        self.assertIn("#main_index #search_view:focus-within", containment)

    def test_semantic_color_contrast(self):
        css = _read("searx/static/themes/simple/goreecloud.css")
        light = _css_variables(css, r":root")
        dark = _css_variables(css, r"html\.theme-dark,\s*html\.theme-black")
        pairs = (
            (light, "--gc-text", "--gc-canvas"),
            (light, "--gc-muted", "--gc-canvas"),
            (light, "--gc-accent", "--gc-canvas"),
            (light, "--gc-on-accent", "--gc-accent"),
            (light, "--gc-on-accent", "--gc-accent-secondary"),
            (dark, "--gc-text", "--gc-canvas"),
            (dark, "--gc-muted", "--gc-canvas"),
            (dark, "--gc-accent", "--gc-canvas"),
            (dark, "--gc-on-accent", "--gc-accent"),
            (dark, "--gc-on-accent", "--gc-accent-secondary"),
        )
        for palette, foreground, background in pairs:
            with self.subTest(foreground=foreground, background=background):
                self.assertGreaterEqual(
                    _contrast_ratio(palette[foreground], palette[background]),
                    4.5,
                )

    def test_browser_contract(self):
        acceptance = _read("goreecloud/browser_acceptance.py")
        self.assertIn('"Private search. Total control."', acceptance)
        self.assertIn(".goreecloud-index-brand", acceptance)
        self.assertIn("#categories_container input[type='checkbox']", acceptance)
        self.assertIn("goreecloud-home-preferences-v2.css", acceptance)
        self.assertIn('Appearance("light", "light")', acceptance)
        self.assertIn('Appearance("dark", "dark")', acceptance)
        self.assertIn("Page.captureScreenshot", acceptance)
        self.assertIn("_assert_search_category_separation", acceptance)
        self.assertIn("_assert_footer", acceptance)

    def test_exact_head_evidence(self):
        workflow = _read(".github/workflows/goreecloud-browser-acceptance.yml")
        exact_head = "${{ github.event.pull_request.head.sha || github.sha }}"
        self.assertIn(f"ref: {exact_head}", workflow)
        self.assertIn(f"name: goreecloud-search-ui-{exact_head}", workflow)

    def test_footer_contract(self):
        base = _read("searx/templates/simple/base.html")
        self.assertIn('class="goreecloud-footer"', base)
        self.assertIn("Powered by", base)
        self.assertIn("SearXNG", base)
        self.assertIn("GNU AGPL", base)
        self.assertIn("GoreeCloud/goreecloud-search", base)
        self.assertIn("searxng/searxng", base)


if __name__ == "__main__":
    unittest.main()
