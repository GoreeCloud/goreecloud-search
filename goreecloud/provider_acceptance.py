# SPDX-License-Identifier: AGPL-3.0-or-later
"""Manual real-provider acceptance for GoreeCloud Search.

This check is deliberately separate from normal pull-request CI because external
search providers can throttle or block shared CI runners. It is intended for
explicit acceptance runs and target-environment testing. When requested, the
suite writes sanitized candidate-bound JSON evidence without storing query text,
response content, cookies, credentials, or provider tokens.
"""

from __future__ import annotations

import argparse
import datetime as dt
import html.parser
import json
import pathlib
import re
import sys
import urllib.parse
import urllib.request
from dataclasses import asdict, dataclass


SHA_RE = re.compile(r"^[0-9a-f]{40}$")
IMAGE_RE = re.compile(r"^ghcr\.io/goreecloud/goreecloud-search@sha256:[0-9a-f]{64}$")


@dataclass(frozen=True)
class AcceptanceCase:
    category: str
    query: str


@dataclass(frozen=True)
class AcceptanceResult:
    category: str
    exit_code: int
    http_status: int | None
    product_identity: bool
    result_cards: int
    engine_messages: int
    passed: bool


# These categories are mandatory for first-Stable representative provider
# acceptance because they cover the primary GoreeCloud Search result surfaces
# that received product-specific stabilization work.
RELEASE_REQUIRED_CATEGORIES = frozenset(
    {
        "general",
        "images",
        "videos",
        "news",
        "files",
    }
)


REPRESENTATIVE_SUITE = (
    AcceptanceCase("general", "GoreeCloud private search"),
    AcceptanceCase("images", "open source personal cloud"),
    AcceptanceCase("news", "privacy technology"),
    AcceptanceCase("videos", "self hosted search"),
    AcceptanceCase("files", "open source privacy whitepaper pdf"),
    AcceptanceCase("it", "Python metasearch engine"),
    AcceptanceCase("science", "information retrieval privacy research"),
)


def validate_representative_suite() -> None:
    categories = {case.category for case in REPRESENTATIVE_SUITE}
    missing = sorted(RELEASE_REQUIRED_CATEGORIES - categories)
    if missing:
        raise RuntimeError(
            "Representative provider suite is missing first-Stable category coverage: "
            + ", ".join(missing)
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


def _result(
    case: AcceptanceCase,
    code: int,
    *,
    status: int | None = None,
    product_identity: bool = False,
    result_cards: int = 0,
    engine_messages: int = 0,
) -> AcceptanceResult:
    return AcceptanceResult(
        category=case.category,
        exit_code=code,
        http_status=status,
        product_identity=product_identity,
        result_cards=result_cards,
        engine_messages=engine_messages,
        passed=code == 0,
    )


def run_case(base_url: str, case: AcceptanceCase, minimum_results: int, timeout: float) -> AcceptanceResult:
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
            "User-Agent": "GoreeCloud-Search-Provider-Acceptance/1.1",
            "Accept": "text/html,application/xhtml+xml",
        },
    )

    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            status = response.status
            body = response.read().decode("utf-8", errors="replace")
    except Exception as exc:  # explicit acceptance command should explain external failures
        print(f"[{case.category}] Provider acceptance request failed: {exc}", file=sys.stderr)
        return _result(case, 2)

    parser = ResultCounter()
    parser.feed(body)

    print(f"[{case.category}] URL: {url}")
    print(f"[{case.category}] HTTP status: {status}")
    print(f"[{case.category}] GoreeCloud Search identity: {'yes' if parser.product_identity else 'no'}")
    print(f"[{case.category}] Result cards: {parser.results}")
    print(f"[{case.category}] Engine-message surfaces: {parser.engine_messages}")

    if status != 200:
        print(f"[{case.category}] Expected HTTP 200, received {status}.", file=sys.stderr)
        return _result(
            case,
            3,
            status=status,
            product_identity=parser.product_identity,
            result_cards=parser.results,
            engine_messages=parser.engine_messages,
        )
    if not parser.product_identity:
        print(
            f"[{case.category}] Search response lost the GoreeCloud Search product identity.",
            file=sys.stderr,
        )
        return _result(
            case,
            4,
            status=status,
            product_identity=False,
            result_cards=parser.results,
            engine_messages=parser.engine_messages,
        )
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
        return _result(
            case,
            5,
            status=status,
            product_identity=True,
            result_cards=parser.results,
            engine_messages=parser.engine_messages,
        )

    print(f"[{case.category}] Real-provider acceptance passed.")
    return _result(
        case,
        0,
        status=status,
        product_identity=True,
        result_cards=parser.results,
        engine_messages=parser.engine_messages,
    )


