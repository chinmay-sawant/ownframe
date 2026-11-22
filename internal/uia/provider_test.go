package uia

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

type fakeClient struct {
	p *Provider
}

// walk visits the tree the way a UIA client does: FirstChild then NextSibling.
func (c fakeClient) walk(t *testing.T) []host.AccessibilityNodeID {
	t.Helper()
	var out []host.AccessibilityNodeID
	var visit func(id host.AccessibilityNodeID)
	visit = func(id host.AccessibilityNodeID) {
		out = append(out, id)
		child, ok, err := c.p.Navigate(id, FirstChild)
		if err != nil {
			t.Fatalf("FirstChild(%s): %v", id, err)
		}
		for ok {
			if parent, has, _ := c.p.Navigate(child, Parent); !has || parent != id {
				t.Fatalf("Parent(%s) = %q, want %q", child, parent, id)
			}
			visit(child)
			child, ok, err = c.p.Navigate(child, NextSibling)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	visit(c.p.Root())
	return out
}

type recorder struct {
	mu   sync.Mutex
	reqs []host.AccessibilityRequest
	err  error
}

func (r *recorder) perform(_ context.Context, req host.AccessibilityRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reqs = append(r.reqs, req)
	return r.err
}

func (r *recorder) last(t *testing.T) host.AccessibilityRequest {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.reqs) == 0 {
		t.Fatal("no request reached the page loop")
	}
	return r.reqs[len(r.reqs)-1]
}

func (r *recorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.reqs)
}

func sampleTree() host.AccessibilityTree {
	act := []host.AccessibilityActionKind{host.AccessibilityFocus, host.AccessibilityActivate}
	edit := []host.AccessibilityActionKind{host.AccessibilityFocus, host.AccessibilitySetValue}
	return host.AccessibilityTree{
		Document: 7, Revision: 1, Root: "root", Focused: "email",
		Nodes: []host.AccessibilityNode{
			{ID: "root", Role: host.AccessibilityDocument, Name: "Login", Children: []host.AccessibilityNodeID{"h", "go", "email", "pw", "ro", "tog", "r1", "dis"}},
			{ID: "h", Parent: "root", Role: host.AccessibilityHeading, Name: "Sign in", HasBounds: true, Bounds: host.AccessibilityBounds{X: 10, Y: 20, W: 100, H: 30}},
			{ID: "go", Parent: "root", Role: host.AccessibilityButton, Name: "Go", Description: "Submit", ElementID: "go-btn", Focusable: true, Actions: act, HasBounds: true, Bounds: host.AccessibilityBounds{X: 10, Y: 60, W: 50, H: 20}},
			{ID: "email", Parent: "root", Role: host.AccessibilityTextbox, Name: "Email", Value: "a@b.c", Focusable: true, Focused: true, Actions: edit},
			{ID: "pw", Parent: "root", Role: host.AccessibilityTextbox, Name: "Password", Value: "hunter2", Password: true, Focusable: true, Actions: []host.AccessibilityActionKind{host.AccessibilityFocus}},
			{ID: "ro", Parent: "root", Role: host.AccessibilityTextbox, Name: "Readonly", Value: "fixed", Actions: []host.AccessibilityActionKind{host.AccessibilityFocus}},
			{ID: "tog", Parent: "root", Role: host.AccessibilityCheckbox, Name: "Remember", HasChecked: true, Checked: true, Actions: act},
			{ID: "r1", Parent: "root", Role: host.AccessibilityRadio, Name: "Plan A", HasChecked: true, Actions: act},
			{ID: "dis", Parent: "root", Role: host.AccessibilityButton, Name: "Off", Disabled: true, Actions: act},
		},
	}
}

func newTest(t *testing.T) (*Provider, *recorder) {
	t.Helper()
	rec := &recorder{}
	p := NewProvider(rec.perform)
	p.Update(sampleTree())
	return p, rec
}

func mustEl(t *testing.T, p *Provider, id host.AccessibilityNodeID) Element {
	t.Helper()
	e, err := p.Element(id)
	if err != nil {
		t.Fatalf("Element(%s): %v", id, err)
	}
	return e
}

