# SPDX-License-Identifier: AGPL-3.0-or-later
"""Deterministic provider-failure degradation acceptance for GoreeCloud Search.

The isolated CI runtime contains exactly two providers: one deterministic
healthy offline provider and one intentionally unreachable online provider.
This proves that an upstream-style provider failure is contained inside
GoreeCloud Search while healthy-provider results and the GoreeCloud product
shell remain usable. The healthy provider is deliberately offline so the test
does not weaken Search's outbound-network protections just to create a fixture.
"""

from __future__ import annotations

import argparse
import html.parser
import sys
import urllib.parse
import urllib.request


HEALTHY_ENGINE = "goreecloud acceptance healthy"
FAILING_ENGINE = "goreecloud acceptance failing"
HEALTHY_RESULT_URL = "https://example.invalid/goreecloud-provider-healthy"
FORBIDDEN_FALLBACK_HOSTS = (
    "google.com",
    "www.google.com",
    "bing.com",
    "www.bing.com",
    "duckduckgo.com",
    "www.duckduckgo.com",
    "search.brave.com",
    "perplexity.ai",
    "www.perplexity.ai",
)


class DegradationParser(html.parser.HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self.results = 0
        self.product_identity = False
        self.hrefs: list[str] = []
        self.engine_names: list[str] = []
        self.error_cells = 0
        self._capture_engine_name = False

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        attributes = dict(attrs)
        classes = set((attributes.get("class") or "").split())
        if tag == "article" and "result" in classes:
            self.results += 1
        if tag == "meta" and attributes.get("name") == "application-name":
            self.product_identity = attributes.get("content") == "GoreeCloud Search"
        if tag == "a" and attributes.get("href"):
            self.hrefs.append(attributes["href"] or "")
        if tag == "td" and "engine-name" in classes:
            self._capture_engine_name = True
        if tag == "td" and "response-error" in classes:
            self.error_cells += 1

    def handle_endtag(self, tag: str) -> None:
        if tag == "td":
            self._capture_engine_name = False

    def handle_data(self, data: str) -> None:
        text = data.strip()
        if "GoreeCloud Search" in text:
            self.product_identity = True
        if self._capture_engine_name and text:
            self.engine_names.append(text)


def _host(url: str) -> str:
    try:
        return (urllib.parse.urlparse(url).hostname or "").lower()
    except ValueError:
        return ""


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base-url", default="http://127.0.0.1:8888")
    parser.add_argument("--query", default="GoreeCloud deterministic provider degradation")
    parser.add_argument("--timeout", type=float, default=10.0)
    args = parser.parse_args()

    params = urllib.parse.urlencode(
        {
            "q": args.query,
            "categories": "general",
            "language": "auto",
            "safesearch": "0",
        }
    )
    url = f"{args.base_url.rstrip('/')}/search?{params}"
    request = urllib.request.Request(
        url,
        headers={
            "User-Agent": "GoreeCloud-Search-Provider-Degradation-Acceptance/1.0",
            "Accept": "text/html,application/xhtml+xml",
        },
    )

    try:
        with urllib.request.urlopen(request, timeout=args.timeout) as response:
            status = response.status
            body = response.read().decode("utf-8", errors="replace")
    except Exception as exc:
        print(f"Provider degradation acceptance request failed: {exc}", file=sys.stderr)
        return 2

    parsed = DegradationParser()
    parsed.feed(body)
    body_lower = body.lower()
    observed_engines = set(parsed.engine_names)
    unexpected_engines = observed_engines - {HEALTHY_ENGINE, FAILING_ENGINE}
    fallback_hosts = sorted({_host(href) for href in parsed.hrefs} & set(FORBIDDEN_FALLBACK_HOSTS))

    checks = {
        "http_200": status == 200,
        "goreecloud_identity": parsed.product_identity,
        "healthy_result_rendered": parsed.results >= 1 and HEALTHY_RESULT_URL in body,
        "healthy_engine_visible": HEALTHY_ENGINE in body_lower,
        "failed_engine_visible": FAILING_ENGINE in body_lower,
        "failure_surface_present": parsed.error_cells >= 1,
        "only_fixture_engines_observed": not unexpected_engines,
        "no_external_fallback_links": not fallback_hosts,
    }

    print(f"URL: {url}")
    print(f"HTTP status: {status}")
    print(f"Result cards: {parsed.results}")
    print(f"Engine error cells: {parsed.error_cells}")
    print(f"Observed engine rows: {', '.join(sorted(observed_engines)) or '(none)'}")
    print(f"Unexpected engine rows: {', '.join(sorted(unexpected_engines)) or '(none)'}")
    print(f"Forbidden fallback hosts: {', '.join(fallback_hosts) or '(none)'}")
    for name, passed in checks.items():
        print(f"{name}: {'PASS' if passed else 'FAIL'}")

    failed = [name for name, passed in checks.items() if not passed]
    if failed:
        print("Deterministic provider degradation acceptance failed:", file=sys.stderr)
        for name in failed:
            print(f"- {name}", file=sys.stderr)
        return 3

    print(
        "Deterministic provider degradation acceptance passed: the failing provider stayed bounded "
        "inside GoreeCloud Search, the healthy provider remained usable, and no external search "
        "fallback authority appeared in the response."
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
