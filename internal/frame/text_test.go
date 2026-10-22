package frame

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestTextFindsTheLabel(t *testing.T) {
	t.Parallel()

	op := Text(testDisplay(), page.Box{X: 10, Y: 10, W: 100, H: 20})
	if op == nil || op.Text != "1:08" {
		t.Fatalf("op = %+v", op)
	}

	op.Text = "1:09"
	if Text(testDisplay(), page.Box{X: 200, Y: 100, W: 10, H: 10}) != nil {
		t.Fatal("text found outside the box")
	}
}

func TestNilDisplay(t *testing.T) {
	t.Parallel()

	if Fill(nil, page.Box{}, [3]float64{}) != nil {
		t.Fatal("Fill on a nil display")
	}

	if Text(nil, page.Box{}) != nil {
		t.Fatal("Text on a nil display")
	}
}