func TestWalkMatchesDocumentOrder(t *testing.T) {
	p, _ := newTest(t)
	got := fakeClient{p}.walk(t)
	want := []host.AccessibilityNodeID{"root", "h", "go", "email", "pw", "ro", "tog", "r1", "dis"}
	if !slices.Equal(got, want) {
		t.Fatalf("walk = %v, want %v", got, want)
	}
	if _, ok, _ := p.Navigate("root", Parent); ok {
		t.Fatal("root must not report a parent; the window provider owns it")
	}
	if last, _, _ := p.Navigate("root", LastChild); last != "dis" {
		t.Fatalf("LastChild = %q", last)
	}
	if prev, ok, _ := p.Navigate("dis", PreviousSibling); !ok || prev != "r1" {
		t.Fatalf("PreviousSibling = %q %v", prev, ok)
	}
	if _, ok, _ := p.Navigate("h", PreviousSibling); ok {
		t.Fatal("first child has no previous sibling")
	}
}

func TestControlTypesNamesAndState(t *testing.T) {
	p, _ := newTest(t)
	cases := []struct {
		id        host.AccessibilityNodeID
		ct        ControlType
		localized string
		focusable bool
	}{
		{"root", ControlDocument, "", false},
		{"h", ControlText, "heading", false},
		{"go", ControlButton, "", true},
		{"email", ControlEdit, "", true},
		{"tog", ControlCheckBox, "", false},
		{"r1", ControlRadioButton, "", false},
	}
	for _, c := range cases {
		e := mustEl(t, p, c.id)
		if e.ControlType != c.ct || e.LocalizedType != c.localized || e.Focusable != c.focusable {
			t.Errorf("%s: got type %d %q focusable=%v", c.id, e.ControlType, e.LocalizedType, e.Focusable)
		}
	}
	e := mustEl(t, p, "go")
	if e.Name != "Go" || e.HelpText != "Submit" || e.AutomationID != "go-btn" || !e.Enabled {
		t.Fatalf("go = %+v", e)
	}
	if !mustEl(t, p, "email").Focused || p.Focused() != "email" {
		t.Fatal("email should hold focus")
	}
	if mustEl(t, p, "dis").Enabled {
		t.Fatal("disabled button reports enabled")
	}
}

func TestPatternsPerRole(t *testing.T) {
	p, _ := newTest(t)
	type pat struct{ invoke, value, ro, toggle, sel bool }
	want := map[host.AccessibilityNodeID]pat{
		"go":    {invoke: true},
		"email": {value: true},
		"pw":    {},
		"ro":    {value: true, ro: true},
		"tog":   {toggle: true},
		"r1":    {sel: true},
		"h":     {},
	}
	for id, w := range want {
		e := mustEl(t, p, id)
		got := pat{e.Invoke, e.ValuePattern, e.ReadOnly && e.ValuePattern, e.Toggle, e.SelectionItem}
		if got != w {
			t.Errorf("%s: patterns %+v, want %+v", id, got, w)
		}
	}
	if tog := mustEl(t, p, "tog"); tog.Toggled != ToggleOn {
		t.Fatalf("toggle state = %d", tog.Toggled)
	}
}

func TestPasswordNeverExposed(t *testing.T) {
	p, rec := newTest(t)
	e := mustEl(t, p, "pw")
	if !e.Password || e.Value != "" || e.ValuePattern {
		t.Fatalf("password element leaks: %+v", e)
	}
	if err := p.SetValue("pw", "x"); !errors.Is(err, ErrNotSupported) {
		t.Fatalf("SetValue on password = %v", err)
	}
	if rec.count() != 0 {
		t.Fatal("password write reached the page loop")
	}
	// A hostile snapshot that forgot to blank the value is still redacted.
	tree := sampleTree()
	for i := range tree.Nodes {
		if tree.Nodes[i].ID == "pw" {
			tree.Nodes[i].Actions = append(tree.Nodes[i].Actions, host.AccessibilitySetValue)
		}
	}
	p.Update(tree)
	e = mustEl(t, p, "pw")
	if e.Value != "" || e.ValuePattern {
		t.Fatalf("password leaks after update: %+v", e)
	}
	if err := p.SetValue("pw", "x"); !errors.Is(err, ErrNotSupported) {
		t.Fatalf("SetValue on password with action = %v", err)
	}
	for id := range p.snap.byID {
		el := mustEl(t, p, id)
		if strings.Contains(el.Name+el.HelpText+el.Value, "hunter2") {
			t.Fatalf("%s exposes the secret", id)
		}
	}
}

