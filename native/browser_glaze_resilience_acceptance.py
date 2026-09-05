from __future__ import annotations

import os
from dataclasses import dataclass
from pathlib import Path

from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait

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


RTL_VIEWPORTS = (
    Viewport("compact", 390, 844),
    Viewport("wide", 1600, 1000),
)

PAGES = (
    ("/", "#hero-title", (".brand", "#q", ".search-box button", ".category-row .chip"), "home"),
    (
        "/preferences",
        "#appearance",
        (".brand", "[data-settings-filter]", "[data-settings-nav] a", "select[data-preference]"),
        "preferences",
    ),
    (
        "/search",
        "#results-title",
        (".results-brand", "#results-q", ".results-search button", ".results-categories a"),
        "results",
    ),
    (
        "/search?case=images",
        ".image-results-grid",
        (".results-brand", "#results-q", ".image-result-open"),
        "images",
    ),
)


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


def store_preferences(driver: webdriver.Chrome) -> None:
    driver.get(f"{BASE_URL}/")
    driver.execute_script(
        "localStorage.setItem(arguments[0], JSON.stringify({schema_version: 1, preferences: {'appearance.theme': 'light', 'appearance.result_density': 'comfortable'}}));",
        STORAGE_KEY,
    )


def capture(driver: webdriver.Chrome, name: str) -> None:
    if not SCREENSHOT_DIR:
        return
    output = Path(SCREENSHOT_DIR)
    output.mkdir(parents=True, exist_ok=True)
    driver.save_screenshot(str(output / f"{name}.png"))


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


def assert_reachable_without_document_overflow(driver: webdriver.Chrome, element, context: str) -> None:
    x, width, viewport, intentionally_scrollable = driver.execute_script(
        """
        const element = arguments[0];
        const rect = element.getBoundingClientRect();
        let ancestor = element.parentElement;
        let intentionallyScrollable = false;
        while (ancestor) {
          const style = getComputedStyle(ancestor);
          if ((style.overflowX === 'auto' || style.overflowX === 'scroll') &&
              ancestor.scrollWidth > ancestor.clientWidth + 1) {
            intentionallyScrollable = true;
            break;
          }
          ancestor = ancestor.parentElement;
        }
        return [rect.x, rect.width, window.innerWidth, intentionallyScrollable];
        """,
        element,
    )
    escaped = x < -2 or x + width > viewport + 2
    if escaped and not intentionally_scrollable:
        raise AssertionError(
            f"{context}: control escaped viewport without an intentional horizontal scroller: "
            f"x={x:.1f}, width={width:.1f}, viewport={viewport:.1f}"
        )


def assert_controls(driver: webdriver.Chrome, selectors: tuple[str, ...], context: str) -> None:
    for selector in selectors:
        for index, element in enumerate(driver.find_elements(By.CSS_SELECTOR, selector), start=1):
            if not element.is_displayed():
                continue
            assert_target(element, f"{context} {selector} #{index}")
            assert_reachable_without_document_overflow(
                driver, element, f"{context} {selector} #{index}"
            )


def assert_rtl_resilience() -> None:
    for viewport in RTL_VIEWPORTS:
        driver = driver_for(viewport)
        wait = WebDriverWait(driver, 10)
        try:
            for path, marker, targets, name in PAGES:
                store_preferences(driver)
                driver.get(f"{BASE_URL}{path}")
                wait.until(EC.visibility_of_element_located((By.CSS_SELECTOR, marker)))
                driver.execute_script(
                    "document.documentElement.dir='rtl'; document.documentElement.lang='ar';"
                )
                context = f"rtl/{viewport.name}/{name}"
                root = driver.find_element(By.TAG_NAME, "html")
                if root.get_attribute("dir") != "rtl":
                    raise AssertionError(f"{context}: RTL document direction was not applied")
                if root.get_attribute("data-glaze-version") != "1.1":
                    raise AssertionError(f"{context}: Glaze UI V1.1 contract was lost")
                assert_no_horizontal_overflow(driver, context)
                assert_controls(driver, targets, context)
                capture(driver, f"glaze-v1-1-rtl-{viewport.name}-{name}")
        finally:
            driver.quit()


def assert_two_x_scale_reflow() -> None:
    physical = Viewport("physical-1280x900", 1280, 900)
    driver = driver_for(physical)
    wait = WebDriverWait(driver, 10)
    try:
        store_preferences(driver)
        driver.execute_cdp_cmd(
            "Emulation.setDeviceMetricsOverride",
            {
                "width": 640,
                "height": 450,
                "deviceScaleFactor": 2,
                "mobile": False,
            },
        )
        inner_width, pixel_ratio = driver.execute_script(
            "return [window.innerWidth, window.devicePixelRatio];"
        )
        if abs(float(inner_width) - 640.0) > 2:
            raise AssertionError(f"2x scale: CSS viewport is {inner_width!r}, want approximately 640")
        if float(pixel_ratio) < 1.9:
            raise AssertionError(f"2x scale: devicePixelRatio is {pixel_ratio!r}, want approximately 2")

        for path, marker, targets, name in PAGES:
            driver.get(f"{BASE_URL}{path}")
            wait.until(EC.visibility_of_element_located((By.CSS_SELECTOR, marker)))
            context = f"2x-scale/640-css/{name}"
            root = driver.find_element(By.TAG_NAME, "html")
            if root.get_attribute("data-glaze-version") != "1.1":
                raise AssertionError(f"{context}: Glaze UI V1.1 contract was lost")
            assert_no_horizontal_overflow(driver, context)
            assert_controls(driver, targets, context)
            capture(driver, f"glaze-v1-1-2x-scale-{name}")
    finally:
        driver.execute_cdp_cmd("Emulation.clearDeviceMetricsOverride", {})
        driver.quit()


def main() -> int:
    assert_rtl_resilience()
    assert_two_x_scale_reflow()
    print("native Glaze UI V1.1 resilience browser acceptance passed")
    print("note: RTL is structural stress evidence, not localization acceptance")
    print("note: 2x device-scale/640 CSS reflow is deterministic CI evidence, not a substitute for manual browser 200% zoom acceptance")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
