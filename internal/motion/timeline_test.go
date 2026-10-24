package motion

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// timelineFixture describes one target element inside a page-scroller
// ancestor. The target's animation follows timeline; the parent and the target
// declare the named timelines.
type timelineFixture struct {
	parentTimelines []layout.TimelineDef
	targetTimelines []layout.TimelineDef
	timeline        string
	rangeStart      string
	rangeEnd        string
	rangeCenter     string
	targetY         float64
	parentY         float64
	parentH         float64
}

// timelineDisplay builds the fixture. One fill operation belongs to the
// target, and the target's keyframes translate it 100px, so the sampled x is
// one hundred times the animation's progress.
func timelineDisplay(f timelineFixture) *layout.Display {
	if f.timeline == "" {
		f.timeline = "--page"
	}

	size := 20.0
	return &layout.Display{
		Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{{
			ID: testOpID, Kind: layout.DisplayOpFillRect,
			X: 10, Y: f.targetY, W: size, H: size, R: 1, Alpha: 1,
		}},
		OpOwners: []int{1},
		Boxes: []layout.Box{
			{
				ID: testPageID, Tag: "main", ParentIndex: -1,
				X: 0, Y: f.parentY, W: testWidth, H: f.parentH,
				Timelines: f.parentTimelines,
			},
			{
				ID: testChipID, Tag: "div", ParentIndex: 0,
				X: 10, Y: f.targetY, W: size, H: size,
				OperationIDs: []uint64{testOpID},
				Timelines:    f.targetTimelines,
				Animations: []layout.Animation{{
					Name: testAnimation, DurationMS: testDurationMS,
					TimingFunction: "linear", IterationCount: 1, FillMode: "both",
					Timeline:    f.timeline,
					RangeStart:  f.rangeStart,
					RangeEnd:    f.rangeEnd,
					RangeCenter: f.rangeCenter,
				}},
			},
		},
		Keyframes: []layout.KeyframesRule{{
			Name: testAnimation,
			Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: testTransform, Value: testStartValue},
				}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: testTransform, Value: testTransformShift},
				}},
			},
		}},
	}
}

const (
	testPageID           = "page"
	testTransformShift   = "translateX(100px)"
	timelineShiftPixels  = 100.0
	timelineScrollHeight = 1000.0
)

// scrollTimelineDef declares the page scroller under a name.
func scrollTimelineDef(name string, axis string) layout.TimelineDef {
	return layout.TimelineDef{Name: name, Kind: timelineScroll, Axis: axis}
}

// viewTimelineDef declares a view timeline whose subject is id.
func viewTimelineDef(name, id string, insetStart, insetEnd float64) layout.TimelineDef {
	return layout.TimelineDef{
		Name: name, Kind: timelineView, Axis: "y",
		SubjectID: id, InsetStart: insetStart, InsetEnd: insetEnd,
	}
}

// pageView is the standard fixture viewport: 200x100 with a tall page.
func pageView(contentHeight float64) View {
	return View{
		Width: testWidth, Height: 100,
		ContentWidth: testWidth, ContentHeight: contentHeight,
	}
}

// sampleTimeline binds d, installs the view, and samples once at the test
// clock. Timeline-driven effects ignore the clock, so every assertion below is
// a pure function of the page geometry.
func sampleTimeline(t *testing.T, d *layout.Display, view View) float64 {
	t.Helper()

	r := New()
	r.now = testClock
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.SetView(view)
	r.Step(d, testClock())

	return movedBy(&d.Ops[0])
}

// wantTimelineX asserts one sampled translation against the progress the
// geometry should produce.
func wantTimelineX(t *testing.T, got, progress float64) {
	t.Helper()

	want := timelineShiftPixels * progress
	if math.Abs(got-want) > 0.01 {
		t.Fatalf("sampled translation = %v, want %v (progress %.4f)", got, want, progress)
	}
}

// TestNamedScrollTimelineSamplesPageProgress covers the three anchors of a
// named scroll timeline: top, middle, and bottom of the page scroller.
func TestNamedScrollTimelineSamplesPageProgress(t *testing.T) {
	for _, c := range []struct {
		name   string
		offset float64
		want   float64
	}{{"top", 0, 0}, {"middle", 450, 0.5}, {"bottom", 900, 1}} {
		t.Run(c.name, func(t *testing.T) {
			d := timelineDisplay(timelineFixture{
				parentTimelines: []layout.TimelineDef{scrollTimelineDef("--page", "y")},
			})
			view := pageView(timelineScrollHeight)
			view.ScrollY = c.offset

			wantTimelineX(t, sampleTimeline(t, d, view), c.want)
		})
	}
}

