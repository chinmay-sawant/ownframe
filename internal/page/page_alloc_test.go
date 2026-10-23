package page

import (
	"testing"
	"time"
)

func TestAllocDueGap(t *testing.T) {
	t.Parallel()

	p := &Page{}
	start := time.Unix(0, 0)
	if !p.allocDue(start) {
		t.Fatal("first sample")
	}
	if p.allocDue(start.Add(499 * time.Millisecond)) {
		t.Fatal("inside the gap")
	}
	if !p.allocDue(start.Add(allocSampleGap)) {
		t.Fatal("at the gap")
	}
}
