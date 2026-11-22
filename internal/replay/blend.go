package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"math"
	"sync"
	"sync/atomic"
)

// paintItem is one step of a display's paint plan: one op, or a nested group.
type paintItem struct {
	op    *layout.DisplayOp
	child *groupNode
}

// groupNode buffers one blend or isolation group. The root has a nil group.
type groupNode struct {
	group          *layout.DisplayGroup
	items          []paintItem
	box            image.Rectangle
	maskBox        image.Rectangle
	backdropBox    image.Rectangle
	backdropMark   *layout.DisplayOp
	backdrop       backdropFilter
	backdropRadius float64
	maskTex        *ebiten.Image
	maskMu         sync.Mutex
	known          bool
	hasMask        bool
	hasBackdropBox bool
	hasBackdrop    bool
	unbounded      bool
}

// planEntry keeps one display's plan. A frame that repaints several rects
// builds the group tree once. plan is the chunk-aware paint order built on
// demand from the same tree: the first tile paint lazily folds it, so a frame
// that never touches chunks pays no fold cost. gen is the display shape the
// plans were built from (see displayFingerprint); an element splice changes
// the shape and rebuilds.
type planEntry struct {
	d    *layout.Display
	gen  uint64
	root *groupNode
	plan *chunkPlan
}

// groupPlanBuilds counts group-tree builds. Tests prove plan reuse by
// observing no build across identical repaints, not by timing.
var groupPlanBuilds atomic.Int64

// groupPlanBuildCount reports group-tree builds.
func groupPlanBuildCount() int64 {
	return groupPlanBuilds.Load()
}

type planCache struct {
	mu      sync.Mutex
	entries [2]planEntry
	next    int
}

// planFor returns the display's group tree, from a small cache keyed on the
// display pointer and its generation fingerprint. A display is replaced
// wholesale on redraw, so a pointer match with the same shape is a plan
// match; an element splice changes lengths or canvas size, changes the
// fingerprint, and rebuilds. The chunk plan stays nil until the first tile
// paint folds it lazily from the same tree. Call on the draw path only; the
// display must not be mutated concurrently. Tick patches ops on the same
// loop, so never share a Page across goroutines. The lock stays held across
// a miss build, so two concurrent callers serialize instead of building
// twice and evicting.
func (renderer *Renderer) planFor(d *layout.Display) *groupNode {
	gen := displayFingerprint(d)
	renderer.plans.mu.Lock()
	defer renderer.plans.mu.Unlock()

	for i := range renderer.plans.entries {
		if renderer.plans.entries[i].d == d && renderer.plans.entries[i].gen == gen {
			return renderer.plans.entries[i].root
		}
	}

	root := buildPlan(d)
	groupPlanBuilds.Add(1)

	disposePlan(renderer.plans.entries[renderer.plans.next].root)
	renderer.plans.entries[renderer.plans.next] = planEntry{d: d, gen: gen, root: root}
	renderer.plans.next = (renderer.plans.next + 1) % len(renderer.plans.entries)

	return root
}

// EvictDisplay clears the cached group and chunk plans for d.
func (renderer *Renderer) EvictDisplay(d *layout.Display) {
	if renderer == nil || d == nil {
		return
	}
	renderer.plans.evictDisplay(d)
}

func (c *planCache) evictDisplay(d *layout.Display) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.entries {
		if c.entries[i].d == d {
			disposePlan(c.entries[i].root)
			c.entries[i] = planEntry{}
		}
	}
}

// hasGroups reports a display with ops inside a blend or isolation group.
func hasGroups(d *layout.Display) bool {
	for i := range d.Ops {
		if d.Ops[i].Group() != nil {
			return true
		}
	}

	return false
}

// buildPlan buckets the paint order by owning group. A group's item sits in
// its parent where the group first paints, so nesting follows the engine's
// group pointers rather than marker positions.
func buildPlan(d *layout.Display) *groupNode {
	root := &groupNode{}
	nodes := map[*layout.DisplayGroup]*groupNode{}

	for _, i := range d.Order {
		if i < 0 || i >= len(d.Ops) {
			continue
		}

		op := &d.Ops[i]
		if op.GroupBoundary() != 0 {
			if op.IsGroupBegin() && op.Group() != nil {
				group := op.Group()
				if group.Mask != nil || group.BackdropFilter != "" {
					node := attachNode(root, nodes, group)
					if group.Mask != nil {
						node.maskBox = masking.GroupBounds(op, d)
						node.hasMask = true
					}
					if group.BackdropFilter != "" {
						node.backdropBox = masking.GroupBounds(op, d)
						node.backdropMark = op
						node.hasBackdropBox = true
						node.backdrop, node.hasBackdrop = parseBackdropFilter(group.BackdropFilter)
					}
				}
			}
			continue
		}

		node := attachNode(root, nodes, op.Group())
		node.items = append(node.items, paintItem{op: op})
	}
	setBackdropRadii(root, d.PixelPerPoint)

	return root
}

