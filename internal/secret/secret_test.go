package secret

import (
	"encoding"
	"encoding/json"
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

func TestMarshalJSON(t *testing.T) {
	t.Parallel()

	s := Secret("hunter2")
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	if string(b) != `"`+redacted+`"` {
		t.Fatalf("json.Marshal got %s, want %s", string(b), `"`+redacted+`"`)
	}
	if strings.Contains(string(b), "hunter2") {
		t.Fatalf("json leaked plaintext")
	}

	type config struct {
		Password Secret `json:"password"`
	}
	cb, err := json.Marshal(config{Password: s})
	if err != nil {
		t.Fatalf("json.Marshal struct failed: %v", err)
	}
	if strings.Contains(string(cb), "hunter2") {
		t.Fatalf("json struct leaked plaintext: %s", string(cb))
	}
	if !strings.Contains(string(cb), `"`+redacted+`"`) {
		t.Fatalf("json struct missing redacted marker: %s", string(cb))
	}
}

func TestMarshalText(t *testing.T) {
	t.Parallel()

	s := Secret("hunter2")
	b, err := s.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText failed: %v", err)
	}
	if string(b) != redacted {
		t.Fatalf("MarshalText got %s, want %s", string(b), redacted)
	}
	if strings.Contains(string(b), "hunter2") {
		t.Fatalf("text leaked plaintext")
	}

	var tm encoding.TextMarshaler = s
	tmb, err := tm.MarshalText()
	if err != nil {
		t.Fatalf("TextMarshaler failed: %v", err)
	}
	if string(tmb) != redacted {
		t.Fatalf("TextMarshaler got %s, want %s", string(tmb), redacted)
	}
}
