package atspi

import (
	"errors"
	"testing"
)

func TestFakeBusRecordsClaimAndPublish(t *testing.T) {
	bus := NewFakeBus()
	if err := bus.Claim("org.a11y.Bus"); !errors.Is(err, ErrName) {
		t.Fatalf("registry: %v", err)
	}
	if err := bus.Claim(DefaultName); err != nil {
		t.Fatal(err)
	}
	app := Map(sampleTree())
	if err := bus.Publish(app); err != nil {
		t.Fatal(err)
	}
	if err := bus.Emit(Event{Path: app.RootPath, Kind: "children-changed"}); err != nil {
		t.Fatal(err)
	}
	if got := bus.Claimed(); len(got) != 1 || got[0] != DefaultName {
		t.Fatalf("claimed %v", got)
	}
	if len(bus.Apps()) != 1 || len(bus.Events()) != 1 {
		t.Fatal("record")
	}
	if !bus.Live() {
		t.Fatal("fake live")
	}
	if err := bus.Close(); err != nil || bus.Live() {
		t.Fatal("close")
	}
}

func TestCheckNameRejectsForeignBus(t *testing.T) {
	for _, name := range []string{
		"", "org.a11y.Bus", "org.a11y.atspi.Registry",
		"org.a11y.atspi.ownframe.", "org.a11y.atspi.ownframe.app/x",
		"org.a11y.atspi.ownframe.org.a11y.Bus",
	} {
		if err := CheckName(name); !errors.Is(err, ErrName) {
			t.Fatalf("%q: %v", name, err)
		}
	}
	if err := CheckName("org.a11y.atspi.ownframe.test"); err != nil {
		t.Fatal(err)
	}
}