func setBackdropRadii(node *groupNode, pixelsPerPoint float64) {
	if node.group != nil && node.hasBackdrop {
		for i := range node.items {
			item := &node.items[i]
			if item.op == nil || item.op.Kind != layout.DisplayOpFillRect {
				continue
			}

			op := item.op
			radius := max(op.Radius, op.RadiusTopLeft, op.RadiusTopRight,
				op.RadiusBottomRight, op.RadiusBottomLeft, op.RadiusTopLeftY,
				op.RadiusTopRightY, op.RadiusBottomRightY, op.RadiusBottomLeftY)
			node.backdropRadius = radius * pixelsPerPoint
			break
		}
	}

	for i := range node.items {
		if child := node.items[i].child; child != nil {
			setBackdropRadii(child, pixelsPerPoint)
		}
	}
}

// attachNode finds or creates the node for g and attaches it, with any
// missing ancestor, to its parent's items.
func attachNode(root *groupNode, nodes map[*layout.DisplayGroup]*groupNode, g *layout.DisplayGroup) *groupNode {
	if g == nil {
		return root
	}

	if node := nodes[g]; node != nil {
		return node
	}

	node := &groupNode{group: g}
	nodes[g] = node

	parent := attachNode(root, nodes, g.Parent)
	parent.items = append(parent.items, paintItem{child: node})

	return node
}

// compositeLayer draws one buffered group onto dst at at, once, with the
// group's mix-blend-mode. A normal-mode layer is plain source-over, and so is
// a blended layer when the shader cannot compile; a blended layer with the
// shader reads the backdrop already on dst and rewrites the region with the
// CSS compositing formula.
func compositeLayer(dst *ebiten.Image, layer *ebiten.Image, at image.Point, mode string) {
	var geo ebiten.GeoM
	geo.Translate(float64(at.X), float64(at.Y))
	compositeLayerImage(dst, layer, geo, 1, ebiten.FilterNearest, mode)
}

// compositeLayerPose draws a retained compositor layer through its animation
// matrix and group alpha. The texture already holds the base pose; geo maps it
// to the current page position. The blend shader path only accepts a
// translation, which resolveLayer enforces for non-normal modes.
func compositeLayerPose(dst *ebiten.Image, layer *ebiten.Image, geo ebiten.GeoM, alpha float64, mode string) {
	compositeLayerImage(dst, layer, geo, alpha, ebiten.FilterLinear, mode)
}

// compositeLayerImage is the shared group and layer composite. A non-normal
// mode reads dst as the backdrop and runs the blend shader over the layer's
// destination rectangle; any other mode is a source-over blit.
func compositeLayerImage(dst *ebiten.Image, layer *ebiten.Image, geo ebiten.GeoM,
	alpha float64, filter ebiten.Filter, mode string) {
	if dst == nil || layer == nil {
		return
	}

	opts := &ebiten.DrawImageOptions{GeoM: geo, Filter: filter}
	if alpha < 1 {
		opts.ColorScale.ScaleAlpha(float32(alpha))
	}

	shader := (*ebiten.Shader)(nil)
	if !normalMode(mode) {
		shader = groupShader()
	}
	if shader == nil {
		dst.DrawImage(layer, opts)

		return
	}

	at, ok := geomTranslation(geo)
	if !ok {
		// The shader path cannot rotate or scale its destination rectangle.
		// resolveLayer keeps a blended layer a translation, so this is only a
		// defensive fallback for a caller that bypassed it.
		dst.DrawImage(layer, opts)

		return
	}

	backdrop := ebiten.NewImage(layer.Bounds().Dx(), layer.Bounds().Dy())
	defer backdrop.Deallocate()

	back := &ebiten.DrawImageOptions{}
	back.GeoM.Translate(
		float64(dst.Bounds().Min.X)-at[0],
		float64(dst.Bounds().Min.Y)-at[1],
	)
	backdrop.DrawImage(dst, back)

	blend := &ebiten.DrawRectShaderOptions{}
	blend.GeoM = geo
	blend.Images[0] = layer
	blend.Images[1] = backdrop
	blend.Blend = ebiten.BlendCopy
	if alpha < 1 {
		blend.ColorScale.ScaleAlpha(float32(alpha))
	}
	blend.Uniforms = map[string]any{"Mode": float32(groupModeID(mode))}
	dst.DrawRectShader(layer.Bounds().Dx(), layer.Bounds().Dy(), shader, blend)
}

