package motion

import (
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"strings"
	"testing"
	"time"
)

// layoutDisplay builds a two-box page whose boxes animate width, and nothing
// else. The display carries no operations: a layout player owns no paint
// target, which is the shape the queue has to handle.
func layoutDisplay(ids ...string) *Display {
	boxes := make([]layout.Box, len(ids))
	for i, id := range ids {
		boxes[i] = layout.Box{
			ID: id,
			Animations: []layout.Animation{{
				Name: "widen", DurationMS: 1000, IterationCount: 1, FillMode: "both",
			}},
		}
	}

	return &Display{
		Keyframes: []layout.KeyframesRule{{
			Name: "widen",
			Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: "width", Value: "40px"},
				}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: "width", Value: "120px"},
				}},
			},
		}},
		Boxes: boxes,
	}
}

// layoutRuntime binds a page at a fixed clock so the samples are deterministic.
func layoutRuntime(t *testing.T, ids ...string) (*Runtime, time.Time) {
	t.Helper()

	start := time.Unix(1000, 0)
	runtime := New()
	runtime.now = func() time.Time { return start }
	if err := runtime.BindAfter(nil, layoutDisplay(ids...)); err != nil {
		t.Fatal(err)
	}
	if got := runtime.Stats().LayoutBound; got != len(ids) {
		t.Fatalf("bound layout elements = %d, want %d", got, len(ids))
	}

	return runtime, start
}

// TestStepLayoutRunsOneElementPerFrame checks the budget and the element-id
// queue: one request per frame, the rest reported as waiting, and the deferred
// element sampling on the frame that runs it.
func TestStepLayoutRunsOneElementPerFrame(t *testing.T) {
	t.Parallel()

	runtime, start := layoutRuntime(t, "a", "b")

	requests := runtime.StepLayout(start.Add(500 * time.Millisecond))
	if len(requests) != 1 {
		t.Fatalf("requests = %+v, want one relayout", requests)
	}
	if requests[0].ElementID != "a" {
		t.Fatalf("first element = %q, want the id-ordered head a", requests[0].ElementID)
	}
	if got := requests[0].Overrides[0].Value; got != "80px" {
		t.Fatalf("width at 50%% = %q, want 80px", got)
	}
	stats := runtime.Stats()
	if stats.LayoutQueued != 1 || !stats.LayoutBudgetHit {
		t.Fatalf("queued = %d, hit = %v, want b waiting", stats.LayoutQueued, stats.LayoutBudgetHit)
	}

	// b was not sampled on the first frame, so its first request is its value
	// at this frame's clock.
	requests = runtime.StepLayout(start.Add(750 * time.Millisecond))
	if len(requests) != 1 || requests[0].ElementID != "b" {
		t.Fatalf("second frame = %+v, want b", requests)
	}
	if got := requests[0].Overrides[0].Value; got != "100px" {
		t.Fatalf("b width at the later frame = %q, want 100px", got)
	}

	// a comes back round after b, still live at the later clock.
	requests = runtime.StepLayout(start.Add(900 * time.Millisecond))
	if len(requests) != 1 || requests[0].ElementID != "a" {
		t.Fatalf("third frame = %+v, want a again", requests)
	}
}

// TestStepLayoutStopsAFinishedElement checks that a completed animation leaves
// the queue: the page stops relaying it out instead of replaying the end pose
// on every tick.
func TestStepLayoutStopsAFinishedElement(t *testing.T) {
	t.Parallel()

	runtime, start := layoutRuntime(t, "a")
	if requests := runtime.StepLayout(start.Add(2 * time.Second)); len(requests) != 1 {
		t.Fatalf("first sample = %+v, want the end pose", requests)
	}

	if requests := runtime.StepLayout(start.Add(3 * time.Second)); len(requests) != 0 {
		t.Fatalf("a finished element sampled again: %+v", requests)
	}
	if got := runtime.Stats().LayoutQueued; got != 0 {
		t.Fatalf("queued = %d after the animation finished, want 0", got)
	}
}