// TestNamedViewTimelineSamplesEntryCoverExit samples the same named view
// timeline at a quarter, half, and three quarters of its cover range.
func TestNamedViewTimelineSamplesEntryCoverExit(t *testing.T) {
	for _, c := range []struct {
		name   string
		offset float64
		want   float64
	}{{"entry", 30, 0.25}, {"cover", 60, 0.5}, {"exit", 90, 0.75}} {
		t.Run(c.name, func(t *testing.T) {
			d := timelineDisplay(timelineFixture{
				timeline:        "--card",
				targetTimelines: []layout.TimelineDef{viewTimelineDef("--card", testChipID, 0, 0)},
				targetY:         100,
			})
			view := pageView(400)
			view.ScrollY = c.offset

			wantTimelineX(t, sampleTimeline(t, d, view), c.want)
		})
	}
}

// TestNamedViewTimelineRanges checks the four named ranges against the same
// geometry. At a quarter of the cover range the entry range has finished, the
// cover range is a quarter done, and the exit range has not started.
func TestNamedViewTimelineRanges(t *testing.T) {
	for _, c := range []struct {
		name  string
		start string
		end   string
		want  float64
	}{
		{"entry", "entry", "entry", 1},
		{"cover", "cover", "cover", 0.25},
		{"exit", "exit", "exit", 0},
		{"keyword start", "entry 50%", "exit 50%", 0.2},
	} {
		t.Run(c.name, func(t *testing.T) {
			d := timelineDisplay(timelineFixture{
				timeline:        "--card",
				targetTimelines: []layout.TimelineDef{viewTimelineDef("--card", testChipID, 0, 0)},
				targetY:         100,
				rangeStart:      c.start,
				rangeEnd:        c.end,
			})
			view := pageView(400)
			view.ScrollY = 30

			wantTimelineX(t, sampleTimeline(t, d, view), c.want)
		})
	}
}

// TestViewTimelineInsetShiftsProgress checks that a start inset moves the
// adjusted scrollport's leading edge, so the same page offset is further
// through the subject's visibility range.
func TestViewTimelineInsetShiftsProgress(t *testing.T) {
	plain := timelineDisplay(timelineFixture{
		timeline:        "--card",
		targetTimelines: []layout.TimelineDef{viewTimelineDef("--card", testChipID, 0, 0)},
		targetY:         100,
	})
	inset := timelineDisplay(timelineFixture{
		timeline:        "--card",
		targetTimelines: []layout.TimelineDef{viewTimelineDef("--card", testChipID, 20, 0)},
		targetY:         100,
	})
	view := pageView(400)
	view.ScrollY = 50

	// 50 / (100 + 20) without the inset; (150 - 100) / (80 + 20) with it.
	wantTimelineX(t, sampleTimeline(t, plain, view), 50.0/120.0)
	wantTimelineX(t, sampleTimeline(t, inset, view), 50.0/100.0)
}

// TestViewTimelineEndInsetShiftsProgress checks the trailing edge inset, which
// moves the scrollport's end edge and shortens the span.
func TestViewTimelineEndInsetShiftsProgress(t *testing.T) {
	inset := timelineDisplay(timelineFixture{
		timeline:        "--card",
		targetTimelines: []layout.TimelineDef{viewTimelineDef("--card", testChipID, 0, 20)},
		targetY:         100,
	})
	view := pageView(400)
	view.ScrollY = 50

	// The end edge is at 50 + 100 - 20 = 130, so progress is 30 / 100.
	wantTimelineX(t, sampleTimeline(t, inset, view), 30.0/100.0)
}

