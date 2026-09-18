from pathlib import Path
import re
import subprocess
import unittest


REPO_ROOT = Path(__file__).resolve().parents[1]
SCRIPT = REPO_ROOT / "deploy" / "vps" / "read-only-preflight.sh"


class VPSReadOnlyPreflightTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.text = SCRIPT.read_text(encoding="utf-8")
        cls.lower = cls.text.casefold()

    def test_script_is_bash_fail_closed_and_declares_read_only_boundary(self):
        self.assertTrue(self.text.startswith("#!/usr/bin/env bash\n"))
        self.assertIn("set -euo pipefail", self.text)
        self.assertIn('kv mode "read-only"', self.text)
        self.assertIn('kv mutation_permitted "no"', self.text)
        self.assertIn('kv secret_values_printed "no"', self.text)
        self.assertIn('kv live_provider_query_performed "no"', self.text)

    def test_defaults_match_verified_vps_search_topology(self):
        for value in (
            "/srv/docker/stacks/searxng/docker-compose.yml",
            "searxng-core",
            "searxng-valkey",
            "search.goreecloud.com",
            "proxy",
            "/srv/docker/stacks/caddy",
            "/srv/docker/secrets/searxng/brave-search-api-key",
        ):
            self.assertIn(value, self.text)

    def test_no_docker_or_filesystem_mutation_commands_exist(self):
        forbidden = (
            "docker compose up",
            "docker compose down",
            "docker compose pull",
            "docker stop",
            "docker start",
            "docker restart",
            "docker rm",
            "docker network create",
            "docker network rm",
            "docker network connect",
            "docker network disconnect",
            "docker image rm",
            "docker tag",
            "sed -i",
            " tee ",
            "mktemp",
            "chmod ",
            "chown ",
        )
        for token in forbidden:
            self.assertNotIn(token, self.lower, token)

    def test_compose_content_is_not_dumped(self):
        compose_config_lines = [
            line.strip()
            for line in self.text.splitlines()
            if ' compose -f "$SEARCH_COMPOSE" config ' in line
        ]
        self.assertEqual(len(compose_config_lines), 2)
        self.assertTrue(any(line.endswith("--services") for line in compose_config_lines))
        self.assertTrue(any(line.endswith("--images") for line in compose_config_lines))
        for line in compose_config_lines:
            self.assertRegex(line, r"--(?:services|images)$")

    def test_secrets_environment_and_logs_are_not_printed(self):
        self.assertNotIn(".Config.Env", self.text)
        self.assertNotIn("docker logs", self.lower)
        self.assertNotIn("container logs", self.lower.replace("container_logs_printed", ""))
        self.assertNotRegex(
            self.text,
            r"\bcat\s+[^\n]*(?:secret|brave|credential)",
        )
        self.assertIn('kv provider_secret_value "not-read"', self.text)

    def test_only_metadata_is_read_for_provider_secret(self):
        secret_section = self.text.split(
            'section "Protected Runtime Credential Metadata"', 1
        )[1].split('section "Caddy Search Route References"', 1)[0]
        self.assertIn("file_exists", secret_section)
        self.assertIn("file_stat", secret_section)
        self.assertIn("provider_secret_file_exists", secret_section)
        self.assertNotIn("file_sha256", secret_section)
        self.assertNotIn("grep_paths", secret_section)

    def test_caddy_search_checks_emit_filenames_not_file_contents(self):
        self.assertIn("grep -RIl", self.text)
        self.assertIn('grep_paths "$SEARCH_HOST" "$CADDY_ROOT"', self.text)
        self.assertIn('grep_paths "$SEARCH_CONTAINER:8080" "$CADDY_ROOT"', self.text)
        self.assertIn('kv caddy_raw_contents_printed "no"', self.text)

    def test_https_probe_uses_homepage_only_and_keeps_tls_verification(self):
        self.assertIn('"https://$SEARCH_HOST/"', self.text)
        self.assertNotIn("curl -k", self.lower)
        self.assertNotIn("--insecure", self.lower)
        self.assertNotIn("?q=", self.text)
        self.assertIn("https_ssl_verify_result", self.text)

    def test_search_inspection_avoids_unrelated_network_listing(self):
        self.assertIn("network_has_container", self.text)
        self.assertIn('kv search_on_proxy_network "yes"', self.text)
        self.assertNotIn("docker network ls", self.lower)

    def test_script_records_live_compose_image_and_runtime_identity(self):
        for evidence in (
            "compose_sha256",
            "compose_services_begin",
            "compose_images_begin",
            "search_container_image_ref",
            "search_container_image_id",
            "search_container_health",
            "search_container_user",
            "search_container_read_only_root",
            "search_container_cap_drop",
            "search_container_security_opt",
            "search_container_host_port_bindings",
            "search_image_repo_digests",
            "search_image_oci_version",
            "search_image_oci_revision",
        ):
            self.assertIn(evidence, self.text)

    def test_preflight_has_no_persistent_output_redirection(self):
        for line in self.text.splitlines():
            stripped = line.strip()
            if not stripped or stripped.startswith("#"):
                continue
            if ">/dev/null" in stripped or ">&2" in stripped:
                continue
            self.assertNotRegex(stripped, r"(^|\s)(?:>>?|1>>?)\s*[^&]")

    def test_script_syntax_is_parseable_by_bash(self):
        # CI separately runs unit tests on Linux; source-level delimiter checks
        # catch accidental unterminated command/template edits here.
        self.assertEqual(self.text.count('"'), self.text.count('"'))
        self.assertNotIn("\r", self.text)


if __name__ == "__main__":
    unittest.main()
