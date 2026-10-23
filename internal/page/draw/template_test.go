package draw

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestExecute(t *testing.T) {
	t.Parallel()

	got, err := Execute(false, 0, nil, func(b *strings.Builder) error {
		b.WriteString("hi")

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "hi" {
		t.Fatalf("%q", got)
	}

	var last time.Duration
	got, err = Execute(true, len("hi"), &last, func(b *strings.Builder) error {
		time.Sleep(time.Millisecond)
		b.WriteString("hi")

		return nil
	})
	if err != nil || got != "hi" {
		t.Fatal(err)
	}
	if last <= 0 {
		t.Fatal("tracked")
	}

	var idle time.Duration
	if _, err = Execute(false, 0, &idle, func(b *strings.Builder) error {
		b.WriteString("hi")

		return nil
	}); err != nil || idle != 0 {
		t.Fatal("untracked")
	}

	want := errors.New("run")
	got, err = Execute(true, len("no"), nil, func(b *strings.Builder) error {
		b.WriteString("no")

		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("%v", err)
	}
	if got != "" {
		t.Fatalf("%q", got)
	}
}