def run_suite(base_url: str, minimum_results: int, timeout: float) -> tuple[int, list[AcceptanceResult]]:
    validate_representative_suite()
    results: list[AcceptanceResult] = []
    for case in REPRESENTATIVE_SUITE:
        result = run_case(base_url, case, minimum_results, timeout)
        results.append(result)
        print()

    failures = [result for result in results if not result.passed]
    if failures:
        print("GoreeCloud Search representative provider suite did not fully pass:", file=sys.stderr)
        for result in failures:
            print(f"- {result.category}: exit code {result.exit_code}", file=sys.stderr)
        print(
            "Classify each failure as application/runtime, provider throttling/blocking, engine initialization, "
            "or genuinely empty results before deciding whether the release candidate is defective.",
            file=sys.stderr,
        )
        return 6, results

    print("GoreeCloud Search representative real-provider suite passed.")
    return 0, results


def _require_candidate_identity(source: str, image: str) -> tuple[str, str]:
    if not SHA_RE.fullmatch(source):
        raise ValueError("--expected-source must be a lowercase 40-character Git SHA")
    if not IMAGE_RE.fullmatch(image):
        raise ValueError(
            "--expected-image must be ghcr.io/goreecloud/goreecloud-search pinned by sha256 digest"
        )
    return source, image


def write_evidence(
    path: str,
    expected_source: str,
    expected_image: str,
    minimum_results: int,
    suite_code: int,
    results: list[AcceptanceResult],
) -> None:
    source, image = _require_candidate_identity(expected_source, expected_image)
    result_by_category = {result.category: result for result in results}
    required_passed = all(
        category in result_by_category and result_by_category[category].passed
        for category in RELEASE_REQUIRED_CATEGORIES
    )
    evidence = {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "generated_at": dt.datetime.now(dt.timezone.utc).isoformat().replace("+00:00", "Z"),
        "candidate": {
            "source_revision": source,
            "image": image,
        },
        "minimum_results": minimum_results,
        "required_categories": sorted(RELEASE_REQUIRED_CATEGORIES),
        "results": [asdict(result) for result in results],
        "scope": {
            "real_provider_requests_performed": True,
            "all_required_categories_passed": required_passed,
            "full_diagnostic_suite_passed": suite_code == 0,
            "query_text_persisted": False,
            "response_content_persisted": False,
            "production_cutover_authorized": False,
            "statement": (
                "This artifact records sanitized real-provider acceptance for one exact GoreeCloud Search "
                "candidate. Provider availability can change after capture, and this artifact does not "
                "independently authorize production cutover."
            ),
        },
    }
    pathlib.Path(path).write_text(json.dumps(evidence, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def main() -> int:
    validate_representative_suite()
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base-url", default="http://127.0.0.1:8888")
    parser.add_argument("--query", default="GoreeCloud private search")
    parser.add_argument("--category", default="general")
    parser.add_argument("--minimum-results", type=int, default=1)
    parser.add_argument("--timeout", type=float, default=45.0)
    parser.add_argument(
        "--suite",
        action="store_true",
        help=(
            "Run the representative general/images/news/videos/files/IT/science "
            "real-provider acceptance suite."
        ),
    )
    parser.add_argument("--expected-source", default="")
    parser.add_argument("--expected-image", default="")
    parser.add_argument(
        "--evidence-json",
        default="",
        help="Write sanitized candidate-bound suite evidence to this JSON path.",
    )
    args = parser.parse_args()

    if args.evidence_json and not args.suite:
        parser.error("--evidence-json requires --suite")
    if args.evidence_json and (not args.expected_source or not args.expected_image):
        parser.error("--evidence-json requires --expected-source and --expected-image")

    if args.suite:
        code, results = run_suite(args.base_url, args.minimum_results, args.timeout)
        if args.evidence_json:
            try:
                write_evidence(
                    args.evidence_json,
                    args.expected_source,
                    args.expected_image,
                    args.minimum_results,
                    code,
                    results,
                )
            except ValueError as exc:
                parser.error(str(exc))
        return code

    result = run_case(
        args.base_url,
        AcceptanceCase(args.category, args.query),
        args.minimum_results,
        args.timeout,
    )
    return result.exit_code


if __name__ == "__main__":
    raise SystemExit(main())
