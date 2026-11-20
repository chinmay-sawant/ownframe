package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"math"
	"reflect"
	"sort"
	"sync/atomic"
)

// chunkPaint is one step of a display's chunk plan: a direct op, one cached
// run, or a non-normal group node painted as a unit.
type chunkPaint struct {
	op    *layout.DisplayOp
	run   *chunkRun
	group *groupNode
}

// chunkRun is a consecutive run of ops that share one element box. sig and
// box are computed once at plan build; writeGen is the op-write generation
// they were built from, so a repaint reuses them without rehashing while the
// tick-writable fields still match.
type chunkRun struct {
	id       int
	ops      []*layout.DisplayOp
	sig      uint64
	box      image.Rectangle
	writeGen uint64
	hasSig   bool
}

// chunkPlan is a display's paint order with eligible runs folded into chunks.
// boxes are the element boxes a chunk may use, smallest first, and grid
// indexes them so an op only tests the boxes near it. tiles buckets each
// item by 256 px tile once at build, so a tile repaint looks up candidates
// instead of scanning the whole plan per tile. gen is the display shape the
// plan was built from.
type chunkPlan struct {
	items   []chunkPaint
	content image.Rectangle
	boxes   []chunkBox
	grid    *chunkGrid
	tiles   *planTileIndex
	gen     uint64
}

// chunkPlanBuilds counts chunk-plan folds. Tests prove reuse by observing no
// build across identical repaints, not by timing.
var chunkPlanBuilds atomic.Int64 //nolint:gochecknoglobals // reason: one process-wide counter read by chunkPlanBuildCount so tests can prove plan reuse.

// chunkPlanBuildCount reports chunk-plan folds.
func chunkPlanBuildCount() int64 {
	return chunkPlanBuilds.Load()
}

// buildChunks flattens the group tree into a paint plan and folds runs of
// ops that share an element box into chunks. A non-normal or masked group
// stays a unit, so its subtree keeps direct ops and the group painters never
// meet a chunk. Signatures compute once here; tile buckets index every item
// once here.
func buildChunks(d *layout.Display, root *groupNode) *chunkPlan {
	p := &chunkPlan{content: contentRect(d), boxes: chunkBoxes(d), gen: displayFingerprint(d)}
	p.grid = newChunkGrid(p.content, p.boxes)
	next := 0
	flattenNode(d, root, p, &next)
	sealChunkSigs(d, p)
	indexPlanTiles(d, p)
	chunkPlanBuilds.Add(1)

	return p
}

// sealChunkSigs computes each run's signature and box once at plan build and
// records the write generation. A run that cannot bake keeps hasSig false.
func sealChunkSigs(d *layout.Display, p *chunkPlan) {
	for i := range p.items {
		c := p.items[i].run
		if c == nil {
			continue
		}

		sig, box := chunkSig(d, c, p.content)
		if box.Empty() {
			continue
		}

		c.sig = sig
		c.box = box
		c.writeGen = runWriteGen(c)
		c.hasSig = true
	}
}

// indexPlanTiles buckets every plan item by tile once at build. Groups use
// their page box; unbounded ink becomes a global visited by every tile.
func indexPlanTiles(d *layout.Display, p *chunkPlan) {
	t := &planTileIndex{origin: p.content.Min}
	t.cols = planTileSpan(p.content.Dx())
	t.rows = planTileSpan(p.content.Dy())
	t.buckets = make([][]int, t.cols*t.rows)

	ppt := 0.0
	if d != nil {
		ppt = d.PointsPerPixel
	}

	for i := range p.items {
		it := &p.items[i]

		switch {
		case it.run != nil:
			if it.run.hasSig {
				t.add(it.run.box, i, true)
			} else {
				t.add(image.Rectangle{}, i, false)
			}
		case it.op != nil:
			box, ok := opBounds(it.op, ppt)
			t.add(box, i, ok)
		case it.group != nil:
			box, ok := nodeBox(it.group, ppt)
			t.add(box, i, ok)
		}
	}

	p.tiles = t
}

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

