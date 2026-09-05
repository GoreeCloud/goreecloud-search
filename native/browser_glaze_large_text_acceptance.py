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
TEXT_SCALE = 2.0


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

PAGES = (
    (
        "/",
        "#hero-title",
        (".brand", "#q", ".search-box button", ".category-row .chip"),
        "home",
    ),
    (
        "/preferences",
        "#appearance",
        (
            ".brand",
            "[data-settings-filter]",
            "[data-settings-nav] a",
            "select[data-preference]",
        ),
        "preferences",
    ),
    (
        "/search",
        "#results-title",
        (
            ".results-brand",
            "#results-q",
            ".results-search button",
            ".results-categories a",
        ),
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


def horizontal_overflowers(driver: webdriver.Chrome) -> list[dict[str, object]]:
    return driver.execute_script(
        """
        const viewport = document.documentElement.clientWidth;
        const problems = [];
        for (const element of document.body.querySelectorAll('*')) {
          if (problems.length >= 12) break;
          const style = getComputedStyle(element);
          if (style.display === 'none' || style.visibility === 'hidden') continue;
          const rect = element.getBoundingClientRect();
          if (rect.width <= 0 || rect.height <= 0) continue;
          if (rect.left < -2 || rect.right > viewport + 2) {
            problems.push({
              tag: element.tagName.toLowerCase(),
              id: element.id || '',
              className: typeof element.className === 'string' ? element.className : '',
              left: Math.round(rect.left * 10) / 10,
              right: Math.round(rect.right * 10) / 10,
              width: Math.round(rect.width * 10) / 10,
              viewport,
              text: (element.innerText || '').trim().slice(0, 80),
            });
          }
        }
        return problems;
        """
    )


def assert_no_horizontal_overflow(driver: webdriver.Chrome, context: str) -> None:
    scroll_width, client_width = driver.execute_script(
        "return [document.documentElement.scrollWidth, document.documentElement.clientWidth];"
    )
    if scroll_width > client_width + 2:
        raise AssertionError(
            f"{context}: horizontal overflow {scroll_width}px > {client_width}px; "
            f"offenders={horizontal_overflowers(driver)!r}"
        )


def assert_target(element, context: str) -> None:
    rect = element.rect
    width = float(rect.get("width", 0))
    height = float(rect.get("height", 0))
    if width + 0.5 < TARGET_FLOOR or height + 0.5 < TARGET_FLOOR:
        raise AssertionError(
            f"{context}: target {width:.1f}x{height:.1f}px; expected >= 48px"
        )


def viewport_rect(driver: webdriver.Chrome, element) -> tuple[float, float, float]:
    x, width, viewport = driver.execute_script(
        "const r=arguments[0].getBoundingClientRect(); return [r.x, r.width, window.innerWidth];",
        element,
    )
    return float(x), float(width), float(viewport)


def assert_reachable_without_document_overflow(
    driver: webdriver.Chrome, element, context: str
) -> None:
    x, width, viewport = viewport_rect(driver, element)
    escaped = x < -2 or x + width > viewport + 2
    if escaped:
        driver.execute_script(
            "arguments[0].scrollIntoView({block:'nearest', inline:'nearest', behavior:'auto'});",
            element,
        )
        assert_no_horizontal_overflow(driver, f"{context} after reachability scroll")
        x, width, viewport = viewport_rect(driver, element)
        escaped = x < -2 or x + width > viewport + 2
    if escaped:
        raise AssertionError(
            f"{context}: control cannot be brought into the viewport without document overflow: "
            f"x={x:.1f}, width={width:.1f}, viewport={viewport:.1f}"
        )


def assert_controls(
    driver: webdriver.Chrome, selectors: tuple[str, ...], context: str
) -> None:
    for selector in selectors:
        for index, element in enumerate(
            driver.find_elements(By.CSS_SELECTOR, selector), start=1
        ):
            if not element.is_displayed():
                continue
            assert_target(element, f"{context} {selector} #{index}")
            assert_reachable_without_document_overflow(
                driver, element, f"{context} {selector} #{index}"
            )


def apply_two_x_text_scale(driver: webdriver.Chrome, marker_selector: str) -> None:
    marker = driver.find_element(By.CSS_SELECTOR, marker_selector)
    before = float(
        driver.execute_script(
            "return parseFloat(getComputedStyle(arguments[0]).fontSize);", marker
        )
    )
    count = driver.execute_script(
        """
        const nodes = [document.body, ...document.body.querySelectorAll('*')];
        const original = nodes.map((element) => [
          element,
          parseFloat(getComputedStyle(element).fontSize),
        ]);
        let changed = 0;
        for (const [element, size] of original) {
          if (Number.isFinite(size) && size > 0) {
            element.style.fontSize = `${size * 2}px`;
            changed += 1;
          }
        }
        document.documentElement.dataset.goreecloudTextScale = '200';
        return changed;
        """
    )
    if int(count) < 1:
        raise AssertionError("200% text scale: no rendered font sizes were updated")
    after = float(
        driver.execute_script(
            "return parseFloat(getComputedStyle(arguments[0]).fontSize);", marker
        )
    )
    if before <= 0 or after + 0.5 < before * TEXT_SCALE:
        raise AssertionError(
            f"200% text scale: marker font-size changed from {before:.1f}px to "
            f"{after:.1f}px; expected at least {before * TEXT_SCALE:.1f}px"
        )


def assert_no_hidden_text_clipping(driver: webdriver.Chrome, context: str) -> None:
    problems = driver.execute_script(
        """
        const clippingValues = new Set(['hidden', 'clip']);
        const scrollValues = new Set(['hidden', 'clip', 'auto', 'scroll']);
        const problems = [];

        function nestedClipBetween(node, ancestor, axis) {
          let current = node.parentElement;
          while (current && current !== ancestor) {
            const style = getComputedStyle(current);
            const overflow = axis === 'x' ? style.overflowX : style.overflowY;
            if (scrollValues.has(overflow)) return true;
            current = current.parentElement;
          }
          return false;
        }

        for (const element of document.body.querySelectorAll('*')) {
          if (problems.length >= 12) break;
          if (element.classList.contains('sr-only')) continue;
          const style = getComputedStyle(element);
          if (style.display === 'none' || style.visibility === 'hidden') continue;
          const clipsX = clippingValues.has(style.overflowX);
          const clipsY = clippingValues.has(style.overflowY);
          if (!clipsX && !clipsY) continue;

          const elementRect = element.getBoundingClientRect();
          const walker = document.createTreeWalker(element, NodeFilter.SHOW_TEXT);
          let textNode = walker.nextNode();
          while (textNode && problems.length < 12) {
            if ((textNode.textContent || '').trim()) {
              const nestedX = clipsX && nestedClipBetween(textNode, element, 'x');
              const nestedY = clipsY && nestedClipBetween(textNode, element, 'y');
              const range = document.createRange();
              range.selectNodeContents(textNode);
              for (const rect of range.getClientRects()) {
                const clippedX = clipsX && !nestedX &&
                  (rect.left < elementRect.left - 2 || rect.right > elementRect.right + 2);
                const clippedY = clipsY && !nestedY &&
                  (rect.top < elementRect.top - 2 || rect.bottom > elementRect.bottom + 2);
                if (clippedX || clippedY) {
                  problems.push({
                    tag: element.tagName.toLowerCase(),
                    id: element.id || '',
                    className: typeof element.className === 'string' ? element.className : '',
                    clippedX,
                    clippedY,
                    text: (textNode.textContent || '').trim().slice(0, 80),
                    elementLeft: Math.round(elementRect.left * 10) / 10,
                    elementRight: Math.round(elementRect.right * 10) / 10,
                    textLeft: Math.round(rect.left * 10) / 10,
                    textRight: Math.round(rect.right * 10) / 10,
                  });
                  break;
                }
              }
            }
            textNode = walker.nextNode();
          }
        }
        return problems;
        """
    )
    if problems:
        raise AssertionError(f"{context}: clipped visible text detected: {problems!r}")


def assert_two_x_text_resilience() -> None:
    for viewport in VIEWPORTS:
        driver = driver_for(viewport)
        wait = WebDriverWait(driver, 10)
        try:
            for path, marker, targets, name in PAGES:
                store_preferences(driver)
                driver.get(f"{BASE_URL}{path}")
                wait.until(EC.visibility_of_element_located((By.CSS_SELECTOR, marker)))
                context = f"text-200/{viewport.name}/{name}"
                root = driver.find_element(By.TAG_NAME, "html")
                if root.get_attribute("data-glaze-version") != "1.1":
                    raise AssertionError(f"{context}: Glaze UI V1.1 contract was lost")
                apply_two_x_text_scale(driver, marker)
                if root.get_attribute("data-goreecloud-text-scale") != "200":
                    raise AssertionError(f"{context}: deterministic 200% text marker missing")
                capture(driver, f"glaze-v1-1-text-200-{viewport.name}-{name}")
                assert_no_horizontal_overflow(driver, context)
                assert_controls(driver, targets, context)
                assert_no_hidden_text_clipping(driver, context)
        finally:
            driver.quit()


def main() -> int:
    assert_two_x_text_resilience()
    print("native Glaze UI V1.1 deterministic 200% text stress acceptance passed")
    print(
        "note: deterministic text-only scaling is resilience evidence, not a substitute "
        "for manual browser zoom, assistive-technology, localization, or physical-device acceptance"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
