# GoreeCloud Search ranking

GoreeCloud Search uses a deterministic, privacy-preserving metasearch ranking layer on top of the SearXNG provider and normalization foundation.

The ranking goal is not to imitate a commercial search engine's behavioral model. It is to make the information already returned by configured providers more useful by combining provider position, source agreement, query relevance, and modest source diversity without creating a user profile.

## Signals

The current web-result score combines:

1. **Reciprocal provider position** — higher provider positions contribute more than lower positions, using a bounded reciprocal-rank formula instead of the previous raw `1 / position` accumulation.
2. **Bounded provider weight** — configured engine weights are clamped and averaged across engines. They are not multiplied together, preventing multi-engine duplicates from receiving an exponential weight effect.
3. **Multi-engine agreement** — a modest logarithmic bonus rewards a result found by more than one configured engine.
4. **Title relevance** — an exact normalized query phrase in the title receives the strongest lexical bonus; all-term and partial title coverage receive smaller bounded bonuses.
5. **Hostname relevance** — query terms appearing in the hostname receive a small bonus.
6. **Snippet relevance** — query terms in the normalized result content receive a smaller supporting bonus.
7. **Priority semantics** — existing high/low result priority remains authoritative. High-priority results receive an explicit boost; low-priority results remain at the bottom.
8. **First-viewport source diversity** — after score sorting, the first result viewport allows at most two results from the same hostname when other scored hosts are available. Deferred results remain in score order after that small window.

## Privacy properties

The current query is passed only to the in-memory `ResultContainer` that already owns the returned search results. The ranking layer does not:

- persist query history;
- record click behavior;
- create a user identifier or profile;
- call a remote ranking service;
- call a language model;
- use advertising or sponsored-ranking signals;
- use account history;
- send the query to any provider that was not already selected by the normal GoreeCloud Search orchestration path.

The Browser and application privacy boundaries therefore remain unchanged by ranking.

## Tuning constants

The implementation keeps its constants near the top of `searx/results.py`. They are deliberately bounded and reviewable rather than hidden in a model. Changes to these constants should be accompanied by deterministic tests and representative provider/runtime acceptance.

Current categories of constants include reciprocal-rank damping, consensus bonus, title exact/all-term/coverage bonuses, hostname and content coverage bonuses, high-priority bonus, and the first-viewport diversity window.

## Result presentation

The interface does not expose the internal numeric ranking score. It uses a restrained `Top match` marker for the first ordinary web result and a source-consensus chip when multiple engines found the same result. Provider/engine names remain visible for transparency.

The results header states that ranking is local and non-behavioral. This is an architectural statement, not a claim that external search providers themselves are free of ranking, geographic, or other provider-side behavior.

## Limitations

GoreeCloud Search is still metasearch, not an independent public-Web index. The ranking layer can only rank the results returned by configured external providers. A provider can omit, rate-limit, geographically alter, or reorder its own results before GoreeCloud Search receives them.

Multi-engine agreement is useful evidence but is not proof of source independence or factual correctness. Several engines may depend on overlapping upstream indexes or return the same popular source. GoreeCloud Search ranking is for discovery; source verification remains a separate research task.

## Acceptance

Deterministic ranking tests cover title relevance, multi-engine consensus, priority semantics, domain diversity, and the in-memory query boundary. The dedicated `GoreeCloud ranking acceptance` workflow also validates the result-interface contract before executing the ranking unit tests.

Representative general, technical, news, science, image, and video searches should still be reviewed on the target environment because real provider quality cannot be fully represented by deterministic CI fixtures.
