from __future__ import annotations

import argparse
import asyncio
from dataclasses import asdict
from datetime import date
from enum import Enum
import json
import os
from pathlib import Path
import stat
import sys
from typing import Any

from .brave_provider import BraveWebSearchProvider
from .http_api import (
    CONTAINER_SEARCH_HOST,
    DEFAULT_CONTAINER_SEARCH_PORT,
    DEFAULT_LOCAL_SEARCH_PORT,
    LOCAL_SEARCH_HOST,
    create_container_server,
    create_local_server,
)
from .models import SourceMode
from .query_parser import QueryParseError, parse_query
from .service import SearchCore
from .version import __version__


_MAX_RUNTIME_SECRET_BYTES = 16 * 1024


def _jsonable(value: Any) -> Any:
    if isinstance(value, Enum):
        return value.value
    if isinstance(value, date):
        return value.isoformat()
    if isinstance(value, tuple):
        return [_jsonable(item) for item in value]
    if isinstance(value, dict):
        return {key: _jsonable(item) for key, item in value.items()}
    if isinstance(value, list):
        return [_jsonable(item) for item in value]
    return value


def _add_provider_options(parser: argparse.ArgumentParser) -> None:
    parser.add_argument(
        "--country",
        default=None,
        help="Optional two-letter country code",
    )
    parser.add_argument(
        "--language",
        default=None,
        help="Optional Brave search language",
    )
    parser.add_argument(
        "--provider-safesearch",
        choices=("off", "moderate", "strict"),
        default="moderate",
        help="Filtering requested from the external provider; default: moderate",
    )


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="goreecloud-search",
        description="GoreeCloud Search development CLI",
    )
    parser.add_argument(
        "--version",
        action="version",
        version=__version__,
    )
    subparsers = parser.add_subparsers(
        dest="command",
        required=True,
    )

    parse_command = subparsers.add_parser(
        "parse",
        help="Parse a query without performing network access.",
    )
    parse_command.add_argument(
        "query",
        help="Query text to parse",
    )

    search_command = subparsers.add_parser(
        "search",
        help=(
            "Run an opt-in Brave Web Search through "
            "the GoreeCloud Search pipeline."
        ),
    )
    search_command.add_argument(
        "query",
        help="Query text",
    )
    search_command.add_argument(
        "--limit",
        type=int,
        default=10,
    )
    _add_provider_options(search_command)

    serve_command = subparsers.add_parser(
        "serve",
        help=(
            "Run the Development HTTP API on IPv4 loopback only. "
            "No user-facing web UI is included."
        ),
    )
    serve_command.add_argument(
        "--port",
        type=int,
        default=DEFAULT_LOCAL_SEARCH_PORT,
    )
    _add_provider_options(serve_command)

    container_command = subparsers.add_parser(
        "serve-container",
        help=(
            "Run the Development API on the container network for an approved "
            "reverse proxy. This command is not a host-publication instruction."
        ),
    )
    container_command.add_argument(
        "--port",
        type=int,
        default=DEFAULT_CONTAINER_SEARCH_PORT,
    )
    container_command.add_argument(
        "--service-hostname",
        default="search.goreecloud.com",
        help="Host header accepted from the reverse proxy",
    )
    _add_provider_options(container_command)
    return parser


def _runtime_secret(
    *,
    value_name: str,
    file_name: str,
) -> str:
    direct = os.environ.get(value_name, "").strip()
    path_text = os.environ.get(file_name, "").strip()
    if direct and path_text:
        raise ValueError(
            f"set only one of {value_name} or {file_name}"
        )
    if direct:
        return direct
    if not path_text:
        raise ValueError(
            f"{value_name} or {file_name} is required"
        )

    path = Path(path_text)
    try:
        metadata = path.stat()
    except OSError as exc:
        raise ValueError(
            f"{file_name} could not be read"
        ) from exc
    if not stat.S_ISREG(metadata.st_mode):
        raise ValueError(
            f"{file_name} must reference a regular file"
        )
    if metadata.st_size < 1 or metadata.st_size > _MAX_RUNTIME_SECRET_BYTES:
        raise ValueError(
            f"{file_name} has an invalid size"
        )
    try:
        secret = path.read_text(
            encoding="utf-8",
        ).strip()
    except (OSError, UnicodeError) as exc:
        raise ValueError(
            f"{file_name} could not be read"
        ) from exc
    if not secret:
        raise ValueError(
            f"{file_name} contains no usable value"
        )
    return secret


def _brave_provider(args: argparse.Namespace) -> BraveWebSearchProvider:
    api_key = _runtime_secret(
        value_name="BRAVE_SEARCH_API_KEY",
        file_name="BRAVE_SEARCH_API_KEY_FILE",
    )
    return BraveWebSearchProvider(
        api_key=api_key,
        country=args.country,
        search_lang=args.language,
        safesearch=args.provider_safesearch,
    )


def _brave_core(args: argparse.Namespace) -> SearchCore:
    return SearchCore(
        provider_adapters=(_brave_provider(args),),
    )


async def _run_brave_search(
    args: argparse.Namespace,
) -> int:
    if args.limit < 1:
        raise ValueError("--limit must be positive")

    response = await _brave_core(args).search(
        args.query,
        mode=SourceMode.EXTERNAL_ONLY,
        limit=args.limit,
    )
    print(
        json.dumps(
            _jsonable(asdict(response)),
            indent=2,
            sort_keys=True,
        )
    )
    return 0 if response.results else 3


def _validate_port(port: int) -> None:
    if port < 1 or port > 65535:
        raise ValueError("--port must be between 1 and 65535")


def _run_local_server(args: argparse.Namespace) -> int:
    _validate_port(args.port)
    server = create_local_server(
        _brave_core(args),
        port=args.port,
        mode=SourceMode.EXTERNAL_ONLY,
    )
    print(
        (
            "GoreeCloud Search Development API listening on "
            f"http://{LOCAL_SEARCH_HOST}:{server.server_port}"
        ),
        file=sys.stderr,
    )
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        return 0
    finally:
        server.server_close()
    return 0


def _run_container_server(args: argparse.Namespace) -> int:
    _validate_port(args.port)
    hostname = args.service_hostname.strip().casefold().rstrip(".")
    if not hostname or "/" in hostname or any(
        character.isspace() for character in hostname
    ):
        raise ValueError("--service-hostname is invalid")
    server = create_container_server(
        _brave_core(args),
        port=args.port,
        mode=SourceMode.EXTERNAL_ONLY,
        service_hostname=hostname,
    )
    print(
        (
            "GoreeCloud Search container API listening on "
            f"{CONTAINER_SEARCH_HOST}:{server.server_port}; "
            "backend host-port publication is not implied"
        ),
        file=sys.stderr,
    )
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        return 0
    finally:
        server.server_close()
    return 0


def main(
    argv: list[str] | None = None,
) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    if args.command == "parse":
        try:
            parsed = parse_query(args.query)
        except QueryParseError as exc:
            parser.error(str(exc))
        print(
            json.dumps(
                _jsonable(asdict(parsed)),
                indent=2,
                sort_keys=True,
            )
        )
        return 0
    if args.command == "search":
        try:
            return asyncio.run(
                _run_brave_search(args)
            )
        except (
            QueryParseError,
            ValueError,
        ) as exc:
            parser.error(str(exc))
    if args.command == "serve":
        try:
            return _run_local_server(args)
        except ValueError as exc:
            parser.error(str(exc))
    if args.command == "serve-container":
        try:
            return _run_container_server(args)
        except ValueError as exc:
            parser.error(str(exc))
    parser.error("unsupported command")
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
