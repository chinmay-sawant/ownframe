package main

// The two manual paths in this example are temporary and deliberately
// isolated here:
//
//   - The geometry transition cards. The automatic transition driver captures
//     colour and opacity; the paint-property phase covers those five cards
//     already. The six cards that move `left` or `width` still call
//     csstransition until the geometry phase lands.
//   - The three animation-trigger cards. Their effect is trigger-gated, and
//     the click that starts it also publishes the new card class. The publish
//     rebinds the player, and a triggered effect that has not started is not
//     part of the carried clock, so the auto driver would wait for another
//     event. These cards dispatch the trigger themselves after the publish
//     and step that one player until it finishes.
//
// Every other card is driven by the page runtime. Delete this file, the
// calls in catalog_click.go and demo.tick, and the manual fields on demo when
// the transition track lands and the runtime carries a triggered player
// across a publish.

import (
	"image"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssanimation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/csstransition"
	"github.com/chinmay-sawant/ownframe/internal/csstrigger"
)

// manualSample is one preview the page runtime does not sample.
type manualSample struct {
	key      string
	kind     string
	property string
	button   string
	card     string
	from     []paint
	to       []paint
	origin   time.Time
	player   *animation.Player
	renderer transitionRenderer
	active   bool
	pending  bool
	gen      uint64
	damage   image.Rectangle
}

// transitionRenderer is the small interface the transition renderer offers.
// It keeps the manual sample independent of the concrete package type.
type transitionRenderer interface {
	Step(now time.Time) bool
}

// manualTrigger reports whether the card's preview waits for an
// animation-trigger. The card scopes the name on its root, so a real Start
// click can reach the target.
func manualTrigger(name string) bool {
	switch name {
	case "animation", "animation-trigger", "trigger-scope":
		return true
	}
	return false
}

// manualTransition reports whether a transition card still needs the manual
// renderer. The automatic transition driver captures the paint properties
// (colour and opacity); the geometry properties are the transition track's
// next phase, so those cards keep the manual path until it lands.
func manualTransition(name string) bool {
	switch transitionPreviewProperty(name) {
	case "left", "width":
		return true
	}
	return false
}

// manualPreview reports whether one card uses a manual sample.
func manualPreview(row catalogRow, t target) bool {
	if t.kind == "transition" {
		return manualTransition(row.Name)
	}
	return manualTrigger(row.Name)
}

// prepareManual records the paint a manual preview starts from and marks it
// for the bind that follows the next redraw. It runs before SetData, so a
// transition captures the value the card shows before its class changes.
func (d *demo) prepareManual(row catalogRow) {
	t, ok := targetByKey(row.SampleKey)
	if !ok || !manualPreview(row, t) {
		return
	}
	s := d.manual[t.key]
	if s == nil {
		s = &manualSample{key: t.key, kind: t.kind, property: t.property, button: row.Button, card: row.CardID}
		d.manual[t.key] = s
	}
	if s.from == nil {
		s.from = d.captureTarget(t.id)
	}
	s.origin = time.Now()
	s.active = false
	s.pending = true
	s.gen = 0
	s.player = nil
	s.renderer = nil
	s.damage = d.cardRect(row.CardID)
}

// bindManual starts or rebinds every visible preview that still steps by
// hand. A sample that already targets the current display is left alone; a
// redraw replaces the operations, so an active sample binds again at the
// clock it already reached.
func (d *demo) bindManual(rows []catalogRow) error {
	gen := d.p.Probe().Generation()
	for _, row := range rows {
		if !d.v.Started[row.Button] {
			continue
		}
		t, ok := targetByKey(row.SampleKey)
		if !ok || !manualPreview(row, t) {
			continue
		}
		s := d.manual[t.key]
		if s == nil || (!s.pending && !s.active) || (s.active && s.gen == gen) {
			continue
		}
		box, ok := findBox(d.p.Boxes(), t.id)
		if !ok {
			// The row left the window or a restart hid it. Keep pending so
			// the next publish that shows the target binds it again.
			s.active, s.player, s.renderer, s.gen = false, nil, nil, 0
			continue
		}
		if err := d.bindManualSample(s, box); err != nil {
			return err
		}
	}
	return nil
}

