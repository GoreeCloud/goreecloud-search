from __future__ import annotations

import argparse
import asyncio
from dataclasses import asdict
from datetime import date
from enum import Enum
import json
import os
from pathlib import Path
import sys
from typing import Any

from .brave_provider import BraveWebSearchProvider
from .http_api import (
    CONTAINER_SEARCH_HOST,
    DEFAULT_LOCAL_SEARCH_PORT,
    LOCAL_SEARCH_HOST,
    create_search_server,
)
from .models import SourceMode
from .query_parser import QueryParseError, parse_query
from .service import SearchCore
from .version import __version__


_MAX_SECRET_FILE_BYTES = 16 * 1024


def _jsonable(value: Any) -> Any:
    if isinstance(value, Enum):
        return value.value
    if isinstance(value, date):
        return value.isoformat()
    if isinstance(value, tuple):
        return [_jsonable(item) for item in value]
    if isinstance(value, dict):
        return {
            key: _jsonable(item)
            for key, item in value.items()
        }
    if isinstance(value, list):
        return [_jsonable(item) for item in value]
    return value


def _add_provider_options(
    parser: argparse.ArgumentParser,
) -> None:
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
        help=(
            "Filtering requested from the external "
            "provider; default: moderate"
        ),
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
        help="Parse a query without network access.",
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
            "Run the Development HTTP/API and web UI. "
            "Loopback is the default; 0.0.0.0 is allowed "
            "only for a container/network-isolated deployment."
        ),
    )
    serve_command.add_argument(
        "--host",
        choices=(
            LOCAL_SEARCH_HOST,
            CONTAINER_SEARCH_HOST,
        ),
        default=LOCAL_SEARCH_HOST,
    )
    serve_command.add_argument(
        "--port",
        type=int,
        default=DEFAULT_LOCAL_SEARCH_PORT,
    )
    serve_command.add_argument(
        "--public-base-url",
        default=os.environ.get(
            "GOREECLOUD_SEARCH_PUBLIC_BASE_URL"
        ),
        help=(
            "Canonical HTTP(S) origin used for OpenSearch "
            "and UI identity; required with --host 0.0.0.0."
        ),
    )
    _add_provider_options(serve_command)
    return parser


def _provider_api_key() -> str:
    env_key = os.environ.get(
        "BRAVE_SEARCH_API_KEY",
        "",
    ).strip()
    secret_file = os.environ.get(
        "BRAVE_SEARCH_API_KEY_FILE",
        "",
    ).strip()

    if env_key and secret_file:
        raise ValueError(
            "configure either BRAVE_SEARCH_API_KEY or "
            "BRAVE_SEARCH_API_KEY_FILE, not both"
        )
    if env_key:
        return env_key
    if not secret_file:
        raise ValueError(
            "Brave provider credential is required"
        )

    path = Path(secret_file)
    try:
        if path.is_symlink():
            raise ValueError(
                "Brave provider credential file "
                "must not be a symbolic link"
            )
        stat = path.stat()
        if (
            not path.is_file()
            or stat.st_size < 1
            or stat.st_size > _MAX_SECRET_FILE_BYTES
        ):
            raise ValueError(
                "Brave provider credential file "
                "has an invalid size or type"
            )
        value = path.read_text(
            encoding="utf-8"
        ).strip()
    except OSError as exc:
        raise ValueError(
            "Brave provider credential file "
            "could not be read"
        ) from exc
    if not value:
        raise ValueError(
            "Brave provider credential file is empty"
        )
    return value


def _brave_provider(
    args: argparse.Namespace,
) -> BraveWebSearchProvider:
    return BraveWebSearchProvider(
        api_key=_provider_api_key(),
        country=args.country,
        search_lang=args.language,
        safesearch=args.provider_safesearch,
    )


def _brave_core(
    args: argparse.Namespace,
) -> SearchCore:
    return SearchCore(
        provider_adapters=(
            _brave_provider(args),
        ),
    )


async def _run_brave_search(
    args: argparse.Namespace,
) -> int:
    if args.limit < 1:
        raise ValueError(
            "--limit must be positive"
        )

    response = await _brave_core(args).search(
        args.query,
        mode=SourceMode.EXTERNAL_ONLY,
        limit=args.limit,
    )
    print(
        json.dumps(
            _jsonable(
                asdict(response)
            ),
            indent=2,
            sort_keys=True,
        )
    )
    return 0 if response.results else 3


def _run_server(
    args: argparse.Namespace,
) -> int:
    if (
        args.port < 1
        or args.port > 65535
    ):
        raise ValueError(
            "--port must be between 1 and 65535"
        )

    server = create_search_server(
        _brave_core(args),
        host=args.host,
        port=args.port,
        mode=SourceMode.EXTERNAL_ONLY,
        public_base_url=args.public_base_url,
    )
    print(
        (
            "GoreeCloud Search Development service "
            f"listening on {args.host}:"
            f"{server.server_port}; canonical origin "
            f"{server.public_base_url}"
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
            parsed = parse_query(
                args.query
            )
        except QueryParseError as exc:
            parser.error(str(exc))
        print(
            json.dumps(
                _jsonable(
                    asdict(parsed)
                ),
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
            return _run_server(args)
        except ValueError as exc:
            parser.error(str(exc))
    parser.error("unsupported command")
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
