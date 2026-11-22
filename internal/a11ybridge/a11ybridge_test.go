package a11ybridge_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/a11ybridge"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

const testHTML = `<main>
<h1>Sign in</h1>
<label for="name">Name</label><input id="name" value="Ada">
<label for="pw">Password</label><input id="pw" type="password" value="hunter2">
<button id="go">Go</button>
</main>`

type decoded struct {
	Doc   int64  `json:"doc"`
	Rev   int64  `json:"rev"`
	Root  string `json:"root"`
	Focus string `json:"focus"`
	Nodes []struct {
		ID        string  `json:"id"`
		Element   string  `json:"element"`
		Role      string  `json:"role"`
		Name      string  `json:"name"`
		Value     string  `json:"value"`
		Password  bool    `json:"password"`
		Focusable bool    `json:"focusable"`
		Actions   []int   `json:"actions"`
		HasBounds bool    `json:"bounds"`
		W         float64 `json:"w"`
	} `json:"nodes"`
}

// newPage builds a real page and wires the bridge the way the window loop
// does: the encoded tree is published, and actions reach the page.
func newPage(t *testing.T) *ownframe.Page {
	t.Helper()
	t.Cleanup(a11ybridge.Reset)

	page, err := ownframe.New(ownframe.Config{HTML: testHTML, Width: 400, Height: 300})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := page.Redraw(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	a11ybridge.SetPerformer(func(ctx context.Context, req host.AccessibilityRequest) error {
		return page.Access().AccessibilityPerform(ctx, req)
	})
	publish(t, page)

	return page
}

func publish(t *testing.T, page *ownframe.Page) {
	t.Helper()

	tree, err := page.Access().AccessibilityTree()
	if err != nil {
		t.Fatalf("AccessibilityTree: %v", err)
	}

	a11ybridge.SetTree(a11ybridge.Encode(tree))
}

func decode(t *testing.T) decoded {
	t.Helper()

	var d decoded
	if err := json.Unmarshal([]byte(a11ybridge.Tree()), &d); err != nil {
		t.Fatalf("Tree is not valid JSON: %v\n%s", err, a11ybridge.Tree())
	}

	return d
}

func nodeFor(t *testing.T, d decoded, element string) string {
	t.Helper()

	for _, n := range d.Nodes {
		if n.Element == element {
			return n.ID
		}
	}

	t.Fatalf("no node for element %q", element)

	return ""
}

func TestTreeCarriesTheControls(t *testing.T) {
	newPage(t)
	d := decode(t)

	if d.Root == "" || len(d.Nodes) == 0 {
		t.Fatalf("empty tree: %+v", d)
	}

	var button, textbox bool

	for _, n := range d.Nodes {
		switch {
		case n.Role == "button" && n.Name == "Go":
			button = true

			if !n.Focusable || len(n.Actions) == 0 || !n.HasBounds || n.W <= 0 {
				t.Errorf("button lacks focus, actions, or bounds: %+v", n)
			}
		case n.Role == "textbox" && n.Name == "Name":
			textbox = true

			if n.Value != "Ada" {
				t.Errorf("textbox value = %q, want Ada", n.Value)
			}
		}
	}

	if !button || !textbox {
		t.Fatalf("button=%t textbox=%t in %+v", button, textbox, d.Nodes)
	}
}

func TestTreeNeverCarriesAPasswordValue(t *testing.T) {
	newPage(t)

	if strings.Contains(a11ybridge.Tree(), "hunter2") {
		t.Fatal("the password text appears in the tree JSON")
	}

	for _, n := range decode(t).Nodes {
		if n.Password && n.Value != "" {
			t.Fatalf("password node %q carries a value", n.ID)
		}
	}
}

func TestTreeBeforeAnyPageIsParsable(t *testing.T) {
	a11ybridge.Reset()

	var d decoded
	if err := json.Unmarshal([]byte(a11ybridge.Tree()), &d); err != nil {
		t.Fatalf("empty tree is not valid JSON: %v", err)
	}

	if len(d.Nodes) != 0 {
		t.Fatalf("empty tree has nodes: %+v", d.Nodes)
	}
}

func TestPerformFocusMovesFocus(t *testing.T) {
	page := newPage(t)
	d := decode(t)
	name := nodeFor(t, d, "name")

	if msg := a11ybridge.Perform(d.Doc, name, a11ybridge.ActionFocus, ""); msg != "" {
		t.Fatalf("Perform focus: %s", msg)
	}

	publish(t, page)

	if got := decode(t).Focus; got != name {
		t.Fatalf("focus = %q after focusing %q", got, name)
	}
}

func TestPerformRefusesBadRequests(t *testing.T) {
	newPage(t)
	d := decode(t)
	button, password := nodeFor(t, d, "go"), nodeFor(t, d, "pw")

	cases := map[string]string{
		"unknown action":     a11ybridge.Perform(d.Doc, button, 99, ""),
		"unknown node":       a11ybridge.Perform(d.Doc, "no-such-node", a11ybridge.ActionActivate, ""),
		"stale document":     a11ybridge.Perform(d.Doc+5, button, a11ybridge.ActionActivate, ""),
		"password set value": a11ybridge.Perform(d.Doc, password, a11ybridge.ActionSetValue, "x"),
	}

	for name, msg := range cases {
		if msg == "" {
			t.Errorf("%s: Perform succeeded, want a refusal", name)
		}
	}
}

func TestPerformWithoutAPageIsRefused(t *testing.T) {
	a11ybridge.Reset()

	if msg := a11ybridge.Perform(0, "x", a11ybridge.ActionFocus, ""); msg == "" {
		t.Fatal("Perform with no page attached must refuse")
	}
}

func TestPerformReportsThePagesError(t *testing.T) {
	t.Cleanup(a11ybridge.Reset)

	a11ybridge.SetPerformer(func(context.Context, host.AccessibilityRequest) error {
		return errors.New("page said no")
	})

	if msg := a11ybridge.Perform(1, "n", a11ybridge.ActionActivate, ""); msg != "page said no" {
		t.Fatalf("Perform = %q, want the page's message", msg)
	}
}
