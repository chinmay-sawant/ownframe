package page

import (
	"context"
	"testing"
)

// TestPatchElementRefreshesTheAccessibilityTree checks that the tree
// reports the text the patch set.
func TestPatchElementRefreshesTheAccessibilityTree(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<p id="p">old</p>`)

	tree, err := p.AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	if got := accessibilityTestNode(t, tree, "id:p").Name; got != "old" {
		t.Fatalf("name %q", got)
	}

	text := "new"
	mustPatch(t, p.PatchElement(ctx, "p", ElementEdit{Text: &text}))

	tree, err = p.AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	node := accessibilityTestNode(t, tree, "id:p")
	if node.Name != "new" {
		t.Fatalf("name %q, want new", node.Name)
	}
	if tree.Document == 0 {
		t.Fatal("the tree has no document")
	}
}

// TestPatchElementClearsTheCachedPicture checks that the page does not serve
// the picture from before the patch.
func TestPatchElementClearsTheCachedPicture(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#p{width:40px;height:20px;background:#112233}</style><p id="p">x</p>`)

	if len(p.PNG()) == 0 {
		t.Fatal("the page has no picture")
	}

	mustPatch(t, p.PatchElement(ctx, "p", ElementEdit{Attribute: "style",
		Value: "width:40px;height:20px;background:#445566"}))
	if p.draw.PNG != nil {
		t.Fatal("the patch kept the cached picture")
	}
}
