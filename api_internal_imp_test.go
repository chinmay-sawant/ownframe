package ownframe_test

import (
	"go/ast"
	"strings"
	"testing"
)

func checkNoInternal(t *testing.T, file string, imps []*ast.ImportSpec) {
	t.Helper()
	for _, imp := range imps {
		if imp.Path == nil {
			continue
		}
		p := strings.Trim(imp.Path.Value, `"`)
		if strings.Contains(p, "/internal/") || strings.HasSuffix(p, "/internal") {
			t.Errorf("%s imports %s", file, p)
		}
	}
}
