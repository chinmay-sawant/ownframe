package main

import (
	"context"
	"strings"
	"testing"
)

// first returns the index of the first step whose changes include want, or -1.
func first(logs []StepLog, want string) int {
	for i, l := range logs {
		for _, c := range l.Changes {
			if c == want {
				return i
			}
		}
	}

	return -1
}

func TestDemoScript(t *testing.T) {
	logs, err := runDemo(context.Background(), fixture(t), 880, 700, "", false, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(logs) < 60 {
		t.Fatalf("only %d steps", len(logs))
	}

	// Scroll progress moves: the bar reaches 1.00 at the bottom.
	if first(logs, "scroll-bar --page=1.00") < 0 {
		t.Fatal("scroll bar never reached 1.00")
	}

	// The view-timeline trigger turns on at its range and off again later.
	on, off := first(logs, "trigger-box trigger --show=on"), first(logs, "trigger-box trigger --show=off")
	if on < 0 || off < 0 {
		t.Fatalf("trigger-box on=%d off=%d", on, off)
	}

	// The scoped trigger fires too, which needs timeline-scope to resolve.
	if first(logs, "trigger-longhand trigger --show2=on") < 0 {
		t.Fatal("scoped trigger never turned on")
	}

	// Pointer timeline: inside the box it has a value, away it is idle.
	if first(logs, "pointer-box --ptr=0.50") < 0 && first(logs, "pointer-box --ptr=0.49") < 0 {
		t.Logf("pointer middle changes: %v", logs[stepNamed(t, logs, "pointer-box pointer middle")].Changes)
		t.Fatal("pointer timeline did not read about 0.5 at the middle")
	}

	// Events: click arms the click-only trigger and Enter leaves it armed.
	click := stepNamed(t, logs, "event-box click")
	if !has2(logs[click].Changes, "event-box event --press=armed") {
		t.Fatalf("click did not arm event-box: %v", logs[click].Changes)
	}

	// The longhand trigger starts on click and Enter ends it.
	c2 := stepNamed(t, logs, "event-box-longhand click")
	if !has2(logs[c2].Changes, "event-box-longhand event --press2=armed") {
		t.Fatalf("click did not arm longhand: %v", logs[c2].Changes)
	}

	if !has2(logs[stepNamed(t, logs, "event-box-longhand press Enter")].Changes, "event-box-longhand event --press2=cleared") {
		t.Fatal("Enter did not clear the longhand trigger")
	}

	// Inert panel takes no hit; auto panel does.
	if l := logs[stepNamed(t, logs, "inert-panel click")]; !l.Clicked || l.Hit == "inert-panel" {
		t.Fatalf("inert panel hit: %+v", l)
	}

	if l := logs[stepNamed(t, logs, "auto-panel click")]; l.Hit != "auto-panel" {
		t.Fatalf("auto panel hit: %+v", l)
	}

	// Interest: not started right after the hover begins (the clock has moved
	// less than the start delay), started after it, cleared after the end delay.
	for _, id := range []string{"interest-card", "interest-card-longhand"} {
		start := first(logs, id+" interest=started")
		clear := first(logs, id+" interest=cleared")
		hoverFirst := stepNamed(t, logs, id+" hover")

		if start <= hoverFirst {
			t.Fatalf("%s: interest started at step %d, hover began at %d (must wait the delay)", id, start, hoverFirst)
		}

		if clear <= start || !strings.HasSuffix(logs[clear].Label, "leave") {
			t.Fatalf("%s: cleared at %d (%q), want a leave step after start %d", id, clear, logs[clear].Label, start)
		}
	}

	// The 200ms delay needs at least 2 hover steps at 100ms each.
	h := stepNamed(t, logs, "interest-card hover")
	if first(logs, "interest-card interest=started") < h+2 {
		t.Fatal("interest-card started before its 200ms delay")
	}
}

func stepNamed(t *testing.T, logs []StepLog, label string) int {
	t.Helper()

	for i, l := range logs {
		if l.Label == label {
			return i
		}
	}

	t.Fatalf("no step %q", label)

	return -1
}

func has2(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}

	return false
}