// flattenNode appends one node's paint order to p. A direct op joins the
// current run when it lands in the same element box and the caps still hold;
// anything else flushes the run first.
func flattenNode(d *layout.Display, n *groupNode, p *chunkPlan, next *int) {
	if n.group != nil && (!normalMode(n.group.Mode) || n.group.Mask != nil || n.group.BackdropFilter != "") {
		p.items = append(p.items, chunkPaint{group: n})

		return
	}

	run := make([]*layout.DisplayOp, 0, chunkMaxOps)
	box, union := -1, image.Rectangle{}

	flush := func() {
		if len(run) > 1 {
			// The chunk keeps this backing array; a fresh run starts
			// below, so no copy is needed to protect the emitted ops.
			ops := run
			p.items = append(p.items, chunkPaint{run: &chunkRun{id: *next, ops: ops}})
			*next++
			run = make([]*layout.DisplayOp, 0, chunkMaxOps)
		} else if len(run) == 1 {
			p.items = append(p.items, chunkPaint{op: run[0]})
		}

		run, box, union = run[:0], -1, image.Rectangle{}
	}

	for i := range n.items {
		it := &n.items[i]
		if it.child != nil {
			flush()
			flattenNode(d, it.child, p, next)

			continue
		}

		idx, b, ok := chunkBoxFor(d, it.op, p)
		if !ok {
			flush()
			p.items = append(p.items, chunkPaint{op: it.op})

			continue
		}

		if len(run) > 0 && (idx != box || len(run) >= chunkMaxOps || !chunkFits(union.Union(b))) {
			flush()
		}

		if len(run) == 0 {
			box = idx
		}

		union = union.Union(b)
		run = append(run, it.op)
	}

	flush()
}

// paintChunked replays display in paint order with element runs served from
// the chunk cache. It falls back to the direct walk when no plan is cached.
func (renderer *Renderer) paintChunked(dst *ebiten.Image, d *layout.Display, pick opPicker) {
	if dst == nil || d == nil {
		return
	}

	p := renderer.chunkPlanFor(d)
	if p == nil {
		renderer.paintDisplay(dst, d, pick)

		return
	}

	for i := range p.items {
		it := &p.items[i]

		switch {
		case it.group != nil:
			renderer.paintGroup(dst, d, it.group, pick)
		case it.run != nil:
			renderer.paintRun(dst, d, p, it.run, pick)
		case it.op != nil:
			if dx, dy, ok := pick(it.op); ok {
				countOp()
				renderer.drawOp(dst, it.op, dx, dy)
			}
		}
	}
}

// paintRun paints one chunk when the pick accepts any member, at the
// translation the picked members share. A mixed translation, a fractional
// one, or a translucent member keeps the per-op path: an image blit can only
// copy whole pixels, and a translucent op cannot promise byte parity.
func (renderer *Renderer) paintRun(dst *ebiten.Image, d *layout.Display, p *chunkPlan, c *chunkRun, pick opPicker) {
	tx, ty, picked, shared := pickRun(c, pick)
	if !picked {
		return
	}

	if !shared || tx != math.Trunc(tx) || ty != math.Trunc(ty) || !runOpaque(c) {
		renderer.drawRunDirect(dst, c, pick)

		return
	}

	e := renderer.chunkRaster(d, p, c)
	if e == nil {
		renderer.drawRunDirect(dst, c, pick)

		return
	}

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(e.box.Min.X)+tx, float64(e.box.Min.Y)+ty)
	countBlit()
	dst.DrawImage(e.img, opts)
}

// pickRun returns the shared translation of the run's picked ops. picked is
// false when none is; shared is false when two picked ops disagree, which no
// painter the plan feeds does.
func pickRun(c *chunkRun, pick opPicker) (tx, ty float64, picked, shared bool) {
	shared = true

	for _, op := range c.ops {
		dx, dy, ok := pick(op)
		if !ok {
			continue
		}

		if picked && (dx != tx || dy != ty) {
			shared = false
		}

		tx, ty, picked = dx, dy, true
	}

	return tx, ty, picked, shared
}

// runOpaque reports a run whose every op is fully opaque. A tick can lower an
// op's paint opacity after the plan was built, so the run rechecks it.
func runOpaque(c *chunkRun) bool {
	for _, op := range c.ops {
		if op.Opacity() < 1 {
			return false
		}
	}

	return true
}

// drawRunDirect paints the picked ops one by one.
func (renderer *Renderer) drawRunDirect(dst *ebiten.Image, c *chunkRun, pick opPicker) {
	for _, op := range c.ops {
		if dx, dy, ok := pick(op); ok {
			countOp()
			renderer.drawOp(dst, op, dx, dy)
		}
	}
}

// fnvOffsetBasis is the FNV-1a 64-bit offset basis starting a signature.
const fnvOffsetBasis uint64 = 14695981039346656037

// fnvPrime is the FNV-1a 64-bit prime folding each word.
const fnvPrime uint64 = 1099511628211