// geomTranslation returns a GeoM's translation when its linear part is the
// identity.
func geomTranslation(geo ebiten.GeoM) ([2]float64, bool) {
	const eps = 1e-9
	close := func(a, b float64) bool {
		d := a - b
		if d < 0 {
			d = -d
		}

		return d < eps
	}
	if !close(geo.Element(0, 0), 1) || !close(geo.Element(0, 1), 0) ||
		!close(geo.Element(1, 0), 0) || !close(geo.Element(1, 1), 1) {
		return [2]float64{}, false
	}

	return [2]float64{geo.Element(0, geomTransCol), geo.Element(1, geomTransCol)}, true
}

// layerGeoM maps a layer texture's pixels into dst. Box.Min is the texture's
// page origin and Matrix is the animation delta in display point space, so its
// translation is converted from points to pixels before the page offset.
func layerGeoM(layer *Layer, d *layout.Display, dx, dy float64) ebiten.GeoM {
	var geo ebiten.GeoM
	if layer == nil {
		return geo
	}

	m := layer.Matrix
	ppp := layerPixelPerPoint(d)
	box := layer.Box

	geo.SetElement(0, 0, m.A)
	geo.SetElement(0, 1, m.C)
	geo.SetElement(0, geomTransCol, m.A*float64(box.Min.X)+m.C*float64(box.Min.Y)+m.E*ppp+dx)
	geo.SetElement(1, 0, m.B)
	geo.SetElement(1, 1, m.D)
	geo.SetElement(1, geomTransCol, m.B*float64(box.Min.X)+m.D*float64(box.Min.Y)+m.F*ppp+dy)

	return geo
}

// normalMode reports a group that needs no blend step.
func normalMode(mode string) bool {
	return masking.BlendNormal(mode)
}

// groupModeID maps a CSS mix-blend-mode to the shader's mode number. Only
// the separable modes reach a compositor that asks.
func groupModeID(mode string) int {
	return masking.BlendModeID(mode)
}

// firstPick returns the translation of the node's first visible op, or ok
// false when nothing in the node paints.
func firstPick(n *groupNode, pick opPicker) ([2]float64, bool) {
	for i := range n.items {
		it := &n.items[i]
		if it.child != nil {
			if base, ok := firstPick(it.child, pick); ok {
				return base, true
			}

			continue
		}

		if dx, dy, ok := pick(it.op); ok {
			return [2]float64{dx, dy}, true
		}
	}

	return [2]float64{}, false
}

// resetNodeBox clears cached group bounds so nodeBox recomputes them after
// in-place motion or frame mutations change descendant geometry or transforms.
func resetNodeBox(n *groupNode) {
	if n == nil {
		return
	}
	n.known = false
	n.unbounded = false
	n.box = image.Rectangle{}
	for i := range n.items {
		if n.items[i].child != nil {
			resetNodeBox(n.items[i].child)
		}
	}
}

// nodeBox is the node's page box in CSS pixels, the union of its ops and
// children. ok is false when any op is unbounded or the node is empty; the
// caller then uses the whole destination.
func nodeBox(n *groupNode, ppt float64) (image.Rectangle, bool) {
	if !n.known {
		n.known = true

		for i := range n.items {
			it := &n.items[i]

			var (
				box image.Rectangle
				ok  bool
			)

			if it.child != nil {
				box, ok = nodeBox(it.child, ppt)
			} else {
				box, ok = opBounds(it.op, ppt)
			}

			if !ok {
				n.unbounded = true
			}

			n.box = n.box.Union(box)
		}
	}

	return n.box, !n.unbounded
}

// paintGroup draws one buffered group: the visible members into a layer the
// size of the group's window, then one composite with the group's mode.
func (renderer *Renderer) paintGroup(dst *ebiten.Image, d *layout.Display, n *groupNode, pick opPicker) {
	rect, base, ok := groupWindow(dst, d, n, pick)
	if !ok {
		return
	}
	if n.hasBackdrop && n.hasBackdropBox {
		renderer.applyBackdropFilter(dst, d, n, base)
	}

	layer := ebiten.NewImage(rect.Dx(), rect.Dy())
	defer layer.Deallocate()

	inner := func(op *layout.DisplayOp) (float64, float64, bool) {
		dx, dy, ok := pick(op)

		return dx - float64(rect.Min.X), dy - float64(rect.Min.Y), ok
	}

	renderer.paintItems(layer, d, n, inner, true)
	if n.group.Mask != nil {
		mask := groupMaskTexture(n)
		if mask == nil {
			return
		}
		opts := &ebiten.DrawImageOptions{Blend: ebiten.BlendDestinationIn}
		opts.GeoM.Translate(float64(n.maskBox.Min.X)+base[0]-float64(rect.Min.X),
			float64(n.maskBox.Min.Y)+base[1]-float64(rect.Min.Y))
		layer.DrawImage(mask, opts)
	}
	compositeLayer(dst, layer, rect.Min, n.group.Mode)
}

