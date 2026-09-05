package preferences

import "testing"

func TestAppearanceIncludesStableGlazeV11Modes(t *testing.T) {
	definition, ok := Find("appearance.theme")
	if !ok {
		t.Fatal("missing appearance preference")
	}
	if definition.Default != "system" {
		t.Fatalf("appearance default = %v, want system", definition.Default)
	}
	required := map[string]bool{
		"system":    false,
		"light":     false,
		"dark":      false,
		"deep-dark": false,
	}
	for _, choice := range definition.Choices {
		if _, exists := required[choice]; exists {
			required[choice] = true
		}
	}
	for choice, present := range required {
		if !present {
			t.Fatalf("Stable Glaze V1.1 appearance %q is not represented", choice)
		}
	}
}
