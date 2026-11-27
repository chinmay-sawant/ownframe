package patch

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

type editFake struct {
	text       map[string]string
	attr       map[string][2]string
	remAttr    map[string]string
	addClass   map[string]string
	remClass   map[string]string
	togClass   map[string]string
	removed    map[string]bool
	innerHTML  map[string]string
	appendHTML map[string]string
	htmlErr    error
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

func (f *editFake) RemoveAttribute(id, name string) bool {
	if id == "" || name == "" {
		return false
	}
	f.remAttr[id] = name

	return true
}

func (f *editFake) AddClass(id, className string) bool {
	if id == "" || className == "" {
		return false
	}
	f.addClass[id] = className

	return true
}

func (f *editFake) RemoveClass(id, className string) bool {
	if id == "" || className == "" {
		return false
	}
	f.remClass[id] = className

	return true
}

func (f *editFake) ToggleClass(id, className string) (bool, bool) {
	if id == "" || className == "" {
		return false, false
	}
	f.togClass[id] = className

	return true, true
}

func (f *editFake) RemoveElement(id string) bool {
	if id == "" {
		return false
	}
	f.removed[id] = true

	return true
}

func (f *editFake) SetInnerHTML(parentID, fragment string) (bool, error) {
	if f.htmlErr != nil {
		return false, f.htmlErr
	}
	if parentID == "" {
		return false, nil
	}
	f.innerHTML[parentID] = fragment

	return true, nil
}

func (f *editFake) AppendHTML(parentID, fragment string) (bool, error) {
	if f.htmlErr != nil {
		return false, f.htmlErr
	}
	if parentID == "" {
		return false, nil
	}
	f.appendHTML[parentID] = fragment

	return true, nil
}

func TestElementEditApply(t *testing.T) {
	f := &editFake{
		text:       map[string]string{},
		attr:       map[string][2]string{},
		remAttr:    map[string]string{},
		addClass:   map[string]string{},
		remClass:   map[string]string{},
		togClass:   map[string]string{},
		removed:    map[string]bool{},
		innerHTML:  map[string]string{},
		appendHTML: map[string]string{},
	}
	text := "hi"
	if ok, err := Apply(f, "box", []ElementEdit{{Text: &text}, {ID: "kid", Attribute: "class", Value: "on"}}); err != nil || !ok {
		t.Fatalf("apply = (%v, %v), want (true, nil)", ok, err)
	}
	if f.text["box"] != "hi" || f.attr["kid"] != [2]string{"class", "on"} {
		t.Fatalf("text %v attr %v", f.text, f.attr)
	}

	both := "x"
	if ok, err := Apply(f, "box", []ElementEdit{{Attribute: "id", Value: "nope", Text: &both}}); err != nil || !ok {
		t.Fatalf("text-and-attribute edit = (%v, %v)", ok, err)
	}
	if _, ok := f.attr["box"]; ok {
		t.Fatal("attribute won over text")
	}
	if f.text["box"] != "x" {
		t.Fatalf("text = %q", f.text["box"])
	}

	frag := "<span>new</span>"
	ok, err := Apply(f, "box", []ElementEdit{
		{ID: "a", RemoveAttribute: "disabled"},
		{ID: "b", AddClass: "active"},
		{ID: "c", RemoveClass: "stale"},
		{ID: "d", ToggleClass: "open"},
		{ID: "e", Remove: true},
		{ID: "f", InnerHTML: &frag},
		{ID: "g", AppendHTML: "<li>x</li>"},
	})
	if err != nil || !ok {
		t.Fatalf("extended apply = (%v, %v)", ok, err)
	}
	if f.remAttr["a"] != "disabled" || f.addClass["b"] != "active" ||
		f.remClass["c"] != "stale" || f.togClass["d"] != "open" ||
		!f.removed["e"] || f.innerHTML["f"] != frag || f.appendHTML["g"] != "<li>x</li>" {
		t.Fatalf("unexpected fake state: %+v", f)
	}

	if ok, err := Apply(f, "box", nil); err != nil || ok {
		t.Fatalf("nil edits = (%v, %v)", ok, err)
	}
	if ok, err := Apply(f, "box", []ElementEdit{{Value: "on"}}); err != nil || ok {
		t.Fatalf("empty edit = (%v, %v)", ok, err)
	}

	f.htmlErr = errors.New("parse fail")
	if _, err := Apply(f, "box", []ElementEdit{{AppendHTML: "<div>"}}); err == nil {
		t.Fatal("expected parse error from AppendHTML")
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
