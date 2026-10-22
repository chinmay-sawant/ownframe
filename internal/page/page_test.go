package page_test

import (
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestNewRejectsEmptyHTML(t *testing.T) {
	t.Parallel()

	_, err := page.New(page.Config{Width: 320, Height: 400})
	if !errors.Is(err, page.ErrEmptyHTML) {
		t.Fatalf("err = %v", err)
	}
}

func TestSetSizeClamps(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{
		HTML:      "<p>Hi</p>",
		Width:     480,
		Height:    640,
		MinWidth:  320,
		MinHeight: 400,
		MaxWidth:  800,
		MaxHeight: 900,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetSize(10, 10)
	width, height := screen.Size()
	if width != 320 || height != 400 {
		t.Fatalf("min clamp = %d x %d", width, height)
	}

	// The max is a window bound: the frame follows the window past it.
	screen.SetSize(5000, 5000)
	width, height = screen.Size()
	if width != 5000 || height != 5000 {
		t.Fatalf("frame past the max = %d x %d", width, height)
	}

	maxW, maxH := screen.MaxSize()
	if maxW != 800 || maxH != 900 {
		t.Fatalf("window bound = %d x %d", maxW, maxH)
	}
}

func TestNewRejectsMinAboveMax(t *testing.T) {
	t.Parallel()

	_, err := page.New(page.Config{
		HTML:      "<p>Hi</p>",
		Width:     480,
		Height:    640,
		MinWidth:  900,
		MaxWidth:  800,
		MinHeight: 400,
		MaxHeight: 900,
	})
	if !errors.Is(err, page.ErrBadSize) {
		t.Fatalf("err = %v, want ErrBadSize", err)
	}
}
