from __future__ import annotations

import base64
import os
from dataclasses import dataclass
from pathlib import Path

from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait

BASE_URL = os.environ.get("GOREECLOUD_SEARCH_NATIVE_RESULTS_TEST_URL", "http://127.0.0.1:8091").rstrip("/")
SCREENSHOT_DIR = os.environ.get("GOREECLOUD_SEARCH_NATIVE_RESULTS_SCREENSHOT_DIR")


@dataclass(frozen=True)
class Viewport:
    name: str
    width: int
    height: int
    touch: bool = False


VIEWPORTS = (
    Viewport("compact", 390, 844, True),
    Viewport("medium", 768, 900, True),
    Viewport("expanded", 1280, 900, False),
    Viewport("wide", 1600, 1000, False),
)


@dataclass(frozen=True)
class Appearance:
    name: str
    value: str


APPEARANCES = (Appearance("light", "light"), Appearance("dark", "dark"))


def driver_for(viewport: Viewport) -> webdriver.Chrome:
    options = webdriver.ChromeOptions()
    options.add_argument("--headless=new")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    options.add_argument(f"--window-size={viewport.width},{viewport.height}")
    options.add_argument("--force-device-scale-factor=1")
    driver = webdriver.Chrome(options=options)
    driver.set_window_size(viewport.width, viewport.height)
    return driver


def set_media(driver: webdriver.Chrome, *, scheme: str, motion: str = "no-preference", contrast: str = "no-preference", forced_colors: str = "none") -> None:
    driver.execute_cdp_cmd(
        "Emulation.setEmulatedMedia",
        {
            "media": "screen",
            "features": [
                {"name": "prefers-color-scheme", "value": scheme},
                {"name": "prefers-reduced-motion", "value": motion},
                {"name": "prefers-contrast", "value": contrast},
                {"name": "forced-colors", "value": forced_colors},
            ],
        },
    )


def capture(driver: webdriver.Chrome, name: str) -> None:
    if not SCREENSHOT_DIR:
        return
    output = Path(SCREENSHOT_DIR)
    output.mkdir(parents=True, exist_ok=True)
    metrics = driver.execute_cdp_cmd("Page.getLayoutMetrics", {})
    size = metrics["cssContentSize"]
    width = max(1, min(int(size["width"]), 4096))
    height = max(1, min(int(size["height"]), 16384))
    result = driver.execute_cdp_cmd(
        "Page.captureScreenshot",
        {
            "format": "png",
            "captureBeyondViewport": True,
            "fromSurface": True,
            "clip": {"x": 0, "y": 0, "width": width, "height": height, "scale": 1},
        },
    )
    (output / f"{name}.png").write_bytes(base64.b64decode(result["data"]))


def assert_no_horizontal_overflow(driver: webdriver.Chrome, context: str) -> None:
    scroll_width, client_width = driver.execute_script(
        "return [document.documentElement.scrollWidth, document.documentElement.clientWidth];"
    )
    if scroll_width > client_width + 2:
        raise AssertionError(f"{context}: horizontal overflow {scroll_width}px > {client_width}px")


def assert_min_target(element, minimum: float, context: str) -> None:
    rect = element.rect
    width = float(rect.get("width", 0))
    height = float(rect.get("height", 0))
    if width + 0.5 < minimum or height + 0.5 < minimum:
        raise AssertionError(f"{context}: target {width:.1f}x{height:.1f}px; expected >= {minimum:.0f}px")


def assert_solid_content_canvas(driver: webdriver.Chrome, context: str) -> None:
    color = driver.execute_script("return getComputedStyle(document.body).backgroundColor")
    if color in {"transparent", "rgba(0, 0, 0, 0)"}:
        raise AssertionError(f"{context}: body content canvas is transparent: {color}")


def assert_result_rows_not_card_wall(driver: webdriver.Chrome, context: str) -> None:
    row = driver.find_element(By.CSS_SELECTOR, ".result-card")
    background, radius, shadow = driver.execute_script(
        """
        const s = getComputedStyle(arguments[0]);
        return [s.backgroundColor, s.borderRadius, s.boxShadow];
        """,
        row,
    )
    if background not in {"rgba(0, 0, 0, 0)", "transparent"}:
        raise AssertionError(f"{context}: result row gained card background {background}")
    if radius not in {"0px", "0px 0px 0px 0px"}:
        raise AssertionError(f"{context}: result row gained card radius {radius}")
    if shadow != "none":
        raise AssertionError(f"{context}: result row gained elevated card shadow {shadow}")