func groupMaskTexture(n *groupNode) *ebiten.Image {
	if n == nil || n.group == nil || n.group.Mask == nil || !n.hasMask {
		return nil
	}
	n.maskMu.Lock()
	defer n.maskMu.Unlock()
	if n.maskTex != nil {
		return n.maskTex
	}
	box := image.Rect(0, 0, n.maskBox.Dx(), n.maskBox.Dy())
	pixels, err := masking.Image(box, n.group.Mask)
	if err != nil {
		return nil
	}
	n.maskTex = ebiten.NewImageFromImage(pixels)
	return n.maskTex
}

func disposePlan(n *groupNode) {
	if n == nil {
		return
	}
	n.maskMu.Lock()
	if n.maskTex != nil {
		n.maskTex.Deallocate()
		n.maskTex = nil
	}
	n.maskMu.Unlock()
	for i := range n.items {
		disposePlan(n.items[i].child)
	}
}

// opPicker decides whether an op paints in this call and where. ok false
// skips the op; a group paints when any member is picked.
type opPicker func(op *layout.DisplayOp) (dx, dy float64, ok bool)

// paintDisplay replays display in paint order, compositing each non-normal
// group once. A page with no groups takes the direct path and allocates
// nothing.
func (renderer *Renderer) paintDisplay(dst *ebiten.Image, d *layout.Display, pick opPicker) {
	if dst == nil || d == nil {
		return
	}

	if !hasGroups(d) {
		for _, i := range d.Order {
			if i < 0 || i >= len(d.Ops) {
				continue
			}

			if dx, dy, ok := pick(&d.Ops[i]); ok {
				renderer.drawOp(dst, &d.Ops[i], dx, dy)
			}
		}

		return
	}

	groupPick := func(op *layout.DisplayOp) (float64, float64, bool) {
		dx, dy, ok := pick(op)
		if ok && op.Fixed {
			// Fixed ops keep their viewport position. Group layers apply a
			// second, local offset when they draw into the offscreen image.
			dx, dy = 0, 0
		}
		return dx, dy, ok
	}
	renderer.paintNode(dst, d, renderer.planFor(d), groupPick, false)
}

// paintNode paints one node: a non-normal group buffers its members into a
// layer and composites once; every other node paints item by item.
func (renderer *Renderer) paintNode(dst *ebiten.Image, d *layout.Display, n *groupNode, pick opPicker, local bool) {
	if n.group != nil && (!normalMode(n.group.Mode) || n.group.Mask != nil || n.group.BackdropFilter != "") {
		renderer.paintGroup(dst, d, n, pick)

		return
	}

	renderer.paintItems(dst, d, n, pick, local)
}

// paintItems paints a node's ops in order, recursing into child groups.
func (renderer *Renderer) paintItems(dst *ebiten.Image, d *layout.Display, n *groupNode, pick opPicker, local bool) {
	for i := range n.items {
		it := &n.items[i]
		if it.child != nil {
			renderer.paintNode(dst, d, it.child, pick, local)

			continue
		}

		if dx, dy, ok := pick(it.op); ok {
			if local {
				renderer.drawOpInLayer(dst, it.op, dx, dy)
			} else {
				renderer.drawOp(dst, it.op, dx, dy)
			}
		}
	}
}

// groupWindow returns the group's translated paint box clipped to dst.
func groupWindow(dst *ebiten.Image, d *layout.Display, n *groupNode,
	pick opPicker) (image.Rectangle, [2]float64, bool) {
	base, ok := firstPick(n, pick)
	if !ok && n.backdropMark != nil {
		if dx, dy, picked := pick(n.backdropMark); picked {
			base, ok = [2]float64{dx, dy}, true
		}
	}
	if !ok {
		return image.Rectangle{}, [2]float64{}, false
	}
	box, bounded := nodeBox(n, d.PointsPerPixel)
	if n.group != nil && n.group.Mask != nil && n.hasMask {
		box, bounded = n.maskBox, true
	}
	if bounded && box.Empty() && n.hasBackdropBox {
		box, bounded = n.backdropBox, true
	}
	if !bounded {
		box = dst.Bounds()
	} else {
		box = image.Rect(int(math.Floor(float64(box.Min.X)+base[0])),
			int(math.Floor(float64(box.Min.Y)+base[1])),
			int(math.Ceil(float64(box.Max.X)+base[0])),
			int(math.Ceil(float64(box.Max.Y)+base[1])))
	}
	box = box.Intersect(dst.Bounds())
	if box.Empty() {
		return image.Rectangle{}, [2]float64{}, false
	}
	return box, base, true
}
