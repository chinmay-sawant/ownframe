package main

import (
	"strings"
	"testing"
)

func TestCatalogDescribesTheVisibleResult(t *testing.T) {
	cases := map[string]string{
		"animation":                         "diagonally",
		"-webkit-animation":                 "diagonally",
		"-webkit-animation-delay":           "350ms",
		"-webkit-animation-direction":       "backward",
		"-webkit-animation-duration":        "1.2 seconds",
		"-webkit-animation-fill-mode":       "keeps its coral finish",
		"-webkit-animation-iteration-count": "three times",
		"animation-duration":                "1.6 seconds",
		"animation-range-start":             "start range",
		"animation-timeline":                "preview clock",
		"animation-trigger":                 "Start launches",
		"transition":                        "right",
		"transition-property":               "Expands",
		"transition-duration":               "Fades",
		"transition-delay":                  "orange tile to teal",
		"transition-behavior":               "allow-discrete",
		"view-transition-name":              "transition name",
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
