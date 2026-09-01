package search

import (
	"net"
	"net/url"
	"path"
	"strings"
	"unicode"
)

type queryPart struct {
	text   string
	quoted bool
}

type intentLexeme struct {
	text   string
	quoted bool
}

type queryIntent struct {
	normalized         string
	tokens             []string
	phrases            []string
	siteHosts          []string
	domainTargets      []string
	fileTypes          []string
	freshnessRequested bool
}

func parseQueryIntent(raw string) queryIntent {
	intent := queryIntent{}
	lexemes := make([]intentLexeme, 0)
	for _, part := range splitQueryParts(raw) {
		trimmed := strings.TrimSpace(part.text)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "site:") {
			if host := normalizeDomainTarget(trimmed[len("site:"):]); host != "" {
				intent.siteHosts = appendUniqueString(intent.siteHosts, host)
			}
			continue
		}
		if strings.HasPrefix(lower, "filetype:") {
			if fileType := normalizeFileType(trimmed[len("filetype:"):]); fileType != "" {
				intent.fileTypes = appendUniqueString(intent.fileTypes, fileType)
			}
			continue
		}
		if strings.HasPrefix(lower, "ext:") {
			if fileType := normalizeFileType(trimmed[len("ext:"):]); fileType != "" {
				intent.fileTypes = appendUniqueString(intent.fileTypes, fileType)
			}
			continue
		}

		if host := normalizeDomainTarget(trimmed); host != "" {
			intent.domainTargets = appendUniqueString(intent.domainTargets, host)
			continue
		}

		normalized := normalizeSearchText(trimmed)
		if normalized == "" {
			continue
		}
		if part.quoted && strings.Contains(normalized, " ") {
			intent.phrases = appendUniqueString(intent.phrases, normalized)
		}
		for _, token := range strings.Fields(normalized) {
			lexemes = append(lexemes, intentLexeme{text: token, quoted: part.quoted})
		}
	}

	var lexicalSkip map[int]bool
	intent.freshnessRequested, lexicalSkip = temporalIntentLexemes(lexemes)
	normalizedTokens := make([]string, 0, len(lexemes))
	for index, lexeme := range lexemes {
		if lexicalSkip[index] {
			continue
		}
		normalizedTokens = append(normalizedTokens, lexeme.text)
	}
	intent.normalized = strings.Join(normalizedTokens, " ")
	intent.tokens = uniqueTokens(intent.normalized)
	return intent
}

// temporalIntentLexemes separates clear temporal modifiers from subject terms.
// Quoted terms remain literal. Ambiguous/content-bearing terms such as "news",
// "updated", and "updates" may request freshness but remain in lexical ranking.
func temporalIntentLexemes(lexemes []intentLexeme) (bool, map[int]bool) {
	unquoted := make([]int, 0, len(lexemes))
	for index, lexeme := range lexemes {
		if !lexeme.quoted {
			unquoted = append(unquoted, index)
		}
	}

	requested := false
	skip := map[int]bool{}
	for position, index := range unquoted {
		token := lexemes[index].text
		switch token {
		case "latest", "recent", "recently", "today", "breaking", "newest":
			requested = true
			skip[index] = true
		case "updated", "updates", "news":
			requested = true
		case "current":
			// Leading "current" is normally a temporal modifier (for example,
			// "current weather"). A trailing noun use such as "electric current"
			// remains ordinary lexical content and does not activate freshness.
			if position == 0 {
				requested = true
				skip[index] = true
			}
		case "this":
			if position+1 < len(unquoted) {
				nextIndex := unquoted[position+1]
				next := lexemes[nextIndex].text
				if next == "week" || next == "month" {
					requested = true
					skip[index] = true
					skip[nextIndex] = true
				}
			}
		}
	}
	return requested, skip
}

func splitQueryParts(raw string) []queryPart {
	parts := make([]queryPart, 0)
	var builder strings.Builder
	quoted := false
	partQuoted := false
	flush := func() {
		if builder.Len() == 0 {
			partQuoted = false
			return
		}
		parts = append(parts, queryPart{text: builder.String(), quoted: partQuoted})
		builder.Reset()
		partQuoted = false
	}

	for _, r := range raw {
		switch {
		case r == '"':
			if quoted {
				quoted = false
				partQuoted = true
				flush()
			} else {
				flush()
				quoted = true
				partQuoted = true
			}
		case unicode.IsSpace(r) && !quoted:
			flush()
		default:
			builder.WriteRune(r)
		}
	}
	flush()
	return parts
}

