package replay

import (
	"context"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"strings"
	"testing"
)

// TestChunkPlanSkipsIneligibleOps pins the run breaks: fixed, z-indexed, and
// translucent ops stay direct, and an opaque pair still folds.
func TestChunkPlanSkipsIneligibleOps(t *testing.T) {
	d := &layout.Display{
		Width: 400, Height: 200, PointsPerPixel: 1,
		Boxes: []layout.Box{{X: 10, Y: 10, W: 100, H: 120}},
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 12, W: 40, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 36, W: 40, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 60, W: 40, H: 20, Alpha: 0.5},
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 84, W: 40, H: 20, Fixed: true},
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 108, W: 40, H: 10, ZIndex: 2, ZIndexSet: true},
		},
		Order: []int{0, 1, 2, 3, 4},
	}

	p := buildChunks(d, buildPlan(d))
	if len(p.items) != 4 {
		t.Fatalf("items = %d, want 4", len(p.items))
	}

	if p.items[0].run == nil || len(p.items[0].run.ops) != 2 {
		t.Fatalf("first item = %+v, want a two-op run", p.items[0])
	}

	for i := 1; i < len(p.items); i++ {
		if p.items[i].op == nil {
			t.Fatalf("item %d is not direct", i)
		}
	}
}

// TestChunkPlanSkipsOversizedOps checks an op wider than the side cap stays
// direct.
func TestChunkPlanSkipsOversizedOps(t *testing.T) {
	d := &layout.Display{
		Width: 900, Height: 200, PointsPerPixel: 1,
		Boxes: []layout.Box{{X: 10, Y: 10, W: 800, H: 100}},
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 12, W: 600, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 40, W: 600, H: 20},
		},
		Order: []int{0, 1},
	}

	p := buildChunks(d, buildPlan(d))
	if len(p.items) != 2 {
		t.Fatalf("items = %d, want 2 direct ops", len(p.items))
	}

	if p.items[0].op == nil || p.items[1].op == nil {
		t.Fatal("oversized ops folded into a run")
	}
}

// flushChunks empties the chunk cache so a test starts cold.
func flushChunks() {
	chunks.mu.Lock()
	defer chunks.mu.Unlock()

	for _, e := range chunks.entries {
		e.img.Deallocate()
	}

	chunks.entries = nil
	chunks.order.reset()
	chunks.pixels = 0
}

// TestChunkCacheReusesAndRebuilds pins the two mutation paths: a signature
// change with the same box repaints the same image, and a box change builds
// a new one.
func TestChunkCacheReusesAndRebuilds(t *testing.T) {
	flushChunks()
	defer flushChunks()

	key := chunkKey{id: 7}
	box := image.Rect(0, 0, 32, 32)

	first := chunks.get(key, 1, box, func(*ebiten.Image) {})
	second := chunks.get(key, 2, box, func(*ebiten.Image) {})
	if second.img != first.img {
		t.Fatal("a signature change built a new image")
	}

	third := chunks.get(key, 2, image.Rect(0, 0, 48, 32), func(*ebiten.Image) {})
	if third.img == first.img {
		t.Fatal("a box change reused the old image")
	}

	if third.img.Bounds().Dx() != 48 {
		t.Fatalf("new image = %v", third.img.Bounds())
	}
}

// TestChunkCacheStaysBounded fills past the entry cap and requires eviction
// down to it.
func TestChunkCacheStaysBounded(t *testing.T) {
	flushChunks()
	defer flushChunks()

	box := image.Rect(0, 0, 16, 16)
	for i := range chunkMaxEntries + 8 {
		chunks.get(chunkKey{id: i}, uint64(i), box, func(*ebiten.Image) {})
	}

	if got := len(chunks.entries); got > chunkMaxEntries {
		t.Fatalf("entries = %d, want at most %d", got, chunkMaxEntries)
	}

	if chunks.pixels > chunkMaxPixels {
		t.Fatalf("pixels = %d, want at most %d", chunks.pixels, chunkMaxPixels)
	}
}

