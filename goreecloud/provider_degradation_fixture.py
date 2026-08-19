# SPDX-License-Identifier: AGPL-3.0-or-later
"""Deterministic local provider fixture for GoreeCloud Search acceptance."""

from __future__ import annotations

import argparse
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse


class FixtureHandler(BaseHTTPRequestHandler):
    server_version = "GoreeCloudProviderFixture/1.0"

    def do_GET(self) -> None:  # noqa: N802 - BaseHTTPRequestHandler API
        parsed = urlparse(self.path)
        if parsed.path == "/healthz":
            self._write_json({"status": "ok"})
            return
        if parsed.path != "/search":
            self.send_error(404)
            return

        query = parse_qs(parsed.query).get("q", [""])[0]
        self._write_json(
            {
                "results": [
                    {
                        "url": "https://example.invalid/goreecloud-provider-healthy",
                        "title": "GoreeCloud provider degradation healthy result",
                        "content": (
                            "Healthy provider result remained usable while a sibling provider failed. "
                            f"Query: {query}"
                        ),
                    }
                ]
            }
        )

    def log_message(self, format: str, *args: object) -> None:
        return

    def _write_json(self, payload: dict[str, object]) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bind", default="127.0.0.1")
    parser.add_argument("--port", type=int, default=8899)
    args = parser.parse_args()

    server = ThreadingHTTPServer((args.bind, args.port), FixtureHandler)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        server.server_close()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
