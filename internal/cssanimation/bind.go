package cssanimation

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// Bind builds an animation player from one box's Blinkless CSS metadata. It
// targets every captured operation with the full paint mask, which is the
// whole-subtree behaviour the catalog previews drive by hand. A page's
// automatic runtime uses BindBox instead, so a colour keyframe only writes the
// operations its property owns.
func Bind(d *layout.Display, box layout.Box, indices []int, initial []cssmotion.Paint,
	start time.Time) (*animation.Player, error) {
	if d == nil || len(box.Animations) == 0 {
		return nil, errors.New("cssanimation: box has no animation metadata")
	}
	ops := cssmotion.Resolve(d, indices)
	if len(ops) != len(indices) {
		return nil, errors.New("cssanimation: captured operations are missing from display")
	}
	if len(ops) == 0 || len(initial) != len(ops) {
		return nil, errors.New("cssanimation: target operation paints do not match")
	}
	targets := make([]animation.Target, len(ops))
	for i := range targets {
		targets[i] = animation.Target{
			Op:      ops[i],
			From:    initial[i],
			Parent:  box.Transform.Parent,
			World:   box.Transform.World,
			OriginX: box.Transform.OriginX,
			OriginY: box.Transform.OriginY,

			HasElementTransform: box.Transform.HasMatrix,
			HasTransform3D:      box.Transform.HasTransform3D,
		}
	}
	size := transformSize{width: box.W, height: box.H, pointsPerPixel: displayScale(d), hasBox: true}
	effects := make([]animation.Effect, 0, len(box.Animations))
	for _, item := range box.Animations {
		var frames []animation.Keyframe
		if item.Name != "none" && item.Name != "" {
			frames = keyframes(d.Keyframes, item.Name, size.pointsPerPixel, box.W, box.H)
		}
		effects = append(effects, animation.Effect{Frames: frames,
			Timing: timing(item), Trigger: item.Trigger})
	}
	player, err := animation.NewComposite(targets, effects)
	if err != nil {
		return nil, err
	}
	player.Start(start)
	return player, nil
}

// Bound is one player bound to one target set: the operations it writes, the
// paint fields it may write, and the player itself.
type Bound struct {
	Player  *animation.Player
	Indices []int
	Mask    cssmotion.Mask
}

// BindBox builds the players for one element's CSS animations, one player per
// distinct target set. boxIndex is the box's position in d.Boxes, or -1 when
// the element has no id the runtime can trust; an owner-scoped colour then
// falls back to the box's subtree and the binder reports the fallback.
//
// A property is split out of its animation by target set and field mask: a
// background colour writes only the element's own fill operations, a text
// colour writes only its own text operations, and a transform or opacity writes
// the whole subtree. Two classes with the same target set share one player, so
// two animations with disjoint fields still compose on one operation.
//
// A property this phase cannot write is left at the element's base paint and
// reported as a fidelity gap. The returned gaps are ordered and deduplicated
// per element and reason.
func BindBox(d *layout.Display, box layout.Box, boxIndex int, start time.Time) ([]Bound, []string, error) {
	if d == nil || len(box.Animations) == 0 {
		return nil, nil, errors.New("cssanimation: box has no animation metadata")
	}
	size := transformSize{width: box.W, height: box.H, pointsPerPixel: displayScale(d), hasBox: true}
	var report gapList

	current, hasCurrent := Color{}, false
	for i := range box.Animations {
		if usesCurrentColor(ruleFor(d.Keyframes, box.Animations[i].Name)) {
			current, hasCurrent = elementColor(d, box, boxIndex)
			break
		}
	}

	var groups []targetGroup
	for i := range box.Animations {
		item := box.Animations[i]
		source := ruleFor(d.Keyframes, item.Name)
		if len(source) == 0 {
			continue
		}
		reportDeferred(&report, box, source)
		for _, class := range effectClasses(source) {
			frames := keyframesFor(source, size, class, current, hasCurrent)
			if len(frames) == 0 {
				continue
			}
			if err := frameError(frames); err != nil {
				return nil, report.list, fmt.Errorf("keyframe %s: %w", class.name(), err)
			}
			if class.kind != cssmotion.KindSubtree && boxIndex < 0 {
				report.add(fmt.Sprintf(
					"box %q has no display index; owner-scoped colours fall back to its subtree", box.ID))
			}
			indices := cssmotion.Select(d, box, boxIndex, class.kind)
			if len(indices) == 0 {
				continue
			}
			reportColorAlpha(&report, box, frames)
			group := groupFor(&groups, indices)
			group.mask |= class.mask
			group.effects = append(group.effects, animation.Effect{
				Frames: frames, Timing: timing(item), Trigger: item.Trigger,
			})
		}
	}

	bounds := make([]Bound, 0, len(groups))
	for i := range groups {
		targets, err := groups[i].targets(d, box)
		if err != nil {
			return nil, report.list, err
		}
		player, err := animation.NewComposite(targets, groups[i].effects)
		if err != nil {
			return nil, report.list, err
		}
		player.Start(start)
		bounds = append(bounds, Bound{Player: player, Indices: groups[i].indices, Mask: groups[i].mask})
	}
	return bounds, report.list, nil
}

