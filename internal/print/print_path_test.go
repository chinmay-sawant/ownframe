package print

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestPrintPathTable(t *testing.T) {
	cases := []struct {
		name string
		path string
		ok   bool
	}{
		{"absolute", "/tmp/r.pdf", true},
		{"abs dot", "/tmp/./r.pdf", true},
		{"benign relative", "reports/out.pdf", true},
		{"dot relative", "./out.pdf", true},
		{"inner dot", "a/./b.pdf", true},
		{"traversal", "../secret.pdf", false},
		{"embedded", "a/../../secret.pdf", false},
		{"control", "a\nb.pdf", false},
		{"empty", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidPath(tc.path)
			if tc.ok && err != nil {
				t.Fatalf("ValidPath(%q) = %v", tc.path, err)
			}
			if !tc.ok && !errors.Is(err, ErrBadPath) {
				t.Fatalf("ValidPath(%q) = %v, want ErrBadPath", tc.path, err)
			}
		})
	}
}

func TestRunCanonicalizesRelative(t *testing.T) {
	calls := fakeRunner(t, map[string]bool{"lp": true})
	rel := filepath.Join("reports", "out.pdf")
	if err := run(context.Background(), rel, linuxCommands); err != nil {
		t.Fatal(err)
	}
	got := *calls
	if len(got) != 1 {
		t.Fatalf("calls = %+v", got)
	}
	want, _ := canonicalPath(rel)
	if got[0].args[1] != want || !filepath.IsAbs(got[0].args[1]) {
		t.Fatalf("args = %v, want %q", got[0].args, want)
	}
	if err := run(context.Background(), "../secret.pdf", linuxCommands); !errors.Is(err, ErrBadPath) {
		t.Fatalf("err = %v, want ErrBadPath", err)
	}
}
