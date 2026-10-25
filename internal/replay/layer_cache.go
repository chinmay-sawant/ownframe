package replay

import (
	"image"
	"sync"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// layerMaxEntries bounds the number of retained layers one renderer keeps.
	layerMaxEntries = 16
	// layerMaxPixels bounds the retained layer pixels one renderer keeps.
	layerMaxPixels = 32 << 20

	layerChurnMisses    = 2
	layerChurnCooldown  = 60
	layerReasonCountCap = "layer count cap"
	layerReasonPixelCap = "layer pixel cap"
	layerReasonEvicted  = "evicted by last use"
	// layerReasonContentChurn names a layer whose content changed on every
	// frame. Re-baking it per frame costs more than painting its operations
	// into the page, so it stays direct until its content settles.
	layerReasonContentChurn = "content churn"
	// layerReasonRest names a new layer spec whose pose is the rest pose. An
	// element already held as a layer keeps it through that pose, so an
	// alternate animation does not demote and re-promote at every turn.
	layerReasonRest = "no transform or opacity motion"
	// layerReasonEnded names a promoted element that left the specification
	// list. It covers both tracks: an animation that finished, paused, or was
	// replaced, and a transition that ended, cancelled, or was replaced.
	layerReasonEnded = "track ended"

	// layerBytesPerPixel converts retained pixels to the bytes a Probe
	// report shows.
	layerBytesPerPixel = 4
)

// layerKey names one layer cache slot. The page generation invalidates every
// layer after a redraw; the box index separates anonymous elements that share
// the empty id.
type layerKey struct {
	id  string
	box int
	gen uint64
}

// layerStore owns the retained layer textures. It sits beside chunkStore and
// is driven by BeginLayers once per frame.
type layerStore struct {
	mu      sync.Mutex
	entries map[layerKey]*Layer
	pixels  int
	use     uint64

	stats   LayerStats
	reasons map[string]int
	damage  []image.Rectangle
	// topology is the subset of damage where layer membership changed, not
	// just a pose: a promotion, a demotion, or a moved texture box. The
	// compositor repaints the page background only for these.
	topology []image.Rectangle
	// churn holds the frame each churn key may be promoted again.
	churn map[layerKey]uint64

	frame layerFrame
}

// BeginLayers resolves the requested promotions, keeps or bakes their
// textures, and returns the old ∪ new damage of every layer that moved,
// appeared, changed alpha or mode, or disappeared. The paint walk that follows
// skips the returned frame's owned operations and composites each layer once.
func (renderer *Renderer) BeginLayers(d *layout.Display, gen uint64, specs []LayerSpec) []image.Rectangle {
	store := &renderer.layers
	store.mu.Lock()
	defer store.mu.Unlock()

	store.resetFrameStats()
	store.damage = store.damage[:0]
	store.topology = store.topology[:0]
	store.frame.reset()

	if d == nil {
		store.demoteAllLocked(layerReasonEnded)

		return store.damage
	}
	if len(specs) == 0 {
		store.demoteAllLocked(layerReasonEnded)

		return store.damage
	}

	// Resolve every request first, rejecting nested or ineligible elements.
	claimed := make([]bool, len(d.Ops))
	ranks := layerRanks(d)
	candidates := make([]layerCandidate, 0, min(len(specs), layerMaxEntries))
	for _, spec := range specs {
		candidate, reason := resolveLayer(d, spec, gen, claimed, ranks)
		if reason != "" {
			store.demote(reason)
			continue
		}
		if candidate.matrix.IsIdentity() && candidate.alpha == 1 && store.entries[candidate.key] == nil {
			store.demote(layerReasonRest)
			continue
		}
		for _, index := range candidate.owned {
			claimed[index] = true
		}
		candidates = append(candidates, candidate)
	}

	candidates = store.capCandidates(candidates)
	keep := make(map[layerKey]bool, len(candidates))
	for i := range candidates {
		keep[candidates[i].key] = true
	}

	active := make([]*Layer, 0, len(candidates))
	for i := range candidates {
		cand := &candidates[i]
		old := store.entries[cand.key]
		oldPose := image.Rectangle{}
		oldMatrix, oldAlpha, oldMode := layout.Matrix2D{}, 0.0, ""
		if old != nil {
			oldPose = old.pose
			oldMatrix, oldAlpha, oldMode = old.Matrix, old.Alpha, old.Mode
		}

		if until, ok := store.churn[cand.key]; ok && store.use < until {
			store.demote(layerReasonContentChurn)
			continue
		}
		layer, hit, ok, reason := store.getLocked(*cand, keep, renderer, d)
		if !ok {
			store.demote(reason)
			continue
		}
		if old == nil {
			store.stats.Promoted++
		}
		if !hit {
			store.stats.Rebaked++
		}

		layer.Matrix = cand.matrix
		layer.Alpha = cand.alpha
		layer.Mode = cand.mode
		layer.firstRank = cand.firstRank
		layer.pose = layerPose(layer, d)
		layer.LastUse = store.use

		if old == nil {
			store.addDamage(layer.pose)
			store.addTopology(layer.pose)
		} else if !hit || oldPose != layer.pose ||
			oldMatrix != layer.Matrix || oldAlpha != layer.Alpha || oldMode != layer.Mode {
			store.addDamage(oldPose.Union(layer.pose))
		}

		active = append(active, layer)
	}

	store.demoteMissingLocked(keep)
	store.buildFrame(d, active)
	store.stats.Active = len(active)
	store.stats.Pixels = store.pixels
	store.stats.Bytes = store.pixels * layerBytesPerPixel

	return store.damage
}