// TestReturnLayoutPutsTheSampleBack checks the retained-state debt path: a
// request the page cannot apply yet runs first on the next frame.
func TestReturnLayoutPutsTheSampleBack(t *testing.T) {
	t.Parallel()

	runtime, start := layoutRuntime(t, "a", "b")
	requests := runtime.StepLayout(start.Add(200 * time.Millisecond))
	if len(requests) != 1 || requests[0].ElementID != "a" {
		t.Fatalf("requests = %+v, want a", requests)
	}

	runtime.ReturnLayout(requests)
	if got := runtime.Stats().LayoutQueued; got != 2 {
		t.Fatalf("queued = %d, want both elements", got)
	}

	again := runtime.StepLayout(start.Add(300 * time.Millisecond))
	if len(again) != 1 || again[0].ElementID != "a" {
		t.Fatalf("returned sample = %+v, want a first", again)
	}
}

// TestStepLayoutSamplesEveryPropertyOnce checks the grouping rule: an element
// that animates width and height sends one request with one override each.
func TestStepLayoutSamplesEveryPropertyOnce(t *testing.T) {
	t.Parallel()

	display := layoutDisplay("card")
	display.Keyframes[0].Keyframes[0].Declarations = []layout.KeyframeDeclaration{
		{Property: "width", Value: "40px"},
		{Property: "height", Value: "20px"},
	}
	display.Keyframes[0].Keyframes[1].Declarations = []layout.KeyframeDeclaration{
		{Property: "width", Value: "120px"},
		{Property: "height", Value: "60px"},
	}

	start := time.Unix(2000, 0)
	runtime := New()
	runtime.now = func() time.Time { return start }
	if err := runtime.BindAfter(nil, display); err != nil {
		t.Fatal(err)
	}

	requests := runtime.StepLayout(start.Add(250 * time.Millisecond))
	if len(requests) != 1 {
		t.Fatalf("requests = %+v, want one element", requests)
	}
	if got := len(requests[0].Overrides); got != 2 {
		t.Fatalf("overrides = %d, want one per property", got)
	}
	want := map[string]string{"width": "60px", "height": "30px"}
	for _, override := range requests[0].Overrides {
		if want[override.Property] != override.Value {
			t.Fatalf("%s = %q, want %q", override.Property, override.Value, want[override.Property])
		}
	}
}

// TestStepLayoutPausedElementSamplesOnce keeps a frozen layout pose from
// dirtying the page every tick.
func TestStepLayoutPausedElementSamplesOnce(t *testing.T) {
	t.Parallel()

	display := layoutDisplay("card")
	display.Boxes[0].Animations[0].PlayState = "paused"
	start := time.Unix(3000, 0)
	runtime := New()
	runtime.now = func() time.Time { return start }
	if err := runtime.BindAfter(nil, display); err != nil {
		t.Fatal(err)
	}

	if requests := runtime.StepLayout(start.Add(100 * time.Millisecond)); len(requests) != 1 {
		t.Fatalf("first sample = %+v, want the frozen pose once", requests)
	}
	if requests := runtime.StepLayout(start.Add(200 * time.Millisecond)); len(requests) != 0 {
		t.Fatalf("paused element sampled again: %+v", requests)
	}
	if got := runtime.Stats().LayoutQueued; got != 0 {
		t.Fatalf("queued = %d after the paused sample, want 0", got)
	}
}

// TestStepLayoutReportsTheParseErrorOnce keeps a malformed layout value from
// being dropped silently.
func TestBadLayoutValueFailsTheBindOnce(t *testing.T) {
	t.Parallel()

	display := layoutDisplay("card")
	display.Keyframes[0].Keyframes[1].Declarations[0].Value = "banana"

	runtime := New()
	err := runtime.BindAfter(nil, display)
	if err == nil {
		t.Fatal("the bind accepted a malformed width")
	}
	if !strings.Contains(err.Error(), "invalid layout value") {
		t.Fatalf("error = %v, want the invalid layout value", err)
	}
	if got := runtime.Stats().LayoutBound; got != 0 {
		t.Fatalf("bound layout elements = %d, want the failed box left out", got)
	}
}

