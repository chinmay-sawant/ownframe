package workbook

import (
	"errors"
	"testing"
)

func TestPending(t *testing.T) {
	p := NewPending(2)

	if p.Capacity() != 2 || p.Len() != 0 || p.Full() {
		t.Fatalf("fresh pending = %+v", p)
	}

	first := Command{Sheet: 1, Edits: []CellEdit{{Pos: Pos{0, 0}, Cell: ParseInput("1")}}}
	second := Command{Sheet: 1, Edits: []CellEdit{{Pos: Pos{0, 1}, Cell: ParseInput("2")}}}
	third := Command{Sheet: 1, Edits: []CellEdit{{Pos: Pos{0, 2}, Cell: ParseInput("3")}}}

	if err := p.Push(first); err != nil {
		t.Fatal(err)
	}

	if err := p.Push(second); err != nil {
		t.Fatal(err)
	}

	if err := p.Push(third); !errors.Is(err, ErrPendingFull) {
		t.Fatalf("push past capacity = %v, want ErrPendingFull", err)
	}

	if p.Len() != 2 || !p.Full() {
		t.Fatalf("queue dropped a command: len=%d", p.Len())
	}

	got, ok := p.Pop()
	if !ok || got.Edits[0].Pos != first.Edits[0].Pos {
		t.Fatalf("pop = %+v", got)
	}

	got, ok = p.Pop()
	if !ok || got.Edits[0].Pos != second.Edits[0].Pos {
		t.Fatalf("pop order = %+v", got)
	}

	if _, ok := p.Pop(); ok {
		t.Fatal("pop from empty queue succeeded")
	}

	if err := p.Push(third); err != nil {
		t.Fatalf("push after drain = %v", err)
	}
}
