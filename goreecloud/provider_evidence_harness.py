# SPDX-License-Identifier: AGPL-3.0-or-later
"""Temporary candidate-bound real-provider evidence harness for RC #09.

This runner keeps the published candidate immutable. It reuses the established
provider cases, result parser, candidate identity contract, and sanitized evidence
writer while treating the application's real ``/healthz`` response as the runtime
health authority. Docker HEALTHCHECK metadata is not required because the RC #09
image does not define it.
"""

from __future__ import annotations

import argparse
import json
import pathlib
import sys
import urllib.request

from goreecloud import provider_acceptance


def verify_http_health(base_url: str, timeout: float) -> None:
    """Require the staged candidate's real HTTP health endpoint to return 200."""
    url = f"{base_url.rstrip('/')}/healthz"
    request = urllib.request.Request(
        url,
        headers={
            "User-Agent": "GoreeCloud-Search-Provider-Evidence/1.0",
            "Accept": "text/plain,*/*;q=0.1",
        },
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            status = response.status
    except Exception as exc:
        raise ValueError(f"provider evidence HTTP health check failed: {exc}") from exc
    if status != 200:
        raise ValueError(f"provider evidence HTTP health check returned {status}, expected 200")


def verify_runtime_identity(
    base_url: str,
    container: str,
    expected_source: str,
    expected_image: str,
    timeout: float,
) -> provider_acceptance.RuntimeIdentity:
    """Bind evidence to the exact running candidate and its HTTP health endpoint."""
    source, image = provider_acceptance._require_candidate_identity(  # pylint: disable=protected-access
        expected_source,
        expected_image,
    )
    if not container.strip():
        raise ValueError("--container is required for candidate-bound provider evidence")

    published_port = provider_acceptance._loopback_binding(  # pylint: disable=protected-access
        base_url,
        container,
    )
    running = provider_acceptance._docker_output(  # pylint: disable=protected-access
        "inspect",
        "-f",
        "{{.State.Running}}",
        container,
    )
    if running != "true":
        raise ValueError(f"provider evidence container {container!r} is not running")

    observed_ref = provider_acceptance._docker_output(  # pylint: disable=protected-access
        "inspect",
        "-f",
        "{{.Config.Image}}",
        container,
    )
    observed_id = provider_acceptance._docker_output(  # pylint: disable=protected-access
        "inspect",
        "-f",
        "{{.Image}}",
        container,
    )
    expected_id = provider_acceptance._docker_output(  # pylint: disable=protected-access
        "image",
        "inspect",
        "-f",
        "{{.Id}}",
        image,
    )
    if observed_ref != image:
        raise ValueError("provider evidence container image reference is not the immutable expected image")
    if observed_id != expected_id:
        raise ValueError("provider evidence container image ID does not match the expected digest image")

    def label(name: str) -> str:
        return provider_acceptance._docker_output(  # pylint: disable=protected-access
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

    verify_http_health(base_url, timeout)

    return provider_acceptance.RuntimeIdentity(
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


def _verify_unchanged_runtime(
    baseline: provider_acceptance.RuntimeIdentity,
    base_url: str,
    container: str,
    expected_source: str,
    expected_image: str,
    timeout: float,
) -> provider_acceptance.RuntimeIdentity:
    observed = verify_runtime_identity(
        base_url,
        container,
        expected_source,
        expected_image,
        timeout,
    )
    if observed != baseline:
        raise ValueError("provider evidence runtime identity changed during provider requests")
    return observed


def _write_http_bound_evidence(
    path: str,
    expected_source: str,
    expected_image: str,
    minimum_results: int,
    suite_code: int,
    results: list[provider_acceptance.AcceptanceResult],
    runtime_identity: provider_acceptance.RuntimeIdentity,
) -> None:
    provider_acceptance.write_evidence(
        path,
        expected_source,
        expected_image,
        minimum_results,
        suite_code,
        results,
        runtime_identity,
    )
    evidence_path = pathlib.Path(path)
    evidence = json.loads(evidence_path.read_text(encoding="utf-8"))
    evidence["runtime_binding"]["http_health_endpoint"] = "/healthz"
    evidence["runtime_binding"]["http_health_verified_before_and_after_each_request"] = True
    evidence["runtime_binding"]["docker_health_metadata_required"] = False
    evidence["scope"]["http_health_verified_during_provider_requests"] = True
    evidence["scope"]["statement"] = (
        "This artifact records sanitized real-provider acceptance against one loopback-staged, "
        "identity-verified GoreeCloud Search candidate. Exact image and OCI identity plus the real "
        "/healthz endpoint are verified before and after every provider request. Docker HEALTHCHECK "
        "metadata is not used as a substitute for application health. Provider availability can change "
        "after capture, and this artifact does not independently authorize production cutover."
    )
    evidence_path.write_text(
        json.dumps(evidence, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )


def run_candidate_bound_suite(
    base_url: str,
    container: str,
    expected_source: str,
    expected_image: str,
    minimum_results: int,
    timeout: float,
    evidence_json: str,
) -> int:
    """Run all representative provider cases with runtime binding around each request."""
    provider_acceptance.validate_representative_suite()
    baseline = verify_runtime_identity(
        base_url,
        container,
        expected_source,
        expected_image,
        timeout,
    )
    results: list[provider_acceptance.AcceptanceResult] = []

    for case in provider_acceptance.REPRESENTATIVE_SUITE:
        _verify_unchanged_runtime(
            baseline,
            base_url,
            container,
            expected_source,
            expected_image,
            timeout,
        )
        result = provider_acceptance.run_case(base_url, case, minimum_results, timeout)
        results.append(result)
        _verify_unchanged_runtime(
            baseline,
            base_url,
            container,
            expected_source,
            expected_image,
            timeout,
        )
        print()

    failures = [result for result in results if not result.passed]
    suite_code = 6 if failures else 0
    final_identity = _verify_unchanged_runtime(
        baseline,
        base_url,
        container,
        expected_source,
        expected_image,
        timeout,
    )
    _write_http_bound_evidence(
        evidence_json,
        expected_source,
        expected_image,
        minimum_results,
        suite_code,
        results,
        final_identity,
    )

    if failures:
        print("GoreeCloud Search representative provider suite did not fully pass:", file=sys.stderr)
        for result in failures:
            print(f"- {result.category}: exit code {result.exit_code}", file=sys.stderr)
        return suite_code

    print("GoreeCloud Search representative real-provider suite passed.")
    return 0


def main() -> int:
    """Run the temporary RC #09 candidate-bound provider evidence harness."""
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base-url", default="http://127.0.0.1:8888")
    parser.add_argument("--container", required=True)
    parser.add_argument("--expected-source", required=True)
    parser.add_argument("--expected-image", required=True)
    parser.add_argument("--minimum-results", type=int, default=1)
    parser.add_argument("--timeout", type=float, default=45.0)
    parser.add_argument("--evidence-json", required=True)
    args = parser.parse_args()

    try:
        return run_candidate_bound_suite(
            args.base_url,
            args.container,
            args.expected_source,
            args.expected_image,
            args.minimum_results,
            args.timeout,
            args.evidence_json,
        )
    except ValueError as exc:
        parser.error(str(exc))


if __name__ == "__main__":
    raise SystemExit(main())