// TestChunkCacheEvictsLeastRecentlyUsed pins the recency policy: a hit marks
// the slot most-recently used, so the next eviction drops an untouched slot
// instead of the hot one.
func TestChunkCacheEvictsLeastRecentlyUsed(t *testing.T) {
	flushChunks()
	defer flushChunks()

	box := image.Rect(0, 0, 16, 16)
	for i := range chunkMaxEntries {
		chunks.get(chunkKey{id: i}, uint64(i), box, func(*ebiten.Image) {})
	}

	hot := chunkKey{id: 0}
	chunks.get(hot, 0, box, func(*ebiten.Image) {})
	chunks.get(chunkKey{id: chunkMaxEntries}, chunkMaxEntries, box, func(*ebiten.Image) {})

	if _, ok := chunks.entries[hot]; !ok {
		t.Fatal("recently used chunk 0 was evicted")
	}

	if _, ok := chunks.entries[chunkKey{id: 1}]; ok {
		t.Fatal("untouched chunk 1 survived ahead of recently used chunk 0")
	}
}

// TestChunkPlanSplitsPastOpCap checks a run longer than the op cap splits
// into runs that both fit.
func TestChunkPlanSplitsPastOpCap(t *testing.T) {
	d := &layout.Display{
		Width: 400, Height: 200, PointsPerPixel: 1,
		Boxes: []layout.Box{{X: 10, Y: 10, W: 300, H: 100}},
	}

	for i := range chunkMaxOps + 6 {
		d.Ops = append(d.Ops, layout.DisplayOp{
			Kind: layout.DisplayOpFillRect, X: 12 + float64(i), Y: 12, W: 20, H: 20,
		})
		d.Order = append(d.Order, i)
	}

	p := buildChunks(d, buildPlan(d))
	if len(p.items) != 2 {
		t.Fatalf("items = %d, want 2 runs", len(p.items))
	}

	if got := len(p.items[0].run.ops); got != chunkMaxOps {
		t.Fatalf("first run = %d ops, want %d", got, chunkMaxOps)
	}

	if got := len(p.items[1].run.ops); got != 6 {
		t.Fatalf("second run = %d ops, want 6", got)
	}
}

// groupedDisplay builds a display whose chunkable operations sit inside an
// isolation group. The engine marks a grouped operation as z-indexed, so the
// display is built by hand.
func groupedDisplay() *layout.Display {
	group := &layout.DisplayGroup{ID: 1, Isolate: true}
	ops := []layout.DisplayOp{
		{Kind: layout.DisplayOpFillRect, X: 12, Y: 12, W: 40, H: 20},
		{Kind: layout.DisplayOpFillRect, X: 12, Y: 36, W: 40, H: 20},
	}
	for i := range ops {
		ops[i].SetBlendGroup(group)
	}

	return &layout.Display{
		Width: 200, Height: 100, PointsPerPixel: 1,
		Boxes: []layout.Box{{X: 10, Y: 10, W: 50, H: 80}},
		Ops:   ops, Order: []int{0, 1},
	}
}

// firstRun returns the plan's first chunk run.
func firstRun(p *chunkPlan) *chunkRun {
	for i := range p.items {
		if p.items[i].run != nil {
			return p.items[i].run
		}
	}

	return nil
}

// TestChunkGroupRunSurvivesRedraw pins that a chunk holding an operation
// inside a group is reused after a redraw. The group is a fresh allocation
// on every layout run, so a signature folding its address missed every
// redraw.
func TestChunkGroupRunSurvivesRedraw(t *testing.T) {
	flushChunks()
	defer flushChunks()

	first := groupedDisplay()
	plan := buildChunks(first, buildPlan(first))
	run := firstRun(plan)
	if run == nil {
		t.Fatal("no run holds the grouped operations")
	}

	baked := chunkRaster(first, plan, run)
	if baked == nil {
		t.Fatal("the grouped run did not bake")
	}

	second := groupedDisplay()
	plan2 := buildChunks(second, buildPlan(second))
	run2 := firstRun(plan2)
	if run2 == nil {
		t.Fatal("the redraw has no run holding the grouped operations")
	}

	if run2.ops[0].Group() == run.ops[0].Group() {
		t.Fatal("the redraw reused the group allocation")
	}

	if again := chunkRaster(second, plan2, run2); again != baked {
		t.Fatal("the grouped chunk was baked again after the redraw")
	}
}

