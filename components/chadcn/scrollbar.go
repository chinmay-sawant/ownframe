package chadcn

import (
	"context"
	"math"
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

const minimumScrollThumb = 24

// ScrollMetrics describes a vertical range in consistent logical units, such as
// rows or CSS pixels. Track and thumb geometry are always in CSS pixels.
type ScrollMetrics struct {
	Content  int
	Viewport int
	Offset   int
}

// MaxOffset is the last valid offset. An unmeasured viewport cannot scroll.
func (m ScrollMetrics) MaxOffset() int {
	if m.Viewport <= 0 || m.Content <= m.Viewport {
		return 0
	}
	return m.Content - m.Viewport
}

// Thumb computes proportional geometry, retaining a usable minimum thumb size.
// A fitting range or unmeasured track has no visible thumb.
func (m ScrollMetrics) Thumb(track int) (top, height int, visible bool) {
	maximum := m.MaxOffset()
	if track <= 0 || maximum == 0 {
		return 0, 0, false
	}
	height = int(float64(track) * float64(m.Viewport) / float64(m.Content))
	height = min(max(height, minimumScrollThumb), track)
	offset := min(max(m.Offset, 0), maximum)
	top = int(float64(track-height) * float64(offset) / float64(maximum))
	return top, height, true
}

// ScrollState owns one range, fractional input, and a thumb drag. The zero value
// is an empty range. Use it on the page loop; do not share it across panes.
type ScrollState struct {
	metrics       ScrollMetrics
	wheelFraction float64
	dragFraction  float64
	dragging      bool
	dragTrack     int
}

// Metrics returns a detached snapshot.
func (s *ScrollState) Metrics() ScrollMetrics { return s.metrics }

// Offset returns the current logical position.
func (s *ScrollState) Offset() int { return s.metrics.Offset }

// SetRange measures content and viewport in the same units and clamps position.
// Changed extents cancel a drag and discard input fractions from the old range.
func (s *ScrollState) SetRange(content, viewport int) bool {
	before := s.metrics
	s.metrics.Content, s.metrics.Viewport = max(content, 0), max(viewport, 0)
	s.metrics.Offset = min(max(s.metrics.Offset, 0), s.metrics.MaxOffset())
	if before.Content != s.metrics.Content || before.Viewport != s.metrics.Viewport {
		s.wheelFraction = 0
		s.EndDrag()
	}
	return before != s.metrics
}

// ScrollTo sets a logical position and clears pending input and drag state.
func (s *ScrollState) ScrollTo(offset int) bool {
	s.wheelFraction = 0
	s.EndDrag()
	return s.moveTo(offset)
}

// ScrollBy moves by logical units, clamping before addition can overflow.
func (s *ScrollState) ScrollBy(delta int) bool {
	maximum := s.metrics.MaxOffset()
	if delta > 0 && delta >= maximum-s.Offset() {
		return s.ScrollTo(maximum)
	}
	if delta < 0 && delta <= -s.Offset() {
		return s.ScrollTo(0)
	}
	return s.ScrollTo(s.Offset() + delta)
}

// Wheel accepts the host's vertical wheel delta: positive moves up. Fractional
// trackpad input accumulates until it advances one logical unit.
func (s *ScrollState) Wheel(delta, unitsPerNotch float64) bool {
	if !finite(delta) || !finite(unitsPerNotch) || unitsPerNotch <= 0 {
		return false
	}
	amount := -delta * unitsPerNotch
	if !finite(amount) {
		return false
	}
	s.EndDrag()
	return s.advance(amount, &s.wheelFraction)
}

// BeginDrag starts a thumb drag when the measured track permits movement.
func (s *ScrollState) BeginDrag(track int) bool {
	_, height, visible := s.metrics.Thumb(track)
	s.dragging = visible && track > height
	s.dragTrack = track
	s.wheelFraction = 0
	s.dragFraction = 0
	return s.dragging
}

// DragBy converts a pixel delta to logical units using the current thumb travel.
func (s *ScrollState) DragBy(delta float64, track int) bool {
	if track != s.dragTrack {
		s.EndDrag()
	}
	if !s.dragging || !finite(delta) {
		return false
	}
	_, height, visible := s.metrics.Thumb(track)
	if !visible || track <= height {
		return false
	}
	amount := delta * float64(s.metrics.MaxOffset()) / float64(track-height)
	if !finite(amount) {
		return false
	}
	return s.advance(amount, &s.dragFraction)
}

// EndDrag releases a thumb drag.
func (s *ScrollState) EndDrag() { s.dragging = false; s.dragFraction = 0 }

// Page moves by one viewport toward direction. Zero leaves the range alone.
func (s *ScrollState) Page(direction int) bool {
	if direction < 0 {
		return s.ScrollBy(-s.metrics.Viewport)
	}
	if direction > 0 {
		return s.ScrollBy(s.metrics.Viewport)
	}
	return false
}

// Key handles vertical arrows, Page Up/Down, Home, and End. Bind it only while
// this scrollbar's thumb has focus. step is in logical units.
func (s *ScrollState) Key(key string, step int) bool {
	switch key {
	case "arrowup":
		return s.ScrollBy(-max(step, 1))
	case "arrowdown":
		return s.ScrollBy(max(step, 1))
	case "pageup":
		return s.Page(-1)
	case "pagedown":
		return s.Page(1)
	case "home":
		return s.ScrollTo(0)
	case "end":
		return s.ScrollTo(s.metrics.MaxOffset())
	default:
		return false
	}
}

func (s *ScrollState) moveTo(offset int) bool {
	before := s.metrics.Offset
	s.metrics.Offset = min(max(offset, 0), s.metrics.MaxOffset())
	return before != s.metrics.Offset
}

func (s *ScrollState) advance(amount float64, fraction *float64) bool {
	total := *fraction + amount
	if total >= float64(s.metrics.MaxOffset()-s.Offset()) {
		*fraction = 0
		return s.moveTo(s.metrics.MaxOffset())
	}
	if total <= float64(-s.Offset()) {
		*fraction = 0
		return s.moveTo(0)
	}
	whole := int(total)
	*fraction = total - float64(whole)
	return s.moveTo(s.Offset() + whole)
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

// ScrollbarOptions connects a pane's state, measured track, and render callback.
// ID must be unique. ThumbID defaults to ID+"-thumb". A nil State is empty.
// OnChange is called after a track click changes the offset. Route wheel,
// keyboard, and drag input to the same ScrollState and redraw on change.
type ScrollbarOptions struct {
	ID          string
	ThumbID     string
	ControlsID  string
	Label       string
	State       *ScrollState
	TrackPixels int
	OnChange    func(context.Context) error
	Overrides   theme.Overrides
}

// Scrollbar overlays a vertical scrollbar on a positioned, clipped viewport.
// It disappears when content fits. Track regions page the state, and the thumb
// is the keyboard focus target. The host routes native drag and wheel events.
func Scrollbar(options ScrollbarOptions) *ui.Element {
	if options.State == nil {
		return nil
	}
	metrics := options.State.Metrics()
	top, height, visible := metrics.Thumb(options.TrackPixels)
	if !visible {
		return nil
	}
	thumbID := options.ThumbID
	if thumbID == "" {
		thumbID = options.ID + "-thumb"
	}
	label := options.Label
	if label == "" {
		label = "Scroll"
	}
	thumb := ui.Button("").ID(thumbID).Class("of-scrollbar-thumb").Attr("type", "button").Attr("role", "scrollbar").
		Attr("aria-label", label).Attr("aria-controls", options.ControlsID).Attr("aria-orientation", "vertical").
		Attr("aria-valuemin", "0").Attr("aria-valuemax", strconv.Itoa(metrics.MaxOffset())).
		Attr("aria-valuenow", strconv.Itoa(metrics.Offset)).Position(ui.PositionAbsolute).
		Top(ui.Px(float64(top))).Height(ui.Px(float64(height))).Width(ui.Percent(100)).
		Padding(ui.Px(0)).BoxSizing(ui.BoxSizingBorderBox).Border(ui.SolidBorder(ui.Px(0), ui.Transparent())).
		BorderRadius(ui.LengthVar("--of-scrollbar-radius", ui.Px(4))).BackgroundColor(theme.Scrollbar())
	before := scrollPageButton(options, -1, 0, top)
	after := scrollPageButton(options, 1, top+height, options.TrackPixels-top-height)
	return options.Overrides.Apply(ui.Div().ID(options.ID).Position(ui.PositionAbsolute).
		Top(ui.Px(0)).Right(ui.LengthVar("--of-scrollbar-inset", ui.Px(2))).
		Width(ui.LengthVar("--of-scrollbar-width", ui.Px(8))).Height(ui.Percent(100)).ZIndex(ui.Number(2)).
		Children(before, thumb, after))
}

func scrollPageButton(options ScrollbarOptions, direction, top, height int) *ui.Element {
	if height <= 0 {
		return nil
	}
	suffix := "-after"
	if direction < 0 {
		suffix = "-before"
	}
	button := ui.Button("").ID(options.ID+suffix).Attr("type", "button").Attr("tabindex", "-1").Attr("aria-hidden", "true").
		Position(ui.PositionAbsolute).Top(ui.Px(float64(top))).Width(ui.Percent(100)).Height(ui.Px(float64(height))).
		Padding(ui.Px(0)).Border(ui.SolidBorder(ui.Px(0), ui.Transparent())).BackgroundColor(ui.Transparent())
	if options.OnChange == nil {
		return button.Attr("disabled", "disabled")
	}
	return button.OnClick(func(ctx context.Context) error {
		if options.State.Page(direction) {
			return options.OnChange(ctx)
		}
		return nil
	})
}
