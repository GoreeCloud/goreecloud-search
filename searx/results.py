# SPDX-License-Identifier: AGPL-3.0-or-later
# pylint: disable=missing-module-docstring, missing-class-docstring

import math
import re
import typing as t
import warnings
from collections import defaultdict
from threading import RLock

from searx import logger as log
import searx.engines
from searx.metrics import histogram_observe, counter_add
from searx.result_types import Result, LegacyResult, MainResult
from searx.result_types.answer import AnswerSet, BaseAnswer

# GoreeCloud ranking is deliberately local, deterministic, and non-personalized. It
# combines bounded provider confidence, reciprocal-rank evidence, multi-engine
# consensus, query/title/domain/snippet relevance, and a light top-results domain
# diversity pass. It never calls an external reranker or records user behavior.
RRF_K = 20.0
CONSENSUS_BONUS = 0.18
TITLE_EXACT_BONUS = 0.90
TITLE_ALL_TERMS_BONUS = 0.55
TITLE_COVERAGE_BONUS = 0.35
HOST_COVERAGE_BONUS = 0.18
CONTENT_COVERAGE_BONUS = 0.12
HIGH_PRIORITY_BONUS = 5.0
DIVERSITY_WINDOW = 8
DIVERSITY_MAX_PER_HOST = 2
TAG_RE = re.compile(r"<[^>]+>")
TOKEN_RE = re.compile(r"\w+", re.UNICODE)


def _plain_text(value: str | None) -> str:
    return TAG_RE.sub(" ", value or "").casefold()


def _query_terms(query: str) -> tuple[str, list[str]]:
    normalized = " ".join(TOKEN_RE.findall(_plain_text(query)))
    terms = [term for term in normalized.split() if len(term) > 1]
    if not terms and normalized:
        terms = normalized.split()
    return normalized, terms


def _coverage(terms: list[str], text: str) -> float:
    if not terms:
        return 0.0
    matched = sum(1 for term in terms if term in text)
    return matched / len(terms)


def _bounded_engine_weight(result: MainResult | LegacyResult) -> float:
    weights: list[float] = []
    for result_engine in result["engines"]:
        engine = searx.engines.engines.get(result_engine)
        try:
            weight = float(getattr(engine, "weight", 1.0))
        except (TypeError, ValueError):
            weight = 1.0
        weights.append(min(2.0, max(0.5, weight)))
    if not weights:
        return 1.0
    return sum(weights) / len(weights)


def _lexical_relevance(result: MainResult | LegacyResult, query: str) -> float:
    normalized_query, terms = _query_terms(query)
    if not normalized_query or not terms:
        return 0.0

    title = _plain_text(result.title)
    content = _plain_text(result.content)
    hostname = ""
    if result.parsed_url:
        hostname = result.parsed_url.hostname or result.parsed_url.netloc or ""
    hostname = hostname.casefold().removeprefix("www.")

    score = 0.0
    if normalized_query in title:
        score += TITLE_EXACT_BONUS
    if all(term in title for term in terms):
        score += TITLE_ALL_TERMS_BONUS
    score += TITLE_COVERAGE_BONUS * _coverage(terms, title)
    score += HOST_COVERAGE_BONUS * _coverage(terms, hostname)
    score += CONTENT_COVERAGE_BONUS * _coverage(terms, content)
    return score


def calculate_score(
    result: MainResult | LegacyResult,
    priority: MainResult.PriorityType,
    query: str = "",
) -> float:
    """Calculate deterministic GoreeCloud metasearch relevance.

    The upstream position signal remains important, but provider weights are averaged
    and bounded instead of multiplied across duplicates. Agreement by independent
    engines receives a modest bonus, while query/title/domain/snippet matching can
    distinguish results with otherwise similar provider positions.
    """
    if priority == "low":
        return 0.0

    positions = [position for position in result["positions"] if isinstance(position, int) and position > 0]
    if not positions:
        positions = [1]

    reciprocal_rank = sum(RRF_K / (RRF_K + position) for position in positions)
    provider_weight = _bounded_engine_weight(result)
    engine_count = len({engine for engine in result["engines"] if engine})
    consensus = CONSENSUS_BONUS * math.log2(1 + max(1, engine_count))
    lexical = _lexical_relevance(result, query)

    score = reciprocal_rank * provider_weight + consensus + lexical
    if priority == "high":
        score += HIGH_PRIORITY_BONUS
    return score


def _hostname(result: MainResult | LegacyResult) -> str:
    if not result.parsed_url:
        return ""
    return (result.parsed_url.hostname or result.parsed_url.netloc or "").casefold().removeprefix("www.")


