package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestProbePatchFix2Codeblock locks the incremental patch path for CodeBlock
// code-text edits: the code body carries a stable id derived from ID, so a
// same-shape text edit commits one generation without reparsing, and an
// identical rebuild is a no-op.
func TestProbePatchFix2Codeblock(t *testing.T) {
	t.Parallel()

	code := func(text string) *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.CodeBlock(chadcn.CodeBlockOptions{
			ID: "code-1", Filename: "main.go", Language: "go", Code: text,
		}))
	}

	t.Run("code-text-edit-patches", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, code("a := 1"))
		expectProbePatch(t, v, ctx, code("a := 2"))
	})

	t.Run("identical-noop", func(t *testing.T) {
		t.Parallel()
		options := chadcn.CodeBlockOptions{ID: "code-1", Filename: "main.go", Language: "go", Code: "a := 1"}
		root := func() *ui.Element {
			return ui.Div().ID("probe-root").Children(chadcn.CodeBlock(options))
		}
		v, ctx := mountProbeView(t, root())
		expectProbeNoop(t, v, ctx, root())
	})
}
