package atspi

import (
	"errors"
	"testing"
)

func TestLiveBusDisabledBus(t *testing.T) {
	t.Setenv("AT_SPI_BUS_ADDRESS", "disabled")
	bus := NewLiveBus()
	if bus.Live() {
		t.Fatal("live")
	}
	st := bus.Status()
	if st.Live || st.Reason == "" {
		t.Fatalf("%+v", st)
	}
	if err := bus.Claim(DefaultName); !errors.Is(err, ErrLiveUnavailable) {
		t.Fatalf("claim %v", err)
	}
	if err := bus.Publish(Application{}); !errors.Is(err, ErrLiveUnavailable) {
		t.Fatalf("publish %v", err)
	}
	if err := bus.Emit(Event{Kind: "children-changed"}); !errors.Is(err, ErrLiveUnavailable) {
		t.Fatalf("emit %v", err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}