// TestDuplicateTimelineNameNearestAncestorWins puts a scroll timeline on the
// inner parent and a view timeline on the outer box. The inner declaration is
// nearest, so the sampled value is the page progress, not the outer subject's
// visibility.
func TestDuplicateTimelineNameNearestAncestorWins(t *testing.T) {
	d := timelineDisplay(timelineFixture{
		parentTimelines: []layout.TimelineDef{scrollTimelineDef("--page", "y")},
		parentY:         0,
		parentH:         100,
	})
	// Insert an outer box that also declares --page, as a view timeline whose
	// subject is itself. The inner parent keeps the scroll declaration.
	outer := layout.Box{
		ID: "outer", Tag: "section", ParentIndex: -1, X: 0, Y: 0, W: testWidth, H: 40,
		Timelines: []layout.TimelineDef{viewTimelineDef("--page", "outer", 0, 0)},
	}
	d.Boxes[0].ParentIndex = 1
	d.Boxes = append(d.Boxes, outer)

	view := pageView(400)
	view.ScrollY = 90

	// The inner scroll declaration wins: 90 / 300. The outer view timeline
	// would read (90 + 100 - 0) / 140, which clamps to 1.
	wantTimelineX(t, sampleTimeline(t, d, view), 90.0/300.0)
}

// TestMissingTimelineNameStaysInactive checks that an unresolved dashed ident
// leaves the animation off, with a counter, and never advances on the clock.
func TestMissingTimelineNameStaysInactive(t *testing.T) {
	d := timelineDisplay(timelineFixture{
		timeline:        "--missing",
		parentTimelines: []layout.TimelineDef{scrollTimelineDef("--page", "y")},
	})

	r := New()
	r.now = testClock
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.SetView(pageView(timelineScrollHeight))
	r.Step(d, testClock().Add(10*time.Second))

	if got := movedBy(&d.Ops[0]); got != 0 {
		t.Fatalf("missing timeline moved by %v, want no motion", got)
	}
	if got := r.Stats().UnresolvedTimelines; got != 1 {
		t.Fatalf("unresolved timelines = %d, want 1", got)
	}
	if r.Active() {
		t.Fatal("an inactive timeline asked for another frame")
	}
}

// TestZeroRangeScrollerStaysInactive checks a named scroll timeline on a page
// that cannot scroll: no division by zero, no motion, and one inactive count.
func TestZeroRangeScrollerStaysInactive(t *testing.T) {
	d := timelineDisplay(timelineFixture{
		parentTimelines: []layout.TimelineDef{scrollTimelineDef("--page", "y")},
	})

	r := New()
	r.now = testClock
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.SetView(pageView(100))
	r.Step(d, testClock())

	if got := movedBy(&d.Ops[0]); got != 0 {
		t.Fatalf("zero-range timeline moved by %v, want no motion", got)
	}
	if got := r.Stats().InactiveTimelines; got != 1 {
		t.Fatalf("inactive timelines = %d, want 1", got)
	}
}

// TestAnimationRangeStartEndCenterSampleGeometry feeds animation-range-start,
// -end, and -center through the existing resolver with the host's concrete
// view ranges.
func TestAnimationRangeStartEndCenterSampleGeometry(t *testing.T) {
	startEnd := timelineDisplay(timelineFixture{
		timeline:        "--card",
		targetTimelines: []layout.TimelineDef{viewTimelineDef("--card", testChipID, 0, 0)},
		targetY:         100,
		rangeStart:      "25%",
		rangeEnd:        "75%",
	})
	center := timelineDisplay(timelineFixture{
		timeline:        "--card",
		targetTimelines: []layout.TimelineDef{viewTimelineDef("--card", testChipID, 0, 0)},
		targetY:         100,
		rangeCenter:     "25%",
	})
	view := pageView(400)
	view.ScrollY = 60

	wantTimelineX(t, sampleTimeline(t, startEnd, view), 0.5)
	// centerProgress maps the 25% center to two half-ranges: at half progress
	// the first half is 25% of the range, so the sample lands at 2/3.
	wantTimelineX(t, sampleTimeline(t, center, view), 2.0/3.0)
}

// TestAnimationRangeCenterSourceUsesSubjectProgress gives the view timeline a
// subject that is not the animated element, so source and target progress are
// different. The center then interpolates between them through the existing
// center resolver.
func TestAnimationRangeCenterSourceUsesSubjectProgress(t *testing.T) {
	d := timelineDisplay(timelineFixture{
		timeline:        "--card",
		parentTimelines: []layout.TimelineDef{viewTimelineDef("--card", testPageID, 0, 0)},
		parentY:         0,
		parentH:         100,
		targetY:         100,
		rangeCenter:     "source 50%",
	})
	view := pageView(400)
	view.ScrollY = 60

	// Subject progress 0.8, target progress 0.5, so the center is 0.65. The
	// timeline progress is the subject's 0.8, which lands two thirds of the
	// way through the second half.
	wantTimelineX(t, sampleTimeline(t, d, view), 5.0/7.0)
}

