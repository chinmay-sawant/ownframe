package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestReadingFlowReordersByPosition(t *testing.T) {
	m := demoModel(t, readDemoHTML)

	for parent, want := range map[string][]string{
		"r-src":  {"s-a", "s-b"},
		"r-rows": {"p2", "p1", "p3"},
		"r-cols": {"q2", "q3", "q1"},
		"r-flex": {"r-c", "r-b", "r-a"},
	} {
		if got := readIDs(m, parent); !reflect.DeepEqual(got, want) {
			t.Errorf("%s reads %v, want %v", parent, got, want)
		}
	}

	// Without the property the order is the source order.
	plain := demoModel(t, strings.ReplaceAll(readDemoHTML, "reading-flow:grid-rows;", ""))
	if got := readIDs(plain, "r-rows"); !reflect.DeepEqual(got, []string{"p1", "p2", "p3"}) {
		t.Errorf("no reading-flow reads %v", got)
	}
}

func TestReadingOrderSortsByNumber(t *testing.T) {
	m := demoModel(t, readDemoHTML)

	if got := readIDs(m, "r-ord"); !reflect.DeepEqual(got, []string{"o2", "o3", "o1"}) {
		t.Errorf("reading-order reads %v", got)
	}

	// reading-order is ignored when the container keeps reading-flow normal.
	off := demoModel(t, strings.Replace(readDemoHTML, "reading-flow:source-order;width:300px;background:#f7f3e3", "width:300px", 1))
	if got := readIDs(off, "r-ord"); !reflect.DeepEqual(got, []string{"o1", "o2", "o3"}) {
		t.Errorf("without reading-flow reads %v", got)
	}

	// Descendants follow their parent in the sequence.
	nested := demoModel(t, `<div id="c" style="reading-flow:source-order"><div id="x" style="reading-order:2"><p id="x1">a</p></div><div id="y" style="reading-order:1">b</div></div>`)

	var seq []string
	for _, i := range nested.readingSequence() {
		seq = append(seq, nested.Items[i].ID)
	}

	if got := strings.Join(seq, ","); !strings.Contains(got, "y,x,x1") {
		t.Errorf("sequence %s", got)
	}
}

func TestReadOverlayAttachmentAndTbd(t *testing.T) {
	m := demoModel(t, readDemoHTML)
	decl := func(id, p string) string { return m.Items[m.ByID(id)].Declared[p] }

	if !readTopLayer(false, decl("ov-auto", "overlay")) || readTopLayer(false, decl("ov-none", "overlay")) || !readTopLayer(true, "none") {
		t.Error("overlay keeps a closing element in the top layer only for auto")
	}

	if decl("bg-local", "background-attachment") != "local, fixed" {
		t.Errorf("background-attachment stored %q", decl("bg-local", "background-attachment"))
	}

	if readBackgroundShift("fixed", 100, 30) != 100 || readBackgroundShift("local", 100, 30) != -30 || readBackgroundShift("scroll", 100, 30) != 0 {
		t.Error("background shift per attachment")
	}

	if decl("bg-tbd", "background-tbd") == "" {
		t.Error("background-tbd must reach the host")
	}
}

func TestReadDemosShowTheSequence(t *testing.T) {
	if d := readDemoFlow(); !strings.Contains(d, "1 p2") || !strings.Contains(d, "2 p1") {
		t.Error("flow demo lacks the numbered chips")
	}

	if d := readDemoOrder(); !strings.Contains(d, "1 o2") {
		t.Error("order demo lacks the numbered chips")
	}
}