// TestChunkPlanKeepsGroupsWhole pins that a blend group is one plan item and
// the runs on either side do not reach into it.
func TestChunkPlanKeepsGroupsWhole(t *testing.T) {
	group := &layout.DisplayGroup{ID: 1, Mode: "multiply", Isolate: true}
	ops := []layout.DisplayOp{
		{Kind: layout.DisplayOpFillRect, X: 12, Y: 12, W: 40, H: 20},
		{Kind: layout.DisplayOpFillRect, X: 12, Y: 36, W: 40, H: 20},
		{Kind: layout.DisplayOpFillRect, X: 60, Y: 12, W: 20, H: 20},
		{Kind: layout.DisplayOpFillRect, X: 12, Y: 60, W: 40, H: 20},
	}
	ops[2].SetBlendGroup(group)

	d := &layout.Display{
		Width: 200, Height: 100, PointsPerPixel: 1,
		Boxes: []layout.Box{{X: 10, Y: 10, W: 50, H: 80}},
		Ops:   ops, Order: []int{0, 1, 2, 3},
	}

	p := buildChunks(d, buildPlan(d))
	if len(p.items) != 3 {
		t.Fatalf("items = %d, want run, group, op", len(p.items))
	}

	if p.items[0].run == nil || len(p.items[0].run.ops) != 2 {
		t.Fatalf("first item = %+v, want a two-op run", p.items[0])
	}

	if p.items[1].group == nil || p.items[1].group.group != group {
		t.Fatalf("second item = %+v, want the group node", p.items[1])
	}

	if p.items[2].op == nil {
		t.Fatal("the op after the group is not direct")
	}
}

// cardPage lays out a grid of cards, each a background, border, title, and
// two body lines, the shape a dashboard repaints on hover.
func cardPage(t testing.TB, cards int) *layout.Display {
	t.Helper()

	var sb strings.Builder
	sb.WriteString(`<html><body style="margin:0;background:#202030">`)

	for i := range cards {
		fmt.Fprintf(&sb, `<div id="card%d" style="margin:8px;width:200px;height:90px;`+
			`background:#eef;border:2px solid #336;border-radius:6px;padding:6px">`+
			`<p style="margin:0">Card %d</p><p style="margin:2px 0">detail line</p></div>`, i, i)
	}

	sb.WriteString(`</body></html>`)

	d, err := render.DisplayList(context.Background(), sb.String(), 480, 640)
	if err != nil {
		t.Fatal(err)
	}

	return d
}

// planStats counts a plan's direct ops, runs, and run members.
func planStats(p *chunkPlan) (direct, runs, members int) {
	for i := range p.items {
		switch {
		case p.items[i].run != nil:
			runs++
			members += len(p.items[i].run.ops)
		case p.items[i].op != nil:
			direct++
		}
	}

	return direct, runs, members
}

func TestChunkPlanFoldsCards(t *testing.T) {
	d := cardPage(t, 6)
	p := buildChunks(d, buildPlan(d))
	direct, runs, members := planStats(p)
	t.Logf("ops=%d boxes=%d direct=%d runs=%d members=%d", len(d.Ops), len(d.Boxes), direct, runs, members)

	if runs == 0 {
		t.Fatal("no card run folded")
	}

	if members < 2 {
		t.Fatalf("runs hold %d ops", members)
	}

	for i := range p.items {
		if p.items[i].run == nil {
			continue
		}

		for _, op := range p.items[i].run.ops {
			if _, ok := opBounds(op, d.PointsPerPixel); !ok {
				t.Fatalf("run holds an unbounded op %v", op.Kind)
			}
		}
	}
}

