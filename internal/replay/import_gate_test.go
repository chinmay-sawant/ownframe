package replay

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const replayBannedRender = "github.com/chinmay-sawant/ownframe/internal/render"

var replayRenderAllow = map[string]bool{}

func TestReplayRenderGate(t *testing.T) {
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
		if e.IsDir() || !strings.HasSuffix(n, ".go") {
			continue
		}
		if strings.HasSuffix(n, "_test.go") {
			continue
		}
		if replayRenderAllow[n] {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, n), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if p == replayBannedRender {
				t.Errorf("%s imports %s; keep replay below render", n, p)
			}
		}
	}
}