// layoutFrame builds one keyframe that sets a single numeric layout value.
func layoutFrame(offset float64, property string, number float64, unit string) Keyframe {
	return Keyframe{
		Offset: offset,
		Layout: []LayoutValue{{
			Property: property,
			Terms:    []LayoutTerm{{Number: number, Unit: unit}},
		}},
	}
}

// layoutEffect wraps frames in one replacing effect with a linear clock.
func layoutEffect(frames ...Keyframe) Effect {
	return Effect{
		Frames: frames,
		Timing: Timing{
			Duration:   time.Second,
			Iterations: 1,
			Fill:       FillBoth,
		},
	}
}

// TestLayoutWidthInterpolatesLinearly checks the common length pair.
func TestLayoutWidthInterpolatesLinearly(t *testing.T) {
	t.Parallel()

	effect := layoutEffect(
		layoutFrame(0, "width", 40, "px"),
		layoutFrame(1, "width", 120, "px"),
	)

	values := effect.sampleLayout(0.25)
	if len(values) != 1 || values[0].Property != "width" {
		t.Fatalf("values = %+v, want one width", values)
	}
	if got := values[0].Terms[0]; got.Number != 60 || got.Unit != "px" {
		t.Fatalf("width = %+v, want 60px", got)
	}
}

// TestLayoutPercentInterpolatesInItsOwnUnit keeps the percentage unresolved:
// the engine resolves it against the containing block, so the sampler must not
// bake the element box into it.
func TestLayoutPercentInterpolatesInItsOwnUnit(t *testing.T) {
	t.Parallel()

	effect := layoutEffect(
		layoutFrame(0, "width", 10, "%"),
		layoutFrame(1, "width", 50, "%"),
	)

	values := effect.sampleLayout(0.5)
	if got := values[0].Terms[0]; got.Number != 30 || got.Unit != "%" {
		t.Fatalf("width = %+v, want 30%%", got)
	}
}

// TestLayoutKeywordValuesStepAtMidpoint checks the discrete rule: a keyword
// pair has no continuous interpolation and switches at 50%.
func TestLayoutKeywordValuesStepAtMidpoint(t *testing.T) {
	t.Parallel()

	effect := layoutEffect(
		Keyframe{Offset: 0, Layout: []LayoutValue{{
			Property: "height", Terms: []LayoutTerm{{Keyword: "auto"}},
		}}},
		Keyframe{Offset: 1, Layout: []LayoutValue{{
			Property: "height", Terms: []LayoutTerm{{Number: 80, Unit: "px"}},
		}}},
	)

	before := effect.sampleLayout(0.49)[0].CSS()
	after := effect.sampleLayout(0.51)[0].CSS()
	if before != "auto" {
		t.Fatalf("height before the midpoint = %q, want auto", before)
	}
	if after != "80px" {
		t.Fatalf("height after the midpoint = %q, want 80px", after)
	}
}

// TestLayoutUnitlessZeroTakesTheOtherUnit keeps the common 0 to 100px pair
// continuous instead of making it discrete.
func TestLayoutUnitlessZeroTakesTheOtherUnit(t *testing.T) {
	t.Parallel()

	effect := layoutEffect(
		layoutFrame(0, "right", 0, ""),
		layoutFrame(1, "right", 100, "px"),
	)

	values := effect.sampleLayout(0.25)
	if got := values[0].Terms[0]; got.Number != 25 || got.Unit != "px" {
		t.Fatalf("right = %+v, want 25px", got)
	}
}

