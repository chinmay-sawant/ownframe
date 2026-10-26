package main

import (
	"strings"
	"testing"
)

func TestDiffWrapUnified(t *testing.T) {
	t.Parallel()

	if got := wrapDiffRows(nil, 10, 6, true); got != nil {
		t.Fatalf("nil rows = %+v, want nil", got)
	}

	if got := wrapDiffRows([]diffRow{}, 10, 6, true); len(got) != 0 {
		t.Fatalf("empty rows = %+v, want none", got)
	}

	rows := []diffRow{
		{Kind: rowAdd, NewNo: "12", Left: []segment{{Text: "one two three four", Class: classPlain}}},
		{Kind: rowDel, OldNo: "9", Left: []segment{{Text: "abcdefghijklmno", Class: classPlain}}},
		{Kind: rowCtx, OldNo: "3", NewNo: "4", Left: []segment{{Text: "1234567890", Class: classPlain}}},
		{Kind: rowCtx, OldNo: "5", NewNo: "6"},
	}

	got := wrapDiffRows(rows, 10, 6, true)
	if len(got) != 6 {
		t.Fatalf("visual rows = %d, want 6: %+v", len(got), got)
	}

	add := got[:2]
	if add[0].Kind != rowAdd || add[0].NewNo != "12" || add[0].OldNo != "" || add[0].Cont {
		t.Fatalf("first add row = %+v", add[0])
	}

	if add[1].Kind != rowAdd || add[1].NewNo != "" || add[1].OldNo != "" || !add[1].Cont {
		t.Fatalf("add continuation = %+v", add[1])
	}

	if text := lineText(add[0].Left); text != "one two " {
		t.Fatalf("first add chunk = %q, want a space break", text)
	}

	if text := lineText(add[1].Left); text != "three four" {
		t.Fatalf("second add chunk = %q", text)
	}

	del := got[2:4]
	if del[0].Kind != rowDel || del[0].OldNo != "9" || del[0].Cont {
		t.Fatalf("first del row = %+v", del[0])
	}

	if del[1].Kind != rowDel || del[1].OldNo != "" || del[1].NewNo != "" || !del[1].Cont {
		t.Fatalf("del continuation = %+v", del[1])
	}

	if text := lineText(del[0].Left); text != "abcdefghij" {
		t.Fatalf("first del chunk = %q, want a hard break", text)
	}

	if text := lineText(del[1].Left); text != "klmno" {
		t.Fatalf("second del chunk = %q", text)
	}

	exact := got[4]
	if exact.Kind != rowCtx || exact.OldNo != "3" || exact.NewNo != "4" || exact.Cont {
		t.Fatalf("exact-width row = %+v", exact)
	}

	if text := lineText(exact.Left); text != "1234567890" {
		t.Fatalf("exact-width text = %q, want no ellipsis", text)
	}

	empty := got[5]
	if empty.Kind != rowCtx || empty.OldNo != "5" || empty.NewNo != "6" || empty.Cont {
		t.Fatalf("empty row = %+v", empty)
	}

	if width := diffWrapTestWidth(empty.Left); width != 0 {
		t.Fatalf("empty row width = %d", width)
	}

	for _, row := range got {
		if width := diffWrapTestWidth(row.Left); width > 10 {
			t.Fatalf("row %q is %d runes, max 10", lineText(row.Left), width)
		}
	}

	if joined := lineText(add[0].Left) + lineText(add[1].Left); joined != "one two three four" {
		t.Fatalf("add text lost across chunks: %q", joined)
	}
}

