package patch

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// ElementEdit is one change to an element in the retained tree: an attribute
// to set, or the text to replace. A nil Text and an empty Attribute change
// nothing.
//
// ID names the element the edit applies to. An empty ID means the element
// the patch is for, and an edit on a descendant rides along with the patch.
// Text replaces the text content and drops the children.
//
// Text set this way does not go through the emoji pass. The pass runs on the
// source before it is parsed, and a text edit lands on the parsed tree, so an
// emoji in patched text paints as text rather than as the color image the
// same emoji gets from the template. The engine's tree API only writes a text
// node, so the image cannot be built inside the tree.
type ElementEdit struct {
	ID        string
	Attribute string
	Value     string
	Text      *string
}

// editor is the cached tree an element edit writes.
type editor interface {
	SetText(id, text string) bool
	SetAttribute(id, name, value string) bool
}

// ElementIn is one element patch. Display is the retained list. Edits are
// applied to Cache before the engine re-lays the element out. State is read
// after the edits. Adopt receives the patched display. Rollback runs when
// the engine refuses the edit, before the error returns.
type ElementIn struct {
	Refused  error
	ID       string
	Edits    []ElementEdit
	Cache    *render.Cache
	Display  *layout.Display
	Retain   bool
	Width    int
	Height   int
	State    func() render.State
	Adopt    func(*layout.Display)
	Rollback func()
}

// Apply writes edits into the cached tree. An empty edit ID uses id.
// Text wins when an edit sets both text and an attribute. changed is true
// when any write lands.
func Apply(ed editor, id string, edits []ElementEdit) bool {
	changed := false
	for _, e := range edits {
		target := e.ID
		if target == "" {
			target = id
		}
		switch {
		case e.Text != nil:
			changed = ed.SetText(target, *e.Text) || changed
		case e.Attribute != "":
			changed = ed.SetAttribute(target, e.Attribute, e.Value) || changed
		}
	}

	return changed
}

// Ready reports why a patch cannot start. refused is the sentinel callers
// unwrap. A nil page is the caller's check; Ready does not see it.
func Ready(refused error, id string, hasCache, hasDisplay, retain bool) error {
	if id == "" {
		return fmt.Errorf("empty id: %w", refused)
	}
	if !hasCache || !hasDisplay {
		return fmt.Errorf("no cache: %w", refused)
	}
	if !retain {
		return fmt.Errorf("retain state off: %w", refused)
	}

	return nil
}

// Refuse wraps err with refused. A nil err returns refused itself.
func Refuse(refused, err error) error {
	if err == nil {
		return refused
	}

	return fmt.Errorf("%w: %w", refused, err)
}

// AdoptedDisplay is the patched page, or false when the engine handed
// nothing back to draw.
func AdoptedDisplay(ep *layout.ElementPatch) (*layout.Display, bool) {
	if ep == nil || ep.Display == nil {
		return nil, false
	}

	return ep.Display, true
}

// ApplyElement edits one element and re-lays it out. A refused patch does
// not commit the styled document. The caller rolls the tree back through
// ElementIn.Rollback.
func ApplyElement(ctx context.Context, in ElementIn) error {
	if err := Ready(in.Refused, in.ID, in.Cache != nil, in.Display != nil, in.Retain); err != nil {
		return err
	}
	if !Apply(in.Cache, in.ID, in.Edits) {
		return fmt.Errorf("no change: %w", in.Refused)
	}

	state := in.State()
	styled, err := in.Cache.RelayoutFor(ctx, in.Width, in.Height, state)
	if err != nil {
		return failElement(in, err)
	}

	ep, err := layout.RelayoutElement(ctx, styled, in.Display, in.ID, layout.Options{Images: state.Images})
	if err != nil {
		return failElement(in, err)
	}

	d, ok := AdoptedDisplay(ep)
	if !ok {
		return failElement(in, nil)
	}

	in.Adopt(d)
	in.Cache.SetStyled(styled)

	return nil
}

func failElement(in ElementIn, err error) error {
	in.Rollback()

	return Refuse(in.Refused, err)
}