// bindManualSample points one manual sample at the current display.
func (d *demo) bindManualSample(s *manualSample, box ownframe.Box) error {
	display := d.p.Probe().Display()
	indices := cssmotion.OperationIndices(display, box)
	if len(indices) == 0 || len(s.from) != len(indices) {
		s.active, s.pending = false, false
		return nil
	}
	switch {
	case s.renderer != nil || s.kind == "transition":
		if s.to == nil {
			ops := cssmotion.Resolve(display, indices)
			s.to = make([]paint, len(ops))
			for i := range ops {
				s.to[i] = cssmotion.Read(ops[i])
			}
		}
		renderer, err := csstransition.BindTo(display, box, s.property, indices, s.from, s.to, s.origin)
		if err != nil {
			return err
		}
		s.renderer, s.player = renderer, nil
	default:
		player, err := cssanimation.Bind(display, box, indices, s.from, s.origin)
		if err != nil {
			return err
		}
		if err := d.dispatchTrigger(s, box, player); err != nil {
			return err
		}
		s.player, s.renderer = player, nil
	}
	s.damage = d.cardRect(s.card)
	s.active, s.pending, s.gen = true, false, d.p.Probe().Generation()
	return nil
}

// dispatchTrigger fires the names the target declares from its Start button.
func (d *demo) dispatchTrigger(s *manualSample, box ownframe.Box, player *animation.Player) error {
	boxes := d.p.Boxes()
	target := boxIndex(boxes, box.ID)
	source := boxIndex(boxes, s.button)
	if target < 0 || source < 0 {
		return nil
	}
	names := triggerNames(box)
	if len(names) == 0 {
		return nil
	}
	for _, name := range names {
		if err := csstrigger.Dispatch(boxes, csstrigger.Event{Name: name, Source: source, Active: true},
			map[int]*animation.Player{target: player}, s.origin); err != nil {
			return err
		}
	}
	return nil
}

// stepManual advances the samples the page runtime does not drive and marks
// the card they paint. It reports whether one still has work.
func (d *demo) stepManual(now time.Time) bool {
	live := false
	for _, s := range d.manual {
		if !s.active {
			continue
		}
		switch {
		case s.renderer != nil:
			s.active = s.renderer.Step(now)
		case s.player != nil:
			s.active = s.player.Step(now)
		}
		if !s.active {
			continue
		}
		live = true
		d.p.Probe().MarkRect(s.damage)
	}
	return live
}

// manualLive reports whether a manual sample is running.
func (d *demo) manualLive() bool {
	for _, s := range d.manual {
		if s.active {
			return true
		}
	}
	return false
}

// captureTarget reads the target's declared paint from the current display.
func (d *demo) captureTarget(id string) []paint {
	box, ok := findBox(d.p.Boxes(), id)
	if !ok {
		return nil
	}
	return capture(d.p.Probe().Display(), box)
}

// damagePad is the slack around a card for a paint that overshoots its box.
const damagePad = 2

// cardRect is the region a manual sample paints. A transition can move its
// target across the stage and a keyframe can translate it, so the card is the
// smallest rectangle that always covers the result.
func (d *demo) cardRect(id string) image.Rectangle {
	box, ok := findBox(d.p.Boxes(), id)
	if !ok {
		return image.Rectangle{}
	}
	return image.Rect(int(box.X)-damagePad, int(box.Y)-damagePad,
		int(box.X+box.W)+damagePad, int(box.Y+box.H)+damagePad)
}

// boxIndex returns the index of id in boxes, or -1.
func boxIndex(boxes []ownframe.Box, id string) int {
	for i := range boxes {
		if boxes[i].ID == id {
			return i
		}
	}
	return -1
}

// triggerNames splits one animation-trigger declaration into its dashed-ident
// event names.
func triggerNames(box ownframe.Box) []string {
	seen := map[string]bool{}
	var names []string
	for _, item := range box.Animations {
		for _, token := range strings.Fields(strings.ReplaceAll(item.Trigger, ",", " ")) {
			if strings.HasPrefix(token, "--") && !seen[token] {
				seen[token] = true
				names = append(names, token)
			}
		}
	}
	return names
}
