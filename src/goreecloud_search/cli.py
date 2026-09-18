from __future__ import annotations

import argparse
import asyncio
from dataclasses import asdict
from datetime import date
from enum import Enum
import json
import os
from typing import Any

from .brave_provider import BraveWebSearchProvider
from .models import SourceMode
from .query_parser import QueryParseError, parse_query
from .service import SearchCore
from .version import __version__


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
    search_command.add_argument(
        "--country",
        default=None,
        help="Optional two-letter country code",
    )
    search_command.add_argument(
        "--language",
        default=None,
        help="Optional Brave search language",
    )
    search_command.add_argument(
        "--provider-safesearch",
        choices=("off", "moderate", "strict"),
        default="moderate",
        help="Filtering requested from the external provider; default: moderate",
    )
    return parser


async def _run_brave_search(
    args: argparse.Namespace,
) -> int:
    api_key = os.environ.get(
        "BRAVE_SEARCH_API_KEY",
        "",
    ).strip()
    if not api_key:
        raise ValueError(
            "BRAVE_SEARCH_API_KEY is required "
            "for the Brave provider"
        )
    if args.limit < 1:
        raise ValueError("--limit must be positive")

    provider = BraveWebSearchProvider(
        api_key=api_key,
        country=args.country,
        search_lang=args.language,
        safesearch=args.provider_safesearch,
    )
    core = SearchCore(
        provider_adapters=(provider,),
    )
    response = await core.search(
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
    parser.error("unsupported command")
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
