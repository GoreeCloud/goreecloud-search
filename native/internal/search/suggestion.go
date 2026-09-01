package search

import (
	"sort"
	"strings"
)

const minCorrectionProviderAgreement = 2

type correctionTarget struct {
	partIndex int
	token     string
}

type correctionChoice struct {
	token       string
	sourceCount int
}

// suggestQueryCorrection derives at most one explicit, user-visible correction
// from result-title evidence already present in the request. It never changes
// the query submitted to providers. A candidate correction must be exactly one
// bounded edit/transposition away and must appear in titles from at least two
// independent normalized providers. Quoted text, operators, and navigational
// domain targets are never rewritten.
func suggestQueryCorrection(raw string, candidates []Result) string {
	parts := splitQueryParts(raw)
	targets := correctionTargets(parts)
	if len(targets) == 0 {
		return ""
	}

	titleEvidence := titleTokenProviderEvidence(candidates)
	if len(titleEvidence) == 0 {
		return ""
	}

	selectedTarget := -1
	selectedToken := ""
	for _, target := range targets {
		if providers := titleEvidence[target.token]; len(providers) > 0 {
			// If a result title already contains the submitted token, Search has no
			// basis to tell the user that token is misspelled.
			continue
		}
		choice, ok := chooseCorrection(target.token, titleEvidence)
		if !ok {
			continue
		}
		if selectedTarget != -1 {
			// Keep the feature conservative: never propose a multi-token rewrite
			// from result evidence in one step.
			return ""
		}
		selectedTarget = target.partIndex
		selectedToken = choice.token
	}
	if selectedTarget < 0 || selectedToken == "" {
		return ""
	}

	rendered := make([]string, 0, len(parts))
	for index, part := range parts {
		text := part.text
		if index == selectedTarget {
			text = selectedToken
		}
		if part.quoted {
			text = `"` + text + `"`
		}
		rendered = append(rendered, text)
	}
	suggestion := strings.Join(rendered, " ")
	if normalizeSearchText(suggestion) == normalizeSearchText(raw) {
		return ""
	}
	return suggestion
}

func correctionTargets(parts []queryPart) []correctionTarget {
	targets := make([]correctionTarget, 0, len(parts))
	for index, part := range parts {
		if part.quoted {
			continue
		}
		trimmed := strings.TrimSpace(part.text)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "site:") || strings.HasPrefix(lower, "filetype:") || strings.HasPrefix(lower, "ext:") {
			continue
		}
		if normalizeDomainTarget(trimmed) != "" {
			continue
		}
		normalized := normalizeSearchText(trimmed)
		fields := strings.Fields(normalized)
		if len(fields) != 1 || len([]rune(fields[0])) < 5 {
			continue
		}
		targets = append(targets, correctionTarget{partIndex: index, token: fields[0]})
	}
	return targets
}

func titleTokenProviderEvidence(candidates []Result) map[string]map[string]struct{} {
	evidence := map[string]map[string]struct{}{}
	for _, result := range candidates {
		provider, ok := normalizeProviderName(result.Provider)
		if !ok {
			continue
		}
		for _, token := range uniqueTokens(normalizeSearchText(result.Title)) {
			if len([]rune(token)) < 4 {
				continue
			}
			providers := evidence[token]
			if providers == nil {
				providers = map[string]struct{}{}
				evidence[token] = providers
			}
			providers[provider] = struct{}{}
		}
	}
	return evidence
}

func chooseCorrection(queryToken string, evidence map[string]map[string]struct{}) (correctionChoice, bool) {
	choices := make([]correctionChoice, 0)
	for token, providers := range evidence {
		if len(providers) < minCorrectionProviderAgreement || !oneEditOrTranspositionApart(queryToken, token) {
			continue
		}
		choices = append(choices, correctionChoice{token: token, sourceCount: len(providers)})
	}
	if len(choices) == 0 {
		return correctionChoice{}, false
	}
	sort.Slice(choices, func(i, j int) bool {
		if choices[i].sourceCount != choices[j].sourceCount {
			return choices[i].sourceCount > choices[j].sourceCount
		}
		return choices[i].token < choices[j].token
	})
	if len(choices) > 1 && choices[0].sourceCount == choices[1].sourceCount {
		// Equal provider support for different one-edit alternatives is
		// ambiguous; Search should not guess.
		return correctionChoice{}, false
	}
	return choices[0], true
}