func normalizeDomainTarget(raw string) string {
	trimmed := strings.ToLower(strings.Trim(strings.TrimSpace(raw), "()[]{}<>,;\"'"))
	if trimmed == "" {
		return ""
	}

	candidate := trimmed
	if !strings.Contains(candidate, "://") {
		candidate = "https://" + candidate
	}
	parsed, err := url.Parse(candidate)
	if err != nil || parsed.Hostname() == "" || parsed.User != nil {
		return ""
	}
	host := strings.TrimSuffix(strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www."), ".")
	if host == "" {
		return ""
	}
	if net.ParseIP(host) != nil {
		return host
	}

	labels := strings.Split(host, ".")
	if len(labels) < 2 {
		return ""
	}
	for _, label := range labels {
		if !validDomainLabel(label) {
			return ""
		}
	}
	tld := labels[len(labels)-1]
	if !validDomainTLD(tld) {
		return ""
	}
	return host
}

func validDomainLabel(label string) bool {
	if label == "" || len([]rune(label)) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
		return false
	}
	for _, r := range label {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			continue
		}
		return false
	}
	return true
}

func validDomainTLD(label string) bool {
	if strings.HasPrefix(label, "xn--") {
		return len(label) > len("xn--") && validDomainLabel(label)
	}
	runes := []rune(label)
	if len(runes) < 2 || len(runes) > 63 {
		return false
	}
	for _, r := range runes {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func normalizeFileType(raw string) string {
	trimmed := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(raw)), ".")
	if trimmed == "" || len(trimmed) > 12 {
		return ""
	}
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			continue
		}
		return ""
	}
	return trimmed
}

func appendUniqueString(values []string, candidate string) []string {
	for _, value := range values {
		if value == candidate {
			return values
		}
	}
	return append(values, candidate)
}

func hostMatchesTarget(host, target string) bool {
	host = strings.TrimSuffix(strings.TrimPrefix(strings.ToLower(host), "www."), ".")
	target = strings.TrimSuffix(strings.TrimPrefix(strings.ToLower(target), "www."), ".")
	return host == target || strings.HasSuffix(host, "."+target)
}

func resultFileType(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	extension := strings.TrimPrefix(strings.ToLower(path.Ext(parsed.Path)), ".")
	return normalizeFileType(extension)
}

func fuzzyTokenCoverageScore(queryTokens []string, normalizedField string, maximum int) int {
	if len(queryTokens) == 0 || normalizedField == "" || maximum <= 0 {
		return 0
	}
	fieldTokens := uniqueTokens(normalizedField)
	exact := make(map[string]struct{}, len(fieldTokens))
	for _, token := range fieldTokens {
		exact[token] = struct{}{}
	}

	matched := 0
	for _, queryToken := range queryTokens {
		if _, ok := exact[queryToken]; ok {
			continue
		}
		if len([]rune(queryToken)) < 5 {
			continue
		}
		for _, fieldToken := range fieldTokens {
			if oneEditOrTranspositionApart(queryToken, fieldToken) {
				matched++
				break
			}
		}
	}
	return matched * maximum / len(queryTokens)
}

func oneEditOrTranspositionApart(left, right string) bool {
	if left == right {
		return false
	}
	leftRunes := []rune(left)
	rightRunes := []rune(right)
	if len(leftRunes) < 4 || len(rightRunes) < 4 {
		return false
	}
	if len(leftRunes)-len(rightRunes) > 1 || len(rightRunes)-len(leftRunes) > 1 {
		return false
	}

	if len(leftRunes) == len(rightRunes) {
		mismatches := make([]int, 0, 2)
		for index := range leftRunes {
			if leftRunes[index] != rightRunes[index] {
				mismatches = append(mismatches, index)
				if len(mismatches) > 2 {
					return false
				}
		}
		}
		if len(mismatches) == 1 {
			return true
		}
		return len(mismatches) == 2 &&
			mismatches[1] == mismatches[0]+1 &&
			leftRunes[mismatches[0]] == rightRunes[mismatches[1]] &&
			leftRunes[mismatches[1]] == rightRunes[mismatches[0]]
	}

	longer := leftRunes
	shorter := rightRunes
	if len(rightRunes) > len(leftRunes) {
		longer = rightRunes
		shorter = leftRunes
	}
	longIndex := 0
	shortIndex := 0
	skipped := false
	for longIndex < len(longer) && shortIndex < len(shorter) {
		if longer[longIndex] == shorter[shortIndex] {
			longIndex++
			shortIndex++
			continue
		}
		if skipped {
			return false
		}
		skipped = true
		longIndex++
	}
	return true
}