func TestDiffWrapSplit(t *testing.T) {
	t.Parallel()

	long := []segment{{Text: "one two three four", Class: classPlain}}

	rows := []diffRow{
		{
			Kind: "split", LeftNo: "5", RightNo: "9",
			LeftKind: rowDel, RightKind: rowAdd,
			Left:  []segment{{Text: "cat", Class: classPlain}},
			Right: long,
		},
		{
			Kind: "split", LeftNo: "10", RightNo: "11",
			LeftKind: rowDel, RightKind: rowAdd,
			Left:  long,
			Right: []segment{{Text: "dog", Class: classPlain}},
		},
		{
			Kind: "split", LeftKind: rowDel, RightKind: rowAdd,
			LeftPad: true,
			Right:   long,
		},
		{Kind: "split", LeftKind: rowCtx, RightKind: rowCtx},
	}

	got := wrapDiffRows(rows, 10, 10, true)
	if len(got) != 7 {
		t.Fatalf("visual rows = %d, want 7: %+v", len(got), got)
	}

	first := got[0:2]
	if first[0].LeftNo != "5" || first[0].RightNo != "9" {
		t.Fatalf("first numbers = %q/%q", first[0].LeftNo, first[0].RightNo)
	}

	if first[0].LeftCont || first[0].RightCont || first[0].LeftPad || first[0].RightPad {
		t.Fatalf("first flags = %+v", first[0])
	}

	if first[0].LeftKind != rowDel || first[0].RightKind != rowAdd {
		t.Fatalf("first kinds = %s/%s", first[0].LeftKind, first[0].RightKind)
	}

	if first[1].LeftNo != "" || first[1].RightNo != "" {
		t.Fatalf("continuation numbers = %q/%q", first[1].LeftNo, first[1].RightNo)
	}

	if first[1].LeftCont || !first[1].RightCont {
		t.Fatalf("continuation flags = left %v right %v", first[1].LeftCont, first[1].RightCont)
	}

	if !first[1].LeftPad || first[1].RightPad {
		t.Fatalf("continuation pads = left %v right %v", first[1].LeftPad, first[1].RightPad)
	}

	if first[1].LeftKind != rowDel || first[1].RightKind != rowAdd {
		t.Fatalf("continuation kinds = %s/%s", first[1].LeftKind, first[1].RightKind)
	}

	if joined := lineText(first[0].Right) + lineText(first[1].Right); joined != "one two three four" {
		t.Fatalf("right text lost across chunks: %q", joined)
	}

	second := got[2:4]
	if second[0].LeftNo != "10" || second[0].RightNo != "11" {
		t.Fatalf("second numbers = %q/%q", second[0].LeftNo, second[0].RightNo)
	}

	if !second[1].LeftCont || second[1].RightCont {
		t.Fatalf("second continuation flags = left %v right %v", second[1].LeftCont, second[1].RightCont)
	}

	if second[1].LeftPad || !second[1].RightPad {
		t.Fatalf("second continuation pads = left %v right %v", second[1].LeftPad, second[1].RightPad)
	}

	padded := got[4:6]
	for i, row := range padded {
		if !row.LeftPad {
			t.Fatalf("padded row %d left pad = false", i)
		}

		if row.LeftNo != "" {
			t.Fatalf("padded row %d left number = %q", i, row.LeftNo)
		}

		if i > 0 && !row.RightCont {
			t.Fatalf("padded row %d right cont = false", i)
		}
	}

	empty := got[6]
	if empty.LeftPad || empty.RightPad || empty.LeftCont || empty.RightCont {
		t.Fatalf("empty split row = %+v", empty)
	}

	if width := diffWrapTestWidth(empty.Left) + diffWrapTestWidth(empty.Right); width != 0 {
		t.Fatalf("empty split width = %d", width)
	}

	for _, row := range got {
		if row.Kind != "split" {
			t.Fatalf("split kind lost: %+v", row)
		}

		if width := diffWrapTestWidth(row.Left); width > 10 {
			t.Fatalf("left cell is %d runes, max 10", width)
		}

		if width := diffWrapTestWidth(row.Right); width > 10 {
			t.Fatalf("right cell is %d runes, max 10", width)
		}
	}
}

func TestDiffWrapHunk(t *testing.T) {
	t.Parallel()

	hunkText := "@@ -10,26 +10,92 @@ cd backend and then read the scheduler"
	metaText := "Binary file. The viewer shows two lines of metadata instead of a diff."

	rows := []diffRow{
		{Kind: rowHunk, Hunk: hunkText},
		{Kind: rowMeta, Hunk: metaText},
		{Kind: rowHunk, Hunk: "short"},
	}

	got := wrapDiffRows(rows, 16, 8, true)

	var hunks, metas []diffRow

	for _, row := range got {
		switch row.Kind {
		case rowHunk:
			hunks = append(hunks, row)
		case rowMeta:
			metas = append(metas, row)
		default:
			t.Fatalf("unexpected kind %q", row.Kind)
		}
	}

	if len(hunks) < 3 {
		t.Fatalf("hunk rows = %d, want a wrapped label and the short row", len(hunks))
	}

	label := hunks[:len(hunks)-1]
	if label[0].Cont {
		t.Fatalf("first hunk row is a continuation: %+v", label[0])
	}

	if label[0].Hunk == "" {
		t.Fatal("first hunk chunk is empty")
	}

	for i, row := range label[1:] {
		if !row.Cont {
			t.Fatalf("hunk continuation %d = %+v", i+1, row)
		}

		if row.Hunk == "" {
			t.Fatalf("hunk continuation %d is empty", i+1)
		}
	}

	joined := ""
	for _, row := range label {
		joined += row.Hunk
	}

	if joined != hunkText {
		t.Fatalf("hunk text lost across chunks: %q", joined)
	}

	short := hunks[len(hunks)-1]
	if short.Hunk != "short" || short.Cont {
		t.Fatalf("short hunk = %+v", short)
	}

	if len(metas) < 2 {
		t.Fatalf("meta rows = %d, want wrapped text", len(metas))
	}

	if metas[0].Cont {
		t.Fatalf("first meta row is a continuation: %+v", metas[0])
	}

	joined = ""

	for i, row := range metas {
		if i > 0 && !row.Cont {
			t.Fatalf("meta continuation %d = %+v", i, row)
		}

		joined += row.Hunk
	}

	if joined != metaText {
		t.Fatalf("meta text lost across chunks: %q", joined)
	}

	for _, row := range got {
		if width := len([]rune(row.Hunk)); width > 16 {
			t.Fatalf("hunk row is %d runes, max 16: %q", width, row.Hunk)
		}

		if row.OldNo != "" || row.NewNo != "" {
			t.Fatalf("hunk row grew numbers: %+v", row)
		}
	}
}

