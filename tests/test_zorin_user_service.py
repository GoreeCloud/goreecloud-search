from pathlib import Path
import subprocess
import unittest


REPO_ROOT = Path(__file__).resolve().parents[1]
DEPLOY = REPO_ROOT / "deploy" / "linux"


class ZorinUserServiceTests(unittest.TestCase):
    def test_unit_is_user_scoped_loopback_development_service(self):
        unit = (DEPLOY / "goreecloud-search.service").read_text()
        self.assertIn(
            "ExecStart=%h/.local/share/goreecloud/search/bin/"
            "goreecloud-search serve --port 8787",
            unit,
        )
        self.assertIn(
            "EnvironmentFile=%h/.config/goreecloud/search.env",
            unit,
        )
        self.assertNotIn("--host", unit)
        self.assertIn("NoNewPrivileges=true", unit)
        self.assertIn("ProtectSystem=strict", unit)
        self.assertIn("ProtectHome=read-only", unit)
        self.assertIn("UMask=0077", unit)

    def test_example_environment_contains_no_secret_value(self):
        env = (DEPLOY / "search.env.example").read_text()
        self.assertIn("BRAVE_SEARCH_API_KEY=\n", env)
        self.assertNotIn("sk-", env)

    def test_install_is_health_gated_and_does_not_modify_searxng(self):
        script = (DEPLOY / "install-user-service.sh").read_text()
        self.assertIn("http://127.0.0.1:8787/healthz", script)
        self.assertIn("requires Python 3.11 or newer", script)
        self.assertIn('cp -R "$REPO_ROOT/src/goreecloud_search"', script)
        self.assertNotIn("pip install", script)
        self.assertIn("systemctl --user enable --now goreecloud-search.service", script)
        self.assertNotIn("searxng.service", script.casefold())
        self.assertNotIn("docker stop", script.casefold())
        self.assertNotIn("docker compose down", script.casefold())

    def test_remove_keeps_configuration_without_explicit_purge(self):
        script = (DEPLOY / "remove-user-service.sh").read_text()
        self.assertIn("--purge", script)
        self.assertIn('if [ "$purge" = true ]', script)

    def test_shell_scripts_parse(self):
        for name in ("install-user-service.sh", "remove-user-service.sh"):
            result = subprocess.run(
                ["sh", "-n", str(DEPLOY / name)],
                check=False,
                capture_output=True,
                text=True,
            )
            self.assertEqual(result.returncode, 0, result.stderr)


if __name__ == "__main__":
    unittest.main()
