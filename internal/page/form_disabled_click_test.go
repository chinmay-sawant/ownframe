package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestDisabledControlsSkipHandler(t *testing.T) {
	t.Parallel()
	check := func(t *testing.T, body, id string, val func(*page.Page) string) {
		t.Helper()
		s := formPage(t, body)
		calls := 0
		s.Handle(page.Handlers{Click: func(context.Context, page.Box) error { calls++; return nil }})
		page.InstallPicker(s, func(context.Context, string) (string, bool) { calls++; return "/tmp/x.txt", true })
		before := val(s)
		gen := s.Generation()
		clickID(t, s, id)
		if calls != 0 {
			t.Fatalf("%s calls = %d, want 0", id, calls)
		}
		if got := val(s); got != before {
			t.Fatalf("%s value = %q, want %q", id, got, before)
		}
		if s.Generation() != gen {
			t.Fatalf("%s gen moved", id)
		}
	}
	t.Run("checkbox", func(t *testing.T) {
		t.Parallel()
		check(t, `<input id="c" type="checkbox" disabled>`, "c", func(s *page.Page) string {
			if s.FormChecked("c") {
				return "1"
			}
			return "0"
		})
	})
	t.Run("text", func(t *testing.T) {
		t.Parallel()
		check(t, `<input id="t" type="text" value="keep" disabled`+wide+`>`, "t", func(s *page.Page) string { return s.FormValue("t") })
	})
	t.Run("file", func(t *testing.T) {
		t.Parallel()
		check(t, `<input id="f" type="file" disabled`+wide+`>`, "f", func(s *page.Page) string { return s.FormValue("f") })
	})
}
