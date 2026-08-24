package preferences

import "testing"

func TestDefinitionsHaveUniqueKeysAndValidScopes(t *testing.T) {
	seen := map[string]bool{}
	for _, definition := range Definitions() {
		if definition.Key == "" || definition.Section == "" || definition.Label == "" {
			t.Fatalf("incomplete preference definition: %#v", definition)
		}
		if seen[definition.Key] {
			t.Fatalf("duplicate preference key %q", definition.Key)
		}
		seen[definition.Key] = true
		switch definition.Scope {
		case ScopeLocal, ScopeAccount, ScopeDeployment:
		default:
			t.Fatalf("invalid scope %q", definition.Scope)
		}
		if definition.Kind == KindChoice && len(definition.Choices) == 0 {
			t.Fatalf("choice preference %q has no choices", definition.Key)
		}
	}
}

func TestPrivacySensitiveConveniencesDefaultOff(t *testing.T) {
	for _, key := range []string{"search.autocomplete", "privacy.recent_queries"} {
		definition, ok := Find(key)
		if !ok {
			t.Fatalf("missing %s", key)
		}
		value, ok := definition.Default.(bool)
		if !ok || value {
			t.Fatalf("%s must default off", key)
		}
	}
}

func TestRequiredSearchCategoriesRemainRepresented(t *testing.T) {
	definition, ok := Find("search.default_category")
	if !ok {
		t.Fatal("missing default category preference")
	}
	required := map[string]bool{"general": false, "images": false, "videos": false, "news": false, "files": false}
	for _, choice := range definition.Choices {
		if _, exists := required[choice]; exists {
			required[choice] = true
		}
	}
	for category, present := range required {
		if !present {
			t.Fatalf("required category %q is not represented", category)
		}
	}
}
