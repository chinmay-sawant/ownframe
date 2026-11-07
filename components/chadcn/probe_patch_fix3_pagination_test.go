package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestProbePatchFix3Pagination covers the Pagination aria-current fix: every
// numbered control always renders aria-current ("page" on the current page,
// "false" elsewhere), so a page move inside a stable slot window is value
// edits only and takes the ui.View.Update patch path instead of falling back
// to a full document render. The identical rebuild is a NOOP.
func TestProbePatchFix3Pagination(t *testing.T) {
	t.Parallel()

	build := func(page int) *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.Pagination(chadcn.PaginationOptions{
			ID: "pp-pag3", Page: page, TotalPages: 5,
		}))
	}

	t.Run("IdenticalNoop", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, build(2))
		expectProbeNoop(t, v, ctx, build(2))
	})

	// SameWindowPageMovePatch flips: pages 2 and 3 mount the same five
	// numbered buttons in the same order with stable ids, and only the two
	// aria-current values change ("page"<->"false"), so the diff emits two
	// value edits and Update commits without reparsing.
	t.Run("SameWindowPageMovePatch", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, build(2))
		expectProbePatch(t, v, ctx, build(3))
		html := v.Page.HTML()
		if !strings.Contains(html, `aria-current="page" aria-label="Page 3"`) {
			t.Fatalf("page 3 is not current in %s", html)
		}
		if !strings.Contains(html, `aria-current="false" aria-label="Page 2"`) {
			t.Fatalf("page 2 did not flip to false in %s", html)
		}
	})
}
