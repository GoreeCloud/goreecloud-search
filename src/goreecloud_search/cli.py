from __future__ import annotations

import argparse
from dataclasses import asdict
from datetime import date
from enum import Enum
import json
from typing import Any

from .query_parser import QueryParseError, parse_query
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
    parser = argparse.ArgumentParser(prog="goreecloud-search", description="GoreeCloud Search development CLI")
    parser.add_argument("--version", action="version", version=__version__)
    subparsers = parser.add_subparsers(dest="command", required=True)
    parse_command = subparsers.add_parser("parse", help="Parse a query without performing network access.")
    parse_command.add_argument("query", help="Query text to parse")
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    if args.command == "parse":
        try:
            parsed = parse_query(args.query)
        except QueryParseError as exc:
            parser.error(str(exc))
        print(json.dumps(_jsonable(asdict(parsed)), indent=2, sort_keys=True))
        return 0
    parser.error("unsupported command")
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
