package ownframe

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var errAliasCaller = errors.New("ownframe: unknown test caller")

func rootTypeAliases() (map[string]bool, error) {
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errAliasCaller
	}
	dir := filepath.Dir(self)
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	got := make(map[string]bool)
	for _, e := range ents {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, n), nil, 0)
		if err != nil {
			return nil, err
		}
		collectAliases(n, f, got)
	}
	return got, nil
}

func collectAliases(file string, f *ast.File, got map[string]bool) {
	for _, d := range f.Decls {
		gen, ok := d.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Assign == 0 || ts.Name == nil {
				continue
			}
			got[file+":"+ts.Name.Name] = true
		}
	}
}
