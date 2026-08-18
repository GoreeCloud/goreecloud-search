# SPDX-License-Identifier: AGPL-3.0-or-later
"""Focused browser acceptance for GoreeCloud Search Preferences.

The shared browser-acceptance suite proves overall Glaze UI behavior. This
module specifically exercises every top-level Preferences surface and the
native radio-group keyboard contract so a visually polished settings page
cannot regress into mouse-only navigation.
"""

from __future__ import annotations

import os
import sys
import time
from dataclasses import dataclass

from selenium import webdriver
from selenium.common.exceptions import WebDriverException
from selenium.webdriver import ActionChains
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.remote.webelement import WebElement
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait


BASE_URL = os.environ.get("GOREECLOUD_SEARCH_TEST_URL", "http://127.0.0.1:8888").rstrip("/")
MIN_TARGET_SIZE = 44
TAB_IDS = ("general", "ui", "privacy", "engines", "query", "cookies")


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


def _driver(viewport: Viewport) -> webdriver.Chrome:
    options = webdriver.ChromeOptions()
    options.add_argument("--headless=new")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    options.add_argument(f"--window-size={viewport.width},{viewport.height}")
    options.add_argument("--force-device-scale-factor=1")
    return webdriver.Chrome(options=options)


def _assert_target_size(element: WebElement, context: str) -> None:
    rect = element.rect
    width = float(rect.get("width", 0))
    height = float(rect.get("height", 0))
    if width + 0.5 < MIN_TARGET_SIZE or height + 0.5 < MIN_TARGET_SIZE:
        raise AssertionError(
            f"{context}: rendered target is {width:.1f}x{height:.1f}px; expected at least "
            f"{MIN_TARGET_SIZE}x{MIN_TARGET_SIZE}px"
        )


def _overflow_state(driver: webdriver.Chrome) -> dict[str, object]:
    return driver.execute_script(
        """
        const viewport = document.documentElement.clientWidth;
        const scrollWidth = document.documentElement.scrollWidth;

        function scrollContainerFor(element) {
          let parent = element.parentElement;
          while (parent && parent !== document.body && parent !== document.documentElement) {
            const style = getComputedStyle(parent);
            const overflowX = style.overflowX;
            const canContain = overflowX === 'auto' || overflowX === 'scroll' || overflowX === 'hidden' || overflowX === 'clip';
            if (canContain) {
              const rect = parent.getBoundingClientRect();
              if (rect.left >= -2 && rect.right <= viewport + 2) return parent;
            }
            parent = parent.parentElement;
          }
          return null;
        }

        const uncontained = Array.from(document.querySelectorAll('body *'))
          .map((element) => {
            const rect = element.getBoundingClientRect();
            const outside = rect.right > viewport + 2 || rect.left < -2;
            if (!outside || scrollContainerFor(element)) return null;
            return {
              tag: element.tagName.toLowerCase(),
              id: element.id || '',
              classes: typeof element.className === 'string' ? element.className : '',
              left: Math.round(rect.left * 10) / 10,
              right: Math.round(rect.right * 10) / 10,
              width: Math.round(rect.width * 10) / 10,
            };
          })
          .filter(Boolean)
          .slice(0, 8);

        return {viewport, scrollWidth, uncontained};
        """
    )


def _assert_no_horizontal_overflow(driver: webdriver.Chrome, context: str) -> None:
    state = _overflow_state(driver)
    uncontained = state.get("uncontained") or []
    if uncontained:
        details = " | ".join(
            f"{item['tag']}{'#' + item['id'] if item['id'] else ''}"
            f"{'.' + str(item['classes']).strip().replace(' ', '.') if item['classes'] else ''} "
            f"[left={item['left']}, right={item['right']}, width={item['width']}]"
            for item in uncontained
        )
        raise AssertionError(
            f"{context}: uncontained horizontal overflow detected: "
            f"scrollWidth={state.get('scrollWidth')}, clientWidth={state.get('viewport')}; offenders: {details}"
        )


