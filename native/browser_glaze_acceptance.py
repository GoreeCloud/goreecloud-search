from __future__ import annotations

import os
from dataclasses import dataclass
from pathlib import Path

from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import Select, WebDriverWait

BASE_URL = os.environ.get(
    "GOREECLOUD_SEARCH_NATIVE_RESULTS_TEST_URL", "http://127.0.0.1:8091"
).rstrip("/")
SCREENSHOT_DIR = os.environ.get("GOREECLOUD_SEARCH_NATIVE_RESULTS_SCREENSHOT_DIR")
STORAGE_KEY = "goreecloud.search.preferences.v1"
TARGET_FLOOR = 48.0


@dataclass(frozen=True)
class Viewport:
    name: str
    width: int
    height: int


VIEWPORTS = (
    Viewport("compact", 390, 844),
    Viewport("medium", 768, 900),
    Viewport("expanded", 1280, 900),
    Viewport("wide", 1600, 1000),
)

APPEARANCES = ("light", "dark", "deep-dark")
EXPECTED_CANVAS = {
    "light": "rgb(245, 247, 250)",
    "dark": "rgb(11, 13, 17)",
    "deep-dark": "rgb(5, 7, 10)",
}


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


def store_preferences(driver: webdriver.Chrome, appearance: str, density: str = "comfortable") -> None:
    driver.get(f"{BASE_URL}/")
    driver.execute_script(
        "localStorage.setItem(arguments[0], JSON.stringify({schema_version: 1, preferences: {'appearance.theme': arguments[1], 'appearance.result_density': arguments[2]}}));",
        STORAGE_KEY,
        appearance,
        density,
    )


def assert_glaze_root(driver: webdriver.Chrome, appearance: str, density: str, context: str) -> None:
    root = driver.find_element(By.TAG_NAME, "html")
    if root.get_attribute("data-glaze-version") != "1.1":
        raise AssertionError(f"{context}: missing Glaze UI V1.1 root contract")
    if root.get_attribute("data-glz-appearance") != appearance:
        raise AssertionError(
            f"{context}: appearance = {root.get_attribute('data-glz-appearance')!r}, want {appearance!r}"
        )
    expected_density = "productive" if density == "compact" else "comfortable"
    if root.get_attribute("data-glaze-density-profile") != expected_density:
        raise AssertionError(
            f"{context}: density = {root.get_attribute('data-glaze-density-profile')!r}, want {expected_density!r}"
        )
    canvas = driver.execute_script("return getComputedStyle(document.body).backgroundColor")
    if canvas != EXPECTED_CANVAS[appearance]:
        raise AssertionError(f"{context}: body canvas = {canvas!r}, want {EXPECTED_CANVAS[appearance]!r}")


def assert_no_horizontal_overflow(driver: webdriver.Chrome, context: str) -> None:
    scroll_width, client_width = driver.execute_script(
        "return [document.documentElement.scrollWidth, document.documentElement.clientWidth];"
    )
    if scroll_width > client_width + 2:
        raise AssertionError(f"{context}: horizontal overflow {scroll_width}px > {client_width}px")


def assert_target(element, context: str) -> None:
    rect = element.rect
    width = float(rect.get("width", 0))
    height = float(rect.get("height", 0))
    if width + 0.5 < TARGET_FLOOR or height + 0.5 < TARGET_FLOOR:
        raise AssertionError(f"{context}: target {width:.1f}x{height:.1f}px; expected >= 48px")


def capture(driver: webdriver.Chrome, name: str) -> None:
    if not SCREENSHOT_DIR:
        return
    output = Path(SCREENSHOT_DIR)
    output.mkdir(parents=True, exist_ok=True)
    driver.save_screenshot(str(output / f"{name}.png"))


def exercise_page(
    driver: webdriver.Chrome,
    wait: WebDriverWait,
    viewport: Viewport,
    appearance: str,
    path: str,
    marker_selector: str,
    target_selectors: tuple[str, ...],
    name: str,
) -> None:
    store_preferences(driver, appearance)
    driver.get(f"{BASE_URL}{path}")
    wait.until(EC.visibility_of_element_located((By.CSS_SELECTOR, marker_selector)))
    context = f"{viewport.name}/{appearance}/{name}"
    assert_glaze_root(driver, appearance, "comfortable", context)
    assert_no_horizontal_overflow(driver, context)
    for selector in target_selectors:
        for index, element in enumerate(driver.find_elements(By.CSS_SELECTOR, selector), start=1):
            if element.is_displayed():
                assert_target(element, f"{context} {selector} #{index}")
    capture(driver, f"glaze-v1-1-{viewport.name}-{appearance}-{name}")


def assert_live_preferences_mapping(driver: webdriver.Chrome, wait: WebDriverWait) -> None:
    store_preferences(driver, "light")
    driver.get(f"{BASE_URL}/preferences")
    wait.until(EC.visibility_of_element_located((By.ID, "appearance")))
    theme = Select(driver.find_element(By.CSS_SELECTOR, 'select[data-preference="appearance.theme"]'))
    density = Select(driver.find_element(By.CSS_SELECTOR, 'select[data-preference="appearance.result_density"]'))

    theme.select_by_value("deep-dark")
    wait.until(lambda d: d.find_element(By.TAG_NAME, "html").get_attribute("data-glz-appearance") == "deep-dark")
    density.select_by_value("compact")
    wait.until(lambda d: d.find_element(By.TAG_NAME, "html").get_attribute("data-glaze-density-profile") == "productive")
    assert_glaze_root(driver, "deep-dark", "compact", "live Preferences mapping")

    envelope = driver.execute_script("return JSON.parse(localStorage.getItem(arguments[0]));", STORAGE_KEY)
    if envelope["preferences"]["appearance.theme"] != "deep-dark":
        raise AssertionError("live Preferences mapping did not persist Deep Dark")
    if envelope["preferences"]["appearance.result_density"] != "compact":
        raise AssertionError("live Preferences mapping did not persist compact density")
    capture(driver, "glaze-v1-1-compact-live-deep-dark-productive")


def main() -> int:
    pages = (
        ("/", "#hero-title", (".brand", "#q", ".search-box button", ".category-row .chip"), "home"),
        ("/preferences", "#appearance", (".brand", "[data-settings-filter]", "select[data-preference]", ".secondary-button"), "preferences"),
        ("/search", "#results-title", (".results-brand", "#results-q", ".results-search button", ".results-categories a"), "results"),
        ("/search?case=images", ".image-results-grid", (".results-brand", "#results-q", ".image-result-open"), "images"),
    )

    for viewport in VIEWPORTS:
        driver = driver_for(viewport)
        wait = WebDriverWait(driver, 10)
        try:
            for appearance in APPEARANCES:
                for path, marker, targets, name in pages:
                    exercise_page(driver, wait, viewport, appearance, path, marker, targets, name)
            if viewport.name == "compact":
                assert_live_preferences_mapping(driver, wait)
        finally:
            driver.quit()

    print("native Glaze UI V1.1 shell browser acceptance passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
