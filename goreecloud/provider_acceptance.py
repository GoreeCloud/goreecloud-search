# SPDX-License-Identifier: AGPL-3.0-or-later
"""Manual real-provider acceptance for GoreeCloud Search.

This check is deliberately separate from normal pull-request CI because external
search providers can throttle or block shared CI runners. It is intended for
explicit acceptance runs and target-environment testing. When requested, the
suite writes sanitized candidate-bound JSON evidence without storing query text,
response content, cookies, credentials, or provider tokens.

Candidate-bound evidence is fail-closed: the provider requests must be sent to a
loopback-published staged container whose running image, image ID, and OCI
revision match the exact candidate before and after the provider suite.
"""

from __future__ import annotations

import argparse
import datetime as dt
import html.parser
import json
import pathlib
import re
import subprocess
import sys
import urllib.parse
import urllib.request
from dataclasses import asdict, dataclass


SHA_RE = re.compile(r"^[0-9a-f]{40}$")
IMAGE_RE = re.compile(r"^ghcr\.io/goreecloud/goreecloud-search@sha256:[0-9a-f]{64}$")
LOOPBACK_HOSTS = frozenset({"127.0.0.1", "localhost", "::1"})


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


@dataclass(frozen=True)
class RuntimeIdentity:
    container: str
    base_url: str
    published_port: str
    image_reference: str
    image_id: str
    oci_title: str
    oci_source: str
    oci_revision: str
    oci_version: str
    oci_licenses: str


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
    """Require every first-Stable provider category in the representative suite."""
    categories = {case.category for case in REPRESENTATIVE_SUITE}
    missing = sorted(RELEASE_REQUIRED_CATEGORIES - categories)
    if missing:
        raise RuntimeError(
            "Representative provider suite is missing first-Stable category coverage: "
            + ", ".join(missing)
        )


class ResultCounter(html.parser.HTMLParser):
    """Count result and provider-message surfaces while checking product identity."""

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


def run_case(
    base_url: str,
    case: AcceptanceCase,
    minimum_results: int,
    timeout: float,
) -> AcceptanceResult:
    """Run one representative provider request and return sanitized result metadata."""
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
            "User-Agent": "GoreeCloud-Search-Provider-Acceptance/1.2",
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


def run_suite(
    base_url: str,
    minimum_results: int,
    timeout: float,
) -> tuple[int, list[AcceptanceResult]]:
    """Run the complete representative suite and retain sanitized result metadata."""
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


def _docker_output(*args: str) -> str:
    try:
        result = subprocess.run(
            ["docker", *args],
            check=True,
            capture_output=True,
            text=True,
        )
    except FileNotFoundError as exc:
        raise ValueError("Docker is required for candidate-bound provider evidence") from exc
    except subprocess.CalledProcessError as exc:
        detail = (exc.stderr or exc.stdout or "docker command failed").strip()
        raise ValueError(f"Docker runtime identity check failed: {detail}") from exc
    return result.stdout.strip()


def _loopback_binding(base_url: str, container: str) -> str:
    parsed = urllib.parse.urlparse(base_url)
    if parsed.scheme not in {"http", "https"} or parsed.hostname not in LOOPBACK_HOSTS:
        raise ValueError(
            "candidate-bound provider evidence requires a loopback staging base URL "
            "(127.0.0.1, localhost, or ::1)"
        )
    port = parsed.port or (443 if parsed.scheme == "https" else 80)
    published = _docker_output("port", container)
    accepted = {f"127.0.0.1:{port}", f"[::1]:{port}"}
    matching_lines = [
        line.strip()
        for line in published.splitlines()
        if any(item in line for item in accepted)
    ]
    if not matching_lines:
        raise ValueError(
            f"container {container!r} is not loopback-published on the provider evidence port {port}"
        )
    return matching_lines[0]