// LayersActive reports whether BeginLayers left active layers for d.
func (renderer *Renderer) LayersActive(d *layout.Display) bool {
	store := &renderer.layers
	store.mu.Lock()
	defer store.mu.Unlock()

	return d != nil && store.frame.display == d && len(store.frame.active) > 0
}

// DemoteLayers drops every retained layer and reports the page area the paint
// path must rebuild. A caller that leaves the tile path uses it so no stale
// texture survives the switch.
func (renderer *Renderer) DemoteLayers() []image.Rectangle {
	store := &renderer.layers
	store.mu.Lock()
	defer store.mu.Unlock()

	store.resetFrameStats()
	store.damage = store.damage[:0]
	store.frame.reset()
	store.demoteAllLocked(layerReasonEnded)

	return store.damage
}

// capCandidates enforces the count and pixel caps by last use. Existing
// entries carry their previous LastUse; a new candidate is the least recently
// used. The caller's slice order is preserved for painting.
func (store *layerStore) capCandidates(candidates []layerCandidate) []layerCandidate {
	if len(candidates) == 0 {
		return candidates
	}

	pixels := 0
	for i := range candidates {
		pixels += candidates[i].box.Dx() * candidates[i].box.Dy()
	}

	drop := make([]bool, len(candidates))
	dropOldest := func(reason string) bool {
		best := -1
		bestUse := uint64(0)
		for i := range candidates {
			if drop[i] {
				continue
			}
			use := uint64(0)
			if entry := store.entries[candidates[i].key]; entry != nil {
				use = entry.LastUse
			}
			if best < 0 || use < bestUse {
				best, bestUse = i, use
			}
		}
		if best < 0 {
			return false
		}
		drop[best] = true
		pixels -= candidates[best].box.Dx() * candidates[best].box.Dy()
		store.demote(reason)

		return true
	}

	for len(candidates)-countDropped(drop) > layerMaxEntries {
		if !dropOldest(layerReasonCountCap) {
			break
		}
	}
	for pixels > layerMaxPixels {
		if !dropOldest(layerReasonPixelCap) {
			break
		}
	}

	kept := make([]layerCandidate, 0, len(candidates))
	for i := range candidates {
		if !drop[i] {
			kept = append(kept, candidates[i])
		}
	}

	return kept
}

func countDropped(drop []bool) int {
	n := 0
	for _, v := range drop {
		if v {
			n++
		}
	}

	return n
}

