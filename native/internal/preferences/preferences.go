package preferences

import "sort"

type Scope string

const (
	ScopeLocal      Scope = "local"
	ScopeAccount    Scope = "account"
	ScopeDeployment Scope = "deployment"
)

type Kind string

const (
	KindBoolean Kind = "boolean"
	KindChoice  Kind = "choice"
	KindText    Kind = "text"
)

type Definition struct {
	Key         string   `json:"key"`
	Section     string   `json:"section"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Kind        Kind     `json:"kind"`
	Scope       Scope    `json:"scope"`
	Default     any      `json:"default"`
	Choices     []string `json:"choices,omitempty"`
}

var definitions = []Definition{
	{Key: "search.default_category", Section: "search", Label: "Default category", Kind: KindChoice, Scope: ScopeLocal, Default: "general", Choices: []string{"general", "images", "videos", "news", "files"}},
	{Key: "search.safe_search", Section: "search", Label: "SafeSearch", Kind: KindChoice, Scope: ScopeLocal, Default: "moderate", Choices: []string{"off", "moderate", "strict"}},
	{Key: "search.autocomplete", Section: "search", Label: "Search suggestions", Description: "Show query suggestions while typing.", Kind: KindBoolean, Scope: ScopeLocal, Default: false},
	{Key: "appearance.theme", Section: "appearance", Label: "Appearance", Kind: KindChoice, Scope: ScopeLocal, Default: "system", Choices: []string{"system", "light", "dark", "deep-dark"}},
	{Key: "appearance.result_density", Section: "appearance", Label: "Result density", Kind: KindChoice, Scope: ScopeLocal, Default: "comfortable", Choices: []string{"comfortable", "compact"}},
	{Key: "privacy.recent_queries", Section: "privacy", Label: "Recent searches", Description: "Keep recent queries on this device only.", Kind: KindBoolean, Scope: ScopeLocal, Default: false},
	{Key: "privacy.media_proxy", Section: "privacy", Label: "Private media loading", Description: "Use the approved Search media-proxy boundary when available.", Kind: KindBoolean, Scope: ScopeDeployment, Default: true},
	{Key: "security.safe_external_links", Section: "security", Label: "Safe external links", Kind: KindBoolean, Scope: ScopeDeployment, Default: true},
	{Key: "data.preference_schema", Section: "data-resilience", Label: "Preference schema", Kind: KindText, Scope: ScopeLocal, Default: "1"},
}

func Definitions() []Definition {
	items := append([]Definition(nil), definitions...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Section == items[j].Section {
			return items[i].Key < items[j].Key
		}
		return items[i].Section < items[j].Section
	})
	return items
}

func Find(key string) (Definition, bool) {
	for _, definition := range definitions {
		if definition.Key == key {
			return definition, true
		}
	}
	return Definition{}, false
}