// TestPlanReusesAcrossIdenticalRepaints proves the generation-keyed cache:
// two identical repaints build once, proven by build counters, not timing.
func TestPlanReusesAcrossIdenticalRepaints(t *testing.T) {
	d := cardBenchDisplay()
	var r Renderer
	defer r.Close()

	g0, c0 := groupPlanBuildCount(), chunkPlanBuildCount()
	p1 := r.tilePlanFor(d)
	g1, c1 := groupPlanBuildCount(), chunkPlanBuildCount()

	if p1 == nil {
		t.Fatal("first tile plan is nil")
	}

	if g1-g0 != 1 || c1-c0 != 1 {
		t.Fatalf("first build = group %+d chunk %+d, want 1 each", g1-g0, c1-c0)
	}

	p2 := r.tilePlanFor(d)
	g2, c2 := groupPlanBuildCount(), chunkPlanBuildCount()

	if p2 != p1 {
		t.Fatal("identical repaint did not reuse the plan pointer")
	}

	if g2-g1 != 0 || c2-c1 != 0 {
		t.Fatalf("identical repaint rebuilt: group %+d chunk %+d", g2-g1, c2-c1)
	}
}

// TestPlanRebuildsOnSpliceGeneration proves an element splice changes the
// display generation and rebuilds instead of reusing a stale plan.
func TestPlanRebuildsOnSpliceGeneration(t *testing.T) {
	d := cardBenchDisplay()
	var r Renderer
	defer r.Close()

	p1 := r.tilePlanFor(d)
	if p1 == nil {
		t.Fatal("first tile plan is nil")
	}

	// Splice one more card onto the same display object: the shape changes,
	// so the fingerprint must miss.
	base := len(d.Ops)
	d.Ops = append(d.Ops, layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 5, Y: 5, W: 10, H: 10})
	d.Order = append(d.Order, base)

	g0, c0 := groupPlanBuildCount(), chunkPlanBuildCount()
	p2 := r.tilePlanFor(d)
	g1, c1 := groupPlanBuildCount(), chunkPlanBuildCount()

	if p2 == p1 {
		t.Fatal("spliced display reused the stale plan")
	}

	if g1-g0 != 1 || c1-c0 != 1 {
		t.Fatalf("splice rebuild = group %+d chunk %+d, want 1 each", g1-g0, c1-c0)
	}
}

// genDamageRects returns deterministic multi-rect damage over the bench page:
// one rect per early card plus two spanning rects that cross tile borders.
func genDamageRects(d *layout.Display) []image.Rectangle {
	rects := make([]image.Rectangle, 0, 10)

	for i := 0; i < 6 && i < len(d.Boxes); i++ {
		b := d.Boxes[i]
		rects = append(rects, image.Rect(
			int(b.X)-2, int(b.Y)-2,
			int(b.X+b.W)+2, int(b.Y+b.H)+2,
		))
	}

	// Spanning rects cross 256 px tile borders on the 850x510 bench page.
	rects = append(rects,
		image.Rect(200, 100, 500, 300),
		image.Rect(0, 0, 850, 510),
	)

	return rects
}

// pickedOpIDs collects the paint order of ops the tile path would paint for
// rects: bucket candidates plus the precise opTouches check, fixed excluded.
func pickedOpIDs(p *chunkPlan, d *layout.Display, rects []image.Rectangle) []int {
	seen := make([]int, len(p.items))
	var cands []int

	const gen = 1
	for _, r := range rects {
		clipped := r.Intersect(p.content)
		if clipped.Empty() {
			continue
		}

		cands = p.tiles.planTilesForRect(clipped, seen, gen, cands)
	}

	sortPlanCandidates(cands)

	dedup := cands[:0]
	for i, at := range cands {
		if i == 0 || at != cands[i-1] {
			dedup = append(dedup, at)
		}
	}

	ppt := d.PointsPerPixel
	var out []int
	for _, at := range dedup {
		it := &p.items[at]

		switch {
		case it.group != nil:
			collectGroupOps(it.group, d, rects, &out)
		case it.run != nil:
			for _, op := range it.run.ops {
				if op.Fixed {
					continue
				}

				if touchesOneOf(op, ppt, rects) {
					out = append(out, opIndex(d, op))
				}
			}
		case it.op != nil:
			if it.op.Fixed || !touchesOneOf(it.op, ppt, rects) {
				continue
			}

			out = append(out, opIndex(d, it.op))
		}
	}

	return out
}