// TestLayoutMismatchedUnitsStep checks that a pair with no common unit stays
// discrete rather than inventing a conversion the engine would resolve
// differently.
func TestLayoutMismatchedUnitsStep(t *testing.T) {
	t.Parallel()

	effect := layoutEffect(
		layoutFrame(0, "width", 10, "px"),
		layoutFrame(1, "width", 50, "%"),
	)

	if got := effect.sampleLayout(0.25)[0].CSS(); got != "10px" {
		t.Fatalf("width = %q, want the discrete 10px", got)
	}
	if got := effect.sampleLayout(0.75)[0].CSS(); got != "50%" {
		t.Fatalf("width = %q, want the discrete 50%%", got)
	}
}

// TestLayoutRadiusInterpolatesAcrossTheSlash checks the separator survives a
// sample and the radii on both sides interpolate.
func TestLayoutRadiusInterpolatesAcrossTheSlash(t *testing.T) {
	t.Parallel()

	effect := Effect{
		Frames: []Keyframe{
			{Offset: 0, Layout: []LayoutValue{{
				Property: "border-radius",
				Terms: []LayoutTerm{
					{Number: 0, Unit: "px"}, {Keyword: "/"}, {Number: 0, Unit: "px"},
				},
			}}},
			{Offset: 1, Layout: []LayoutValue{{
				Property: "border-radius",
				Terms: []LayoutTerm{
					{Number: 20, Unit: "px"}, {Keyword: "/"}, {Number: 40, Unit: "px"},
				},
			}}},
		},
		Timing: Timing{Duration: time.Second, Iterations: 1, Fill: FillBoth},
	}

	got := effect.sampleLayout(0.5)[0].CSS()
	if got != "10px / 20px" {
		t.Fatalf("border-radius = %q, want 10px / 20px", got)
	}
}

// TestLayoutValueCSSPrintsTheEngineForm pins the serialization the override
// list sends: shortest numbers, the authored unit, and spaces between terms.
func TestLayoutValueCSSPrintsTheEngineForm(t *testing.T) {
	t.Parallel()

	value := LayoutValue{Property: "margin", Terms: []LayoutTerm{
		{Number: 1.5, Unit: "px"}, {Number: 0, Unit: "%"}, {Keyword: "auto"},
	}}
	if got := value.CSS(); got != "1.5px 0% auto" {
		t.Fatalf("CSS = %q, want %q", got, "1.5px 0% auto")
	}
}

// TestLayoutAtSamplesThePlayerClock checks the player sampler the motion
// driver reads: the value follows the element clock and reports the effect
// live so the queue keeps it.
func TestLayoutAtSamplesThePlayerClock(t *testing.T) {
	t.Parallel()

	start := time.Unix(0, 0)
	player, err := NewLayout([]Effect{layoutEffect(
		layoutFrame(0, "width", 40, "px"),
		layoutFrame(1, "width", 120, "px"),
	)})
	if err != nil {
		t.Fatal(err)
	}
	player.Start(start)

	values, live := player.LayoutAt(FrameInput{Now: start.Add(500 * time.Millisecond)}, nil)
	if !live {
		t.Fatal("the midpoint sample did not report the effect live")
	}
	if len(values) != 1 {
		t.Fatalf("values = %+v, want one property", values)
	}
	if got := values[0].CSS(); got != "80px" {
		t.Fatalf("width at the midpoint = %q, want 80px", got)
	}

	// A second sample on the same clock returns the same pose; the sampler is
	// a read, not a step.
	again, _ := player.LayoutAt(FrameInput{Now: start.Add(500 * time.Millisecond)}, nil)
	if got := again[0].CSS(); got != "80px" {
		t.Fatalf("the repeated sample = %q, want the same pose", got)
	}
}

