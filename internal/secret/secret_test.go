package secret

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestRedacted(t *testing.T) {
	t.Parallel()

	s := Secret("hunter2")
	for _, got := range []string{fmt.Sprint(s), fmt.Sprintf("%q", s), fmt.Sprintf("%x", s), s.GoString()} {
		if strings.Contains(got, "hunter2") {
			t.Fatalf("leaked in %q", got)
		}
	}

	if got := slog.AnyValue(s).String(); strings.Contains(got, "hunter2") {
		t.Fatalf("slog leaked %q", got)
	}

	if s.Reveal() != "hunter2" {
		t.Fatal("reveal broke")
	}
}
