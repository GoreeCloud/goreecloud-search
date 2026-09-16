"""GoreeCloud Search core package."""

from .content_policy import (
    ContentPolicyAction,
    ContentPolicyDecision,
    ContentPolicyEngine,
    ContentPolicyError,
    ContentPolicyHook,
    ContentPolicyOutcome,
    ContentPolicyReport,
    DomainPolicyHook,
    SafeSearchMode,
)
from .execution import (
    ExecutionPolicy,
    ExecutionReport,
    ProviderAttempt,
    ProviderExecutionError,
    ProviderExecutionStatus,
    SearchAvailability,
    SearchExecutor,
)
from .index_contract import (
    INDEX_CONTRACT_VERSION,
    IndexCapabilities,
    IndexContractError,
    IndexDocumentCandidate,
    IndexSearchRequest,
    IndexSearchResponse,
)
from .index_provider import GoreeCloudIndexProvider, IndexTransport
from .lenses import (
    Lens,
    LensApplicationReport,
    LensError,
    LensExclusion,
    LensRegistry,
    LensRule,
    LensRuleAction,
    LensRuleTarget,
    apply_lens,
)
from .models import ParsedQuery, ProviderDescriptor, ProviderOrigin, QueryDisclosureBudget, QueryFilters, SearchCategory, SourceMode, SourcePlan, SourcePlanStep
from .normalization import NormalizedResult, ResultNormalizationError, ResultProvenance, canonicalize_url, normalize_and_deduplicate
from .planner import SourcePlanningError, plan_sources
from .providers import ProviderSearchBatch, ResultCandidate, SearchProvider
from .query_parser import QueryParseError, parse_query
from .ranking import RankedResult, RankingSignal, rank_results
from .service import SearchCore, SearchResponse
from .version import __version__

__all__ = [
    "__version__", "INDEX_CONTRACT_VERSION", "ContentPolicyAction",
    "ContentPolicyDecision", "ContentPolicyEngine", "ContentPolicyError",
    "ContentPolicyHook", "ContentPolicyOutcome", "ContentPolicyReport",
    "DomainPolicyHook", "SafeSearchMode", "Lens", "LensRule", "LensRuleAction",
    "LensRuleTarget", "LensRegistry", "LensError", "LensExclusion",
    "LensApplicationReport", "apply_lens", "ParsedQuery", "ProviderDescriptor",
    "ProviderOrigin", "QueryDisclosureBudget", "QueryFilters", "SearchCategory", "SourceMode", "SourcePlan",
    "SourcePlanStep", "QueryParseError", "SourcePlanningError", "IndexContractError",
    "ResultNormalizationError", "ProviderExecutionError", "IndexCapabilities",
    "IndexDocumentCandidate", "IndexSearchRequest", "IndexSearchResponse",
    "GoreeCloudIndexProvider", "IndexTransport", "ResultCandidate",
    "ProviderSearchBatch", "SearchProvider", "ResultProvenance", "NormalizedResult",
    "canonicalize_url", "normalize_and_deduplicate", "RankingSignal", "RankedResult",
    "rank_results", "ExecutionPolicy", "ProviderExecutionStatus", "ProviderAttempt",
    "ExecutionReport", "SearchAvailability", "SearchExecutor", "SearchResponse",
    "parse_query", "plan_sources", "SearchCore",
]
