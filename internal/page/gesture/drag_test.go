package gesture

import (
	"errors"
	"testing"
)

func TestDragBeginMoveEnd(t *testing.T) {
	t.Parallel()

	var d Drag
	next, dx, dy, send := d.Move(4, 5)
	if send || dx != 0 || dy != 0 || next.active {
		t.Fatalf("idle move %+v %v %v %v", next, dx, dy, send)
	}

	d = d.Begin(1, 1)
	d, dx, dy, send = d.Move(1, 1)
	if !d.active || send || dx != 0 || dy != 0 || d.x != 1 || d.y != 1 {
		t.Fatalf("still point %+v %v %v %v", d, dx, dy, send)
	}

	d, dx, dy, send = d.Move(4, -2)
	if !send || dx != 3 || dy != -3 || d.x != 4 || d.y != -2 {
		t.Fatalf("delta %+v %v %v %v", d, dx, dy, send)
	}

	d, active := d.End()
	if !active || d.active {
		t.Fatalf("end active=%v drag=%+v", active, d)
	}

	_, active = d.End()
	if active {
		t.Fatal("second end was active")
	}
}

func TestDragClaimed(t *testing.T) {
	t.Parallel()

	if !Claimed(nil, true) {
		t.Fatal("clean claim")
	}

	if Claimed(errors.New("no"), true) || Claimed(nil, false) {
		t.Fatal("error or refusal claimed the drag")
	}
}
