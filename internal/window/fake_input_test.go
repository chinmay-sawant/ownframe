package window

import (
	"context"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

func (f *fakeScreen) Image() image.Image       { return nil }
func (f *fakeScreen) Display() *layout.Display { return nil }
func (f *fakeScreen) PNG() []byte              { return nil }
func (f *fakeScreen) Generation() uint64 {
	if f.redraws < 0 {
		return 0
	}
	return uint64(f.redraws)
}
func (f *fakeScreen) Boxes() []layout.Box                           { return f.boxes }
func (f *fakeScreen) Click(context.Context, float64, float64) error { return nil }
func (f *fakeScreen) Hover(context.Context, float64, float64) error { return nil }
func (f *fakeScreen) Press(context.Context, float64, float64) error { return nil }
func (f *fakeScreen) Release(context.Context) error                 { return nil }
func (f *fakeScreen) Type(context.Context, string) error            { return nil }
func (f *fakeScreen) KeyDown(context.Context, string) error         { return nil }
func (f *fakeScreen) KeyUp(context.Context, string) error           { return nil }
func (f *fakeScreen) Backspace(context.Context) error               { return nil }
func (f *fakeScreen) DeleteWord(context.Context) error              { return nil }
func (f *fakeScreen) Submit(context.Context) error                  { return nil }
func (f *fakeScreen) Copy(context.Context) (string, bool, error)    { return "", false, nil }
func (f *fakeScreen) Cut(context.Context) (string, bool, error)     { return "", false, nil }
func (f *fakeScreen) Paste(context.Context, string) error           { return nil }
func (f *fakeScreen) SelectAll(context.Context) error               { return nil }
func (f *fakeScreen) Undo(context.Context) error                    { return nil }
func (f *fakeScreen) Redo(context.Context) error                    { return nil }
