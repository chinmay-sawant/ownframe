package cssanimation

import (
	"errors"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
)

// layoutDecl builds one keyframe rule with a single declaration, which is the
// shape the binder parses.
func layoutDecl(offset float64, property, value string) layout.Keyframe {
	return layout.Keyframe{
		Offset:       offset,
		Declarations: []layout.KeyframeDeclaration{{Property: property, Value: value}},
	}
}

// parseLayoutOne parses one rule with one declaration and returns its value.
func parseLayoutOne(t *testing.T, property, value string) ([]animation.LayoutTerm, error) {
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
		want     []animation.LayoutTerm
	}{
		{"width px", "width", "40px", []animation.LayoutTerm{{Number: 40, Unit: "px"}}},
		{"width inches", "width", "2in", []animation.LayoutTerm{{Number: 192, Unit: "px"}}},
		{"width percent", "width", "50%", []animation.LayoutTerm{{Number: 50, Unit: "%"}}},
		{"width em", "width", "2em", []animation.LayoutTerm{{Number: 2, Unit: "em"}}},
		{"height auto", "height", "auto", []animation.LayoutTerm{{Keyword: "auto"}}},
		{"max width none", "max-width", "none", []animation.LayoutTerm{{Keyword: "none"}}},
		{"margin shorthand", "margin", "1px 2px 3px 4px", []animation.LayoutTerm{
			{Number: 1, Unit: "px"}, {Number: 2, Unit: "px"},
			{Number: 3, Unit: "px"}, {Number: 4, Unit: "px"},
		}},
		{"margin auto", "margin", "auto", []animation.LayoutTerm{{Keyword: "auto"}}},
		{"inset pair", "inset-block", "10% auto", []animation.LayoutTerm{
			{Number: 10, Unit: "%"}, {Keyword: "auto"},
		}},
		{"padding negative not allowed", "padding-top", "4px", []animation.LayoutTerm{{Number: 4, Unit: "px"}}},
		{"gap normal", "gap", "normal", []animation.LayoutTerm{{Keyword: "normal"}}},
		{"row gap length", "row-gap", "8px", []animation.LayoutTerm{{Number: 8, Unit: "px"}}},
		{"flex basis percent", "flex-basis", "25%", []animation.LayoutTerm{{Number: 25, Unit: "%"}}},
		{"flex grow", "flex-grow", "2", []animation.LayoutTerm{{Number: 2}}},
		{"border width keyword", "border-top-width", "thin", []animation.LayoutTerm{{Keyword: "thin"}}},
		{"border width pair", "border-inline-width", "1px 2px", []animation.LayoutTerm{
			{Number: 1, Unit: "px"}, {Number: 2, Unit: "px"},
		}},
		{"radius slash", "border-radius", "10px / 20%", []animation.LayoutTerm{
			{Number: 10, Unit: "px"}, {Keyword: "/"}, {Number: 20, Unit: "%"},
		}},
		{"font size em", "font-size", "1.5em", []animation.LayoutTerm{{Number: 1.5, Unit: "em"}}},
		{"line height number", "line-height", "1.5", []animation.LayoutTerm{{Number: 1.5}}},
		{"line height normal", "line-height", "normal", []animation.LayoutTerm{{Keyword: "normal"}}},
		{"letter spacing negative", "letter-spacing", "-1px", []animation.LayoutTerm{{Number: -1, Unit: "px"}}},
		{"word spacing normal", "word-spacing", "normal", []animation.LayoutTerm{{Keyword: "normal"}}},
		{"logical longhand", "inset-inline-start", "12px", []animation.LayoutTerm{{Number: 12, Unit: "px"}}},
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
