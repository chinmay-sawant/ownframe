package page

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func exportedPageSynonyms() (map[string]bool, error) {
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errCaller
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
		collectSynonyms(f, got)
	}
	return got, nil
}

func collectSynonyms(f *ast.File, got map[string]bool) {
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name == nil || !fn.Name.IsExported() {
			continue
		}
		recv, ok := pageRecvName(fn)
		if !ok || fn.Body == nil {
			continue
		}
		if isPageSynonym(recv, fn.Body) {
			got[fn.Name.Name] = true
		}
	}
}