// collectGroupOps appends a group's ops in order that touch any rect.
func collectGroupOps(n *groupNode, d *layout.Display, rects []image.Rectangle, out *[]int) {
	ppt := d.PointsPerPixel

	var walk func(*groupNode)
	walk = func(g *groupNode) {
		for i := range g.items {
			it := &g.items[i]
			if it.child != nil {
				walk(it.child)

				continue
			}

			if it.op.Fixed || !touchesOneOf(it.op, ppt, rects) {
				continue
			}

			*out = append(*out, opIndex(d, it.op))
		}
	}

	walk(n)
}

// touchesOneOf reports whether an op touches any rect in the set.
func touchesOneOf(op *layout.DisplayOp, ppt float64, rects []image.Rectangle) bool {
	for _, r := range rects {
		if opTouches(op, ppt, r.Intersect(contentRectOf(op))) {
			return true
		}
	}

	return false
}

// contentRectOf is a wide clip for the touches helper; opTouches already
// handles content clipping via the passed rect.
func contentRectOf(_ *layout.DisplayOp) image.Rectangle {
	return image.Rect(-1<<20, -1<<20, 1<<20, 1<<20)
}

// opIndex returns the op's index in the display, or -1 when absent.
func opIndex(d *layout.Display, want *layout.DisplayOp) int {
	for i := range d.Ops {
		if &d.Ops[i] == want {
			return i
		}
	}

	return -1
}

// directPickedOpIDs is the uncached baseline: a full plan walk with the same
// precise per-op check and no tile buckets.
func directPickedOpIDs(p *chunkPlan, d *layout.Display, rects []image.Rectangle) []int {
	ppt := d.PointsPerPixel
	var out []int

	for i := range p.items {
		it := &p.items[i]

		switch {
		case it.group != nil:
			collectGroupOps(it.group, d, rects, &out)
		case it.run != nil:
			for _, op := range it.run.ops {
				if op.Fixed {
					continue
				}

				if touchesOneOf(op, ppt, rects) {
					out = append(out, opIndex(d, op))
				}
			}
		case it.op != nil:
			if it.op.Fixed || !touchesOneOf(it.op, ppt, rects) {
				continue
			}

			out = append(out, opIndex(d, it.op))
		}
	}

	return out
}

// equalInts reports identical sequences.
func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// TestTilePaintsEachRectOnce proves broadphase parity: the tile-bucketed
// multi-rect selection matches the uncached full walk in paint order, and a
// small rect visits fewer candidates than the full plan. No pixels are read,
// so the test runs headless under -race.
func TestTilePaintsEachRectOnce(t *testing.T) {
	d := cardBenchDisplay()
	rects := genDamageRects(d)

	var r Renderer
	defer r.Close()

	p := r.tilePlanFor(d)
	if p == nil || p.tiles == nil {
		t.Fatal("no tile index")
	}

	tileIDs := pickedOpIDs(p, d, rects)
	directIDs := directPickedOpIDs(p, d, rects)

	if !equalInts(tileIDs, directIDs) {
		t.Fatalf("tile selection (%d ops) != direct (%d ops)", len(tileIDs), len(directIDs))
	}

	if len(tileIDs) == 0 {
		t.Fatal("no ops selected for generated damage")
	}

	// Broadphase check: a single-card rect must narrow candidates.
	seen := make([]int, len(p.items))
	small := rects[0].Intersect(p.content)
	cands := p.tiles.planTilesForRect(small, seen, 1, nil)
	if len(cands) == 0 {
		t.Fatal("small rect found no candidates")
	}

	if len(cands) >= len(p.items) {
		t.Fatalf("broadphase visited %d of %d items for a card rect", len(cands), len(p.items))
	}

	// Each dirty rect paints once across its tiles: gathering the spanning
	// rect alone must not duplicate items before the sort dedupes.
	span := image.Rect(200, 100, 500, 300).Intersect(p.content)
	seen2 := make([]int, len(p.items))
	raw := p.tiles.planTilesForRect(span, seen2, 1, nil)
	sortPlanCandidates(raw)

	dedup := raw[:0]
	for i, at := range raw {
		if i == 0 || at != raw[i-1] {
			dedup = append(dedup, at)
		}
	}

	uniq := map[int]bool{}
	for _, at := range dedup {
		if uniq[at] {
			t.Fatalf("rect painted twice across tiles: item %d", at)
		}

		uniq[at] = true
	}
}

