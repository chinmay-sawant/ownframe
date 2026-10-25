package animation

import (
	"strings"
	"testing"
	"time"
)

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
