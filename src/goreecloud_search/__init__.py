"""GoreeCloud Search core package."""

from .models import ParsedQuery, ProviderDescriptor, ProviderOrigin, QueryFilters, SearchCategory, SourceMode, SourcePlan, SourcePlanStep
from .planner import SourcePlanningError, plan_sources
from .query_parser import QueryParseError, parse_query
from .service import SearchCore
from .version import __version__

__all__ = [
    "__version__", "ParsedQuery", "ProviderDescriptor", "ProviderOrigin", "QueryFilters",
    "SearchCategory", "SourceMode", "SourcePlan", "SourcePlanStep", "QueryParseError",
    "SourcePlanningError", "parse_query", "plan_sources", "SearchCore",
]