// getLocked returns the layer record for cand, baking when the slot is missing,
// its signature changed, or its box moved. keep holds the keys BeginLayers
// intends to use this frame, so an LRU eviction never drops one of them.
func (store *layerStore) getLocked(cand layerCandidate, keep map[layerKey]bool,
	renderer *Renderer, d *layout.Display) (*Layer, bool, bool, string) {
	area := cand.box.Dx() * cand.box.Dy()
	if area > layerMaxPixels {
		return nil, false, false, layerReasonPixelCap
	}
	if store.entries == nil {
		store.entries = make(map[layerKey]*Layer)
	}

	old := store.entries[cand.key]
	if old != nil && old.Box == cand.box && old.Valid {
		if old.Sig == cand.sig {
			old.Ops = setLayerOps(old.Ops, cand.owned)
			old.Box = cand.box
			old.misses = 0

			return old, true, true, ""
		}
		old.misses++
		if old.misses >= layerChurnMisses {
			// The content changes every frame. Painting it into the page is
			// cheaper than clearing and baking a texture per frame, so keep
			// it direct for a while and let the topology damage repair the
			// pose this layer leaves behind.
			store.addDamage(old.pose)
			store.addTopology(old.pose)
			store.addTopology(cand.pose(d))
			store.dropLocked(cand.key)
			if store.churn == nil {
				store.churn = make(map[layerKey]uint64)
			}
			store.churn[cand.key] = store.use + layerChurnCooldown

			return nil, false, false, layerReasonContentChurn
		}
		old.Tex.Clear()
		store.bake(renderer, d, old.Tex, cand)
		old.Sig = cand.sig
		old.Ops = setLayerOps(old.Ops, cand.owned)
		old.Valid = true

		return old, false, true, ""
	}

	// A new box needs room. Evict the least recently used slot that this
	// frame is not keeping.
	for old == nil && (len(store.entries) >= layerMaxEntries || store.pixels+area > layerMaxPixels) {
		if !store.evictLRU(cand.key, keep) {
			return nil, false, false, layerReasonPixelCap
		}
		old = store.entries[cand.key]
	}

	if old != nil {
		store.addTopology(old.pose)
		store.addTopology(cand.pose(d))
		store.pixels -= old.Box.Dx() * old.Box.Dy()
		old.Tex.Deallocate()
		delete(store.entries, cand.key)
	}

	img := ebiten.NewImage(cand.box.Dx(), cand.box.Dy())
	store.bake(renderer, d, img, cand)

	layer := &Layer{
		ElementID: cand.spec.ElementID,
		Gen:       cand.key.gen,
		Box:       cand.box,
		Tex:       img,
		Sig:       cand.sig,
		Ops:       setLayerOps(nil, cand.owned),
		Alpha:     cand.alpha,
		Matrix:    cand.matrix,
		LastUse:   store.use,
		Valid:     true,
	}
	store.entries[cand.key] = layer
	store.pixels += area

	return layer, false, true, ""
}

// dropLocked releases a layer texture without counting an eviction.
func (store *layerStore) dropLocked(key layerKey) {
	entry := store.entries[key]
	if entry == nil {
		return
	}

	store.pixels -= entry.Box.Dx() * entry.Box.Dy()
	entry.Tex.Deallocate()
	entry.Valid = false
	delete(store.entries, key)
}

// bake renders cand's owned operations in their captured base pose into img.
// The texture origin is the box minimum, so the composite matrix later adds
// the animation pose without touching these pixels.
func (store *layerStore) bake(renderer *Renderer, d *layout.Display, img *ebiten.Image, cand layerCandidate) {
	dx := -float64(cand.box.Min.X)
	dy := -float64(cand.box.Min.Y)
	for _, index := range cand.owned {
		if index < 0 || index >= len(d.Ops) {
			continue
		}
		store.bakeOp(renderer, img, &d.Ops[index], cand.base, cand.alpha, dx, dy)
	}
}

// bakeOp draws one owned operation into the layer texture with the sampled
// group alpha divided out, so the composite applies it once.
//
// DisplayOp keeps its rare payload behind a shared pointer, so a struct copy
// that calls SetPaintOpacity or SetXform writes through to the live display.
// The bake therefore borrows the two fields on the live operation and restores
// them before returning, including on a panic. The draw reads the borrowed
// values through the same op, so no alias is needed.
func (store *layerStore) bakeOp(renderer *Renderer, img *ebiten.Image, op *layout.DisplayOp,
	base layout.Matrix2D, groupAlpha, dx, dy float64) {
	savedOpacity := layerElementOpacity(op)
	bakeOpacity := layerBakeOpacity(op, groupAlpha)
	savedXform, savedXformSet := op.Transform(), op.XformSet
	borrowXform := !base.IsIdentity()

	defer func() {
		op.SetPaintOpacity(savedOpacity)
		if borrowXform {
			op.SetXform(savedXform)
			op.XformSet = savedXformSet
		}
	}()

	if bakeOpacity != savedOpacity {
		op.SetPaintOpacity(bakeOpacity)
	}
	if borrowXform {
		current := layout.IdentityMatrix()
		if savedXformSet {
			current = savedXform
		}
		op.SetXform(base.Mul(current))
	}
	countOp()
	renderer.drawOp(img, op, dx, dy)
}

// evictLRU drops the least recently used entry that is not protected. It
// reports false when every entry is protected.
func (store *layerStore) evictLRU(protected layerKey, keep map[layerKey]bool) bool {
	victim := layerKey{}
	found := false
	for key, entry := range store.entries {
		if key == protected || keep[key] {
			continue
		}
		if !found || entry.LastUse < store.entries[victim].LastUse {
			victim, found = key, true
		}
	}
	if !found {
		return false
	}
	store.drop(victim)

	return true
}

func (store *layerStore) drop(key layerKey) {
	entry := store.entries[key]
	if entry == nil {
		return
	}
	store.addDamage(entry.pose)
	store.addTopology(entry.pose)
	store.pixels -= entry.Box.Dx() * entry.Box.Dy()
	entry.Tex.Deallocate()
	entry.Valid = false
	delete(store.entries, key)
	store.stats.Evicted++
	store.demote(layerReasonEvicted)
}