// targetGroup collects the effects of one element that write the same target
// set, so they sample through one player and compose in declaration order.
type targetGroup struct {
	indices []int
	mask    cssmotion.Mask
	effects []animation.Effect
}

// groupFor finds the group already targeting this exact set, or appends a new
// one.
func groupFor(groups *[]targetGroup, indices []int) *targetGroup {
	for i := range *groups {
		if slices.Equal((*groups)[i].indices, indices) {
			return &(*groups)[i]
		}
	}
	*groups = append(*groups, targetGroup{indices: indices})
	return &(*groups)[len(*groups)-1]
}

// targets captures one operation per index and stamps this group's field mask
// on its starting paint. The mask is what the player writes with: a transform
// group never touches colour, and a colour group never touches the matrix.
func (g targetGroup) targets(d *layout.Display, box layout.Box) ([]animation.Target, error) {
	targets := make([]animation.Target, 0, len(g.indices))
	for _, index := range g.indices {
		if index < 0 || index >= len(d.Ops) {
			return nil, fmt.Errorf("cssanimation: operation %d is missing from display", index)
		}
		from := cssmotion.Read(&d.Ops[index])
		from.Mask = g.mask

		targets = append(targets, animation.Target{
			Op:      &d.Ops[index],
			From:    from,
			Parent:  box.Transform.Parent,
			World:   box.Transform.World,
			OriginX: box.Transform.OriginX,
			OriginY: box.Transform.OriginY,

			HasElementTransform: box.Transform.HasMatrix,
			HasTransform3D:      box.Transform.HasTransform3D,
		})
	}
	return targets, nil
}

// elementColor resolves currentColor from the element's own text operation. The
// engine does not export the computed colour on the box, so a box whose text is
// painted by something else reports false and a currentColor keyframe becomes a
// bind error.
func elementColor(d *layout.Display, box layout.Box, boxIndex int) (Color, bool) {
	for _, index := range cssmotion.Select(d, box, boxIndex, cssmotion.KindInk) {
		op := &d.Ops[index]
		return Color{R: op.R, G: op.G, B: op.B, A: maxChannel}, true
	}
	return Color{}, false
}

// usesCurrentColor reports whether a rule animates any colour to or from
// currentColor. Only then does the binder need the element's own ink colour.
func usesCurrentColor(source []layout.Keyframe) bool {
	for i := range source {
		for _, declaration := range source[i].Declarations {
			if _, ok := propertyClass(declaration.Property); !ok {
				continue
			}
			if strings.EqualFold(strings.TrimSpace(declaration.Value), "currentcolor") {
				return true
			}
		}
	}
	return false
}

// reportDeferred counts the keyframe properties that belong to a later phase.
// The declarations stay at the element's base paint, so the animation runs but
// that property does not move.
func reportDeferred(report *gapList, box layout.Box, source []layout.Keyframe) {
	for i := range source {
		for _, declaration := range source[i].Declarations {
			phase, ok := deferredPhase(declaration.Property)
			if !ok {
				continue
			}
			report.add(fmt.Sprintf("box %q: %s is deferred to phase %02d",
				box.ID, declaration.Property, phase))
		}
	}
}

// reportColorAlpha counts a keyframed colour whose alpha the paint phase cannot
// write. The sRGB channels animate; the composite alpha stays at the element's
// base paint until an alpha field exists.
func reportColorAlpha(report *gapList, box layout.Box, frames []animation.Keyframe) {
	for i := range frames {
		if !frames[i].ColorAlphaSet {
			continue
		}
		report.add(fmt.Sprintf("box %q: keyframed colour alpha is left at base paint", box.ID))
		return
	}
}

// gapList collects fidelity gaps once each, in the order they appear.
type gapList struct {
	list []string
	seen map[string]struct{}
}

func (g *gapList) add(reason string) {
	if g.seen == nil {
		g.seen = make(map[string]struct{})
	}
	if _, ok := g.seen[reason]; ok {
		return
	}
	g.seen[reason] = struct{}{}
	g.list = append(g.list, reason)
}

// displayScale returns the display's point-to-pixel factor, or 1 for a
// hand-built display that left it zero. A zero factor would collapse every
// percentage transform length to zero.
func displayScale(d *layout.Display) float64 {
	if d != nil && d.PointsPerPixel > 0 {
		return d.PointsPerPixel
	}
	return 1
}
