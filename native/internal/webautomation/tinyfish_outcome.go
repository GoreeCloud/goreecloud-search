package webautomation

import (
	"bytes"
	"encoding/json"
	"strings"
)

// classifyTinyFishCompletedOutcome distinguishes provider execution completion
// from successful completion of the requested GoreeCloud goal. TinyFish
// documents that a run may reach COMPLETED while its result still describes a
// CAPTCHA, access denial, authentication problem, or another task failure.
//
// This check is intentionally conservative. It recognizes explicit structured
// failure fields and a narrow set of unambiguous free-form failure statements;
// it does not treat every occurrence of words such as "blocked" or "login" as
// a failure because successful diagnostic output may legitimately mention them.
func classifyTinyFishCompletedOutcome(raw json.RawMessage) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}

	var text string
	if err := json.Unmarshal(trimmed, &text); err == nil {
		return classifyTinyFishOutcomeText(text)
	}

	var value any
	if err := json.Unmarshal(trimmed, &value); err != nil {
		// Normal output validation will reject malformed JSON separately.
		return nil
	}
	return classifyTinyFishOutcomeValue(value)
}

func classifyTinyFishOutcomeValue(value any) error {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			normalizedKey := strings.ToLower(strings.TrimSpace(key))
			switch normalizedKey {
			case "goal_succeeded", "goal_success", "succeeded", "success":
				if flag, ok := child.(bool); ok && !flag {
					return ErrGoalFailed
				}
			case "status", "outcome", "result_status":
				if state, ok := child.(string); ok {
					if err := classifyTinyFishOutcomeState(state); err != nil {
						return err
					}
				}
			}
		}
		for _, child := range typed {
			if err := classifyTinyFishOutcomeValue(child); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range typed {
			if err := classifyTinyFishOutcomeValue(child); err != nil {
				return err
			}
		}
	case string:
		return classifyTinyFishOutcomeText(typed)
	}
	return nil
}

func classifyTinyFishOutcomeState(raw string) error {
	state := normalizeOutcomeToken(raw)
	switch state {
	case "blocked", "site_blocked", "captcha", "access_denied", "bot_blocked":
		return ErrSiteBlocked
	case "failed", "failure", "task_failed", "goal_failed", "incomplete", "not_authenticated", "authentication_required", "unauthorized":
		return ErrGoalFailed
	default:
		return nil
	}
}

func classifyTinyFishOutcomeText(raw string) error {
	text := strings.ToLower(strings.TrimSpace(raw))
	if text == "" {
		return nil
	}

	// Explicit site-barrier statements. Avoid generic keyword matching so
	// successful diagnostics such as "no CAPTCHA was present" remain valid.
	siteBarrierPhrases := []string{
		"the site presented a captcha",
		"the page presented a captcha",
		"encountered a captcha",
		"blocked by a captcha",
		"the site returned access denied",
		"the page returned access denied",
		"encountered access denied",
		"blocked by the site",
		"site blocked the automation",
		"bot block prevented",
	}
	for _, phrase := range siteBarrierPhrases {
		if strings.Contains(text, phrase) {
			return ErrSiteBlocked
		}
	}

	// Explicit task/authentication failure statements observed in provider
	// completion output. These are intentionally phrased as inability/failure
	// statements rather than generic authentication vocabulary.
	goalFailurePhrases := []string{
		"the task cannot be completed",
		"this task cannot be completed",
		"cannot complete this task",
		"i cannot complete this task",
		"i cannot complete the task",
		"unable to complete this task",
		"unable to complete the task",
		"could not complete this task",
		"could not complete the task",
		"no credentials are configured",
		"no saved credentials are available",
		"browser session is not authenticated",
		"session is not authenticated",
		"not authenticated to github",
		"not logged into github",
		"not signed in to github",
		"not authenticated to cloudflare",
		"not logged into cloudflare",
		"not signed in to cloudflare",
	}
	for _, phrase := range goalFailurePhrases {
		if strings.Contains(text, phrase) {
			return ErrGoalFailed
		}
	}
	return nil
}

func normalizeOutcomeToken(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	value = strings.NewReplacer("-", "_", " ", "_").Replace(value)
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	return value
}