func TestDiffWrapOffClips(t *testing.T) {
	t.Parallel()

	rows := []diffRow{
		{Kind: rowCtx, Left: []segment{{Text: "abcdefgh", Class: classPlain}}},
		{Kind: rowHunk, Hunk: "abcdefgh"},
		{
			Kind: "split",
			Left: []segment{{Text: "abcdefgh", Class: classPlain}}, Right: []segment{{Text: "ijklmnop", Class: classPlain}},
		},
	}

	got := wrapDiffRows(rows, 5, 4, false)
	if len(got) != len(rows) {
		t.Fatalf("rows = %d, want %d clipped rows", len(got), len(rows))
	}

	if text := lineText(got[0].Left); text != "abcde…" {
		t.Fatalf("unified clip = %q", text)
	}

	if got[1].Hunk != "abcde…" {
		t.Fatalf("hunk clip = %q", got[1].Hunk)
	}

	if text := lineText(got[2].Left); text != "abcd…" {
		t.Fatalf("split left clip = %q", text)
	}

	if text := lineText(got[2].Right); text != "ijkl…" {
		t.Fatalf("split right clip = %q", text)
	}

	for _, row := range got {
		if row.Cont || row.LeftCont || row.RightCont {
			t.Fatalf("clip marked a continuation: %+v", row)
		}
	}

	if text := lineText(rows[0].Left); text != "abcdefgh" {
		t.Fatalf("clip mutated the input row: %q", text)
	}
}

func TestDiffWrapPreservesClasses(t *testing.T) {
	t.Parallel()

	segs := []segment{
		{Text: "func ", Class: classKw},
		{Text: "mainLoop", Class: classFn},
		{Text: "() {", Class: classPunc},
	}

	want := diffWrapTestKey(segs)

	unified := wrapDiffRows([]diffRow{{Kind: rowAdd, NewNo: "1", Left: segs}}, 5, 5, true)
	if len(unified) < 2 {
		t.Fatalf("unified rows = %d, want a wrap", len(unified))
	}

	gotKey := ""
	for _, row := range unified {
		gotKey += diffWrapTestKey(row.Left)
	}

	if gotKey != want {
		t.Fatalf("unified classes changed:\n got %q\nwant %q", gotKey, want)
	}

	joined := ""
	for _, row := range unified {
		joined += lineText(row.Left)
	}

	if joined != "func mainLoop() {" {
		t.Fatalf("unified text order changed: %q", joined)
	}

	split := wrapDiffRows([]diffRow{{
		Kind: "split", LeftNo: "1", RightNo: "1",
		LeftKind: rowDel, RightKind: rowAdd,
		Left: segs, Right: segs,
	}}, 5, 5, true)

	leftKey, rightKey := "", ""

	for _, row := range split {
		leftKey += diffWrapTestKey(row.Left)
		rightKey += diffWrapTestKey(row.Right)
	}

	if leftKey != want || rightKey != want {
		t.Fatalf("split classes changed:\nleft %q\nright %q\nwant %q", leftKey, rightKey, want)
	}

	for _, row := range append(unified, split...) {
		for _, seg := range append(append([]segment(nil), row.Left...), row.Right...) {
			if strings.ContainsRune(seg.Text, '…') {
				t.Fatalf("wrapped segment carries a clip ellipsis: %q", seg.Text)
			}
		}
	}
}

// diffWrapTestWidth counts the runes in one styled cell.
func diffWrapTestWidth(segs []segment) int {
	width := 0

	for _, seg := range segs {
		width += len([]rune(seg.Text))
	}

	return width
}

// diffWrapTestKey records each rune with its class so a wrap can be checked
// for lost text or changed highlighting.
func diffWrapTestKey(segs []segment) string {
	var b strings.Builder

	for _, seg := range segs {
		for _, r := range seg.Text {
			b.WriteRune(r)
			b.WriteByte(':')
			b.WriteString(seg.Class)
			b.WriteByte('|')
		}
	}

	return b.String()
}
