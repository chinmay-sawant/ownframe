package eyes

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestSideTapMovesInEitherDirection(t *testing.T) {
	a, c := newT(t)
	tapHalf(t, a, .25)
	if a.st.mood != len(poses)-1 {
		t.Fatal("left tap did not select the previous expression immediately")
	}
	tk(t, a, c, doubleTapWindow+1)
	tapHalf(t, a, .75)
	tk(t, a, c, doubleTapWindow+1)
	if a.st.mood != 0 {
		t.Fatal("right tap did not select the next expression")
	}
}

func TestDoubleSideTapMovesOnce(t *testing.T) {
	for _, test := range []struct {
		name string
		x    float64
		want int
	}{{"left", .25, len(poses) - 1}, {"right", .75, 1}} {
		t.Run(test.name, func(t *testing.T) {
			a, c := newT(t)
			for range 2 {
				tapHalf(t, a, test.x)
				c.adv(90 * time.Millisecond)
			}
			if a.st.mood != test.want {
				t.Fatal("double tap should move one expression in its side's direction")
			}
		})
	}
}

func tapHalf(t *testing.T, a *App, x float64) {
	t.Helper()
	w, h := a.page.Size()
	box := ownframe.Box{ID: "eye-left"}
	_, err := a.onPress(t.Context(), box, float64(w)*x, float64(h)/2)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.onTap(t.Context(), box); err != nil {
		t.Fatal(err)
	}
}
