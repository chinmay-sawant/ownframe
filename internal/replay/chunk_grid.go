package replay

import "image"

// chunkCell is the grid cell size in CSS pixels. An op only tests the boxes
// registered in the cell its core box starts in.
const chunkCell = 128

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