func TestBoundsUseOriginAndScale(t *testing.T) {
	p, _ := newTest(t)
	p.SetGeometry(Geometry{OriginX: 100, OriginY: 200, Scale: 1.5, Width: 800, Height: 600})
	r, ok, err := p.Bounds("go")
	if err != nil || !ok {
		t.Fatalf("Bounds(go) = %v %v", ok, err)
	}
	if want := (Rect{X: 115, Y: 290, W: 75, H: 30}); r != want {
		t.Fatalf("Bounds(go) = %+v, want %+v", r, want)
	}
	if root, ok, _ := p.Bounds("root"); !ok || root != (Rect{100, 200, 800, 600}) {
		t.Fatalf("root bounds = %+v %v", root, ok)
	}
	if _, ok, _ := p.Bounds("email"); ok {
		t.Fatal("node without bounds must report none")
	}
}

func TestActionsDispatchOnPageLoop(t *testing.T) {
	p, rec := newTest(t)
	steps := []struct {
		name string
		do   func() error
		node host.AccessibilityNodeID
		kind host.AccessibilityActionKind
		val  string
	}{
		{"invoke", func() error { return p.Invoke("go") }, "go", host.AccessibilityActivate, ""},
		{"toggle", func() error { return p.Toggle("tog") }, "tog", host.AccessibilityActivate, ""},
		{"select", func() error { return p.Select("r1") }, "r1", host.AccessibilityActivate, ""},
		{"setvalue", func() error { return p.SetValue("email", "new@x.y") }, "email", host.AccessibilitySetValue, "new@x.y"},
		{"focus", func() error { return p.SetFocus("go") }, "go", host.AccessibilityFocus, ""},
	}
	for _, s := range steps {
		if err := s.do(); err != nil {
			t.Fatalf("%s: %v", s.name, err)
		}
		req := rec.last(t)
		if req.Document != 7 || req.Node != s.node || req.Kind != s.kind || req.Value != s.val {
			t.Errorf("%s: request %+v", s.name, req)
		}
	}
}

func TestActionGuards(t *testing.T) {
	p, rec := newTest(t)
	cases := []struct {
		name string
		err  error
		want error
	}{
		{"invoke disabled", p.Invoke("dis"), ErrNotEnabled},
		{"invoke a heading", p.Invoke("h"), ErrNotSupported},
		{"toggle a button", p.Toggle("go"), ErrNotSupported},
		{"set read-only", p.SetValue("ro", "x"), ErrReadOnly},
		{"set on button", p.SetValue("go", "x"), ErrNotSupported},
		{"missing node", p.Invoke("nope"), ErrNotAvailable},
		{"focus without action", p.SetFocus("h"), ErrNotSupported},
	}
	for _, c := range cases {
		if !errors.Is(c.err, c.want) {
			t.Errorf("%s = %v, want %v", c.name, c.err, c.want)
		}
	}
	if rec.count() != 0 {
		t.Fatalf("%d guarded calls reached the page loop", rec.count())
	}
	// Selecting the already-selected radio is a no-op.
	tree := sampleTree()
	for i := range tree.Nodes {
		if tree.Nodes[i].ID == "r1" {
			tree.Nodes[i].Checked = true
		}
	}
	p.Update(tree)
	if err := p.Select("r1"); err != nil || rec.count() != 0 {
		t.Fatalf("Select of selected radio: err=%v calls=%d", err, rec.count())
	}
}

