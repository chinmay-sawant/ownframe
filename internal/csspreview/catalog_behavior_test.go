package csspreview

import (
	"strings"
	"testing"
)

func TestCatalogDescribesTheVisibleResult(t *testing.T) {
	cases := map[string]string{
		"animation":             "glides right",
		"-webkit-animation":     "hops upward",
		"animation-duration":    "1.6 seconds",
		"animation-range-start": "start point",
		"animation-timeline":    "preview clock",
		"animation-trigger":     "Start launches",
		"transition":            "horizontally",
		"transition-property":   "Expands",
		"transition-duration":   "Fades",
		"transition-delay":      "coral to blue",
		"transition-behavior":   "allow-discrete",
		"view-transition-name":  "view-transition name",
	}
	for name, expected := range cases {
		key, action := catalogAction(name)
		if key == "" || !strings.Contains(strings.ToLower(action), strings.ToLower(expected)) {
			t.Errorf("%s: key=%q action=%q, want %q", name, key, action, expected)
		}
	}
	for _, name := range catalogProperties {
		if key, action := catalogAction(name); key == "" || action == "" {
			t.Errorf("%s has no working preview description", name)
		}
	}
}
