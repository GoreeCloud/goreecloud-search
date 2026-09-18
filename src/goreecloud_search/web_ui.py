from __future__ import annotations

from html import escape
from urllib.parse import urlsplit

from .service import SearchResponse
from .version import __version__


GLAZE_UI_TARGET_VERSION = "1.5.1"
GLAZE_UI_ACCEPTED = False


def _host_label(url: str) -> str:
    parts = urlsplit(url)
    return (parts.hostname or url).casefold().rstrip(".")


def _status_label(response: SearchResponse) -> str:
    return response.execution.availability.value.replace("_", " ").title()


def _result_html(response: SearchResponse) -> str:
    if not response.results:
        return (
            '<section class="empty-state" aria-live="polite">'
            "<h2>No results</h2>"
            "<p>The configured source returned no visible results for this query.</p>"
            "</section>"
        )

    items: list[str] = []
    for item in response.results:
        result = item.result
        provenance = ", ".join(
            sorted({source.provider for source in result.provenance}, key=str.casefold)
        )
        explanations = "".join(
            f"<li>{escape(explanation)}</li>"
            for explanation in item.explanations
        )
        why = (
            '<details class="why-result">'
            "<summary>Why this result?</summary>"
            f"<ul>{explanations}</ul>"
            "</details>"
            if explanations
            else ""
        )
        source_agreement = (
            f'<span class="meta-chip">{result.source_agreement} sources agree</span>'
            if result.source_agreement > 1
            else ""
        )
        items.append(
            '<article class="result" role="listitem">'
            '<div class="result-meta">'
            f'<span class="result-host">{escape(_host_label(result.canonical_url))}</span>'
            f'<span class="meta-chip">via {escape(provenance or "configured source")}</span>'
            f"{source_agreement}"
            "</div>"
            f'<h2><a href="{escape(result.canonical_url, quote=True)}" '
            'rel="noreferrer noopener" referrerpolicy="no-referrer">'
            f"{escape(result.title)}</a></h2>"
            f'<p class="result-url">{escape(result.canonical_url)}</p>'
            f'<p class="result-snippet">{escape(result.snippet)}</p>'
            f"{why}"
            "</article>"
        )
    return '<section class="results" role="list">' + "".join(items) + "</section>"


def render_search_page(
    *,
    query: str = "",
    response: SearchResponse | None = None,
    error: str | None = None,
) -> bytes:
    query_value = escape(query, quote=True)
    status = ""
    results = ""
    if response is not None:
        disclosure = (
            "External provider query disclosure active"
            if response.plan.third_party_query_disclosure
            else "No third-party query disclosure"
        )
        status = (
            '<section class="search-status" aria-live="polite">'
            f'<span class="status-pill">{escape(_status_label(response))}</span>'
            f"<span>{escape(disclosure)}</span>"
            f"<span>{len(response.results)} result"
            f'{"s" if len(response.results) != 1 else ""}</span>'
            "</section>"
        )
        results = _result_html(response)
    elif error:
        results = (
            '<section class="error-state" role="alert">'
            "<h2>Search unavailable</h2>"
            f"<p>{escape(error)}</p>"
            "</section>"
        )

    return f"""<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="color-scheme" content="light dark">
  <meta name="robots" content="noindex,nofollow,noarchive">
  <title>GoreeCloud Search</title>
  <link rel="stylesheet" href="/assets/search.css">
  <link rel="search" type="application/opensearchdescription+xml"
        title="GoreeCloud Search" href="/opensearch.xml">
</head>
<body>
  <a class="skip-link" href="#results">Skip to results</a>
  <main class="shell">
    <header class="masthead">
      <div class="brand" aria-label="GoreeCloud Search">
        <span class="brand-mark" aria-hidden="true">G</span>
        <div>
          <p class="eyebrow">GoreeCloud</p>
          <h1>Search</h1>
        </div>
      </div>
      <span class="development-badge">Development</span>
    </header>

    <section class="search-surface" aria-labelledby="search-heading">
      <h2 id="search-heading" class="visually-hidden">Search the web</h2>
      <form class="search-form" action="/search" method="post" role="search">
        <label class="visually-hidden" for="query">Search query</label>
        <input id="query" name="q" type="search" value="{query_value}"
               placeholder="Search with GoreeCloud" autocomplete="off"
               autocapitalize="none" spellcheck="false" maxlength="2048"
               required>
        <button type="submit">Search</button>
      </form>
      <p class="privacy-note">
        This Development profile uses an explicitly configured external provider.
        The page form submits by POST so the query is not placed in the page URL.
        Query disclosure to the configured provider is shown with each result set.
      </p>
    </section>

    <div id="results">
      {status}
      {results}
    </div>

    <footer>
      <span>GoreeCloud Search {escape(__version__)}</span>
      <span>Glaze UI target {GLAZE_UI_TARGET_VERSION} · adoption not yet accepted</span>
      <span>Private VPS service target · production acceptance pending</span>
    </footer>
  </main>
</body>
</html>
""".encode("utf-8")


def render_opensearch(*, service_origin: str) -> bytes:
    origin = service_origin.rstrip("/")
    template = f"{origin}/search?q={{searchTerms}}"
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<OpenSearchDescription xmlns="http://a9.com/-/spec/opensearch/1.1/">
  <ShortName>GoreeCloud Search</ShortName>
  <Description>Search with GoreeCloud Search</Description>
  <InputEncoding>UTF-8</InputEncoding>
  <Url type="text/html" method="get" template="{escape(template, quote=True)}"/>
  <SearchForm>{escape(origin, quote=True)}/</SearchForm>
</OpenSearchDescription>
""".encode("utf-8")
