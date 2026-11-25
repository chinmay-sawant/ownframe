package draw

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func TestExecute(t *testing.T) {
	t.Parallel()

	got, err := Execute(false, 0, nil, func(b *bytes.Buffer) error {
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
	got, err = Execute(true, len("hi"), &last, func(b *bytes.Buffer) error {
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
	if _, err = Execute(false, 0, &idle, func(b *bytes.Buffer) error {
		b.WriteString("hi")

		return nil
	}); err != nil || idle != 0 {
		t.Fatal("untracked")
	}

	want := errors.New("run")
	got, err = Execute(true, len("no"), nil, func(b *bytes.Buffer) error {
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

func TestExecuteReuse(t *testing.T) {
	t.Parallel()

	prev := "<p>unchanged</p>"
	got, err := ExecuteReuse(false, len(prev), prev, nil, func(b *bytes.Buffer) error {
		b.WriteString(prev)

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != prev {
		t.Fatalf("got %q, want %q", got, prev)
	}

	// Changed output returns a fresh string and never aliases the pooled buffer.
	changed, err := ExecuteReuse(false, len(prev), prev, nil, func(b *bytes.Buffer) error {
		b.WriteString("<p>updated</p>")

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if changed != "<p>updated</p>" {
		t.Fatalf("got %q, want <p>updated</p>", changed)
	}

	// Overwriting the pooled buffer on the next call must not mutate changed.
	next, err := ExecuteReuse(false, len(prev), changed, nil, func(b *bytes.Buffer) error {
		b.WriteString("<p>third</p>")

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if changed != "<p>updated</p>" || next != "<p>third</p>" {
		t.Fatalf("buffer reuse aliased previous output: changed=%q next=%q", changed, next)
	}
}
