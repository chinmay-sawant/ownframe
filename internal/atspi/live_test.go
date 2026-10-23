package atspi

import (
	"errors"
	"runtime"
	"testing"
)

func TestLiveBusNeverReportsConnected(t *testing.T) {
	bus := NewLiveBus()
	if bus.Live() {
		t.Fatal("live")
	}
	st := bus.Status()
	if st.Live || st.Platform != runtime.GOOS || st.Reason == "" {
		t.Fatalf("%+v", st)
	}
	if err := bus.Claim(DefaultName); !errors.Is(err, ErrLiveUnavailable) {
		t.Fatalf("claim %v", err)
	}
	if err := bus.Publish(Application{}); !errors.Is(err, ErrLiveUnavailable) {
		t.Fatalf("publish %v", err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}