def diversify_top_domains(
    results: list[MainResult | LegacyResult],
    window: int = DIVERSITY_WINDOW,
    max_per_host: int = DIVERSITY_MAX_PER_HOST,
) -> list[MainResult | LegacyResult]:
    """Gently prevent one hostname from monopolizing the first result viewport.

    Only the first ``window`` slots are diversified. Strong deferred results keep
    their score order immediately after that window, so diversity does not become a
    broad domain penalty across the result set.
    """
    if window <= 0 or max_per_host <= 0 or len(results) <= max_per_host:
        return list(results)

    remaining = list(results)
    selected: list[MainResult | LegacyResult] = []
    host_counts: dict[str, int] = defaultdict(int)

    while remaining and len(selected) < min(window, len(results)):
        chosen_index: int | None = None
        for index, candidate in enumerate(remaining):
            host = _hostname(candidate)
            if not host or host_counts[host] < max_per_host:
                chosen_index = index
                break
        if chosen_index is None:
            break
        candidate = remaining.pop(chosen_index)
        selected.append(candidate)
        host = _hostname(candidate)
        if host:
            host_counts[host] += 1

    # If every remaining item is from already saturated hosts, fill the window in
    # original score order rather than hiding valid results.
    while remaining and len(selected) < min(window, len(results)):
        selected.append(remaining.pop(0))

    selected.extend(remaining)
    return selected


class Timing(t.NamedTuple):
    engine: str
    total: float
    load: float


class UnresponsiveEngine(t.NamedTuple):
    engine: str
    error_type: str
    suspended: bool


class ResultContainer:
    """In the result container, the results are collected, sorted and duplicates
    will be merged."""

    # pylint: disable=too-many-statements

    main_results_map: dict[int, MainResult | LegacyResult]
    infoboxes: list[LegacyResult]
    suggestions: set[str]
    answers: AnswerSet
    corrections: set[str]

    def __init__(self, query: str = ""):
        self.main_results_map = {}
        self.infoboxes = []
        self.suggestions = set()
        self.answers = AnswerSet()
        self.corrections = set()
        self.query = query

        self.engine_data: dict[str, dict[str, str]] = defaultdict(dict)
        self._closed: bool = False
        self.paging: bool = False
        self.unresponsive_engines: set[UnresponsiveEngine] = set()
        self.timings: list[Timing] = []
        self.redirect_url: str | None = None
        self.on_result: t.Callable[[Result | LegacyResult], bool] = lambda _: True
        self._lock: RLock = RLock()
        self._main_results_sorted: list[MainResult | LegacyResult] = None  # type: ignore

    def extend(
        self, engine_name: str | None, results: list[Result | LegacyResult]
    ):  # pylint: disable=too-many-branches
        if self._closed:
            log.debug("container is closed, ignoring results: %s", results)
            return
        main_count = 0

        for result in list(results):
            if isinstance(result, Result):
                result.engine = result.engine or engine_name
                result.normalize_result_fields()
                if not self.on_result(result):
                    continue

                if isinstance(result, BaseAnswer):
                    self.answers.add(result)
                elif isinstance(result, MainResult):
                    main_count += 1
                    self._merge_main_result(result, main_count)
                else:
                    raise NotImplementedError(f"no handler implemented to process the result of type {result}")

            else:
                result["engine"] = result.get("engine") or engine_name or ""
                result = LegacyResult(result)
                result.normalize_result_fields()

                if "suggestion" in result:
                    if self.on_result(result):
                        self.suggestions.add(result["suggestion"])
                    continue

                if "answer" in result:
                    if self.on_result(result):
                        warnings.warn(
                            f"answer results from engine {result.engine}"
                            " are without typification / migrate to Answer class.",
                            DeprecationWarning,
                        )
                        self.answers.add(result)  # type: ignore
                    continue

                if "correction" in result:
                    if self.on_result(result):
                        self.corrections.add(result["correction"])
                    continue

                if "infobox" in result:
                    if self.on_result(result):
                        self._merge_infobox(result)
                    continue

                if "engine_data" in result:
                    if self.on_result(result) and result.engine:
                        self.engine_data[result.engine][result["key"]] = result["engine_data"]
                    continue

                if self.on_result(result):
                    main_count += 1
                    self._merge_main_result(result, main_count)
                    continue

        if engine_name in searx.engines.engines:
            eng = searx.engines.engines[engine_name]
            histogram_observe(main_count, "engine", eng.name, "result", "count")
            if not self.paging and eng.paging:
                self.paging = True

    def _merge_infobox(self, new_infobox: LegacyResult):
        add_infobox = True
        new_id = getattr(new_infobox, "id", None)
        if new_id is not None:
            with self._lock:
                for existing_infobox in self.infoboxes:
                    if new_id == getattr(existing_infobox, "id", None):
                        merge_two_infoboxes(existing_infobox, new_infobox)
                        add_infobox = False
        if add_infobox:
            self.infoboxes.append(new_infobox)

    def _merge_main_result(self, result: MainResult | LegacyResult, position: int):
        result_hash = hash(result)
        with self._lock:
            merged = self.main_results_map.get(result_hash)
            if not merged:
                result.positions = [position]
                self.main_results_map[result_hash] = result
                return
            merge_two_main_results(merged, result)
            merged.positions.append(position)

    def close(self):
        self._closed = True
        for result in self.main_results_map.values():
            result.score = calculate_score(result, result.priority, self.query)
            for eng_name in result.engines:
                counter_add(result.score, "engine", eng_name, "score")

    def get_ordered_results(self) -> list[MainResult | LegacyResult]:
        """Returns a sorted list of results to be displayed in the main result area."""
        if not self._closed:
            self.close()
        if self._main_results_sorted:
            return self._main_results_sorted

        # Pass 1: deterministic relevance score, followed by a light first-viewport
        # hostname diversity pass. This is deliberately not personalization.
        results = sorted(self.main_results_map.values(), key=lambda x: x.score, reverse=True)
        results = diversify_top_domains(results)

        # Pass 2: preserve the existing SearXNG category/template grouping behavior.
        gresults: list[MainResult | LegacyResult] = []
        categoryPositions: dict[str, t.Any] = {}
        max_count = 8
        max_distance = 20

        for res in results:
            engine = searx.engines.engines.get(res.engine or "")
            if engine:
                res.category = engine.categories[0] if len(engine.categories) > 0 else ""

            category = f"{res.category}:{res.template}:{'img_src' if (res.thumbnail or res.img_src) else ''}"
            grp = categoryPositions.get(category)
            if (grp is not None) and (grp["count"] > 0) and (len(gresults) - grp["index"] < max_distance):
                index = grp["index"]
                gresults.insert(index, res)
                for item in categoryPositions.values():
                    if item["index"] >= index:
                        item["index"] += 1
                grp["count"] -= 1
            else:
                gresults.append(res)
                categoryPositions[category] = {"index": len(gresults), "count": max_count}
                continue

        self._main_results_sorted = gresults
        return self._main_results_sorted

    def add_unresponsive_engine(self, engine_name: str, error_type: str, suspended: bool = False):
        with self._lock:
            if self._closed:
                log.error("call to ResultContainer.add_unresponsive_engine after ResultContainer.close")
                return
            if searx.engines.engines[engine_name].display_error_messages:
                self.unresponsive_engines.add(UnresponsiveEngine(engine_name, error_type, suspended))

    def add_timing(self, engine_name: str, engine_time: float, page_load_time: float):
        with self._lock:
            if self._closed:
                log.error("call to ResultContainer.add_timing after ResultContainer.close")
                return
            self.timings.append(Timing(engine_name, total=engine_time, load=page_load_time))

    def get_timings(self) -> list[Timing]:
        with self._lock:
            if not self._closed:
                log.error("call to ResultContainer.get_timings before ResultContainer.close")
                return []
            return self.timings


