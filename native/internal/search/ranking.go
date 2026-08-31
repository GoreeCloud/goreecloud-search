package search

import (
	"net/url"
	"sort"
	"strings"
	"unicode"
)

const (
	maxRankScore            = 30000
	maxPhraseScore          = 2500
	maxProviderScoreSignal  = 300
	maxConsensusBonus       = 900
	consensusBonusPerSource = 300
	topDiversityWindow      = 10
	maxHostResultsInWindow  = 2
)

type resultCluster struct {
	result    Result
	baseScore int
	sources   map[string]struct{}
}

// rankResults owns cross-provider ranking for the native Search engine. Provider
// scores are treated as bounded supporting evidence rather than a shared global
// relevance scale. Query/result text remains local to the request and this
// ranking path does not use click history, behavioral profiles, or telemetry.
func rankResults(query string, candidates []Result) []Result {
	intent := parseQueryIntent(query)
	clusters := make(map[string]*resultCluster, len(candidates))
	for _, candidate := range candidates {
		baseScore := relevanceScoreIntent(intent, candidate)
		cluster, exists := clusters[candidate.URL]
		if !exists {
			cluster = &resultCluster{
				result:    candidate,
				baseScore: baseScore,
				sources:   map[string]struct{}{},
			}
			clusters[candidate.URL] = cluster
		}
		if candidate.Provider != "" {
			cluster.sources[candidate.Provider] = struct{}{}
		}
		if baseScore > cluster.baseScore || (baseScore == cluster.baseScore && resultBetterThan(candidate, cluster.result)) {
			cluster.result = candidate
			cluster.baseScore = baseScore
		}
	}

	ranked := make([]Result, 0, len(clusters))
	for _, cluster := range clusters {
		sources := make([]string, 0, len(cluster.sources))
		for source := range cluster.sources {
			sources = append(sources, source)
		}
		sort.Strings(sources)

		result := cluster.result
		result.Sources = sources
		result.SourceCount = len(sources)
		result.Score = clampRankScore(cluster.baseScore + consensusBonus(result.SourceCount))
		ranked = append(ranked, result)
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			if ranked[i].URL == ranked[j].URL {
				return ranked[i].Provider < ranked[j].Provider
			}
			return ranked[i].URL < ranked[j].URL
		}
		return ranked[i].Score > ranked[j].Score
	})
	return diversifyTopResultsIntent(intent, ranked)
}

func relevanceScore(query string, result Result) int {
	return relevanceScoreIntent(parseQueryIntent(query), result)
}

func relevanceScoreIntent(intent queryIntent, result Result) int {
	title := normalizeSearchText(result.Title)
	snippet := normalizeSearchText(result.Snippet)
	urlText := normalizeURLText(result.URL)
	host := resultHost(result)

	score := providerScoreSignal(result.Score)
	if len(intent.tokens) > 0 {
		titleCoverage := tokenCoverageScore(intent.tokens, title, 2600)
		score += titleCoverage
		score += tokenCoverageScore(intent.tokens, snippet, 1000)
		score += tokenCoverageScore(intent.tokens, urlText, 700)
		score += fuzzyTokenCoverageScore(intent.tokens, title, 700)
		score += fuzzyTokenCoverageScore(intent.tokens, snippet, 250)
		score += fuzzyTokenCoverageScore(intent.tokens, urlText, 300)

		if title != "" && title == intent.normalized {
			score += 3600
		}
		if len(intent.tokens) > 1 && title != "" && strings.Contains(title, intent.normalized) {
			score += 1800
		}
		if titleCoverage == 2600 {
			score += 900
		}
		if title != "" && strings.HasPrefix(title, intent.normalized) {
			score += 350
		}
	}
	if title == "" {
		score -= 250
	}

	score += phraseMatchScore(intent.phrases, title, snippet, urlText)

	if len(intent.siteHosts) > 0 {
		matched := false
		for _, target := range intent.siteHosts {
			if hostMatchesTarget(host, target) {
				matched = true
				break
			}
		}
		if matched {
			score += 2800
		} else {
			score -= 1200
		}
	}

	for _, target := range intent.domainTargets {
		if hostMatchesTarget(host, target) {
			score += 2400
			break
		}
	}

	if len(intent.fileTypes) > 0 {
		fileType := resultFileType(result.URL)
		matched := false
		for _, target := range intent.fileTypes {
			if fileType == target {
				matched = true
				break
			}
		}
		if matched {
			score += 1800
		} else {
			score -= 500
		}
	}

	return clampRankScore(score)
}