func TestPageErrorsPropagate(t *testing.T) {
	p, rec := newTest(t)
	rec.err = host.ErrAccessibilityStale
	if err := p.Invoke("go"); !errors.Is(err, host.ErrAccessibilityStale) {
		t.Fatalf("Invoke = %v", err)
	}
	if err := NewProvider(nil).Invoke("go"); !errors.Is(err, ErrNotAvailable) {
		t.Fatalf("empty provider Invoke = %v", err)
	}
}

func TestDocumentChangeReplacesNodes(t *testing.T) {
	p, rec := newTest(t)
	before := mustEl(t, p, "go").RuntimeID
	next := sampleTree()
	next.Document, next.Revision = 8, 1
	next.Nodes = next.Nodes[:2]
	next.Nodes[0].Children = []host.AccessibilityNodeID{"h"}
	next.Focused = ""
	ch := p.Update(next)
	if !ch.Document || !ch.Structure {
		t.Fatalf("change = %+v", ch)
	}
	if _, err := p.Element("go"); !errors.Is(err, ErrNotAvailable) {
		t.Fatalf("stale node still resolves: %v", err)
	}
	if err := p.Invoke("go"); !errors.Is(err, ErrNotAvailable) || rec.count() != 0 {
		t.Fatalf("stale invoke = %v calls=%d", err, rec.count())
	}
	if p.Document() != 8 || mustEl(t, p, "h").RuntimeID == 0 {
		t.Fatal("new document not installed")
	}
	_ = before
}

func TestRevisionKeepsRuntimeIDsAndReportsChanges(t *testing.T) {
	p, _ := newTest(t)
	id := mustEl(t, p, "go").RuntimeID

	next := sampleTree()
	next.Revision = 2
	next.Focused = "go"
	for i := range next.Nodes {
		switch next.Nodes[i].ID {
		case "tog":
			next.Nodes[i].Checked = false
		case "email":
			next.Nodes[i].Value = "z"
			next.Nodes[i].Focused = false
		case "go":
			next.Nodes[i].Name = "Go now"
			next.Nodes[i].Focused = true
		}
	}
	ch := p.Update(next)
	if ch.Document || ch.Structure || ch.First {
		t.Fatalf("unexpected change flags: %+v", ch)
	}
	if mustEl(t, p, "go").RuntimeID != id {
		t.Fatal("runtime id changed across revisions")
	}
	if ch.Focus != "go" {
		t.Fatalf("Focus = %q", ch.Focus)
	}
	want := []PropertyChange{{"go", PropName}, {"email", PropValue}, {"tog", PropToggle}}
	for _, w := range want {
		if !slices.Contains(ch.Props, w) {
			t.Errorf("missing %+v in %+v", w, ch.Props)
		}
	}
	if len(ch.Props) != len(want) {
		t.Errorf("extra changes: %+v", ch.Props)
	}

	grow := next
	grow.Revision = 3
	grow.Nodes = append(slices.Clone(next.Nodes), host.AccessibilityNode{ID: "new", Parent: "root", Role: host.AccessibilityText})
	grow.Nodes[0].Children = append(slices.Clone(grow.Nodes[0].Children), "new")
	if ch := p.Update(grow); !ch.Structure || ch.Document {
		t.Fatalf("growing the tree = %+v", ch)
	}
}

func TestStopMakesProviderUnavailable(t *testing.T) {
	p, _ := newTest(t)
	p.Stop()
	if _, err := p.Element("go"); !errors.Is(err, ErrNotAvailable) {
		t.Fatalf("Element after Stop = %v", err)
	}
	if p.Root() != "" || p.Focused() != "" {
		t.Fatal("Stop must hide the tree")
	}
}

func TestConcurrentClientsAndPublisher(t *testing.T) {
	p, _ := newTest(t)
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 200 {
				_, _, _ = p.Navigate("root", FirstChild)
				_, _ = p.Element("go")
				_, _, _ = p.Bounds("go")
				_ = p.Invoke("go")
			}
		}()
	}
	for i := range 100 {
		tree := sampleTree()
		tree.Revision = uint64(i)
		p.Update(tree)
	}
	wg.Wait()
}
