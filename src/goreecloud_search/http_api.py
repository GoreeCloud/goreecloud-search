from __future__ import annotations

import asyncio
from dataclasses import asdict
from datetime import date
from enum import Enum
from http.server import BaseHTTPRequestHandler, HTTPServer
import json
from pathlib import Path
from typing import Any, Iterable
from urllib.parse import parse_qs, urlsplit

from .execution import SearchAvailability
from .models import SourceMode
from .service import SearchCore, SearchResponse
from .version import __version__
from .web_ui import render_opensearch, render_search_page


LOCAL_SEARCH_HOST = "127.0.0.1"
DEFAULT_LOCAL_SEARCH_PORT = 8787
CONTAINER_SEARCH_HOST = "0.0.0.0"
DEFAULT_CONTAINER_SEARCH_PORT = 8080
MAX_HTTP_QUERY_CHARS = 2048
MAX_HTTP_REQUEST_TARGET_CHARS = 4096
MAX_HTTP_FORM_BODY_BYTES = 4096
MAX_HTTP_RESULT_LIMIT = 20
_SEARCH_CSS = (
    Path(__file__)
    .with_name("static")
    .joinpath("search.css")
    .read_bytes()
)


class LocalSearchAPIError(ValueError):
    """Raised when an HTTP search request cannot be accepted safely."""

    def __init__(
        self,
        message: str,
        *,
        code: str = "invalid_search_request",
        parameters: tuple[str, ...] = (),
    ) -> None:
        super().__init__(message)
        self.code = code
        self.parameters = parameters


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


def _normalize_host(value: str) -> str:
    raw = value.strip()
    if not raw:
        return ""
    if raw.startswith("["):
        closing = raw.find("]")
        if closing == -1:
            return ""
        return raw[1:closing].casefold().rstrip(".")
    return raw.rsplit(":", 1)[0].casefold().rstrip(".")


class SearchHTTPServer(HTTPServer):
    """Bounded HTTP boundary for one SearchCore instance."""

    advertised_origin: str

    def __init__(
        self,
        core: SearchCore,
        *,
        host: str,
        port: int,
        mode: SourceMode,
        allowed_hosts: Iterable[str],
        advertised_origin: str | None = None,
    ) -> None:
        if port < 0 or port > 65535:
            raise ValueError("port must be between 0 and 65535")
        normalized = frozenset(
            item.strip().casefold().rstrip(".")
            for item in allowed_hosts
            if item.strip()
        )
        if not normalized:
            raise ValueError("allowed_hosts must not be empty")
        self.search_core = core
        self.source_mode = mode
        self.allowed_hosts = normalized
        super().__init__((host, port), _SearchHandler)
        self.advertised_origin = (
            advertised_origin.rstrip("/")
            if advertised_origin is not None
            else f"http://{host}:{self.server_port}"
        )


class LocalSearchHTTPServer(SearchHTTPServer):
    """Loopback-only Development HTTP boundary."""

    def __init__(
        self,
        core: SearchCore,
        *,
        port: int = DEFAULT_LOCAL_SEARCH_PORT,
        mode: SourceMode = SourceMode.EXTERNAL_ONLY,
    ) -> None:
        super().__init__(
            core,
            host=LOCAL_SEARCH_HOST,
            port=port,
            mode=mode,
            allowed_hosts=("127.0.0.1", "localhost"),
        )
        self.advertised_origin = (
            f"http://{LOCAL_SEARCH_HOST}:{self.server_port}"
        )


class ContainerSearchHTTPServer(SearchHTTPServer):
    """Container-network HTTP boundary for reverse-proxy access."""

    def __init__(
        self,
        core: SearchCore,
        *,
        port: int = DEFAULT_CONTAINER_SEARCH_PORT,
        mode: SourceMode = SourceMode.EXTERNAL_ONLY,
        service_hostname: str = "search.goreecloud.com",
    ) -> None:
        super().__init__(
            core,
            host=CONTAINER_SEARCH_HOST,
            port=port,
            mode=mode,
            allowed_hosts=(
                "127.0.0.1",
                "localhost",
                service_hostname,
            ),
            advertised_origin=f"https://{service_hostname}",
        )


