package remote

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestPhoneUsesFittedScreenWithoutScrolling(t *testing.T) {
	a := newTest(t, WithPhone(true))
	for _, tab := range []string{"tab-pad", "tab-nums", "tab-remote"} {
		clickID(t, a, tab)
		if !a.page.ViewLocked() || a.page.AllowScroll() {
			t.Fatalf("%s enabled scrolling or disabled fitting", tab)
		}
	}
}

func TestPhoneFocusHighlightsButtonWithoutBlueOutline(t *testing.T) {
	a := newTest(t, WithPhone(true))
	ctx := context.Background()
	if err := a.page.Focus(ctx, "tab-pad"); err != nil {
		t.Fatal(err)
	}
	for _, op := range a.page.Display().Ops {
		if op.Kind == layout.DisplayOpStrokeRect && op.R == 138.0/255 && op.G == 180.0/255 && op.B == 1 {
			t.Fatal("focused button has a blue outline")
		}
	}
}
