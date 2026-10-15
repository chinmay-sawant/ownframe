package host

import (
	"context"
	"testing"
)

type pointerFake struct{}

func (pointerFake) Click(context.Context, float64, float64) error { return nil }
func (pointerFake) Hover(context.Context, float64, float64) error { return nil }
func (pointerFake) Press(context.Context, float64, float64) error { return nil }
func (pointerFake) Release(context.Context) error                 { return nil }

type keyboardFake struct{}

func (keyboardFake) Type(context.Context, string) error    { return nil }
func (keyboardFake) KeyDown(context.Context, string) error { return nil }
func (keyboardFake) KeyUp(context.Context, string) error   { return nil }
func (keyboardFake) Backspace(context.Context) error       { return nil }
func (keyboardFake) DeleteWord(context.Context) error      { return nil }
func (keyboardFake) Submit(context.Context) error          { return nil }

type clipboardFake struct{}

func (clipboardFake) Copy(context.Context) (string, bool, error) { return "", false, nil }
func (clipboardFake) Cut(context.Context) (string, bool, error)  { return "", false, nil }
func (clipboardFake) Paste(context.Context, string) error        { return nil }
func (clipboardFake) SelectAll(context.Context) error            { return nil }
func (clipboardFake) Undo(context.Context) error                 { return nil }
func (clipboardFake) Redo(context.Context) error                 { return nil }

var (
	_ Pointer   = pointerFake{}
	_ Keyboard  = keyboardFake{}
	_ Clipboard = clipboardFake{}
)

func TestRolesCoverOneSurfaceEach(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	if err := Pointer(pointerFake{}).Release(ctx); err != nil {
		t.Fatal(err)
	}
	if err := Keyboard(keyboardFake{}).Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := Clipboard(clipboardFake{}).Undo(ctx); err != nil {
		t.Fatal(err)
	}
}
