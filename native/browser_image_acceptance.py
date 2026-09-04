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


def set_media(driver: webdriver.Chrome, scheme: str, forced_colors: str = "none") -> None:
    driver.execute_cdp_cmd(
        "Emulation.setEmulatedMedia",
        {
            "media": "screen",
            "features": [
                {"name": "prefers-color-scheme", "value": scheme},
                {"name": "prefers-reduced-motion", "value": "no-preference"},
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


def assert_min_target(element, context: str) -> None:
    rect = element.rect
    width = float(rect.get("width", 0))
    height = float(rect.get("height", 0))
    if width + 0.5 < TARGET_FLOOR or height + 0.5 < TARGET_FLOOR:
        raise AssertionError(f"{context}: target {width:.1f}x{height:.1f}px; expected >= {TARGET_FLOOR:.0f}px")


def assert_no_horizontal_overflow(driver: webdriver.Chrome, context: str) -> None:
    scroll_width, client_width = driver.execute_script(
        "return [document.documentElement.scrollWidth, document.documentElement.clientWidth];"
    )
    if scroll_width > client_width + 2:
        raise AssertionError(f"{context}: horizontal overflow {scroll_width}px > {client_width}px")


def assert_image_loaded(driver: webdriver.Chrome, image, context: str) -> None:
    complete, width = driver.execute_script("return [arguments[0].complete, arguments[0].naturalWidth]", image)
    if not complete or width <= 0:
        raise AssertionError(f"{context}: image did not load (complete={complete}, naturalWidth={width})")


def assert_image_page(driver: webdriver.Chrome, wait: WebDriverWait, viewport: Viewport, scheme: str) -> None:
    context = f"{viewport.name}/{scheme} images"
    driver.get(f"{BASE_URL}/search?case=images")
    wait.until(EC.visibility_of_element_located((By.CSS_SELECTOR, ".image-results-grid")))

    current = driver.find_elements(By.CSS_SELECTOR, '.results-categories a[aria-current="page"]')
    if len(current) != 1 or current[0].text != "Images":
        raise AssertionError(f"{context}: Images category is not selected")

    cards = driver.find_elements(By.CSS_SELECTOR, ".image-result-card")
    openers = driver.find_elements(By.CSS_SELECTOR, "[data-image-open]")
    dialogs = driver.find_elements(By.CSS_SELECTOR, "[data-image-dialog]")
    if len(cards) != 3 or len(openers) != 3 or len(dialogs) != 3:
        raise AssertionError(f"{context}: expected three image cards/openers/dialogs")

    for index, opener in enumerate(openers):
        assert_min_target(opener, f"{context} opener {index + 1}")
        image = opener.find_element(By.TAG_NAME, "img")
        wait.until(lambda _driver, image=image: _driver.execute_script("return arguments[0].complete", image))
        assert_image_loaded(driver, image, f"{context} thumbnail {index + 1}")

    assert_no_horizontal_overflow(driver, context)
    capture(driver, f"{viewport.name}-{scheme}-native-image-grid")

    first_opener = openers[0]
    first_opener.click()
    first_dialog = dialogs[0]
    wait.until(lambda _driver: _driver.execute_script("return arguments[0].open", first_dialog))
    close = first_dialog.find_element(By.CSS_SELECTOR, "[data-image-close]")
    assert_min_target(close, f"{context} close")
    if driver.switch_to.active_element != close:
        raise AssertionError(f"{context}: viewer did not focus the close control")

    full_image = first_dialog.find_element(By.CSS_SELECTOR, ".image-viewer-stage img")
    assert_image_loaded(driver, full_image, f"{context} full image")
    source = first_dialog.find_element(By.CSS_SELECTOR, ".image-viewer-source")
    previous = first_dialog.find_element(By.CSS_SELECTOR, "[data-image-previous]")
    next_button = first_dialog.find_element(By.CSS_SELECTOR, "[data-image-next]")
    for element, label in ((source, "source"), (previous, "previous"), (next_button, "next")):
        assert_min_target(element, f"{context} {label}")
    if not source.get_attribute("href").startswith("https://photos.goreecloud.example/"):
        raise AssertionError(f"{context}: source action lost result destination")
    capture(driver, f"{viewport.name}-{scheme}-native-image-viewer")

    close.send_keys(Keys.ARROW_RIGHT)
    second_dialog = dialogs[1]
    wait.until(lambda _driver: _driver.execute_script("return arguments[0].open", second_dialog))
    if driver.execute_script("return arguments[0].open", first_dialog):
        raise AssertionError(f"{context}: ArrowRight left the previous dialog open")
    second_close = second_dialog.find_element(By.CSS_SELECTOR, "[data-image-close]")
    if driver.switch_to.active_element != second_close:
        raise AssertionError(f"{context}: next dialog did not focus its close control")

    second_close.send_keys(Keys.ESCAPE)
    wait.until(lambda _driver: not _driver.execute_script("return arguments[0].open", second_dialog))
    if driver.switch_to.active_element.get_attribute("id") != "image-opener-2":
        raise AssertionError(f"{context}: closing the navigated viewer did not restore focus to its opener")
    assert_no_horizontal_overflow(driver, f"{context} after viewer")


def assert_forced_colors() -> None:
    viewport = Viewport("compact", 390, 844)
    driver = driver_for(viewport)
    wait = WebDriverWait(driver, 10)
    try:
        set_media(driver, "light", forced_colors="active")
        driver.get(f"{BASE_URL}/search?case=images")
        wait.until(EC.visibility_of_element_located((By.CSS_SELECTOR, ".image-results-grid")))
        if not driver.execute_script("return matchMedia('(forced-colors: active)').matches"):
            raise AssertionError("image forced-colors emulation is not active")
        assert_no_horizontal_overflow(driver, "image forced colors")
        capture(driver, "compact-light-native-images-forced-colors")
    finally:
        driver.quit()


def main() -> int:
    for viewport in VIEWPORTS:
        driver = driver_for(viewport)
        wait = WebDriverWait(driver, 10)
        try:
            for scheme in ("light", "dark"):
                set_media(driver, scheme)
                assert_image_page(driver, wait, viewport, scheme)
        finally:
            driver.quit()
    assert_forced_colors()
    print("native image results browser acceptance passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
