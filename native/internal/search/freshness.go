package search

import "time"

const (
	maxFreshnessBonus        = 1200
	maxPublishedAtFutureSkew = 24 * time.Hour
)

// normalizeAuthoritativePublishedAt keeps publication metadata only when the
// provider has explicitly opted into the authoritative timestamp contract. It
// also rejects zero/pre-Unix timestamps and implausibly future timestamps so
// untrusted or malformed metadata cannot leak into ranking or result output.
func normalizeAuthoritativePublishedAt(value *time.Time, authoritative bool, now time.Time) (*time.Time, bool) {
	if !authoritative || value == nil || value.IsZero() {
		return nil, false
	}
	published := value.UTC()
	if published.Before(time.Unix(0, 0).UTC()) {
		return nil, false
	}
	if !now.IsZero() && published.After(now.UTC().Add(maxPublishedAtFutureSkew)) {
		return nil, false
	}
	return &published, true
}

// freshnessScore is deliberately a supporting signal. It is active only when
// the query explicitly asks for temporal information or when the News category
// is being ranked, and only for timestamps retained through the authoritative
// provider contract. It never infers time from snippets, URLs, crawl order, or
// arbitrary provider score.
func freshnessScore(intent queryIntent, category string, result Result, now time.Time) int {
	if !result.publishedAtTrusted || result.PublishedAt == nil || now.IsZero() {
		return 0
	}
	explicit := queryRequestsFreshness(intent)
	if !explicit && category != CategoryNews {
		return 0
	}

	age := now.UTC().Sub(result.PublishedAt.UTC())
	if age < -maxPublishedAtFutureSkew {
		return 0
	}
	if age < 0 {
		age = 0
	}

	bonus := 0
	switch {
	case age <= 24*time.Hour:
		bonus = maxFreshnessBonus
	case age <= 3*24*time.Hour:
		bonus = 950
	case age <= 7*24*time.Hour:
		bonus = 700
	case age <= 30*24*time.Hour:
		bonus = 400
	case age <= 90*24*time.Hour:
		bonus = 180
	default:
		return 0
	}

	// News should naturally prefer recent authoritative results, but explicit
	// temporal wording receives the strongest bounded recency signal.
	if category == CategoryNews && !explicit {
		bonus = bonus * 5 / 6
	}
	return bonus
}

func queryRequestsFreshness(intent queryIntent) bool {
	return intent.freshnessRequested
}