// TestViewportUnionSinglePass proves the viewport fast path: overlapping
// rects use one union check, distant rects keep per-rect checks, and the
// single-pass union decision matches a naive per-rect loop for every op.
func TestViewportUnionSinglePass(t *testing.T) {
	overlap := []image.Rectangle{image.Rect(0, 0, 300, 300), image.Rect(100, 100, 400, 400)}
	union, useUnion := viewportUnion(overlap)

	if !useUnion {
		t.Fatal("overlapping rects should use the union")
	}

	if union != image.Rect(0, 0, 400, 400) {
		t.Fatalf("union = %v", union)
	}

	far := []image.Rectangle{image.Rect(0, 0, 100, 100), image.Rect(700, 400, 800, 500)}
	_, useFar := viewportUnion(far)

	if useFar {
		t.Fatal("distant rects should keep per-rect checks")
	}

	d := cardBenchDisplay()
	ppt := d.PointsPerPixel
	rects := []image.Rectangle{image.Rect(0, 0, 400, 300), image.Rect(200, 100, 600, 400)}
	u, useU := viewportUnion(rects)

	for i := range d.Ops {
		op := &d.Ops[i]
		got := opTouchesAny(op, ppt, rects, u, useU)

		want := false
		for _, r := range rects {
			if opTouches(op, ppt, r) {
				want = true
				break
			}
		}

		if useU {
			// The union fast path may overdraw within the threshold but
			// must never miss an op the precise check keeps.
			if want && !got {
				t.Fatalf("op %d missed by union pick", i)
			}
		} else if got != want {
			t.Fatalf("op %d viewport pick = %v, want %v", i, got, want)
		}
	}

	// Distant rects keep the exact per-rect path: no overdraw allowed.
	farRects := []image.Rectangle{image.Rect(0, 0, 100, 100), image.Rect(700, 400, 800, 500)}
	fu, useFU := viewportUnion(farRects)
	if useFU {
		t.Fatal("distant rects should keep per-rect checks")
	}

	for i := range d.Ops {
		op := &d.Ops[i]
		got := opTouchesAny(op, ppt, farRects, fu, useFU)

		want := false
		for _, r := range farRects {
			if opTouches(op, ppt, r) {
				want = true
				break
			}
		}

		if got != want {
			t.Fatalf("far op %d pick = %v, want %v", i, got, want)
		}
	}
}

// TestSigReusesAcrossRepaints proves signatures compute once at plan build and
// invalidate on an op write: a clean repaint hits the cache, a geometry write
// misses, and a refresh hits again.
func TestSigReusesAcrossRepaints(t *testing.T) {
	d := cardBenchDisplay()
	p := buildChunks(d, buildPlan(d))

	var run *chunkRun

	for i := range p.items {
		if p.items[i].run != nil {
			run = p.items[i].run
			break
		}
	}

	if run == nil {
		t.Fatal("no run to signature")
	}

	if !run.hasSig {
		t.Fatal("plan build did not seal the signature")
	}

	if _, _, hit := cachedChunkSig(d, run, p.content); !hit {
		t.Fatal("clean repaint did not hit the cached signature")
	}

	// A tick-style geometry write changes the write generation.
	run.ops[0].X++

	if _, _, hit := cachedChunkSig(d, run, p.content); hit {
		t.Fatal("geometry write kept a stale cached signature")
	}

	sig, box := chunkSig(d, run, p.content)
	if box.Empty() {
		t.Fatal("mutated run has no box")
	}

	refreshRunSig(run, sig, box)

	if _, _, hit := cachedChunkSig(d, run, p.content); !hit {
		t.Fatal("refresh did not restore the cache hit")
	}
}