def _assert_preferences(driver: webdriver.Chrome, wait: WebDriverWait, viewport: Viewport) -> None:
    driver.get(f"{BASE_URL}/preferences")
    form = wait.until(EC.presence_of_element_located((By.CSS_SELECTOR, "form.goreecloud-preferences-form")))
    if form.get_attribute("id") != "search_form":
        raise AssertionError(f"{viewport.name}: Preferences form identity is incomplete")

    body_text = driver.find_element(By.TAG_NAME, "body").text.casefold()
    if "goreecloud search" not in body_text or "search settings" not in body_text or "preferences" not in body_text:
        raise AssertionError(f"{viewport.name}: Preferences product identity is incomplete")

    radios = driver.find_elements(
        By.CSS_SELECTOR,
        '.goreecloud-preferences-form > .tabs > input[type="radio"][name="maintab"]',
    )
    radio_ids = tuple(radio.get_attribute("id") for radio in radios)
    expected_radio_ids = tuple(f"tab-{tab_id}" for tab_id in TAB_IDS)
    if radio_ids != expected_radio_ids:
        raise AssertionError(
            f"{viewport.name}: Preferences main-tab order is {radio_ids!r}; expected {expected_radio_ids!r}"
        )

    first_radio = radios[0]
    driver.execute_script("arguments[0].focus()", first_radio)
    if driver.switch_to.active_element.get_attribute("id") != "tab-general":
        raise AssertionError(f"{viewport.name}: native Preferences radio could not receive keyboard focus")

    ActionChains(driver).send_keys(Keys.ARROW_RIGHT).perform()
    wait.until(lambda _driver: driver.find_element(By.ID, "tab-ui").is_selected())
    if not driver.find_element(By.ID, "tab-content-ui").is_displayed():
        raise AssertionError(f"{viewport.name}: ArrowRight selected User interface but its panel stayed hidden")

    for tab_id in TAB_IDS:
        radio = driver.find_element(By.ID, f"tab-{tab_id}")
        label = driver.find_element(By.CSS_SELECTOR, f'label[for="tab-{tab_id}"]')
        panel_id = radio.get_attribute("aria-controls")
        expected_panel_id = f"tab-content-{tab_id}"
        if panel_id != expected_panel_id:
            raise AssertionError(
                f"{viewport.name} {tab_id}: aria-controls is {panel_id!r}; expected {expected_panel_id!r}"
            )

        panel = driver.find_element(By.ID, expected_panel_id)
        if panel.get_attribute("aria-labelledby") != label.get_attribute("id"):
            raise AssertionError(f"{viewport.name} {tab_id}: panel label relationship is incomplete")
        if panel.get_attribute("role") != "region":
            raise AssertionError(f"{viewport.name} {tab_id}: panel does not expose region semantics")

        driver.execute_script(
            "arguments[0].scrollIntoView({block: 'nearest', inline: 'nearest'});",
            label,
        )
        _assert_target_size(label, f"{viewport.name} Preferences {tab_id} tab")
        label.click()
        wait.until(lambda _driver, target=radio: target.is_selected())
        if not panel.is_displayed():
            raise AssertionError(f"{viewport.name} {tab_id}: selected Preferences panel is not displayed")
        _assert_no_horizontal_overflow(driver, f"{viewport.name} Preferences {tab_id}")

    visible_submit = next(
        (
            element
            for element in driver.find_elements(By.CSS_SELECTOR, 'input[type="submit"], button[type="submit"]')
            if element.is_displayed()
        ),
        None,
    )
    if visible_submit is None:
        raise AssertionError(f"{viewport.name}: Preferences Save action is not visible")
    _assert_target_size(visible_submit, f"{viewport.name} Preferences Save action")

    stylesheet_urls = [
        element.get_attribute("href") for element in driver.find_elements(By.CSS_SELECTOR, 'link[rel="stylesheet"]')
    ]
    if not any(url and "goreecloud-page-compact.css" in url for url in stylesheet_urls):
        raise AssertionError(f"{viewport.name}: keyboard-accessible Preferences style layer is not loaded")


def run() -> None:
    errors: list[str] = []
    for viewport in VIEWPORTS:
        driver: webdriver.Chrome | None = None
        try:
            driver = _driver(viewport)
            driver.set_page_load_timeout(30)
            driver.set_script_timeout(20)
            driver.set_window_size(viewport.width, viewport.height)
            _assert_preferences(driver, WebDriverWait(driver, 20), viewport)
        except (AssertionError, WebDriverException) as exc:
            errors.append(f"{viewport.name}: {exc}")
        finally:
            if driver is not None:
                driver.quit()
            time.sleep(0.25)

    if errors:
        for error in errors:
            print(error, file=sys.stderr)
        raise SystemExit(1)

    print(
        "GoreeCloud Search Preferences acceptance passed all six main settings surfaces "
        "and native keyboard navigation across Compact, Medium, Expanded, and Wide layouts."
    )


if __name__ == "__main__":
    run()
