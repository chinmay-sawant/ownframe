package atspi

import (
	"strings"
	"testing"
)

func TestSummarizeOmitsAccessibleText(t *testing.T) {
	obj := byID(t, Map(sampleTree()), "id:pw")
	s := Summarize(obj)
	if strings.Contains(s, "secret") || strings.Contains(s, "Password") {
		t.Fatalf("leaked %q", s)
	}
	if !strings.Contains(s, string(RolePasswordText)) || !strings.Contains(s, "password") {
		t.Fatalf("summary %q", s)
	}
	name := byID(t, Map(sampleTree()), "id:name")
	got := Summarize(name)
	if strings.Contains(got, "Ada") || strings.Contains(got, "Name") {
		t.Fatalf("leaked %q", got)
	}
}