// chunkSig returns the content signature and the page box of one run. An
// unbounded member or a box past the caps gives an empty box, so the caller
// keeps the run direct.
func chunkSig(d *layout.Display, c *chunkRun, content image.Rectangle) (uint64, image.Rectangle) {
	h := fnvOffsetBasis
	box := image.Rectangle{}

	for _, op := range c.ops {
		b, ok := opBounds(op, d.PointsPerPixel)
		if !ok || b.Empty() {
			return 0, image.Rectangle{}
		}

		box = box.Union(b)
		h = sigOp(h, op)
	}

	box = box.Intersect(content)
	if box.Empty() || !chunkFits(box) {
		return 0, image.Rectangle{}
	}

	return h, box
}

// runWriteGen is the cheap op-write generation for one run: the tick-writable
// geometry, colour, opacity, and transform fields. Motion and frame anims
// patch only these fields in place on the same display; structural payloads
// (text, grid segments, path commands, image bytes, fonts) change only on a
// redraw, which changes the display fingerprint and rebuilds the plan with
// fresh signatures. Equality of the write generation implies the cached
// signature is still valid, so a repaint reuses it instead of rehashing.
func runWriteGen(c *chunkRun) uint64 {
	h := fnvOffsetBasis
	h = mix64(h, uint64(len(c.ops)))

	for _, op := range c.ops {
		h = mix64(h, uint64(op.Kind))
		h = mixF(h, op.X)
		h = mixF(h, op.Y)
		h = mixF(h, op.W)
		h = mixF(h, op.H)
		h = mixF(h, op.R)
		h = mixF(h, op.G)
		h = mixF(h, op.B)
		h = mixF(h, op.Alpha)
		h = mixF(h, op.Opacity())
		h = mixF(h, op.Width)
		h = mixF(h, float64(op.RotateDeg))
		h = mix64(h, bit(op.Fixed))
		h = mix64(h, bit(op.XformSet))
		h = sigXformWrite(h, op)
	}

	return h
}

// cachedChunkSig returns the run's cached signature when its write generation
// still matches, and recomputes both signature and box otherwise. The slow
// path returns fresh values without mutating the cached ones when called
// concurrently; the owning loop refreshes the cache via refreshRunSig.
func cachedChunkSig(d *layout.Display, c *chunkRun, content image.Rectangle) (uint64, image.Rectangle, bool) {
	if c.hasSig && runWriteGen(c) == c.writeGen {
		return c.sig, c.box, true
	}

	sig, box := chunkSig(d, c, content)
	if box.Empty() {
		return 0, image.Rectangle{}, false
	}

	return sig, box, false
}

// refreshRunSig records a freshly computed signature after a write-generation
// miss. Call on the owning Ebiten loop only.
func refreshRunSig(c *chunkRun, sig uint64, box image.Rectangle) {
	c.sig = sig
	c.box = box
	c.writeGen = runWriteGen(c)
	c.hasSig = true
}

// mix64 folds one word into an FNV-1a style signature.
func mix64(h, v uint64) uint64 {
	return (h ^ v) * fnvPrime
}

// mixF folds one float.
func mixF(h uint64, v float64) uint64 {
	return mix64(h, math.Float64bits(v))
}

// mixStr folds a string.
func mixStr(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		h = mix64(h, uint64(s[i]))
	}

	return h
}

// bit packs one bool into a word.
func bit(v bool) uint64 {
	if v {
		return 1
	}

	return 0
}

// sigGrid folds a grid run's segments, or a zero marker for every other op.
func sigGrid(h uint64, op *layout.DisplayOp) uint64 {
	if op.Grid == nil {
		return mix64(h, 0)
	}

	h = mix64(h, uint64(len(op.Grid.Segs))+1)

	for i := range op.Grid.Segs {
		s := &op.Grid.Segs[i]
		for _, v := range [8]float64{s.X, s.Y, s.W, s.H, s.Width, s.R, s.G, s.B} {
			h = mixF(h, v)
		}

		h = mix64(h, uint64(s.LineInset))
	}

	return h
}

// sigPath folds a corner path's command list, or a zero marker for every
// other op. The whole point triple and every verb are folded, so a mutated
// payload cannot keep a stale cached raster.
func sigPath(h uint64, op *layout.DisplayOp) uint64 {
	corner, ok := op.CornerPathValue()
	if !ok || corner == nil {
		return mix64(h, 0)
	}

	h = mix64(h, uint64(len(corner.Cmds))+1)
	h = mix64(h, bit(corner.Stroke))

	for i := range corner.Cmds {
		cmd := &corner.Cmds[i]
		h = mix64(h, uint64(cmd.Verb))

		for _, p := range cmd.Pts {
			h = mixF(h, p.X)
			h = mixF(h, p.Y)
		}
	}

	return h
}