func phraseMatchScore(phrases []string, title, snippet, urlText string) int {
	score := 0
	for _, phrase := range phrases {
		if strings.Contains(title, phrase) {
			score += 1000
			if title == phrase {
				score += 500
			}
		}
		if strings.Contains(snippet, phrase) {
			score += 350
		}
		if strings.Contains(urlText, phrase) {
			score += 200
		}
	}
	if score > maxPhraseScore {
		return maxPhraseScore
	}
	return score
}

func providerScoreSignal(raw int) int {
	if raw <= 0 {
		return 0
	}
	if raw > maxProviderScoreSignal {
		return maxProviderScoreSignal
	}
	return raw
}

func consensusBonus(sourceCount int) int {
	if sourceCount <= 1 {
		return 0
	}
	bonus := (sourceCount - 1) * consensusBonusPerSource
	if bonus > maxConsensusBonus {
		return maxConsensusBonus
	}
	return bonus
}

func clampRankScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > maxRankScore {
		return maxRankScore
	}
	return score
}

func normalizeSearchText(raw string) string {
	var builder strings.Builder
	builder.Grow(len(raw))
	spacePending := false
	wrote := false
	for _, r := range strings.ToLower(raw) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if spacePending && wrote {
				builder.WriteByte(' ')
			}
			builder.WriteRune(r)
			wrote = true
			spacePending = false
			continue
		}
		if wrote {
			spacePending = true
		}
	}
	return builder.String()
}

func uniqueTokens(normalized string) []string {
	fields := strings.Fields(normalized)
	seen := make(map[string]struct{}, len(fields))
	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		tokens = append(tokens, field)
	}
	return tokens
}

func tokenCoverageScore(queryTokens []string, normalizedField string, maximum int) int {
	if len(queryTokens) == 0 || normalizedField == "" || maximum <= 0 {
		return 0
	}
	fieldTokens := uniqueTokens(normalizedField)
	present := make(map[string]struct{}, len(fieldTokens))
	for _, token := range fieldTokens {
		present[token] = struct{}{}
	}
	matched := 0
	for _, token := range queryTokens {
		if _, ok := present[token]; ok {
			matched++
		}
	}
	return matched * maximum / len(queryTokens)
}

func normalizeURLText(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return normalizeSearchText(raw)
	}
	return normalizeSearchText(parsed.Hostname() + " " + parsed.EscapedPath() + " " + parsed.RawQuery)
}

func resultHost(result Result) string {
	parsed, err := url.Parse(result.URL)
	if err != nil {
		return ""
	}
	return strings.ToLower(parsed.Hostname())
}

// diversifyTopResults prevents one hostname from consuming the first viewport
// when other relevant hosts are available. Explicit site/domain-looking queries
// retain pure score order because concentration is likely intentional.
func diversifyTopResults(query string, ranked []Result) []Result {
	return diversifyTopResultsIntent(parseQueryIntent(query), ranked)
}

func diversifyTopResultsIntent(intent queryIntent, ranked []Result) []Result {
	if len(ranked) < 4 || len(intent.siteHosts) > 0 || len(intent.domainTargets) > 0 {
		return ranked
	}
	window := topDiversityWindow
	if len(ranked) < window {
		window = len(ranked)
	}

	selected := make([]bool, len(ranked))
	hostCounts := map[string]int{}
	output := make([]Result, 0, len(ranked))
	for len(output) < window {
		progressed := false
		for index, result := range ranked {
			if selected[index] {
				continue
			}
			host := resultHost(result)
			if host != "" && hostCounts[host] >= maxHostResultsInWindow {
				continue
			}
			selected[index] = true
			output = append(output, result)
			if host != "" {
				hostCounts[host]++
			}
			progressed = true
			if len(output) == window {
				break
			}
		}
		if progressed {
			continue
		}
		for index, result := range ranked {
			if selected[index] {
				continue
			}
			selected[index] = true
			output = append(output, result)
			break
		}
	}
	for index, result := range ranked {
		if !selected[index] {
			output = append(output, result)
		}
	}
	return output
}

func queryTargetsDomain(query string) bool {
	intent := parseQueryIntent(query)
	return len(intent.siteHosts) > 0 || len(intent.domainTargets) > 0
}
