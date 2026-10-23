package patch

import (
	"sync"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestPatchCompareReport(t *testing.T) {
	op := layout.DisplayOp{W: 4, H: 4}
	before := &layout.Display{Width: 40, Height: 40, Ops: []layout.DisplayOp{op}}
	next := &layout.Display{Width: 40, Height: 40, Ops: []layout.DisplayOp{op}}
	same := Compare(before, next)
	if same.Patch.Changed != 0 || same.Patch.Full || same.N != 0 {
		t.Fatalf("same = %+v", same)
	}

	next.Ops[0].X = 8
	moved := Compare(before, next)
	if moved.Patch.Changed != 1 || moved.Patch.Full || moved.N == 0 {
		t.Fatalf("moved = %+v", moved)
	}

	next.Width = 10
	full := Compare(before, next)
	if !full.Patch.Full || full.Patch.Regions != 1 {
		t.Fatalf("full = %+v", full.Patch)
	}
}

func TestPatchPropsShareBacking(t *testing.T) {
	props := []string{"count"}
	got := WithProps(DisplayPatch{Changed: 1}, props)
	if len(got.ChangedProperties) != 1 || got.ChangedProperties[0] != "count" {
		t.Fatalf("props = %v", got.ChangedProperties)
	}
	if &got.Properties[0] != &got.ChangedProperties[0] ||
		&got.Props[0] != &got.ChangedProperties[0] ||
		&got.ChangedProps[0] != &got.ChangedProperties[0] {
		t.Fatal("deprecated fields do not share ChangedProperties")
	}

	got.ChangedProperties[0] = "other"
	if props[0] != "count" {
		t.Fatal("WithProps aliased the caller's slice")
	}
}

func TestPatchOpsPool(t *testing.T) {
	var pool sync.Pool
	pool.New = func() any {
		b := EmptyOps()
		return &b
	}

	buf := Rent(&pool, 1)
	if cap(buf) < opCap {
		t.Fatalf("cap = %d", cap(buf))
	}
	buf = append(buf, layout.DisplayOp{})
	Free(&pool, buf)

	again := Rent(&pool, 1)
	if cap(again) < opCap || len(again) != 0 {
		t.Fatalf("reused cap %d len %d", cap(again), len(again))
	}
}

func TestPatchOpDamage(t *testing.T) {
	d := &layout.Display{Width: 20, Height: 20, Ops: []layout.DisplayOp{{W: 4, H: 4}}}
	if _, ok := OpDamage(d, 3, func(*layout.DisplayOp) {}); ok {
		t.Fatal("out of range reported success")
	}
	if _, ok := OpDamage(nil, 0, func(*layout.DisplayOp) {}); ok {
		t.Fatal("nil display reported success")
	}
	if _, ok := OpDamage(d, 0, nil); ok {
		t.Fatal("nil func reported success")
	}

	r, ok := OpDamage(d, 0, func(o *layout.DisplayOp) { o.X = 6 })
	if !ok || r.Empty() {
		t.Fatalf("damage = %v %v", r, ok)
	}
	if d.Ops[0].X != 6 {
		t.Fatalf("op = %+v", d.Ops[0])
	}
}