def verify_runtime_identity(
    base_url: str,
    container: str,
    expected_source: str,
    expected_image: str,
) -> RuntimeIdentity:
    """Bind provider evidence to the exact healthy staged candidate runtime."""
    source, image = _require_candidate_identity(expected_source, expected_image)
    if not container.strip():
        raise ValueError("--container is required for candidate-bound provider evidence")

    published_port = _loopback_binding(base_url, container)
    running = _docker_output("inspect", "-f", "{{.State.Running}}", container)
    health = _docker_output(
        "inspect",
        "-f",
        "{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}",
        container,
    )
    if running != "true":
        raise ValueError(f"provider evidence container {container!r} is not running")
    if health != "healthy":
        raise ValueError(f"provider evidence container {container!r} is not healthy")

    observed_ref = _docker_output("inspect", "-f", "{{.Config.Image}}", container)
    observed_id = _docker_output("inspect", "-f", "{{.Image}}", container)
    expected_id = _docker_output("image", "inspect", "-f", "{{.Id}}", image)
    if observed_ref != image:
        raise ValueError("provider evidence container image reference is not the immutable expected image")
    if observed_id != expected_id:
        raise ValueError("provider evidence container image ID does not match the expected digest image")

    def label(name: str) -> str:
        return _docker_output(
            "image",
            "inspect",
            "-f",
            f'{{{{ index .Config.Labels "{name}" }}}}',
            image,
        )

    title = label("org.opencontainers.image.title")
    source_url = label("org.opencontainers.image.source")
    revision = label("org.opencontainers.image.revision")
    version = label("org.opencontainers.image.version")
    licenses = label("org.opencontainers.image.licenses")

    if title != "GoreeCloud Search":
        raise ValueError("provider evidence candidate OCI title is not GoreeCloud Search")
    if source_url != "https://github.com/GoreeCloud/goreecloud-search":
        raise ValueError("provider evidence candidate OCI source is not the GoreeCloud Search repository")
    if revision != source:
        raise ValueError("provider evidence candidate OCI revision does not match the expected source")
    if not version:
        raise ValueError("provider evidence candidate OCI version is empty")
    if licenses != "AGPL-3.0-or-later":
        raise ValueError("provider evidence candidate OCI license is not AGPL-3.0-or-later")

    return RuntimeIdentity(
        container=container,
        base_url=base_url.rstrip("/"),
        published_port=published_port,
        image_reference=observed_ref,
        image_id=observed_id,
        oci_title=title,
        oci_source=source_url,
        oci_revision=revision,
        oci_version=version,
        oci_licenses=licenses,
    )


def write_evidence(
    path: str,
    expected_source: str,
    expected_image: str,
    minimum_results: int,
    suite_code: int,
    results: list[AcceptanceResult],
    runtime_identity: RuntimeIdentity,
) -> None:
    """Write sanitized candidate-bound provider evidence."""
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
        "runtime_binding": {
            "verified_before_and_after_requests": True,
            "base_url": runtime_identity.base_url,
            "container": runtime_identity.container,
            "published_port": runtime_identity.published_port,
            "observed_image_reference": runtime_identity.image_reference,
            "observed_image_id": runtime_identity.image_id,
            "oci_revision": runtime_identity.oci_revision,
        },
        "minimum_results": minimum_results,
        "required_categories": sorted(RELEASE_REQUIRED_CATEGORIES),
        "results": [asdict(result) for result in results],
        "scope": {
            "real_provider_requests_performed": True,
            "runtime_identity_verified_during_provider_requests": True,
            "all_required_categories_passed": required_passed,
            "full_diagnostic_suite_passed": suite_code == 0,
            "query_text_persisted": False,
            "response_content_persisted": False,
            "production_cutover_authorized": False,
            "statement": (
                "This artifact records sanitized real-provider acceptance against one loopback-staged, "
                "identity-verified GoreeCloud Search candidate. Runtime identity is checked before and after "
                "the provider requests. Provider availability can change after capture, and this artifact "
                "does not independently authorize production cutover."
            ),
        },
    }
    pathlib.Path(path).write_text(
        json.dumps(evidence, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )


def main() -> int:
    """Run one diagnostic request or the candidate-bound representative suite."""
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
        "--container",
        default="",
        help="Running loopback-staged Docker container to bind candidate provider evidence to.",
    )
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
    if args.evidence_json and not args.container:
        parser.error("--evidence-json requires --container for runtime identity binding")

    if args.suite:
        runtime_before: RuntimeIdentity | None = None
        if args.evidence_json:
            try:
                runtime_before = verify_runtime_identity(
                    args.base_url,
                    args.container,
                    args.expected_source,
                    args.expected_image,
                )
            except ValueError as exc:
                parser.error(str(exc))

        code, results = run_suite(args.base_url, args.minimum_results, args.timeout)

        if args.evidence_json:
            assert runtime_before is not None
            try:
                runtime_after = verify_runtime_identity(
                    args.base_url,
                    args.container,
                    args.expected_source,
                    args.expected_image,
                )
            except ValueError as exc:
                parser.error(str(exc))
            if runtime_after != runtime_before:
                parser.error("provider evidence runtime identity changed during the provider suite")
            try:
                write_evidence(
                    args.evidence_json,
                    args.expected_source,
                    args.expected_image,
                    args.minimum_results,
                    code,
                    results,
                    runtime_after,
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
