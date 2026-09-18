from pathlib import Path
import json
from tempfile import NamedTemporaryFile
from threading import Thread
import unittest
from unittest.mock import patch
from urllib.error import HTTPError
from urllib.request import Request, urlopen

from goreecloud_search import (
    ProviderDescriptor,
    ProviderOrigin,
    ProviderSearchBatch,
    ResultCandidate,
    SearchCategory,
    SearchCore,
    SourceMode,
    create_container_server,
)
from goreecloud_search.cli import _runtime_secret


REPO_ROOT = Path(__file__).resolve().parents[1]


class FakeProvider:
    def __init__(self) -> None:
        self._descriptor = ProviderDescriptor(
            name="fake-external",
            origin=ProviderOrigin.EXTERNAL,
            categories=frozenset({SearchCategory.GENERAL}),
        )

    @property
    def descriptor(self):
        return self._descriptor

    async def search(self, query, *, limit):
        return ProviderSearchBatch(
            candidates=(
                ResultCandidate(
                    title="GoreeCloud Search",
                    url="https://example.com/search",
                    snippet="Native search result",
                    provider=self.descriptor.name,
                ),
            )[:limit]
        )


class ContainerHTTPBoundaryTests(unittest.TestCase):
    def setUp(self) -> None:
        self.server = create_container_server(
            SearchCore(provider_adapters=(FakeProvider(),)),
            port=0,
            mode=SourceMode.EXTERNAL_ONLY,
            service_hostname="search.goreecloud.com",
        )
        self.thread = Thread(
            target=self.server.serve_forever,
            daemon=True,
        )
        self.thread.start()
        self.base = f"http://127.0.0.1:{self.server.server_port}"

    def tearDown(self) -> None:
        self.server.shutdown()
        self.server.server_close()
        self.thread.join(timeout=2)

    def test_container_listener_is_all_interfaces_with_bounded_hosts(self):
        self.assertEqual(self.server.server_address[0], "0.0.0.0")
        self.assertIn("search.goreecloud.com", self.server.allowed_hosts)
        self.assertIn("127.0.0.1", self.server.allowed_hosts)

    def test_container_health_accepts_loopback_host(self):
        with urlopen(f"{self.base}/healthz", timeout=2) as response:
            payload = json.load(response)
        self.assertEqual(response.status, 200)
        self.assertEqual(payload["status"], "ok")

    def test_container_accepts_verified_service_hostname(self):
        request = Request(
            f"{self.base}/healthz",
            headers={"Host": "search.goreecloud.com"},
        )
        with urlopen(request, timeout=2) as response:
            payload = json.load(response)
        self.assertEqual(response.status, 200)
        self.assertEqual(payload["service"], "goreecloud-search")

    def test_container_rejects_unapproved_host_header(self):
        request = Request(
            f"{self.base}/healthz",
            headers={"Host": "unapproved.example"},
        )
        with self.assertRaises(HTTPError) as context:
            urlopen(request, timeout=2)
        error = context.exception
        self.assertEqual(error.code, 421)
        payload = json.load(error)
        self.assertEqual(payload["error"], "misdirected_request")

    def test_mutating_method_also_rejects_unapproved_host(self):
        request = Request(
            f"{self.base}/api/v1/search",
            headers={"Host": "unapproved.example"},
            method="POST",
        )
        with self.assertRaises(HTTPError) as context:
            urlopen(request, timeout=2)
        error = context.exception
        self.assertEqual(error.code, 421)
        payload = json.load(error)
        self.assertEqual(payload["error"], "misdirected_request")


class RuntimeSecretTests(unittest.TestCase):
    def test_direct_environment_value_remains_supported_for_development(self):
        with patch.dict(
            "os.environ",
            {
                "BRAVE_SEARCH_API_KEY": "test-value-a",
                "BRAVE_SEARCH_API_KEY_FILE": "",
            },
            clear=False,
        ):
            value = _runtime_secret(
                value_name="BRAVE_SEARCH_API_KEY",
                file_name="BRAVE_SEARCH_API_KEY_FILE",
            )
        self.assertEqual(value, "test-value-a")

    def test_file_value_is_read_from_runtime_path(self):
        with NamedTemporaryFile("w", encoding="utf-8") as handle:
            handle.write("test-value-b\n")
            handle.flush()
            with patch.dict(
                "os.environ",
                {
                    "BRAVE_SEARCH_API_KEY": "",
                    "BRAVE_SEARCH_API_KEY_FILE": handle.name,
                },
                clear=False,
            ):
                value = _runtime_secret(
                    value_name="BRAVE_SEARCH_API_KEY",
                    file_name="BRAVE_SEARCH_API_KEY_FILE",
                )
        self.assertEqual(value, "test-value-b")

    def test_conflicting_runtime_sources_fail_closed(self):
        with NamedTemporaryFile("w", encoding="utf-8") as handle:
            handle.write("test-value-b\n")
            handle.flush()
            with patch.dict(
                "os.environ",
                {
                    "BRAVE_SEARCH_API_KEY": "test-value-a",
                    "BRAVE_SEARCH_API_KEY_FILE": handle.name,
                },
                clear=False,
            ):
                with self.assertRaisesRegex(ValueError, "set only one"):
                    _runtime_secret(
                        value_name="BRAVE_SEARCH_API_KEY",
                        file_name="BRAVE_SEARCH_API_KEY_FILE",
                    )


class VPSDeploymentArtifactTests(unittest.TestCase):
    def test_dockerfile_is_non_root_and_uses_container_boundary(self):
        dockerfile = (REPO_ROOT / "Dockerfile").read_text()
        self.assertIn("FROM python:3.12.14-slim-bookworm", dockerfile)
        self.assertIn("USER 10001:10001", dockerfile)
        self.assertIn("BRAVE_SEARCH_API_KEY_FILE=/run/secrets/", dockerfile)
        self.assertIn('"serve-container"', dockerfile)
        self.assertNotIn('"--host"', dockerfile)
        self.assertIn("HEALTHCHECK", dockerfile)

    def test_compose_example_preserves_private_proxy_topology(self):
        compose = (
            REPO_ROOT
            / "deploy"
            / "vps"
            / "docker-compose.search.example.yml"
        ).read_text()
        self.assertIn("container_name: searxng-core", compose)
        self.assertIn('expose:\n      - "8080"', compose)
        self.assertNotIn("\n    ports:", compose)
        self.assertIn("no-new-privileges:true", compose)
        self.assertIn("cap_drop:\n      - ALL", compose)
        self.assertIn("read_only: true", compose)
        self.assertIn("searxng-internal", compose)
        self.assertIn("proxy", compose)
        self.assertIn("/srv/docker/secrets/searxng/", compose)
        self.assertIn("tag@sha256 digest", compose)


if __name__ == "__main__":
    unittest.main()
