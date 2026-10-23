package patch

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

type editFake struct {
	text map[string]string
	attr map[string][2]string
}

func (f *editFake) SetText(id, text string) bool {
	if id == "" {
		return false
	}
	f.text[id] = text

	return true
}

func (f *editFake) SetAttribute(id, name, value string) bool {
	if id == "" || name == "" {
		return false
	}
	f.attr[id] = [2]string{name, value}

	return true
}

func TestElementEditApply(t *testing.T) {
	f := &editFake{
		text: map[string]string{},
		attr: map[string][2]string{},
	}
	text := "hi"
	if !Apply(f, "box", []ElementEdit{{Text: &text}, {ID: "kid", Attribute: "class", Value: "on"}}) {
		t.Fatal("apply reported no change")
	}
	if f.text["box"] != "hi" || f.attr["kid"] != [2]string{"class", "on"} {
		t.Fatalf("text %v attr %v", f.text, f.attr)
	}

	both := "x"
	if !Apply(f, "box", []ElementEdit{{Attribute: "id", Value: "nope", Text: &both}}) {
		t.Fatal("text-and-attribute edit reported no change")
	}
	if _, ok := f.attr["box"]; ok {
		t.Fatal("attribute won over text")
	}
	if f.text["box"] != "x" {
		t.Fatalf("text = %q", f.text["box"])
	}

	if Apply(f, "box", nil) || Apply(f, "box", []ElementEdit{{Value: "on"}}) {
		t.Fatal("empty edit reported a change")
	}
}

func TestElementReady(t *testing.T) {
	refused := errors.New("refused")
	err := Ready(refused, "", true, true, true)
	if !errors.Is(err, refused) || err.Error() != "empty id: refused" {
		t.Fatalf("empty = %v", err)
	}
	err = Ready(refused, "p", false, true, true)
	if !errors.Is(err, refused) || err.Error() != "no cache: refused" {
		t.Fatalf("cache = %v", err)
	}
	err = Ready(refused, "p", true, false, true)
	if !errors.Is(err, refused) || err.Error() != "no cache: refused" {
		t.Fatalf("display = %v", err)
	}
	err = Ready(refused, "p", true, true, false)
	if !errors.Is(err, refused) || err.Error() != "retain state off: refused" {
		t.Fatalf("retain = %v", err)
	}
	if err := Ready(refused, "p", true, true, true); err != nil {
		t.Fatal(err)
	}
}

func TestElementNoChange(t *testing.T) {
	refused := errors.New("refused")
	rolled := false
	err := ApplyElement(context.Background(), ElementIn{
		Refused:  refused,
		ID:       "p",
		Edits:    []ElementEdit{{Attribute: "class", Value: "on"}},
		Cache:    &render.Cache{},
		Display:  &layout.Display{},
		Retain:   true,
		Rollback: func() { rolled = true },
	})
	if !errors.Is(err, refused) || err.Error() != "no change: refused" || rolled {
		t.Fatalf("err = %v rolled %v", err, rolled)
	}
}

func TestElementRefusesBeforeRollback(t *testing.T) {
	refused := errors.New("refused")
	rolled := false
	err := ApplyElement(context.Background(), ElementIn{
		Refused:  refused,
		Display:  &layout.Display{},
		Cache:    &render.Cache{},
		Retain:   true,
		Rollback: func() { rolled = true },
	})
	if !errors.Is(err, refused) || err.Error() != "empty id: refused" || rolled {
		t.Fatalf("err = %v rolled %v", err, rolled)
	}

	err = Refuse(refused, nil)
	if err == nil || err.Error() != refused.Error() || !errors.Is(err, refused) {
		t.Fatalf("bare refuse = %v", err)
	}
	wrapped := Refuse(refused, errors.New("engine"))
	if !errors.Is(wrapped, refused) || wrapped.Error() != "refused: engine" {
		t.Fatalf("wrapped = %v", wrapped)
	}
}
