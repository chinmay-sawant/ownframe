package scene

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// newTestSceneHTML builds a scene on a custom template source.
func newTestSceneHTML(t *testing.T, m Model, st Store, opts Options, src string) (*Scene, *testClock) {
	t.Helper()

	clock := &testClock{at: time.Unix(1000, 0)}
	if opts.Now == nil {
		opts.Now = clock.now
	}

	s, err := newFromHTML(m, st, opts, src)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return s, clock
}

// hasText reports whether the display carries a text run containing want.
func hasText(s *Scene, want string) bool {
	d := s.page.Display()
	if d == nil {
		return false
	}

	for i := range d.Ops {
		if d.Ops[i].Kind == ownframe.DisplayOpText && strings.Contains(d.Ops[i].Text, want) {
			return true
		}
	}

	return false
}

// press sends a key press through the page handlers.
func press(t *testing.T, s *Scene, key string) {
	t.Helper()

	if err := s.page.KeyDown(context.Background(), key); err != nil {
		t.Fatal(err)
	}
}
