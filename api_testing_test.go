package ownframe_test

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestTestingTickTakeDirty(t *testing.T) {
	p, err := ownframe.New(ownframe.Config{
		HTML:   `<p id="n">hi</p>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.Probe().TakeDirty()
	var calls int
	p.SetTick(func(context.Context) error {
		calls++
		p.Probe().Invalidate("n")
		return nil
	})
	if err := p.Testing().Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("tick callback calls = %d", calls)
	}
	if rect, ok := p.Probe().TakeDirty(); !ok || rect.Empty() {
		t.Fatal("tick did not expose its invalidated region")
	}
	if _, ok := p.Probe().TakeDirty(); ok {
		t.Fatal("dirty region was not drained")
	}
}

func TestTestingClickAfterRedraw(t *testing.T) {
	p, err := ownframe.New(ownframe.Config{
		HTML:   `<button id="go">n</button>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	var hit string
	p.Handle(ownframe.Handlers{
		Click: func(_ context.Context, b ownframe.Box) error {
			hit = b.ID
			return nil
		},
	})
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	var x, y float64
	for _, b := range p.Boxes() {
		if b.ID == "go" {
			x, y = b.X+b.W/2, b.Y+b.H/2
		}
	}
	if err := p.Click(context.Background(), x, y); err != nil {
		t.Fatal(err)
	}
	if hit != "go" {
		t.Fatalf("hit = %q", hit)
	}
}

func TestNoUnauthorizedPageTestingCalls(t *testing.T) {
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller unknown")
	}
	root := filepath.Dir(self)

	// Pre-existing experimental demo harnesses that still call .Testing() in
	// non-test files. Freeze this set so no new non-test file adopts Testing().
	allowExamples := map[string]bool{
		"examples/blackhole/main.go":             true,
		"examples/blinkless-all/nav_scripted.go": true,
		"examples/blinkless-all/snapshot.go":     true,
		"examples/blinkless-scroll/demo.go":      true,
		"examples/blinkless-scroll/main.go":      true,
	}

	seen := map[string]bool{}
	fset := token.NewFileSet()
	walk := func(dir string, allow map[string]bool) {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "temp" || d.Name() == "frontend" || d.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			f, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			var callsTesting bool
			ast.Inspect(f, func(n ast.Node) bool {
				if callsTesting {
					return false
				}
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if ok && sel.Sel.Name == "Testing" && len(call.Args) == 0 {
					callsTesting = true
					return false
				}
				return true
			})
			if !callsTesting {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			rel = filepath.ToSlash(rel)
			if allow != nil && allow[rel] {
				seen[rel] = true
				return nil
			}
			t.Errorf("%s calls .Testing() in a non-test file", rel)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, sub := range []string{"internal", "components", "ui"} {
		walk(filepath.Join(root, sub), nil)
	}
	walk(filepath.Join(root, "examples"), allowExamples)

	for path := range allowExamples {
		if !seen[path] {
			t.Errorf("allowlisted example %s no longer calls .Testing(); remove it from allowExamples", path)
		}
	}
}
