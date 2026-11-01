package main

import (
	"bytes"
	"go/format"
	"strings"
	"testing"
)

func TestGeneratedStylesGrowWithRegistry(t *testing.T) {
	props := []string{}
	seen := make(map[string]bool)
	for prop := range keywordSpecs() {
		seen[prop] = true
	}
	for _, list := range valueFamilies() {
		for _, prop := range list {
			seen[prop] = true
		}
	}
	for prop := range seen {
		props = append(props, prop)
	}
	props = append(props, "future-property")
	var out bytes.Buffer
	out.WriteString("package ui\n")
	if err := generateStyles(&out, props); err != nil {
		t.Fatal(err)
	}
	if _, err := format.Source(out.Bytes()); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"FutureProperty(v FuturePropertyValue)", `setValue("future-property", v)`} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("new registry entry missing %q", want)
		}
	}
}

func TestGeneratorRejectsAmbiguousNamesAndStaleMetadata(t *testing.T) {
	for _, props := range [][]string{{"-vendor-test", "vendor-test"}, {"display"}} {
		var out bytes.Buffer
		if err := generateStyles(&out, props); err == nil {
			t.Fatal("expected invalid schema error")
		}
	}
}
