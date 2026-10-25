package page

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

var _ host.LayerSource = (*Page)(nil)

// LayerSpecs reports this frame's promoted compositor layers. It is the page
// side of host.LayerSource: the window asks once per tile frame, before it
// plans repaints, so the renderer can bake a promoted element once and
// composite it through the animation or transition matrix instead of replaying
// its operations every frame.
//
// display must be the list the page is drawing. A mismatch means the window
// holds a list the page has replaced, so the box indices the specifications
// name no longer describe the operations the window would paint; the direct
// replay path keeps the frame.
//
// The motion runtime merges the animation and transition registries before this
// method sees them: a live transform or opacity transition becomes one entry,
// and a transition and an animation on the same element share one entry, hence
// one layer. The runtime already drops inactive, paused, layout-moving,
// restamping, 3D, and projection transitions. This method adds the rule a layer
// specification needs that the runtime cannot make from a display: the matrix
// must be invertible so the layer texture can hold the captured pose.
//
// An entry that also writes opacity keeps its Alpha, which is the transition's
// group opacity when the transition owns the field. The renderer divides that
// sampled group opacity out of the bake and applies it once at composite time,
// so a transform and an opacity transition can promote together and the
// overlapping children of an opacity transition fade once, not once per
// operation. A pure opacity entry has an identity matrix and promotes on Alpha
// alone.
//
// Release is not this method's job: a transition that ended or was cancelled
// stops being listed, and the renderer's BeginLayers releases the layer for the
// missing specification and damages its last pose. A page with the animation
// driver off still promotes transitions, because the two switches are
// independent.
//
// The runtime reuses its entry slice, so every field is copied into a new slice
// here. The window consumes the result before the next frame asks again.
// SetPromotedLayers tells the motion driver which elements the compositor
// kept as retained layers on the last frame. Their paint changes land in the
// layer textures, so the next Step leaves their page damage to the layer
// cache; a demotion reports the pose it leaves behind.
func (p *Page) SetPromotedLayers(ids []string) {
	if p == nil || p.motion == nil {
		return
	}

	p.motion.SetPromotedLayers(ids)
}

func (p *Page) LayerSpecs(display *layout.Display) []host.LayerSpec {
	if p == nil || p.motion == nil || (p.animationsOff && p.transitionsOff) {
		return nil
	}
	if display == nil || display != p.draw.Display {
		return nil
	}

	entries := p.motion.LayerEntriesFor(display)
	if len(entries) == 0 {
		return nil
	}

	specs := make([]host.LayerSpec, 0, len(entries))
	for i := range entries {
		entry := &entries[i]
		base, ok := host.InvertLayerMatrix(entry.Matrix)
		if !ok {
			continue
		}
		specs = append(specs, host.LayerSpec{
			ElementID: entry.ElementID,
			BoxIndex:  entry.BoxIndex,
			Matrix:    entry.Matrix,
			Base:      base,
			Alpha:     entry.Alpha,
			Mode:      string(entry.Mode),
		})
	}
	if len(specs) == 0 {
		return nil
	}

	return specs
}
