from __future__ import annotations

import asyncio
from dataclasses import asdict
from datetime import date
from enum import Enum
from http.server import BaseHTTPRequestHandler, HTTPServer
import json
from typing import Any
from urllib.parse import parse_qs, urlsplit

from .models import SourceMode
from .service import SearchCore
from .version import __version__


LOCAL_SEARCH_HOST = "127.0.0.1"
DEFAULT_LOCAL_SEARCH_PORT = 8787
MAX_HTTP_QUERY_CHARS = 2048
MAX_HTTP_REQUEST_TARGET_CHARS = 4096
MAX_HTTP_RESULT_LIMIT = 20


class LocalSearchAPIError(ValueError):
    """Raised when a local API request cannot be accepted safely."""


def _jsonable(value: Any) -> Any:
    if isinstance(value, Enum):
        return value.value
    if isinstance(value, date):
        return value.isoformat()
    if isinstance(value, tuple):
        return [_jsonable(item) for item in value]
    if isinstance(value, list):
        return [_jsonable(item) for item in value]
    if isinstance(value, dict):
        return {str(key): _jsonable(item) for key, item in value.items()}
    return value


class LocalSearchHTTPServer(HTTPServer):
    """Loopback-only Development HTTP boundary for one SearchCore instance."""

    def __init__(
        self,
        core: SearchCore,
        *,
        port: int = DEFAULT_LOCAL_SEARCH_PORT,
        mode: SourceMode = SourceMode.EXTERNAL_ONLY,
    ) -> None:
        if port < 0 or port > 65535:
            raise ValueError("port must be between 0 and 65535")
        self.search_core = core
        self.source_mode = mode
        super().__init__((LOCAL_SEARCH_HOST, port), _LocalSearchHandler)


class _LocalSearchHandler(BaseHTTPRequestHandler):
    server: LocalSearchHTTPServer
    server_version = "GoreeCloudSearch"
    sys_version = ""

    def log_message(self, format: str, *args: object) -> None:
        # Do not emit request targets because search query text is part of the URL.
        del format, args

    def _security_headers(self) -> dict[str, str]:
        return {
            "Cache-Control": "no-store",
            "Content-Security-Policy": (
                "default-src 'none'; frame-ancestors 'none'; base-uri 'none'"
            ),
            "Referrer-Policy": "no-referrer",
            "X-Content-Type-Options": "nosniff",
            "X-Frame-Options": "DENY",
        }

    def _send_json(self, status: int, payload: dict[str, Any]) -> None:
        body = json.dumps(
            _jsonable(payload),
            ensure_ascii=False,
            separators=(",", ":"),
            sort_keys=True,
        ).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Connection", "close")
        for name, value in self._security_headers().items():
            self.send_header(name, value)
        self.end_headers()
        self.wfile.write(body)

    def _single_param(
        self,
        params: dict[str, list[str]],
        name: str,
        *,
        required: bool = False,
    ) -> str | None:
        values = params.get(name)
        if not values:
            if required:
                raise LocalSearchAPIError(f"{name} is required")
            return None
        if len(values) != 1:
            raise LocalSearchAPIError(f"{name} must be supplied once")
        value = values[0].strip()
        if required and not value:
            raise LocalSearchAPIError(f"{name} must not be empty")
        return value

    def _search(self, query_string: str) -> None:
        try:
            params = parse_qs(
                query_string,
                keep_blank_values=True,
                strict_parsing=False,
                max_num_fields=8,
            )
        except ValueError:
            self._send_json(400, {"error": "invalid_query_parameters"})
            return

        unknown = sorted(set(params) - {"q", "limit"})
        if unknown:
            self._send_json(
                400,
                {
                    "error": "unsupported_query_parameter",
                    "parameters": unknown,
                },
            )
            return

        try:
            query = self._single_param(params, "q", required=True)
            assert query is not None
            if len(query) > MAX_HTTP_QUERY_CHARS:
                raise LocalSearchAPIError("q exceeds the local API size limit")
            raw_limit = self._single_param(params, "limit")
            limit = 10 if raw_limit is None or raw_limit == "" else int(raw_limit)
            if limit < 1 or limit > MAX_HTTP_RESULT_LIMIT:
                raise LocalSearchAPIError(
                    f"limit must be between 1 and {MAX_HTTP_RESULT_LIMIT}"
                )
        except (LocalSearchAPIError, ValueError) as exc:
            self._send_json(
                400,
                {
                    "error": "invalid_search_request",
                    "message": str(exc),
                },
            )
            return

        try:
            response = asyncio.run(
                self.server.search_core.search(
                    query,
                    mode=self.server.source_mode,
                    limit=limit,
                )
            )
        except ValueError as exc:
            self._send_json(
                400,
                {
                    "error": "search_rejected",
                    "message": str(exc),
                },
            )
            return
        except Exception:
            self._send_json(503, {"error": "search_unavailable"})
            return

        self._send_json(
            200,
            {
                "api_version": 1,
                "service": "goreecloud-search",
                "version": __version__,
                "response": asdict(response),
            },
        )

    def do_GET(self) -> None:
        if len(self.path) > MAX_HTTP_REQUEST_TARGET_CHARS:
            self._send_json(414, {"error": "request_target_too_large"})
            return

        parsed = urlsplit(self.path)
        if parsed.path == "/healthz":
            if parsed.query:
                self._send_json(
                    400,
                    {"error": "health_endpoint_accepts_no_parameters"},
                )
                return
            self._send_json(
                200,
                {
                    "status": "ok",
                    "service": "goreecloud-search",
                    "version": __version__,
                    "lifecycle": "development",
                },
            )
            return

        if parsed.path == "/api/v1/search":
            self._search(parsed.query)
            return

        self._send_json(404, {"error": "not_found"})

    def do_POST(self) -> None:
        self._send_json(405, {"error": "method_not_allowed"})

    def do_PUT(self) -> None:
        self._send_json(405, {"error": "method_not_allowed"})

    def do_DELETE(self) -> None:
        self._send_json(405, {"error": "method_not_allowed"})


def create_local_server(
    core: SearchCore,
    *,
    port: int = DEFAULT_LOCAL_SEARCH_PORT,
    mode: SourceMode = SourceMode.EXTERNAL_ONLY,
) -> LocalSearchHTTPServer:
    """Create a server that is structurally fixed to IPv4 loopback."""

    return LocalSearchHTTPServer(core, port=port, mode=mode)
