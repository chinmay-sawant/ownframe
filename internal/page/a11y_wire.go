package page

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/a11y"
)

// accessibilityState is the document formFocusState stores.
// The type lives in a11y. The alias keeps the name those files already use.
type accessibilityState = a11y.State

// AccessibilityTree returns current semantics in document order. Call it on
// the page's owning loop; native adapters must queue actions back to that loop.
func (p *Page) AccessibilityTree() (host.AccessibilityTree, error) {
	return a11y.Tree(p.a11yState(), p.accessibilityView())
}

func (p *Page) accessibilityView() a11y.View {
	controls := map[string]a11y.Control{}
	if st := p.formState(); st != nil {
		controls = make(map[string]a11y.Control, len(st.ByID))
		for id, c := range st.ByID {
			controls[id] = accessibilityControl(c)
		}
	}
	return a11y.View{
		Title:    p.title,
		Boxes:    p.boxes,
		Controls: controls,
		Focused:  p.FocusID(),
		Revision: p.generation,
	}
}

func accessibilityControl(c Control) a11y.Control {
	return a11y.Control{
		ID:       c.ID,
		Tag:      c.Tag,
		Type:     c.Type,
		Value:    c.Value,
		Checked:  c.Checked,
		Disabled: c.Disabled,
		Editable: canEdit(c),
	}
}

// AccessibilityPerform runs an advertised action on the page-owning loop.
func (p *Page) AccessibilityPerform(ctx context.Context, request host.AccessibilityRequest) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	tree, err := p.AccessibilityTree()
	if err != nil {
		return err
	}
	return a11y.Perform(
		ctx,
		tree,
		request,
		accessibilityActor{p},
	)
}

// accessibilityActor adapts Page to a11y.Actor.
type accessibilityActor struct{ p *Page }

func (a accessibilityActor) Focus(ctx context.Context, id string) error {
	return a.p.Focus(ctx, id)
}

func (a accessibilityActor) FocusID() string { return a.p.FocusID() }

func (a accessibilityActor) Control(id string) (a11y.Control, bool) {
	c, ok := a.p.control(id)
	if !ok {
		return a11y.Control{}, false
	}
	return accessibilityControl(c), true
}

func (a accessibilityActor) BoxByID(id string) Box { return a.p.boxByID(id) }

func (a accessibilityActor) Lookup(action string) (string, bool) {
	return a.p.nav.Lookup(action)
}

func (a accessibilityActor) Load(ctx context.Context, html string) error {
	return a.p.Load(ctx, html)
}

func (a accessibilityActor) ClickControl(ctx context.Context, box Box) error {
	return a.p.clickControl(ctx, box)
}

func (a accessibilityActor) IMEReplace(ctx context.Context, start, end int, insert string, caret int) error {
	return a.p.IMEReplace(
		ctx,
		start,
		end,
		insert,
		caret,
	)
}

func (p *Page) commitAccessibility(source string) {
	s := p.a11yState()
	if s == nil {
		s = &accessibilityState{}
		p.setA11yState(s)
	}
	s.Commit(source)
}

func (p *Page) invalidateAccessibilityDocument() {
	p.a11yState().Invalidate()
}

// refreshAccessibility puts the patch's edits into the accessibility tree and
// retires the document, so an old tree is not served as current. A tree no
// client asked for yet is built from the source when one does.
func (p *Page) refreshAccessibility(id string, edits []ElementEdit) {
	s := p.a11yState()
	if s == nil {
		return
	}
	s.Apply(id, accessibilityEdits(edits))
	p.invalidateAccessibilityDocument()
	p.commitAccessibility(s.Source())
}

func accessibilityEdits(edits []ElementEdit) []a11y.Edit {
	out := make([]a11y.Edit, len(edits))
	for i, e := range edits {
		out[i] = a11y.Edit{
			ID:        e.ID,
			Attribute: e.Attribute,
			Value:     e.Value,
			Text:      e.Text,
		}
	}
	return out
}