def assert_layout(driver: webdriver.Chrome, viewport: Viewport, context: str) -> None:
    column = driver.find_element(By.CSS_SELECTOR, ".results-column")
    sidebar = driver.find_element(By.CSS_SELECTOR, ".results-sidebar")
    c = column.rect
    s = sidebar.rect
    if viewport.width >= 1024:
        if s["x"] <= c["x"] + c["width"] - 2:
            raise AssertionError(f"{context}: desktop details rail overlaps results column")
    else:
        if s["y"] + 2 < c["y"] + c["height"]:
            raise AssertionError(f"{context}: narrow details region overlaps results column")


def assert_keyboard_focus(driver: webdriver.Chrome, context: str) -> None:
    query = driver.find_element(By.ID, "results-q")
    query.click()
    query.send_keys(Keys.TAB)
    active = driver.switch_to.active_element
    if active.tag_name.lower() != "button":
        raise AssertionError(f"{context}: Tab from results query did not reach Search button")
    outline = driver.execute_script(
        "const s=getComputedStyle(arguments[0]); return [s.outlineStyle, parseFloat(s.outlineWidth) || 0];",
        active,
    )
    if outline[0] == "none" or outline[1] < 2:
        raise AssertionError(f"{context}: keyboard focus is not visibly outlined: {outline}")


def assert_results_page(driver: webdriver.Chrome, wait: WebDriverWait, viewport: Viewport, appearance: Appearance) -> None:
    context = f"{viewport.name}/{appearance.name} results"
    driver.get(f"{BASE_URL}/search")
    wait.until(EC.visibility_of_element_located((By.ID, "results-title")))
    body = driver.find_element(By.TAG_NAME, "body").text
    required = (
        "Results for “goreecloud search privacy”",
        "7 results from the configured native sources.",
        "3 sources agree",
        "Published Aug 31, 2026",
        "Published Aug 28, 2026",
        "Relevant first",
        "trustworthy freshness when requested",
        "Click history is not used.",
        "Some sources unavailable",
        "limit applied",
        "provider_timeout",
        "Production approval not granted",
    )
    for marker in required:
        if marker not in body:
            raise AssertionError(f"{context}: missing {marker!r}")
    if "Score " in body:
        raise AssertionError(f"{context}: internal rank score leaked into results UI")
    if len(driver.find_elements(By.CSS_SELECTOR, ".result-card")) != 7:
        raise AssertionError(f"{context}: expected seven representative result rows")
    published = driver.find_elements(By.CSS_SELECTOR, "time.result-published")
    if len(published) != 2:
        raise AssertionError(f"{context}: expected two trusted publication timestamps, got {len(published)}")
    if published[0].get_attribute("datetime") != "2026-08-31T14:30:00Z":
        raise AssertionError(f"{context}: publication timestamp lost machine-readable UTC value")
    current = driver.find_elements(By.CSS_SELECTOR, '.results-categories a[aria-current="page"]')
    if len(current) != 1 or current[0].text != "General":
        raise AssertionError(f"{context}: selected category semantics are incorrect")

    minimum = 48.0 if viewport.touch else 44.0
    assert_min_target(driver.find_element(By.ID, "results-q"), minimum, f"{context} query")
    assert_min_target(driver.find_element(By.CSS_SELECTOR, ".results-search button"), minimum, f"{context} submit")
    for control in driver.find_elements(By.CSS_SELECTOR, ".results-categories a"):
        assert_min_target(control, minimum, f"{context} category {control.text}")
    for summary in driver.find_elements(By.CSS_SELECTOR, ".result-provenance summary"):
        assert_min_target(summary, minimum, f"{context} source agreement")
    assert_min_target(driver.find_element(By.CSS_SELECTOR, ".provider-summary summary"), minimum, f"{context} source health")
    for action in driver.find_elements(By.CSS_SELECTOR, ".results-topbar .top-actions a"):
        if action.is_displayed():
            assert_min_target(action, minimum, f"{context} top action")

    assert_solid_content_canvas(driver, context)
    assert_result_rows_not_card_wall(driver, context)
    assert_layout(driver, viewport, context)
    assert_no_horizontal_overflow(driver, context)
    assert_keyboard_focus(driver, context)
    capture(driver, f"{viewport.name}-{appearance.name}-native-results")


