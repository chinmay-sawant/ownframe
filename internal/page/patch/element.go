package patch

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// ElementEdit is one change to an element in the retained tree: text, an
// attribute write or removal, a class-list mutation, an element removal, or
// an HTML fragment insertion/replacement. Leaving all action fields empty/nil
// changes nothing.
//
// ID names the element the edit applies to. An empty ID means the element
// the patch is for, and an edit on a descendant rides along with the patch.
// Text replaces the text content and drops the children. Remove detaches the
// target element from its parent (the patch root should be an ancestor that
// remains in the tree).
//
// Text and HTML fragments set this way do not go through the emoji pass. The
// pass runs on the source before it is parsed, and edits land on the parsed
// tree, so an emoji in patched text paints as text rather than as the color
// image the same emoji gets from the template.
type ElementEdit struct {
	ID              string
	Attribute       string
	Value           string
	Text            *string
	RemoveAttribute string
	AddClass        string
	RemoveClass     string
	ToggleClass     string
	Remove          bool
	InnerHTML       *string
	AppendHTML      string
}

// editor is the cached tree an element edit writes.
type editor interface {
	SetText(id, text string) bool
	SetAttribute(id, name, value string) bool
	RemoveAttribute(id, name string) bool
	AddClass(id, className string) bool
	RemoveClass(id, className string) bool
	ToggleClass(id, className string) (bool, bool)
	RemoveElement(id string) bool
	SetInnerHTML(parentID, fragment string) (bool, error)
	AppendHTML(parentID, fragment string) (bool, error)
}

// HasStructuralEdit reports whether any edit in edits changes child element
// structure (removing an element or replacing/appending child HTML).
func HasStructuralEdit(edits []ElementEdit) bool {
	for _, e := range edits {
		if e.Remove || e.InnerHTML != nil || e.AppendHTML != "" {
			return true
		}
	}

	return false
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
// Precedence when multiple fields are set on one ElementEdit is: Remove,
// InnerHTML, AppendHTML, Text, Attribute, RemoveAttribute, AddClass,
// RemoveClass, ToggleClass. changed is true when any write lands.
func Apply(ed editor, id string, edits []ElementEdit) (bool, error) {
	changed := false
	for _, e := range edits {
		target := e.ID
		if target == "" {
			target = id
		}
		switch {
		case e.Remove:
			changed = ed.RemoveElement(target) || changed
		case e.InnerHTML != nil:
			ok, err := ed.SetInnerHTML(target, *e.InnerHTML)
			if err != nil {
				return changed, err
			}
			changed = ok || changed
		case e.AppendHTML != "":
			ok, err := ed.AppendHTML(target, e.AppendHTML)
			if err != nil {
				return changed, err
			}
			changed = ok || changed
		case e.Text != nil:
			changed = ed.SetText(target, *e.Text) || changed
		case e.Attribute != "":
			changed = ed.SetAttribute(target, e.Attribute, e.Value) || changed
		case e.RemoveAttribute != "":
			changed = ed.RemoveAttribute(target, e.RemoveAttribute) || changed
		case e.AddClass != "":
			changed = ed.AddClass(target, e.AddClass) || changed
		case e.RemoveClass != "":
			changed = ed.RemoveClass(target, e.RemoveClass) || changed
		case e.ToggleClass != "":
			_, ok := ed.ToggleClass(target, e.ToggleClass)
			changed = ok || changed
		}
	}

	return changed, nil
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
	changed, err := Apply(in.Cache, in.ID, in.Edits)
	if err != nil {
		if changed {
			return failElement(in, err)
		}

		return Refuse(in.Refused, err)
	}
	if !changed {
		return fmt.Errorf("no change: %w", in.Refused)
	}

	state := in.State()
	styled, err := in.Cache.RelayoutFor(ctx, in.Width, in.Height, state)
	if err != nil {
		return failElement(in, err)
	}

	d, err := render.RelayoutElements(ctx, styled, in.Display, []string{in.ID}, layout.Options{Images: state.Images, RetainState: true})
	if err != nil {
		return failElement(in, err)
	}
	if d == nil {
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
