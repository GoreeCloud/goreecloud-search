# SPDX-License-Identifier: AGPL-3.0-or-later
"""Manual real-provider acceptance for GoreeCloud Search.

This check is deliberately separate from normal pull-request CI because external
search providers can throttle or block shared CI runners. It is intended for
explicit acceptance runs and target-environment testing.
"""

from __future__ import annotations

import argparse
import html.parser
import sys
import urllib.parse
import urllib.request
from dataclasses import dataclass


@dataclass(frozen=True)
class AcceptanceCase:
    category: str
    query: str


REPRESENTATIVE_SUITE = (
    AcceptanceCase("general", "GoreeCloud private search"),
    AcceptanceCase("images", "open source personal cloud"),
    AcceptanceCase("news", "privacy technology"),
    AcceptanceCase("videos", "self hosted search"),
    AcceptanceCase("it", "Python metasearch engine"),
    AcceptanceCase("science", "information retrieval privacy research"),
)


class ResultCounter(html.parser.HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self.results = 0
        self.engine_messages = 0
        self.product_identity = False

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        attributes = dict(attrs)
        classes = attributes.get("class", "") or ""
        class_names = set(classes.split())
        if tag == "article" and "result" in class_names:
            self.results += 1
        if "engines_msg" in class_names:
            self.engine_messages += 1
        if tag == "meta" and attributes.get("name") == "application-name":
            self.product_identity = attributes.get("content") == "GoreeCloud Search"

    def handle_data(self, data: str) -> None:
        if "GoreeCloud Search" in data:
            self.product_identity = True


def run_case(base_url: str, case: AcceptanceCase, minimum_results: int, timeout: float) -> int:
    params = urllib.parse.urlencode(
        {
            "q": case.query,
            "categories": case.category,
            "language": "auto",
            "safesearch": "0",
        }
    )
    url = f"{base_url.rstrip('/')}/search?{params}"
    request = urllib.request.Request(
        url,
        headers={
            "User-Agent": "GoreeCloud-Search-Provider-Acceptance/1.0",
            "Accept": "text/html,application/xhtml+xml",
        },
    )

    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            status = response.status
            body = response.read().decode("utf-8", errors="replace")
    except Exception as exc:  # explicit acceptance command should explain external failures
        print(f"[{case.category}] Provider acceptance request failed: {exc}", file=sys.stderr)
        return 2

    parser = ResultCounter()
    parser.feed(body)

    print(f"[{case.category}] URL: {url}")
    print(f"[{case.category}] HTTP status: {status}")
    print(f"[{case.category}] GoreeCloud Search identity: {'yes' if parser.product_identity else 'no'}")
    print(f"[{case.category}] Result cards: {parser.results}")
    print(f"[{case.category}] Engine-message surfaces: {parser.engine_messages}")

    if status != 200:
        print(f"[{case.category}] Expected HTTP 200, received {status}.", file=sys.stderr)
        return 3
    if not parser.product_identity:
        print(
            f"[{case.category}] Search response lost the GoreeCloud Search product identity.",
            file=sys.stderr,
        )
        return 4
    if parser.engine_messages:
        print(
            f"[{case.category}] Provider degradation is visible inside the intact GoreeCloud Search shell "
            f"({parser.engine_messages} engine message surface(s))."
        )
    if parser.results < minimum_results:
        print(
            f"[{case.category}] Expected at least {minimum_results} result card(s), received {parser.results}. "
            "External engines may be unavailable, throttled, or blocked from this runner. "
            "The product-identity and engine-message lines above distinguish safe degradation from shell failure.",
            file=sys.stderr,
        )
        return 5

    print(f"[{case.category}] Real-provider acceptance passed.")
    return 0


def run_suite(base_url: str, minimum_results: int, timeout: float) -> int:
    failures: list[tuple[str, int]] = []
    for case in REPRESENTATIVE_SUITE:
        result = run_case(base_url, case, minimum_results, timeout)
        if result != 0:
            failures.append((case.category, result))
        print()

    if failures:
        print("GoreeCloud Search representative provider suite did not fully pass:", file=sys.stderr)
        for category, code in failures:
            print(f"- {category}: exit code {code}", file=sys.stderr)
        print(
            "Classify each failure as application/runtime, provider throttling/blocking, engine initialization, "
            "or genuinely empty results before deciding whether the release candidate is defective.",
            file=sys.stderr,
        )
        return 6

    print("GoreeCloud Search representative real-provider suite passed.")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base-url", default="http://127.0.0.1:8888")
    parser.add_argument("--query", default="GoreeCloud private search")
    parser.add_argument("--category", default="general")
    parser.add_argument("--minimum-results", type=int, default=1)
    parser.add_argument("--timeout", type=float, default=45.0)
    parser.add_argument(
        "--suite",
        action="store_true",
        help="Run the representative general/images/news/videos/IT/science acceptance suite.",
    )
    args = parser.parse_args()

    if args.suite:
        return run_suite(args.base_url, args.minimum_results, args.timeout)
    return run_case(
        args.base_url,
        AcceptanceCase(args.category, args.query),
        args.minimum_results,
        args.timeout,
    )


if __name__ == "__main__":
    raise SystemExit(main())
