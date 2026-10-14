package page

import (
	"testing"
	"time"
)

func TestWantFrameAtRoundTrip(t *testing.T) {
	p := &Page{}
	at := time.Unix(123, 456)
	p.WantFrameAt(at)

	if got := p.NextFrameAt(); !got.Equal(at) {
		t.Fatalf("NextFrameAt = %v, want %v", got, at)
	}
	if got := p.NextFrameAt(); !got.IsZero() {
		t.Fatalf("a second read = %v, want the zero time", got)
	}
}

func TestWantFrameInIsFuture(t *testing.T) {
	p := &Page{}
	p.WantFrameIn(50 * time.Millisecond)

	d := time.Until(p.NextFrameAt())
	if d <= 0 || d > 50*time.Millisecond {
		t.Fatalf("deadline %v out of range", d)
	}
}

func TestWantFrameAtZeroClears(t *testing.T) {
	p := &Page{}
	p.WantFrameAt(time.Unix(1, 0))
	p.WantFrameAt(time.Time{})

	if got := p.NextFrameAt(); !got.IsZero() {
		t.Fatalf("NextFrameAt = %v, want the zero time", got)
	}
}