def assert_empty_state(driver: webdriver.Chrome, wait: WebDriverWait) -> None:
    driver.get(f"{BASE_URL}/search?case=empty")
    wait.until(EC.visibility_of_element_located((By.CSS_SELECTOR, ".empty-state")))
    if "No results found" not in driver.find_element(By.TAG_NAME, "body").text:
        raise AssertionError("empty state: missing no-results message")
    assert_no_horizontal_overflow(driver, "empty state")
    capture(driver, "compact-light-native-results-empty")


def assert_error_state(driver: webdriver.Chrome, wait: WebDriverWait) -> None:
    driver.get(f"{BASE_URL}/search?case=error")
    alert = wait.until(EC.visibility_of_element_located((By.CSS_SELECTOR, '.error-state[role="alert"]')))
    if "Search needs attention" not in alert.text:
        raise AssertionError("error state: accessible alert copy is missing")
    assert_no_horizontal_overflow(driver, "error state")
    capture(driver, "expanded-dark-native-results-error")


def assert_accessibility_modes() -> None:
    viewport = Viewport("compact", 390, 844, True)
    driver = driver_for(viewport)
    wait = WebDriverWait(driver, 10)
    try:
        set_media(driver, scheme="dark", motion="reduce")
        driver.get(f"{BASE_URL}/search")
        wait.until(EC.visibility_of_element_located((By.ID, "results-title")))
        if not driver.execute_script("return matchMedia('(prefers-reduced-motion: reduce)').matches"):
            raise AssertionError("reduced-motion emulation is not active")
        behavior = driver.execute_script("return getComputedStyle(document.documentElement).scrollBehavior")
        if behavior != "auto":
            raise AssertionError(f"reduced motion: scroll behavior remained {behavior!r}")
        capture(driver, "compact-dark-native-results-reduced-motion")

        set_media(driver, scheme="light", contrast="more")
        driver.get(f"{BASE_URL}/search")
        wait.until(EC.visibility_of_element_located((By.ID, "results-title")))
        if not driver.execute_script("return matchMedia('(prefers-contrast: more)').matches"):
            raise AssertionError("increased-contrast emulation is not active")
        border = driver.execute_script("return parseFloat(getComputedStyle(document.querySelector('.result-card')).borderBottomWidth)")
        if border < 2:
            raise AssertionError(f"increased contrast: result separator is only {border}px")
        capture(driver, "compact-light-native-results-increased-contrast")

        set_media(driver, scheme="light", forced_colors="active")
        driver.get(f"{BASE_URL}/search")
        wait.until(EC.visibility_of_element_located((By.ID, "results-title")))
        if not driver.execute_script("return matchMedia('(forced-colors: active)').matches"):
            raise AssertionError("forced-colors emulation is not active")
        backdrop = driver.execute_script("return getComputedStyle(document.querySelector('.results-context')).backdropFilter")
        if backdrop != "none":
            raise AssertionError(f"forced colors: category surface still uses backdrop filter {backdrop!r}")
        assert_no_horizontal_overflow(driver, "forced colors")
        capture(driver, "compact-light-native-results-forced-colors")
    finally:
        driver.quit()


def main() -> int:
    for viewport in VIEWPORTS:
        driver = driver_for(viewport)
        wait = WebDriverWait(driver, 10)
        try:
            for appearance in APPEARANCES:
                set_media(driver, scheme=appearance.value)
                assert_results_page(driver, wait, viewport, appearance)
            if viewport.name == "compact":
                set_media(driver, scheme="light")
                assert_empty_state(driver, wait)
            if viewport.name == "expanded":
                set_media(driver, scheme="dark")
                assert_error_state(driver, wait)
        finally:
            driver.quit()
    assert_accessibility_modes()
    print("native results browser acceptance passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
