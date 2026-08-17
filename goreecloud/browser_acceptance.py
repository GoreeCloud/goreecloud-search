# SPDX-License-Identifier: AGPL-3.0-or-later
"""Browser-level acceptance checks for the GoreeCloud Search product shell.

This module intentionally validates durable product behavior rather than pixel
snapshots. It is run by the GoreeCloud browser-acceptance GitHub Actions
workflow against a locally started application instance.
"""

from __future__ import annotations

import os
import sys
import time
from dataclasses import dataclass

from selenium import webdriver
from selenium.common.exceptions import WebDriverException
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.remote.webelement import WebElement
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait


BASE_URL = os.environ.get("GOREECLOUD_SEARCH_TEST_URL", "http://127.0.0.1:8888").rstrip("/")
MIN_TARGET_SIZE = 44


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


def _fetch_text(driver: webdriver.Chrome, url: str) -> dict[str, object]:
    return driver.execute_async_script(
        """
        const url = arguments[0];
        const done = arguments[arguments.length - 1];
        fetch(url, {credentials: 'same-origin'})
          .then((response) => {
            if (!response.ok) throw new Error(`HTTP ${response.status}`);
            return response.text();
          })
          .then((payload) => done({ok: true, payload}))
          .catch((error) => done({ok: false, error: String(error)}));
        """,
        url,
    )


def _overflow_diagnostics(driver: webdriver.Chrome) -> list[str]:
    return driver.execute_script(
        """
        const viewport = document.documentElement.clientWidth;
        return Array.from(document.querySelectorAll('body *'))
          .map((element) => {
            const rect = element.getBoundingClientRect();
            return {
              tag: element.tagName.toLowerCase(),
              id: element.id || '',
              classes: typeof element.className === 'string' ? element.className : '',
              left: Math.round(rect.left * 10) / 10,
              right: Math.round(rect.right * 10) / 10,
              width: Math.round(rect.width * 10) / 10,
            };
          })
          .filter((item) => item.right > viewport + 2 || item.left < -2)
          .sort((a, b) => Math.max(b.right - viewport, -b.left) - Math.max(a.right - viewport, -a.left))
          .slice(0, 8)
          .map((item) => `${item.tag}${item.id ? '#' + item.id : ''}${item.classes ? '.' + item.classes.trim().replace(/\\s+/g, '.') : ''} [left=${item.left}, right=${item.right}, width=${item.width}]`);
        """
    )


def _assert_no_horizontal_overflow(driver: webdriver.Chrome, context: str) -> None:
    scroll_width, client_width = driver.execute_script(
        "return [document.documentElement.scrollWidth, document.documentElement.clientWidth];"
    )
    if scroll_width > client_width + 2:
        details = _overflow_diagnostics(driver)
        suffix = f"; offenders: {' | '.join(details)}" if details else ""
        raise AssertionError(
            f"{context}: horizontal overflow detected: scrollWidth={scroll_width}, clientWidth={client_width}{suffix}"
        )


def _assert_target_size(element: WebElement, context: str) -> None:
    rect = element.rect
    width = float(rect.get("width", 0))
    height = float(rect.get("height", 0))
    if width + 0.5 < MIN_TARGET_SIZE or height + 0.5 < MIN_TARGET_SIZE:
        raise AssertionError(
            f"{context}: rendered target is {width:.1f}x{height:.1f}px; expected at least "
            f"{MIN_TARGET_SIZE}x{MIN_TARGET_SIZE}px"
        )