// TestLayoutAtReplacesAPropertyAcrossEffects checks the declaration-order rule:
// the later effect's value wins, so one element still sends one override per
// property.
func TestLayoutAtReplacesAPropertyAcrossEffects(t *testing.T) {
	t.Parallel()

	start := time.Unix(0, 0)
	player, err := NewLayout([]Effect{
		layoutEffect(layoutFrame(0, "width", 40, "px"), layoutFrame(1, "width", 40, "px")),
		layoutEffect(layoutFrame(0, "height", 10, "px"), layoutFrame(1, "height", 10, "px")),
	})
	if err != nil {
		t.Fatal(err)
	}
	player.Start(start)

	values, _ := player.LayoutAt(FrameInput{Now: start}, nil)
	if len(values) != 2 {
		t.Fatalf("values = %+v, want two distinct properties", values)
	}

	var props []string
	for _, value := range values {
		props = append(props, value.Property)
	}
	if strings.Join(props, ",") != "width,height" {
		t.Fatalf("properties = %v, want declaration order", props)
	}
}

// TestLayoutAtBeforeABackwardsFillKeepsTheEndpoints checks a fill that does not
// cover the delay: no value is emitted before the active interval.
func TestLayoutAtBeforeAFillNoneKeepsNoValue(t *testing.T) {
	t.Parallel()

	start := time.Unix(0, 0)
	effect := layoutEffect(layoutFrame(0, "width", 40, "px"), layoutFrame(1, "width", 120, "px"))
	effect.Timing.Delay = time.Second
	effect.Timing.Fill = FillNone

	player, err := NewLayout([]Effect{effect})
	if err != nil {
		t.Fatal(err)
	}
	player.Start(start)

	values, live := player.LayoutAt(FrameInput{Now: start.Add(100 * time.Millisecond)}, nil)
	if len(values) != 0 {
		t.Fatalf("values before the delay = %+v, want none", values)
	}
	if !live {
		t.Fatal("an effect waiting on its delay is not live for the next frame")
	}
}

// layoutDecl builds one keyframe rule with a single declaration, which is the
// shape the binder parses.
func layoutDecl(offset float64, property, value string) layout.Keyframe {
	return layout.Keyframe{
		Offset:       offset,
		Declarations: []layout.KeyframeDeclaration{{Property: property, Value: value}},
	}
}

// parseLayoutOne parses one rule with one declaration and returns its value.
func parseLayoutOne(t *testing.T, property, value string) ([]LayoutTerm, error) {
	t.Helper()

	rule := layout.KeyframesRule{
		Name:      "probe",
		Keyframes: []layout.Keyframe{layoutDecl(0, property, value)},
	}
	frames, err := LayoutKeyframes([]layout.KeyframesRule{rule}, "probe")
	if err != nil {
		return nil, err
	}
	if len(frames) != 1 || len(frames[0].Layout) != 1 {
		t.Fatalf("frames = %+v, want one layout value", frames)
	}

	return frames[0].Layout[0].Terms, nil
}

