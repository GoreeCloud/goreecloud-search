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


class ResultCounter(html.parser.HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self.results = 0
        self.engine_messages = 0

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        classes = dict(attrs).get("class", "") or ""
        class_names = set(classes.split())
        if tag == "article" and "result" in class_names:
            self.results += 1
        if "engines_msg" in class_names:
            self.engine_messages += 1


def run(base_url: str, query: str, category: str, minimum_results: int, timeout: float) -> int:
    params = urllib.parse.urlencode(
        {
            "q": query,
            "categories": category,
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
        print(f"Provider acceptance request failed: {exc}", file=sys.stderr)
        return 2

    parser = ResultCounter()
    parser.feed(body)

    print(f"URL: {url}")
    print(f"HTTP status: {status}")
    print(f"Category: {category}")
    print(f"Result cards: {parser.results}")
    print(f"Engine-message surfaces: {parser.engine_messages}")

    if status != 200:
        print(f"Expected HTTP 200, received {status}.", file=sys.stderr)
        return 3
    if parser.results < minimum_results:
        print(
            f"Expected at least {minimum_results} result card(s), received {parser.results}. "
            "External engines may be unavailable, throttled, or blocked from this runner.",
            file=sys.stderr,
        )
        return 4

    print("GoreeCloud Search real-provider acceptance passed.")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base-url", default="http://127.0.0.1:8888")
    parser.add_argument("--query", default="GoreeCloud private search")
    parser.add_argument("--category", default="general")
    parser.add_argument("--minimum-results", type=int, default=1)
    parser.add_argument("--timeout", type=float, default=45.0)
    args = parser.parse_args()
    return run(args.base_url, args.query, args.category, args.minimum_results, args.timeout)


if __name__ == "__main__":
    raise SystemExit(main())