def _assert_browser_metadata(driver: webdriver.Chrome, viewport: Viewport) -> None:
    application_name = driver.find_element(By.CSS_SELECTOR, 'meta[name="application-name"]').get_attribute("content")
    if application_name != "GoreeCloud Search":
        raise AssertionError(f"{viewport.name}: browser application name is {application_name!r}")

    robots = driver.find_element(By.CSS_SELECTOR, 'meta[name="robots"]').get_attribute("content").lower()
    if "noindex" not in robots or "nofollow" not in robots:
        raise AssertionError(f"{viewport.name}: private-search robots metadata is incomplete: {robots!r}")

    favicon = driver.find_element(By.CSS_SELECTOR, 'link[rel="icon"]').get_attribute("href")
    if not favicon or not favicon.endswith("favicon.svg"):
        raise AssertionError(f"{viewport.name}: browser icon is not the GoreeCloud SVG mark: {favicon!r}")

    manifest_url = driver.find_element(By.CSS_SELECTOR, 'link[rel="manifest"]').get_attribute("href")
    if not manifest_url:
        raise AssertionError(f"{viewport.name}: web app manifest link is missing")

    manifest = driver.execute_async_script(
        """
        const url = arguments[0];
        const done = arguments[arguments.length - 1];
        fetch(url, {credentials: 'same-origin'})
          .then((response) => {
            if (!response.ok) throw new Error(`HTTP ${response.status}`);
            return response.json();
          })
          .then((payload) => done({ok: true, payload}))
          .catch((error) => done({ok: false, error: String(error)}));
        """,
        manifest_url,
    )
    if not manifest.get("ok"):
        raise AssertionError(f"{viewport.name}: manifest fetch failed: {manifest.get('error')}")
    payload = manifest.get("payload") or {}
    if payload.get("name") != "GoreeCloud Search" or payload.get("short_name") != "GoreeCloud Search":
        raise AssertionError(f"{viewport.name}: manifest GoreeCloud identity is incomplete: {payload!r}")
    if payload.get("description") != "Private metasearch and research for GoreeCloud":
        raise AssertionError(f"{viewport.name}: manifest description is not GoreeCloud-owned: {payload!r}")
    if payload.get("display") != "standalone":
        raise AssertionError(f"{viewport.name}: manifest display mode is {payload.get('display')!r}")
    icons = payload.get("icons") or []
    if len(icons) != 1 or not str(icons[0].get("src", "")).endswith("favicon.svg"):
        raise AssertionError(f"{viewport.name}: manifest advertises non-GoreeCloud artwork: {icons!r}")

    search_provider = driver.find_element(By.CSS_SELECTOR, 'link[rel="search"]').get_attribute("href")
    if not search_provider:
        raise AssertionError(f"{viewport.name}: OpenSearch provider link is missing")
    opensearch = _fetch_text(driver, search_provider)
    if not opensearch.get("ok"):
        raise AssertionError(f"{viewport.name}: OpenSearch fetch failed: {opensearch.get('error')}")
    opensearch_text = str(opensearch.get("payload") or "")
    if "GoreeCloud Search private metasearch" not in opensearch_text:
        raise AssertionError(f"{viewport.name}: OpenSearch long name is not GoreeCloud-owned")
    if "img/favicon.svg" not in opensearch_text:
        raise AssertionError(f"{viewport.name}: OpenSearch does not advertise GoreeCloud SVG artwork")
    if "SearXNG metasearch" in opensearch_text:
        raise AssertionError(f"{viewport.name}: upstream OpenSearch product identity leaked into GoreeCloud Search")

    color_scheme = driver.find_element(By.CSS_SELECTOR, 'meta[name="color-scheme"]').get_attribute("content")
    if "light" not in color_scheme or "dark" not in color_scheme:
        raise AssertionError(f"{viewport.name}: browser color-scheme metadata is incomplete")


def _assert_home(driver: webdriver.Chrome, wait: WebDriverWait, viewport: Viewport) -> None:
    driver.get(f"{BASE_URL}/")
    wait.until(EC.title_is("GoreeCloud Search"))
    heading = wait.until(EC.visibility_of_element_located((By.CSS_SELECTOR, ".index .title h1")))
    if "GoreeCloud Search" not in heading.text:
        raise AssertionError(f"{viewport.name}: expected GoreeCloud Search heading, got {heading.text!r}")
    query = wait.until(EC.visibility_of_element_located((By.ID, "q")))
    submit = wait.until(EC.element_to_be_clickable((By.ID, "send_search")))
    if not query.get_attribute("placeholder"):
        raise AssertionError(f"{viewport.name}: search input has no placeholder")
    if not submit.get_attribute("aria-label"):
        raise AssertionError(f"{viewport.name}: search button has no accessible label")
    _assert_target_size(query, f"{viewport.name} search input")
    _assert_target_size(submit, f"{viewport.name} search submit")
    query.click()
    if driver.switch_to.active_element.get_attribute("id") != "q":
        raise AssertionError(f"{viewport.name}: search input did not receive focus")
    query.send_keys(Keys.TAB)
    if not driver.switch_to.active_element.is_displayed():
        raise AssertionError(f"{viewport.name}: keyboard focus moved to a hidden element")
    stylesheet_urls = [
        element.get_attribute("href") for element in driver.find_elements(By.CSS_SELECTOR, 'link[rel="stylesheet"]')
    ]
    if not any(url and "goreecloud.css" in url for url in stylesheet_urls):
        raise AssertionError(f"{viewport.name}: GoreeCloud Glaze UI stylesheet is not loaded")
    if not any(url and "goreecloud-states.css" in url for url in stylesheet_urls):
        raise AssertionError(f"{viewport.name}: GoreeCloud secondary-state stylesheet is not loaded")
    _assert_browser_metadata(driver, viewport)
    _assert_no_horizontal_overflow(driver, f"{viewport.name} home")