// TestLayoutKeyframesParseTheTier checks the value forms the phase supports:
// absolute lengths resolve to px, percentages keep their unit for the engine's
// containing block, and the keyword and multi-value forms survive.
func TestLayoutKeyframesParseTheTier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		property string
		value    string
		want     []LayoutTerm
	}{
		{"width px", "width", "40px", []LayoutTerm{{Number: 40, Unit: "px"}}},
		{"width inches", "width", "2in", []LayoutTerm{{Number: 192, Unit: "px"}}},
		{"width percent", "width", "50%", []LayoutTerm{{Number: 50, Unit: "%"}}},
		{"width em", "width", "2em", []LayoutTerm{{Number: 2, Unit: "em"}}},
		{"height auto", "height", "auto", []LayoutTerm{{Keyword: "auto"}}},
		{"max width none", "max-width", "none", []LayoutTerm{{Keyword: "none"}}},
		{"margin shorthand", "margin", "1px 2px 3px 4px", []LayoutTerm{
			{Number: 1, Unit: "px"}, {Number: 2, Unit: "px"},
			{Number: 3, Unit: "px"}, {Number: 4, Unit: "px"},
		}},
		{"margin auto", "margin", "auto", []LayoutTerm{{Keyword: "auto"}}},
		{"inset pair", "inset-block", "10% auto", []LayoutTerm{
			{Number: 10, Unit: "%"}, {Keyword: "auto"},
		}},
		{"padding negative not allowed", "padding-top", "4px", []LayoutTerm{{Number: 4, Unit: "px"}}},
		{"gap normal", "gap", "normal", []LayoutTerm{{Keyword: "normal"}}},
		{"row gap length", "row-gap", "8px", []LayoutTerm{{Number: 8, Unit: "px"}}},
		{"flex basis percent", "flex-basis", "25%", []LayoutTerm{{Number: 25, Unit: "%"}}},
		{"flex grow", "flex-grow", "2", []LayoutTerm{{Number: 2}}},
		{"border width keyword", "border-top-width", "thin", []LayoutTerm{{Keyword: "thin"}}},
		{"border width pair", "border-inline-width", "1px 2px", []LayoutTerm{
			{Number: 1, Unit: "px"}, {Number: 2, Unit: "px"},
		}},
		{"radius slash", "border-radius", "10px / 20%", []LayoutTerm{
			{Number: 10, Unit: "px"}, {Keyword: "/"}, {Number: 20, Unit: "%"},
		}},
		{"font size em", "font-size", "1.5em", []LayoutTerm{{Number: 1.5, Unit: "em"}}},
		{"line height number", "line-height", "1.5", []LayoutTerm{{Number: 1.5}}},
		{"line height normal", "line-height", "normal", []LayoutTerm{{Keyword: "normal"}}},
		{"letter spacing negative", "letter-spacing", "-1px", []LayoutTerm{{Number: -1, Unit: "px"}}},
		{"word spacing normal", "word-spacing", "normal", []LayoutTerm{{Keyword: "normal"}}},
		{"logical longhand", "inset-inline-start", "12px", []LayoutTerm{{Number: 12, Unit: "px"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			terms, err := parseLayoutOne(t, test.property, test.value)
			if err != nil {
				t.Fatalf("%s: %v", test.property, err)
			}
			if len(terms) != len(test.want) {
				t.Fatalf("%s = %+v, want %+v", test.property, terms, test.want)
			}
			for i := range terms {
				want := test.want[i]
				if terms[i] != want {
					t.Fatalf("%s term %d = %+v, want %+v", test.property, i, terms[i], want)
				}
			}
		})
	}
}

// TestLayoutKeyframesRejectMalformedValues checks that a value the property
// grammar cannot read fails the parse instead of animating the base value.
func TestLayoutKeyframesRejectMalformedValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		property string
		value    string
	}{
		{"garbage width", "width", "banana"},
		{"unitless nonzero", "width", "10"},
		{"negative width", "width", "-10px"},
		{"two lengths on a longhand", "width", "10px 20px"},
		{"percentage border width", "border-top-width", "10%"},
		{"auto padding", "padding-top", "auto"},
		{"unknown keyword", "width", "wide"},
		{"empty", "width", ""},
		{"unknown unit", "height", "10furlong"},
		{"radius four plus one", "border-radius", "1px 2px 3px 4px 5px"},
		{"negative gap", "gap", "-2px"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := parseLayoutOne(t, test.property, test.value)
			if err == nil {
				t.Fatalf("%s: %q parsed, want an error", test.property, test.value)
			}
			if !errors.Is(err, ErrLayout) {
				t.Fatalf("%s: error = %v, want %v", test.property, err, ErrLayout)
			}
		})
	}
}