// demoteMissingLocked drops every entry this frame did not keep.
func (store *layerStore) demoteMissingLocked(keep map[layerKey]bool) {
	for key, entry := range store.entries {
		if keep[key] {
			continue
		}
		store.addDamage(entry.pose)
		store.addTopology(entry.pose)
		store.pixels -= entry.Box.Dx() * entry.Box.Dy()
		entry.Tex.Deallocate()
		entry.Valid = false
		delete(store.entries, key)
		store.demote(layerReasonEnded)
	}
}

// demoteAllLocked drops every entry and reports its last pose.
func (store *layerStore) demoteAllLocked(reason string) {
	for key, entry := range store.entries {
		store.addDamage(entry.pose)
		store.addTopology(entry.pose)
		entry.Tex.Deallocate()
		entry.Valid = false
		delete(store.entries, key)
		store.demote(reason)
	}
	store.pixels = 0
	store.stats.Active = 0
	store.stats.Pixels = 0
	store.stats.Bytes = 0
	clear(store.churn)
}

// buildFrame prepares the paint skip and the per-order-position composite
// points for the active layers.
func (store *layerStore) buildFrame(d *layout.Display, active []*Layer) {
	frame := &store.frame
	frame.display = d
	frame.active = append(frame.active[:0], active...)
	if cap(frame.skip) < len(d.Ops) {
		frame.skip = make([]bool, len(d.Ops))
	} else {
		frame.skip = frame.skip[:len(d.Ops)]
		clear(frame.skip)
	}
	if frame.start == nil {
		frame.start = make(map[int]*Layer, len(active))
	}
	clear(frame.start)

	for _, layer := range active {
		for _, index := range layer.Ops {
			if index >= 0 && index < len(frame.skip) {
				frame.skip[index] = true
			}
		}
		frame.start[layer.firstRank] = layer
	}
}

func (store *layerStore) addDamage(rect image.Rectangle) {
	if rect.Empty() {
		return
	}
	store.damage = append(store.damage, rect)
}

// addTopology records the page area whose layer membership changed: a
// promotion, a demotion, or a moved layer box. A compositor that keeps the
// background must repaint these, and only these; a layer that merely moves
// needs no background work.
func (store *layerStore) addTopology(rect image.Rectangle) {
	if rect.Empty() {
		return
	}
	store.topology = append(store.topology, rect)
}

// LayerTopologyDamage returns this frame's promotion, demotion, and box-move
// area. BeginLayers resets it, so read it after that call.
func (renderer *Renderer) LayerTopologyDamage() []image.Rectangle {
	store := &renderer.layers
	store.mu.Lock()
	defer store.mu.Unlock()

	return store.topology
}

// ActiveLayerIDs returns the element ids of this frame's retained layers. A
// page hands them to its motion driver, so a sampled element that draws into a
// layer stops damaging the page background.
func (renderer *Renderer) ActiveLayerIDs() []string {
	store := &renderer.layers
	store.mu.Lock()
	defer store.mu.Unlock()

	ids := make([]string, 0, len(store.frame.active))
	for _, layer := range store.frame.active {
		if layer.ElementID != "" {
			ids = append(ids, layer.ElementID)
		}
	}

	return ids
}

func (store *layerStore) resetFrameStats() {
	store.use++
	store.stats = LayerStats{Reasons: store.reasons}
	clear(store.reasons)
}

func (store *layerStore) demote(reason string) {
	store.stats.Demoted++
	if store.reasons == nil {
		store.reasons = make(map[string]int)
		store.stats.Reasons = store.reasons
	}
	store.reasons[reason]++
}

func (store *layerStore) snapshot() LayerStats {
	store.mu.Lock()
	defer store.mu.Unlock()

	out := store.stats
	out.Active = len(store.frame.active)
	out.Pixels = store.pixels
	out.Bytes = store.pixels * layerBytesPerPixel
	out.Reasons = make(map[string]int, len(store.reasons))
	for reason, count := range store.reasons {
		out.Reasons[reason] = count
	}

	return out
}

func (store *layerStore) close() {
	store.mu.Lock()
	defer store.mu.Unlock()

	for key, entry := range store.entries {
		entry.Tex.Deallocate()
		entry.Valid = false
		delete(store.entries, key)
	}
	store.entries = nil
	store.pixels = 0
	store.frame.reset()
	store.stats = LayerStats{}
	store.reasons = nil
}

// setLayerOps reuses dst when it has room and copies src into it.
func setLayerOps(dst, src []int) []int {
	if cap(dst) < len(src) {
		return append([]int(nil), src...)
	}
	dst = dst[:len(src)]
	copy(dst, src)

	return dst
}
