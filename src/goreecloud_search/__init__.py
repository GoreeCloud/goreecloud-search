"""GoreeCloud Search core package."""

from .index_contract import (
    INDEX_CONTRACT_VERSION,
    IndexCapabilities,
    IndexContractError,
    IndexDocumentCandidate,
    IndexSearchRequest,
    IndexSearchResponse,
)
from .index_provider import GoreeCloudIndexProvider, IndexTransport
from .models import ParsedQuery, ProviderDescriptor, ProviderOrigin, QueryFilters, SearchCategory, SourceMode, SourcePlan, SourcePlanStep
from .normalization import NormalizedResult, ResultNormalizationError, ResultProvenance, canonicalize_url, normalize_and_deduplicate
from .planner import SourcePlanningError, plan_sources
from .providers import ResultCandidate, SearchProvider
from .query_parser import QueryParseError, parse_query
from .ranking import RankedResult, RankingSignal, rank_results
from .service import SearchCore
from .version import __version__

__all__ = [
    "__version__", "INDEX_CONTRACT_VERSION", "ParsedQuery", "ProviderDescriptor",
    "ProviderOrigin", "QueryFilters", "SearchCategory", "SourceMode", "SourcePlan",
    "SourcePlanStep", "QueryParseError", "SourcePlanningError", "IndexContractError",
    "ResultNormalizationError", "IndexCapabilities", "IndexDocumentCandidate",
    "IndexSearchRequest", "IndexSearchResponse", "GoreeCloudIndexProvider",
    "IndexTransport", "ResultCandidate", "SearchProvider", "ResultProvenance",
    "NormalizedResult", "canonicalize_url", "normalize_and_deduplicate",
    "RankingSignal", "RankedResult", "rank_results",
    "parse_query", "plan_sources", "SearchCore",
]