// TestDeferredPhaseSkipsTheParsedLayoutTier checks the gap report: the
// properties this phase parses no longer count as deferred, while the tier
// entries it does not parse are still reported against phase 04.
func TestDeferredPhaseSkipsTheParsedLayoutTier(t *testing.T) {
	t.Parallel()

	for _, property := range []string{
		"width", "height", "min-width", "max-height", "inset", "top",
		"margin", "padding", "gap", "flex-basis", "border-width",
		"border-radius", "font-size", "line-height", "letter-spacing",
		"word-spacing",
	} {
		if phase, ok := deferredPhase(property); ok {
			t.Fatalf("%s is still reported as deferred to phase %d", property, phase)
		}
	}
	for _, property := range []string{
		"box-shadow", "text-shadow", "background-position", "border-image-slice",
	} {
		phase, ok := deferredPhase(property)
		if !ok || phase != phaseLayout {
			t.Fatalf("%s = (%d, %v), want the phase 04 gap", property, phase, ok)
		}
	}
}

// TestLayoutKeyframesSkipPaintDeclarations checks that the layout parser only
// keeps its own tier, so a mixed rule does not sample a colour into an
// override.
func TestLayoutKeyframesSkipPaintDeclarations(t *testing.T) {
	t.Parallel()

	rule := layout.KeyframesRule{
		Name: "mixed",
		Keyframes: []layout.Keyframe{{
			Offset: 0.5,
			Declarations: []layout.KeyframeDeclaration{
				{Property: "width", Value: "10px"},
				{Property: "background-color", Value: "red"},
				{Property: "animation-timing-function", Value: "linear"},
			},
		}},
	}

	frames, err := LayoutKeyframes([]layout.KeyframesRule{rule}, "mixed")
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 {
		t.Fatalf("frames = %d, want 1", len(frames))
	}
	if len(frames[0].Layout) != 1 || frames[0].Layout[0].Property != "width" {
		t.Fatalf("layout values = %+v, want only width", frames[0].Layout)
	}
	if frames[0].TimingFunction != "linear" {
		t.Fatalf("timing function = %q, want linear", frames[0].TimingFunction)
	}
}

// TestBindLayoutBuildsOnePlayerPerElement checks that every layout animation on
// an element shares one player, so the element asks the engine for one
// relayout rather than one per property.
func TestBindLayoutBuildsOnePlayerPerElement(t *testing.T) {
	t.Parallel()

	rules := []layout.KeyframesRule{
		{Name: "widen", Keyframes: []layout.Keyframe{
			layoutDecl(0, "width", "40px"),
			layoutDecl(1, "height", "80px"),
		}},
		{Name: "round", Keyframes: []layout.Keyframe{
			layoutDecl(0, "border-radius", "0px"),
			layoutDecl(1, "border-radius", "12px"),
		}},
	}
	display := &layout.Display{Keyframes: rules}
	box := layout.Box{
		ID: "card",
		Animations: []layout.Animation{
			{Name: "widen", DurationMS: 100, IterationCount: 1, FillMode: "both"},
			{Name: "round", DurationMS: 100, IterationCount: 1, FillMode: "both"},
		},
	}

	player, err := BindLayout(display, box, time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if player == nil {
		t.Fatal("BindLayout returned no player for the layout keyframes")
	}
	if got := player.EffectCount(); got != 2 {
		t.Fatalf("effects = %d, want one per animation on the element", got)
	}
	if player.WantsRestamp() {
		t.Fatal("a layout-only player asked for the transform restamp")
	}
}

// TestBindLayoutSkipsPaintOnlyRules checks that a transform-only element binds
// no layout player, which is what keeps a 2D page off the retained-state bill.
func TestBindLayoutSkipsPaintOnlyRules(t *testing.T) {
	t.Parallel()

	display := &layout.Display{Keyframes: []layout.KeyframesRule{{
		Name: "slide",
		Keyframes: []layout.Keyframe{{
			Offset:       0,
			Declarations: []layout.KeyframeDeclaration{{Property: "transform", Value: "translateX(0px)"}},
		}},
	}}}
	box := layout.Box{ID: "chip", Animations: []layout.Animation{{Name: "slide", DurationMS: 100}}}

	player, err := BindLayout(display, box, time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if player != nil {
		t.Fatal("a transform-only rule bound a layout player")
	}
}
