package page

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func synonymAllow() map[string]bool {
	return splitNames("DisplayPatch,ElementRegistry,FocusNext,FocusPrev,MarkRect,PNG,PatchReport,PressedID,Redo,Redraw,RedrawIfChanged,SelectLineAt,SelectWordAt,Undo,WantFrameIn")
}

func splitNames(s string) map[string]bool {
	m := make(map[string]bool)
	for _, n := range splitComma(s) {
		m[n] = true
	}
	return m
}

var errCaller = errors.New("page: unknown test caller")

func pageRecvName(fn *ast.FuncDecl) (string, bool) {
	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return "", false
	}
	f := fn.Recv.List[0]
	star, ok := f.Type.(*ast.StarExpr)
	if !ok {
		return "", false
	}
	id, ok := star.X.(*ast.Ident)
	if !ok || id.Name != "Page" || len(f.Names) != 1 {
		return "", false
	}
	return f.Names[0].Name, true
}

func isPageSynonym(recv string, body *ast.BlockStmt) bool {
	if len(body.List) != 1 {
		return false
	}
	switch s := body.List[0].(type) {
	case *ast.ExprStmt:
		return pageCall(recv, s.X)
	case *ast.ReturnStmt:
		return len(s.Results) == 1 && pageCall(recv, s.Results[0])
	}
	return false
}

func pageCall(recv string, e ast.Expr) bool {
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == recv
}

func TestExportSynonymGate(t *testing.T) {
	got, err := exportedPageSynonyms()
	if err != nil {
		t.Fatal(err)
	}
	allow := synonymAllow()
	for name := range got {
		if !allow[name] {
			t.Errorf("new Page synonym %s; do not add aliases", name)
		}
	}
	for name := range allow {
		if !got[name] {
			t.Errorf("allowlisted synonym %s is gone; drop it", name)
		}
	}
}

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

func splitComma(s string) []string {
	var out []string
	for _, n := range strings.Split(s, ",") {
		n = strings.TrimSpace(n)
		if n != "" {
			out = append(out, n)
		}
	}
	return out
}
