package ownframe_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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

func TestInternalNoTestingImport(t *testing.T) {
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller unknown")
	}
	root := filepath.Dir(self)
	fset := token.NewFileSet()
	// Dedicated test-helper packages that are never imported by runtime code.
	allow := map[string]bool{
		"internal/perfcheck/perfcheck.go": true,
	}
	for _, sub := range []string{"internal", "cmd"} {
		base := filepath.Join(root, sub)
		err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, imp := range f.Imports {
				unquoted, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					return err
				}
				if unquoted == "testing" {
					rel, _ := filepath.Rel(root, path)
					rel = filepath.ToSlash(rel)
					if !allow[rel] {
						t.Errorf("%s imports testing in a non-test file", rel)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestInternalNoRootOwnframeImport(t *testing.T) {
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller unknown")
	}
	root := filepath.Dir(self)
	fset := token.NewFileSet()
	base := filepath.Join(root, "internal")
	err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			unquoted, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if unquoted == "github.com/chinmay-sawant/ownframe" {
				rel, _ := filepath.Rel(root, path)
				rel = filepath.ToSlash(rel)
				t.Errorf("%s imports root ownframe in a non-test file", rel)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