def _assert_preferences(driver: webdriver.Chrome, wait: WebDriverWait, viewport: Viewport) -> None:
    driver.get(f"{BASE_URL}/preferences")
    wait.until(EC.presence_of_element_located((By.ID, "search_form")))
    body_text = driver.find_element(By.TAG_NAME, "body").text
    if "GoreeCloud Search" not in body_text or "Preferences" not in body_text:
        raise AssertionError(f"{viewport.name}: preferences page product identity is incomplete")
    visible_submit = next(
        (element for element in driver.find_elements(By.CSS_SELECTOR, 'input[type="submit"], button[type="submit"]') if element.is_displayed()),
        None,
    )
    if visible_submit is not None:
        _assert_target_size(visible_submit, f"{viewport.name} preferences submit")
    _assert_no_horizontal_overflow(driver, f"{viewport.name} preferences")


def _assert_about(driver: webdriver.Chrome, wait: WebDriverWait, viewport: Viewport) -> None:
    driver.get(f"{BASE_URL}/about")
    wait.until(EC.presence_of_element_located((By.TAG_NAME, "body")))
    body_text = driver.find_element(By.TAG_NAME, "body").text
    required_text = ("About GoreeCloud Search", "Privacy model", "Open-source foundation", "SearXNG")
    missing = [text for text in required_text if text not in body_text]
    if missing:
        raise AssertionError(f"{viewport.name}: about page is missing product contract text: {missing!r}")
    _assert_no_horizontal_overflow(driver, f"{viewport.name} about")


def _assert_not_found(driver: webdriver.Chrome, wait: WebDriverWait, viewport: Viewport) -> None:
    driver.get(f"{BASE_URL}/goreecloud-browser-acceptance-missing")
    heading = wait.until(EC.visibility_of_element_located((By.ID, "goreecloud-not-found-title")))
    if heading.text != "Page not found":
        raise AssertionError(f"{viewport.name}: 404 heading is {heading.text!r}")
    body_text = driver.find_element(By.TAG_NAME, "body").text
    if "GoreeCloud Search" not in body_text or "Return to search" not in body_text:
        raise AssertionError(f"{viewport.name}: 404 recovery surface is missing GoreeCloud product guidance")
    primary_action = wait.until(EC.visibility_of_element_located((By.CSS_SELECTOR, ".goreecloud-not-found .goreecloud-action-primary")))
    _assert_target_size(primary_action, f"{viewport.name} 404 primary action")
    _assert_no_horizontal_overflow(driver, f"{viewport.name} 404")


def run() -> None:
    errors: list[str] = []
    for viewport in VIEWPORTS:
        driver: webdriver.Chrome | None = None
        try:
            driver = _driver(viewport)
            driver.set_page_load_timeout(30)
            driver.set_script_timeout(20)
            driver.set_window_size(viewport.width, viewport.height)
            wait = WebDriverWait(driver, 20)
            _assert_home(driver, wait, viewport)
            _assert_preferences(driver, wait, viewport)
            _assert_about(driver, wait, viewport)
            _assert_not_found(driver, wait, viewport)
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
    print("GoreeCloud Search browser acceptance passed all Glaze UI adaptive layout classes, recovery surfaces, and browser integrations.")


if __name__ == "__main__":
    run()