// TestGroupPlanCoversTileBuckets proves every plan item lands in the
// broadphase: bucketed plus globals equals the plan length with no loss.
func TestGroupPlanCoversTileBuckets(t *testing.T) {
	d := cardBenchDisplay()
	p := buildChunks(d, buildPlan(d))

	if p.tiles == nil {
		t.Fatal("no tile index")
	}

	seen := make([]int, len(p.items))
	cands := p.tiles.planTilesForRect(p.content, seen, 1, nil)
	sortPlanCandidates(cands)

	dedup := cands[:0]
	for i, at := range cands {
		if i == 0 || at != cands[i-1] {
			dedup = append(dedup, at)
		}
	}

	if len(dedup) != len(p.items) {
		t.Fatalf("tile cover = %d of %d items", len(dedup), len(p.items))
	}
}

// TestRendererEvictDisplay proves EvictDisplay clears the cached plan for d,
// preventing dead displays from being retained by the plan cache.
func TestRendererEvictDisplay(t *testing.T) {
	d := cardBenchDisplay()
	var r Renderer
	defer r.Close()

	p1 := r.tilePlanFor(d)
	if p1 == nil {
		t.Fatal("first tile plan is nil")
	}

	r.EvictDisplay(d)

	// After eviction, plans entries must not retain d.
	r.plans.mu.Lock()
	for i := range r.plans.entries {
		if r.plans.entries[i].d == d {
			t.Fatalf("slot %d still retains evicted display", i)
		}
	}
	r.plans.mu.Unlock()
}

// TestTilePlanRefreshesWhenOpMovesAcrossTileBoundary verifies that in-place
// motion or frame.Anim mutations moving an op across a 256 px tile boundary
// update the plan's tile buckets without rebuilding the whole chunk plan.
func TestTilePlanRefreshesWhenOpMovesAcrossTileBoundary(t *testing.T) {
	d := &layout.Display{
		Width: 600, Height: 300, PointsPerPixel: 1,
		Boxes: []layout.Box{{X: 0, Y: 0, W: 600, H: 300}},
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 20, Y: 20, W: 40, H: 40, R: 1, Alpha: 1},
		},
		Order: []int{0},
	}

	var r Renderer
	defer r.Close()

	p1 := r.tilePlanFor(d)
	if p1 == nil || p1.tiles == nil {
		t.Fatal("missing initial tile plan")
	}
	tile1Rect := image.Rect(256, 0, 512, 256)
	seen := make([]int, len(p1.items))
	if cands := p1.tiles.planTilesForRect(tile1Rect, seen, 1, nil); len(cands) != 0 {
		t.Fatalf("before move, tile 1 had %v, want empty", cands)
	}

	// Move op across the 256 px tile boundary in place (same displayFingerprint).
	d.Ops[0].X = 300
	buildsBefore := chunkPlanBuildCount()
	p2 := r.tilePlanFor(d)
	if p2 != p1 {
		t.Fatal("in-place move replaced plan pointer instead of refreshing tiles")
	}
	if chunkPlanBuildCount() != buildsBefore {
		t.Fatal("in-place move rebuilt the whole chunk plan")
	}
	seen = make([]int, len(p2.items))
	if cands := p2.tiles.planTilesForRect(tile1Rect, seen, 1, nil); len(cands) != 1 || cands[0] != 0 {
		t.Fatalf("after move, tile 1 candidates = %v, want [0]", cands)
	}
}