// TestUnnamedViewTimelineSamplesElementProgress checks view() against the
// animated element's own border box, with no declaration needed.
func TestUnnamedViewTimelineSamplesElementProgress(t *testing.T) {
	d := timelineDisplay(timelineFixture{
		timeline: "view()",
		targetY:  100,
	})
	view := pageView(400)
	view.ScrollY = 60

	// (60 + 100 - 100) / (100 + 20).
	wantTimelineX(t, sampleTimeline(t, d, view), 0.5)
}

// TestTimelineNoneLeavesAnimationInactive checks that animation-timeline:none
// disables the animation without an unresolved-name count and without motion.
func TestTimelineNoneLeavesAnimationInactive(t *testing.T) {
	d := timelineDisplay(timelineFixture{timeline: "none"})
	r := New()
	r.now = testClock
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.SetView(pageView(timelineScrollHeight))
	r.Step(d, testClock().Add(10*time.Second))

	if got := movedBy(&d.Ops[0]); got != 0 {
		t.Fatalf("timeline none moved by %v", got)
	}
	stats := r.Stats()
	if stats.UnresolvedTimelines != 0 || stats.InactiveTimelines != 0 {
		t.Fatalf("timeline none counted as a gap: %+v", stats)
	}
	if r.Active() {
		t.Fatal("timeline none asked for another frame")
	}
}

// TestNamedHorizontalScrollTimeline follows scroll-timeline-axis: x on a page
// wider than its viewport.
func TestNamedHorizontalScrollTimeline(t *testing.T) {
	d := timelineDisplay(timelineFixture{
		parentTimelines: []layout.TimelineDef{scrollTimelineDef("--page", "x")},
	})
	view := View{
		Width: 200, Height: 100,
		ContentWidth: 400, ContentHeight: 100,
		ScrollX: 100,
	}

	wantTimelineX(t, sampleTimeline(t, d, view), 0.5)
}

// TestTimeAnimationUnchangedWithoutTimelineDecls keeps the phase-01 contract:
// a page with no timeline declarations samples on the wall clock exactly as
// before, and neither timeline counter moves.
func TestTimeAnimationUnchangedWithoutTimelineDecls(t *testing.T) {
	d := animatedDisplay("running", false)
	r := New()
	r.now = testClock
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.SetView(pageView(timelineScrollHeight))
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	if got, want := movedBy(&d.Ops[0]), float64(testShift/2); got != want {
		t.Fatalf("time sample = %v, want %v", got, want)
	}
	stats := r.Stats()
	if stats.UnresolvedTimelines != 0 || stats.InactiveTimelines != 0 {
		t.Fatalf("timeline counters moved on a time-only page: %+v", stats)
	}
}

// TestTimelineStepAllocFree keeps the named-timeline sample free of per-frame
// allocation, including the map clear and every progress computation. It
// animates opacity so the measurement covers the driver's sample work rather
// than the keyframe sampler's transform decomposition.
func TestTimelineStepAllocFree(t *testing.T) {
	d := timelineDisplay(timelineFixture{
		timeline:        "--card",
		targetTimelines: []layout.TimelineDef{viewTimelineDef("--card", testChipID, 0, 0)},
		parentTimelines: []layout.TimelineDef{scrollTimelineDef("--page", "y")},
		targetY:         100,
	})
	d.Keyframes = []layout.KeyframesRule{{
		Name: testAnimation,
		Keyframes: []layout.Keyframe{
			{Offset: 0, Declarations: []layout.KeyframeDeclaration{
				{Property: testOpacity, Value: "1"},
			}},
			{Offset: 1, Declarations: []layout.KeyframeDeclaration{
				{Property: testOpacity, Value: "0.25"},
			}},
		},
	}}
	r := New()
	r.now = testClock
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	view := pageView(400)
	view.ScrollY = 60
	r.SetView(view)
	r.Step(d, testClock())

	now := testClock().Add(time.Second)
	allocs := testing.AllocsPerRun(100, func() {
		r.Step(d, now)
	})
	if allocs != 0 {
		t.Fatalf("timeline step allocated %.2f objects per run, want 0", allocs)
	}
}
