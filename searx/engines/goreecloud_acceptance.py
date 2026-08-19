# SPDX-License-Identifier: AGPL-3.0-or-later
"""Deterministic offline provider used only by GoreeCloud Search acceptance CI.

The engine is disabled by default and has no external network dependency. It is
loaded only by the isolated provider-degradation settings file so acceptance can
prove that a healthy provider remains usable when a sibling online provider
fails.
"""

from __future__ import annotations

from searx.enginelib import EngineAbout
from searx.result_types import EngineResults


engine_type = "offline"
categories = ["general"]
disabled = True
about = EngineAbout(
    results="deterministic CI fixture",
    description="GoreeCloud Search provider-degradation acceptance fixture.",
)


def init(engine_settings: dict[str, object]) -> bool:  # pylint: disable=unused-argument
    return True


def search(query: str, params: object) -> EngineResults:  # pylint: disable=unused-argument
    results = EngineResults()
    results.add(
        results.types.LegacyResult(
            url="https://example.invalid/goreecloud-provider-healthy",
            title="GoreeCloud provider degradation healthy result",
            content=(
                "Healthy provider result remained usable while a sibling provider failed. "
                f"Query: {query}"
            ),
        )
    )
    return results
