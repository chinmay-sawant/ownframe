package host

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// screenFake proves Screen still builds from the three roles: a fake
// embeds one role each and stubs the window and picture surface.
type screenFake struct {
	Pointer
	Keyboard
	Clipboard
}

func (screenFake) Boxes() []layout.Box          { return nil }
func (screenFake) Title() string                { return "" }
func (screenFake) Size() (int, int)             { return 0, 0 }
func (screenFake) MinSize() (int, int)          { return 0, 0 }
func (screenFake) Clamp(w, h int) (int, int)    { return w, h }
func (screenFake) SetSize(int, int)             {}
func (screenFake) Redraw(context.Context) error { return nil }
func (screenFake) Image() image.Image           { return nil }
func (screenFake) Display() *layout.Display     { return nil }
func (screenFake) PNG() []byte                  { return nil }
func (screenFake) Generation() uint64           { return 0 }

var _ Screen = screenFake{}

func TestScreenNarrowsToRoles(t *testing.T) {
	t.Parallel()
	s := Screen(screenFake{pointerFake{}, keyboardFake{}, clipboardFake{}})
	ctx := context.Background()
	if err := Pointer(s).Press(ctx, 1, 2); err != nil {
		t.Fatal(err)
	}
	if err := Keyboard(s).Type(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if err := Clipboard(s).Paste(ctx, "a"); err != nil {
		t.Fatal(err)
	}
}
