package replay

import (
	"image"
	"sort"
)

// chunkCell is the grid cell size in CSS pixels. An op only tests the boxes
// registered in the cell its core box starts in.
const chunkCell = 128

// planTileSize is the broadphase tile side in CSS pixels, matching the
// window's replay tile grid. The chunk plan buckets each paint item into the
// tiles its box meets once at build, so a tile repaint looks up candidates
// instead of scanning the whole plan per tile.
const planTileSize = 256

// planTileIndex buckets plan items by tile. buckets holds item indexes in
// paint order per tile; globals holds items with no finite box (unbounded
// ink, empty boxes) that every tile must visit. cols and rows cover the
// content; origin is the content's minimum.
type planTileIndex struct {
	cols    int
	rows    int
	origin  image.Point
	buckets [][]int
	globals []int
}

// chunkGrid buckets eligible element boxes into square cells. Boxes enter in
// ascending area order, so the first containing candidate a cell holds is the
// smallest containing box: every box that contains the core is registered in
// the core's first cell.
type chunkGrid struct {
	cols  int
	spans [][]int
}

// newChunkGrid indexes boxes in the content's coordinate space.
func newChunkGrid(content image.Rectangle, boxes []chunkBox) *chunkGrid {
	g := &chunkGrid{cols: chunkSpan(content.Dx())}
	g.spans = make([][]int, g.cols*chunkSpan(content.Dy()))

	for i := range boxes {
		g.add(content, boxes[i].box, i)
	}

	return g
}

// chunkSpan is the number of cells a length covers.
func chunkSpan(v int) int {
	if v <= 0 {
		return 1
	}

	return (v + chunkCell - 1) / chunkCell
}

// add registers one box in every cell it meets.
func (g *chunkGrid) add(content, box image.Rectangle, at int) {
	box = box.Intersect(content)
	if box.Empty() {
		return
	}

	x0, y0 := chunkCellOf(box.Min, content)
	x1, y1 := chunkCellOf(image.Pt(box.Max.X-1, box.Max.Y-1), content)

	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			c := y*g.cols + x
			g.spans[c] = append(g.spans[c], at)
		}
	}
}

func chunkCellOf(p image.Point, content image.Rectangle) (int, int) {
	return (p.X - content.Min.X) / chunkCell, (p.Y - content.Min.Y) / chunkCell
}

// planTileSpan is the number of plan tiles a length covers.
func planTileSpan(v int) int {
	if v <= 0 {
		return 1
	}

	return (v + planTileSize - 1) / planTileSize
}

// buildPlanTiles buckets one box into the tile index, or records a global
// when the box is empty. Caller passes hasBox false for unbounded ink.
func (t *planTileIndex) add(box image.Rectangle, at int, hasBox bool) {
	if !hasBox || box.Empty() {
		t.globals = append(t.globals, at)

		return
	}

	x0 := (box.Min.X - t.origin.X) / planTileSize
	y0 := (box.Min.Y - t.origin.Y) / planTileSize
	x1 := (box.Max.X - 1 - t.origin.X) / planTileSize
	y1 := (box.Max.Y - 1 - t.origin.Y) / planTileSize

	if x0 < 0 {
		x0 = 0
	}

	if y0 < 0 {
		y0 = 0
	}

	if x1 >= t.cols {
		x1 = t.cols - 1
	}

	if y1 >= t.rows {
		y1 = t.rows - 1
	}

	if x1 < 0 || y1 < 0 || x0 >= t.cols || y0 >= t.rows || x1 < x0 || y1 < y0 {
		t.globals = append(t.globals, at)

		return
	}

	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			c := y*t.cols + x
			t.buckets[c] = append(t.buckets[c], at)
		}
	}
}

// planTilesForRect appends the candidate item indexes whose tiles meet rect,
// globals included. The caller sizes seen to the plan length and bumps gen
// for the lookup; duplicates across tiles are skipped. Candidates are not
// assumed sorted: tiles are visited row-major, so callers sort the gathered
// slice into paint order before painting.
func (t *planTileIndex) planTilesForRect(rect image.Rectangle, seen []int, gen int, out []int) []int {
	if t == nil {
		return out
	}

	for _, at := range t.globals {
		if seen[at] != gen {
			seen[at] = gen
			out = append(out, at)
		}
	}

	x0 := (rect.Min.X - t.origin.X) / planTileSize
	y0 := (rect.Min.Y - t.origin.Y) / planTileSize
	x1 := (rect.Max.X - 1 - t.origin.X) / planTileSize
	y1 := (rect.Max.Y - 1 - t.origin.Y) / planTileSize

	if x0 < 0 {
		x0 = 0
	}

	if y0 < 0 {
		y0 = 0
	}

	if x1 >= t.cols {
		x1 = t.cols - 1
	}

	if y1 >= t.rows {
		y1 = t.rows - 1
	}

	if x1 < 0 || y1 < 0 || x0 >= t.cols || y0 >= t.rows {
		return out
	}

	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			for _, at := range t.buckets[y*t.cols+x] {
				if seen[at] != gen {
					seen[at] = gen
					out = append(out, at)
				}
			}
		}
	}

	return out
}

// sortPlanCandidates restores paint order after a tile-gather.
func sortPlanCandidates(out []int) {
	sort.Ints(out)
}

// find returns the index into boxes of the smallest eligible box that
// contains core, or -1.
func (g *chunkGrid) find(content image.Rectangle, boxes []chunkBox, core image.Rectangle) int {
	clip := core.Intersect(content)
	if clip.Empty() {
		return -1
	}

	x, y := chunkCellOf(clip.Min, content)

	for _, bi := range g.spans[y*g.cols+x] {
		if chunkBoxContains(boxes[bi].box, core) {
			return bi
		}
	}

	return -1
}
