package ownframe_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOwnframeTestNoInternal(t *testing.T) {
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller unknown")
	}
	dir := filepath.Dir(self)
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, e := range ents {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, "_test.go") {
			continue
		}
		path := filepath.Join(dir, n)
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		if f.Name.Name != "ownframe_test" {
			continue
		}
		checkNoInternal(t, n, f.Imports)
	}
}