class _SearchHandler(BaseHTTPRequestHandler):
    server: SearchHTTPServer
    server_version = "GoreeCloudSearch"
    sys_version = ""

    def log_message(self, format: str, *args: object) -> None:
        # Never emit request targets because GET integration can place search
        # query text in the URL.
        del format, args

    @staticmethod
    def _security_headers(*, html: bool = False) -> dict[str, str]:
        csp = "default-src 'none'; frame-ancestors 'none'; base-uri 'none'"
        if html:
            csp += "; style-src 'self'; form-action 'self'"
        return {
            "Cache-Control": "no-store",
            "Content-Security-Policy": csp,
            "Cross-Origin-Opener-Policy": "same-origin",
            "Cross-Origin-Resource-Policy": "same-origin",
            "Permissions-Policy": (
                "accelerometer=(), camera=(), geolocation=(), "
                "microphone=(), payment=(), usb=()"
            ),
            "Referrer-Policy": "no-referrer",
            "X-Content-Type-Options": "nosniff",
            "X-Frame-Options": "DENY",
            "X-Robots-Tag": "noindex, nofollow, noarchive",
        }

    def _send_bytes(
        self,
        status: int,
        body: bytes,
        *,
        content_type: str,
        html: bool = False,
    ) -> None:
        self.send_response(status)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Connection", "close")
        for name, value in self._security_headers(html=html).items():
            self.send_header(name, value)
        self.end_headers()
        self.wfile.write(body)

    def _send_json(self, status: int, payload: dict[str, Any]) -> None:
        body = json.dumps(
            _jsonable(payload),
            ensure_ascii=False,
            separators=(",", ":"),
            sort_keys=True,
        ).encode("utf-8")
        self._send_bytes(
            status,
            body,
            content_type="application/json; charset=utf-8",
        )

    def _send_html(
        self,
        status: int,
        *,
        query: str = "",
        response: SearchResponse | None = None,
        error: str | None = None,
    ) -> None:
        self._send_bytes(
            status,
            render_search_page(
                query=query,
                response=response,
                error=error,
            ),
            content_type="text/html; charset=utf-8",
            html=True,
        )

    def _host_allowed(self) -> bool:
        host = self.headers.get("Host", "")
        return _normalize_host(host) in self.server.allowed_hosts

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

    def _parse_params(self, encoded: str) -> dict[str, list[str]]:
        try:
            return parse_qs(
                encoded,
                keep_blank_values=True,
                strict_parsing=False,
                max_num_fields=8,
            )
        except ValueError as exc:
            raise LocalSearchAPIError(
                "invalid query parameters",
                code="invalid_query_parameters",
            ) from exc

    def _search_parameters(
        self,
        params: dict[str, list[str]],
    ) -> tuple[str, int]:
        unknown = sorted(set(params) - {"q", "limit"})
        if unknown:
            raise LocalSearchAPIError(
                "unsupported query parameter(s): "
                + ", ".join(unknown),
                code="unsupported_query_parameter",
                parameters=tuple(unknown),
            )

        query = self._single_param(
            params,
            "q",
            required=True,
        )
        assert query is not None
        if len(query) > MAX_HTTP_QUERY_CHARS:
            raise LocalSearchAPIError(
                "q exceeds the local API size limit"
            )

        raw_limit = self._single_param(params, "limit")
        try:
            limit = (
                10
                if raw_limit is None or raw_limit == ""
                else int(raw_limit)
            )
        except ValueError as exc:
            raise LocalSearchAPIError(
                "limit must be an integer"
            ) from exc
        if limit < 1 or limit > MAX_HTTP_RESULT_LIMIT:
            raise LocalSearchAPIError(
                f"limit must be between 1 and {MAX_HTTP_RESULT_LIMIT}"
            )
        return query, limit

    def _execute_search(
        self,
        query: str,
        *,
        limit: int,
    ) -> SearchResponse:
        return asyncio.run(
            self.server.search_core.search(
                query,
                mode=self.server.source_mode,
                limit=limit,
            )
        )

    def _search_api(self, query_string: str) -> None:
        try:
            params = self._parse_params(query_string)
            query, limit = self._search_parameters(params)
        except LocalSearchAPIError as exc:
            payload: dict[str, Any] = {
                "error": exc.code,
                "message": str(exc),
            }
            if exc.parameters:
                payload["parameters"] = list(exc.parameters)
            self._send_json(400, payload)
            return

        try:
            response = self._execute_search(
                query,
                limit=limit,
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
            self._send_json(
                503,
                {"error": "search_unavailable"},
            )
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

    def _search_html_params(
        self,
        params: dict[str, list[str]],
    ) -> None:
        query = ""
        try:
            query, limit = self._search_parameters(params)
        except LocalSearchAPIError as exc:
            self._send_html(
                400,
                query=query,
                error=str(exc),
            )
            return

        try:
            response = self._execute_search(
                query,
                limit=limit,
            )
        except ValueError as exc:
            self._send_html(
                400,
                query=query,
                error=str(exc),
            )
            return
        except Exception:
            self._send_html(
                503,
                query=query,
                error="The Search service could not complete this request.",
            )
            return

        status = (
            503
            if response.execution.availability
            is SearchAvailability.UNAVAILABLE
            else 200
        )
        self._send_html(
            status,
            query=query,
            response=response,
        )

    def _search_html_query(self, query_string: str) -> None:
        try:
            params = self._parse_params(query_string)
        except LocalSearchAPIError as exc:
            self._send_html(400, error=str(exc))
            return
        self._search_html_params(params)

    def _search_html_form(self) -> None:
        content_type = self.headers.get("Content-Type", "")
        if not content_type.casefold().startswith(
            "application/x-www-form-urlencoded"
        ):
            self._send_html(
                415,
                error="Search form content type is not supported.",
            )
            return

        raw_length = self.headers.get("Content-Length", "")
        try:
            content_length = int(raw_length)
        except ValueError:
            self._send_html(
                411,
                error="Search form content length is required.",
            )
            return
        if (
            content_length < 1
            or content_length > MAX_HTTP_FORM_BODY_BYTES
        ):
            self._send_html(
                413,
                error="Search form request exceeds the size limit.",
            )
            return

        body = self.rfile.read(content_length)
        if len(body) != content_length:
            self._send_html(
                400,
                error="Search form request body is incomplete.",
            )
            return
        try:
            encoded = body.decode("utf-8")
        except UnicodeDecodeError:
            self._send_html(
                400,
                error="Search form request is not valid UTF-8.",
            )
            return
        try:
            params = self._parse_params(encoded)
        except LocalSearchAPIError as exc:
            self._send_html(400, error=str(exc))
            return
        self._search_html_params(params)

    def do_GET(self) -> None:
        if not self._host_allowed():
            self._send_json(421, {"error": "misdirected_request"})
            return

        if len(self.path) > MAX_HTTP_REQUEST_TARGET_CHARS:
            self._send_json(
                414,
                {"error": "request_target_too_large"},
            )
            return

        parsed = urlsplit(self.path)

        if parsed.path == "/healthz":
            if parsed.query:
                self._send_json(
                    400,
                    {
                        "error": (
                            "health_endpoint_accepts_no_parameters"
                        )
                    },
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
            self._search_api(parsed.query)
            return

        if parsed.path == "/":
            if parsed.query:
                self._search_html_query(parsed.query)
            else:
                self._send_html(200)
            return

        if parsed.path == "/search":
            self._search_html_query(parsed.query)
            return

        if parsed.path == "/assets/search.css":
            if parsed.query:
                self._send_json(
                    400,
                    {
                        "error": (
                            "asset_endpoint_accepts_no_parameters"
                        )
                    },
                )
                return
            self._send_bytes(
                200,
                _SEARCH_CSS,
                content_type="text/css; charset=utf-8",
            )
            return

        if parsed.path == "/opensearch.xml":
            if parsed.query:
                self._send_json(
                    400,
                    {
                        "error": (
                            "opensearch_endpoint_accepts_no_parameters"
                        )
                    },
                )
                return
            self._send_bytes(
                200,
                render_opensearch(
                    service_origin=self.server.advertised_origin
                ),
                content_type=(
                    "application/opensearchdescription+xml; charset=utf-8"
                ),
            )
            return

        self._send_json(
            404,
            {"error": "not_found"},
        )

    def _reject_misdirected(self) -> bool:
        if self._host_allowed():
            return False
        self._send_json(421, {"error": "misdirected_request"})
        return True

    def do_POST(self) -> None:
        if self._reject_misdirected():
            return

        parsed = urlsplit(self.path)
        if parsed.path == "/search" and not parsed.query:
            self._search_html_form()
            return

        self._send_json(
            405,
            {"error": "method_not_allowed"},
        )

    def do_PUT(self) -> None:
        if self._reject_misdirected():
            return
        self._send_json(
            405,
            {"error": "method_not_allowed"},
        )

    def do_DELETE(self) -> None:
        if self._reject_misdirected():
            return
        self._send_json(
            405,
            {"error": "method_not_allowed"},
        )


def create_local_server(
    core: SearchCore,
    *,
    port: int = DEFAULT_LOCAL_SEARCH_PORT,
    mode: SourceMode = SourceMode.EXTERNAL_ONLY,
) -> LocalSearchHTTPServer:
    """Create a server that is structurally fixed to IPv4 loopback."""

    return LocalSearchHTTPServer(
        core,
        port=port,
        mode=mode,
    )


def create_container_server(
    core: SearchCore,
    *,
    port: int = DEFAULT_CONTAINER_SEARCH_PORT,
    mode: SourceMode = SourceMode.EXTERNAL_ONLY,
    service_hostname: str = "search.goreecloud.com",
) -> ContainerSearchHTTPServer:
    """Create a Docker-network listener intended for Caddy reverse proxying."""

    return ContainerSearchHTTPServer(
        core,
        port=port,
        mode=mode,
        service_hostname=service_hostname,
    )