// sigImage folds an image payload's identity, or a zero marker. The payload
// hash is the same one the decoded-image cache keys on.
func sigImage(h uint64, op *layout.DisplayOp) uint64 {
	if op.Kind != layout.DisplayOpImage {
		return mix64(h, 0)
	}

	return mix64(h, imageBytesKey(op))
}

// sigXform folds a baked transform, or a zero marker.
func sigXform(h uint64, op *layout.DisplayOp) uint64 {
	m := op.Transform()
	if !op.XformSet {
		var projected bool
		m, projected = textProjection(op)
		if !projected {
			return mix64(h, 0)
		}
	}

	for _, v := range [6]float64{m.A, m.B, m.C, m.D, m.E, m.F} {
		h = mixF(h, v)
	}

	return h
}

// sigXformWrite folds only the transform identity for the write generation:
// whether a matrix is present and its six coefficients. Colour, geometry,
// and opacity live in runWriteGen; the matrix is the remaining tick-writable
// field.
func sigXformWrite(h uint64, op *layout.DisplayOp) uint64 {
	if !op.XformSet {
		m, projected := textProjection(op)
		if !projected {
			return mix64(h, 0)
		}

		for _, v := range [6]float64{m.A, m.B, m.C, m.D, m.E, m.F} {
			h = mixF(h, v)
		}

		return h
	}

	m := op.Transform()
	for _, v := range [6]float64{m.A, m.B, m.C, m.D, m.E, m.F} {
		h = mixF(h, v)
	}

	return h
}

// ptrBits is the address behind an interface holding a pointer, or zero.
func ptrBits(v any) uint64 {
	if v == nil {
		return 0
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return 0
	}

	return uint64(rv.Pointer())
}

// sigOp folds every op field that shapes its pixels into h. Fields a tick can
// rewrite in place are covered, so a mutated op misses the cached raster.
func sigOp(h uint64, op *layout.DisplayOp) uint64 {
	h = mix64(h, uint64(op.Kind))

	for _, v := range [15]float64{
		op.X, op.Y, op.W, op.H, op.R, op.G, op.B, op.Alpha,
		op.Width, op.Size, op.LetterSpacing, op.InkDescent,
		op.Radius, op.RadiusY, op.Opacity(),
	} {
		h = mixF(h, v)
	}

	for _, v := range [4]float64{
		op.RadiusTopLeft, op.RadiusTopRight, op.RadiusBottomRight, op.RadiusBottomLeft,
	} {
		h = mixF(h, v)
	}

	for _, v := range [4]float64{
		op.RadiusTopLeftY, op.RadiusTopRightY, op.RadiusBottomRightY, op.RadiusBottomLeftY,
	} {
		h = mixF(h, v)
	}

	h = mix64(h, uint64(op.StrokeMask))
	h = mix64(h, uint64(op.LineInset))
	h = mix64(h, uint64(op.ZIndex)) //nolint:gosec // reason: int bit pattern kept for hash.
	h = mix64(h, bit(op.Bold)|bit(op.FakeOblique)<<1|bit(op.Fixed)<<2|
		bit(op.ZIndexSet)<<3|bit(op.NoFakeBoldValue())<<4)
	h = mix64(h, uint64(math.Float32bits(op.RotateDeg)))
	h = mixStr(h, op.Text)
	h = mixStr(h, op.TextTransformValue())
	h = mixStr(h, op.TextLanguage())
	h = mix64(h, ptrBits(op.Font))
	h = sigGroup(h, op)

	return sigXform(sigImage(sigGrid(sigPath(h, op), op), op), op)
}

// sigGroup folds the identity of the operation's blend or isolation group.
// The group is a fresh allocation on every layout run, so folding its
// pointer missed on every redraw. The creation-order ID and the group's
// settings describe the same group after a redraw of the same content.
func sigGroup(h uint64, op *layout.DisplayOp) uint64 {
	group := op.Group()
	if group == nil {
		return mix64(h, 0)
	}

	h = mixF(h, float64(group.ID))
	h = mixStr(h, group.Mode)
	h = mix64(h, bit(group.Isolate))

	if parent := group.Parent; parent != nil {
		h = mixF(h, float64(parent.ID))
		h = mixStr(h, parent.Mode)
	}

	return h
}