def merge_two_infoboxes(origin: LegacyResult, other: LegacyResult):
    """Merges the values from ``other`` into ``origin``."""
    # pylint: disable=too-many-branches
    weight1 = getattr(searx.engines.engines[origin.engine], "weight", 1)
    weight2 = getattr(searx.engines.engines[other.engine], "weight", 1)

    if weight2 > weight1:
        origin.engine = other.engine

    origin.engines |= other.engines

    if other.urls:
        url_items = origin.get("urls", [])
        for url2 in other.urls:
            unique_url = True
            entity_url2 = url2.get("entity")
            for url1 in origin.get("urls", []):
                if (entity_url2 is not None and entity_url2 == url1.get("entity")) or url1.get("url") == url2.get("url"):
                    unique_url = False
                    break
            if unique_url:
                url_items.append(url2)
        origin.urls = url_items

    if other.img_src:
        if not origin.img_src:
            origin.img_src = other.img_src
        elif weight2 > weight1:
            origin.img_src = other.img_src

    if other.attributes:
        if not origin.attributes:
            origin.attributes = other.attributes
        else:
            attr_names_1: set[str] = set()
            for attr in origin.attributes:
                label = attr.get("label")
                if label:
                    attr_names_1.add(label)
                entity = attr.get("entity")
                if entity:
                    attr_names_1.add(entity)
            for attr in other.attributes:
                if attr.get("label") not in attr_names_1 and attr.get("entity") not in attr_names_1:
                    origin.attributes.append(attr)

    if other.content:
        if not origin.content:
            origin.content = other.content
        elif len(other.content) > len(origin.content):
            origin.content = other.content


def merge_two_main_results(origin: MainResult | LegacyResult, other: MainResult | LegacyResult):
    """Merges the values from ``other`` into ``origin``."""
    if len(other.content or "") > len(origin.content or ""):
        origin.content = other.content
    if len(other.title or "") > len(origin.title or ""):
        origin.title = other.title

    if isinstance(other, MainResult) and isinstance(origin, MainResult):
        origin.defaults_from(other)
    elif isinstance(other, LegacyResult) and isinstance(origin, LegacyResult):
        origin.defaults_from(other)

    origin.engines.add(other.engine or "")

    if origin.parsed_url and not origin.parsed_url.scheme.endswith("s"):
        if other.parsed_url and other.parsed_url.scheme.endswith("s"):
            origin.parsed_url = origin.parsed_url._replace(scheme=other.parsed_url.scheme)
            origin.url = origin.parsed_url.geturl()
