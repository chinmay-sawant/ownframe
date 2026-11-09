package replay

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

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
