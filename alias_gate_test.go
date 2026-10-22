package ownframe

import (
	"strings"
	"testing"
)

func TestRootAliasAllowlist(t *testing.T) {
	got, err := rootTypeAliases()
	if err != nil {
		t.Fatal(err)
	}
	allow := aliasAllow()
	files := aliasFiles(allow)
	for key := range got {
		file, _, _ := strings.Cut(key, ":")
		if !files[file] {
			t.Errorf("new alias file %s", file)
		}
		if !allow[key] {
			t.Errorf("new type alias %s", key)
		}
	}
	for key := range allow {
		if !got[key] {
			t.Errorf("allowlisted alias %s is gone; drop it", key)
		}
	}
}
