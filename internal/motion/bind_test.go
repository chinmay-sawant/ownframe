package motion

import (
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"math"
	"slices"
	"strings"
	"testing"
	"time"
)

// TestBindBoxWithIndexMatchesDirect pins the shared-index contract at the
// binder: resolving through one per-display index returns the same players,
// target sets, masks, and gaps as the direct scan, across owner, subtree,
// anonymous, svg, and ownerless displays.
func TestBindBoxWithIndexMatchesDirect(t *testing.T) {
	shift := []layout.KeyframesRule{{Name: "shift", Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: "transform", Value: "translate(0px)"}}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: "transform", Value: "translate(20px)"}}},
	}}}
	start := time.Unix(100, 0)

	animations := []layout.Animation{
		{Name: "shift", DurationMS: 1000, IterationCount: 1, FillMode: "both", PlayState: "running"},
		{Name: "tint", DurationMS: 1000, IterationCount: 1, FillMode: "both", PlayState: "running"},
	}
	newDisplay := func(owners []int) *layout.Display {
		d := indexCardDisplay(animations...)
		d.Keyframes = append(append([]layout.KeyframesRule{}, shift...),
			indexColorRule("tint", "background-color", "#ff0000", "#00ff00")...)
		d.OpOwners = owners
		d.Boxes[1].Animations = animations[:1]
		return d
	}

	displays := map[string]*layout.Display{
		"owned":     newDisplay([]int{0, 0, 1}),
		"ownerless": newDisplay(nil),
	}
	for name, d := range displays {
		idx := NewIndex(d)
		for boxIndex := range d.Boxes {
			if len(d.Boxes[boxIndex].Animations) == 0 {
				continue
			}
			want, wantGaps, err := BindBox(d, d.Boxes[boxIndex], boxIndex, start)
			if err != nil {
				t.Fatalf("%s box %d direct bind: %v", name, boxIndex, err)
			}
			got, gotGaps, err := BindBoxWithIndex(d, d.Boxes[boxIndex], boxIndex, start, idx)
			if err != nil {
				t.Fatalf("%s box %d indexed bind: %v", name, boxIndex, err)
			}
			compareBounds(t, name, boxIndex, want, got)
			if !slices.Equal(gotGaps, wantGaps) {
				t.Fatalf("%s box %d gaps = %v, want %v", name, boxIndex, gotGaps, wantGaps)
			}
		}
		anonymous := d.Boxes[0]
		anonymous.ID = ""
		want, wantGaps, err := BindBox(d, anonymous, -1, start)
		if err != nil {
			t.Fatalf("%s anonymous direct bind: %v", name, err)
		}
		got, gotGaps, err := BindBoxWithIndex(d, anonymous, -1, start, idx)
		if err != nil {
			t.Fatalf("%s anonymous indexed bind: %v", name, err)
		}
		compareBounds(t, name, -1, want, got)
		if !slices.Equal(gotGaps, wantGaps) {
			t.Fatalf("%s anonymous gaps = %v, want %v", name, gotGaps, wantGaps)
		}
	}
}

// TestBindBoxWithIndexMatchesDirectSVG covers the raster tier: an SVG paint
// player must resolve its image operation identically through the index.
func TestBindBoxWithIndexMatchesDirectSVG(t *testing.T) {
	d := svgBindDisplay(svgRule("paint", "fill", "#000000", "#ffffff"), "paint")
	idx := NewIndex(d)
	want, wantGaps, err := BindBox(d, d.Boxes[0], 0, svgStart())
	if err != nil {
		t.Fatalf("direct bind: %v", err)
	}
	got, gotGaps, err := BindBoxWithIndex(d, d.Boxes[0], 0, svgStart(), idx)
	if err != nil {
		t.Fatalf("indexed bind: %v", err)
	}
	compareBounds(t, "svg", 0, want, got)
	if !slices.Equal(gotGaps, wantGaps) {
		t.Fatalf("svg gaps = %v, want %v", gotGaps, wantGaps)
	}
}

// indexCardDisplay is one parent card (fill plus label) with a child card, for
// the shared-index equivalence test. Op ids are stable so both resolution
// paths see the same ownership.
func indexCardDisplay(animations ...layout.Animation) *layout.Display {
	return &layout.Display{
		Width: 200, Height: 120, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{
			{ID: 11, Kind: layout.DisplayOpFillRect,
				X: 10, Y: 10, W: 100, H: 60, R: .1, G: .2, B: .3, Alpha: 1},
			{ID: 12, Kind: layout.DisplayOpText,
				X: 20, Y: 30, W: 40, H: 12, R: 1, G: 1, B: 1, Alpha: 1},
			{ID: 13, Kind: layout.DisplayOpFillRect,
				X: 30, Y: 40, W: 40, H: 20, R: .4, G: .5, B: .6, Alpha: 1},
		},
		Boxes: []layout.Box{
			{ID: "parent", X: 10, Y: 10, W: 100, H: 60,
				OperationIDs: []uint64{11, 12, 13},
				Animations:   animations, ParentIndex: -1},
			{ID: "child", X: 30, Y: 40, W: 40, H: 20,
				OperationIDs: []uint64{13}, ParentIndex: 0},
		},
	}
}

// indexColorRule builds one linear two-stop animation rule.
func indexColorRule(name, property, from, to string) []layout.KeyframesRule {
	return []layout.KeyframesRule{{Name: name, Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: property, Value: from}}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: property, Value: to}}},
	}}}
}

func compareBounds(t *testing.T, name string, boxIndex int, want, got []Bound) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s box %d players = %d, want %d", name, boxIndex, len(got), len(want))
	}
	for i := range want {
		if !slices.Equal(got[i].Indices, want[i].Indices) {
			t.Fatalf("%s box %d player %d targets = %v, want %v",
				name, boxIndex, i, got[i].Indices, want[i].Indices)
		}
		if got[i].Mask != want[i].Mask {
			t.Fatalf("%s box %d player %d mask = %d, want %d",
				name, boxIndex, i, got[i].Mask, want[i].Mask)
		}
	}
}

// keyframeTimingSource is a three-stop rule whose timing function is declared
// beside a paint property on some stops and alone on another.
func keyframeTimingSource() []layout.Keyframe {
	return []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{
			{Property: "opacity", Value: "1"},
			{Property: "animation-timing-function", Value: " ease-out "},
		}},
		{Offset: 0.5, Declarations: []layout.KeyframeDeclaration{
			{Property: "animation-timing-function", Value: "steps(2, jump-none)"},
		}},
		{Offset: 0.75, Declarations: []layout.KeyframeDeclaration{
			{Property: "-webkit-animation-timing-function", Value: "ease-in"},
		}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{
			{Property: "opacity", Value: "0"},
		}},
	}
}

// keyframeTimingWants is the function each stop of keyframeTimingSource
// declares, in order.
func keyframeTimingWants() []string {
	return []string{"ease-out", "steps(2, jump-none)", "ease-in", ""}
}

// TestKeyframeTimingFunctionSurvivesClassParsing pins that the timing function
// is read before the paint-class filter: it belongs to the keyframe, not to one
// property family, so every class the rule produces carries it.
func TestKeyframeTimingFunctionSurvivesClassParsing(t *testing.T) {
	size := transformSize{width: 200, height: 100, pointsPerPixel: 1, hasBox: true}
	class := paintClass{kind: KindSubtree, mask: MaskOpacity}

	frames := keyframesFor(keyframeTimingSource(), size, class, Color{}, false)
	wants := keyframeTimingWants()
	if len(frames) != len(wants) {
		t.Fatalf("frames = %d, want %d", len(frames), len(wants))
	}
	for i, want := range wants {
		if frames[i].TimingFunction != want {
			t.Fatalf("frame %d timing function = %q, want %q", i, frames[i].TimingFunction, want)
		}
	}
	if !frames[0].OpacitySet || !frames[3].OpacitySet {
		t.Fatal("the paint declarations were lost while keeping the timing function")
	}
	for i := 1; i < 3; i++ {
		if frames[i].OpacitySet {
			t.Fatalf("timing-function-only frame %d invented an opacity", i)
		}
	}
}

// TestKeyframeTimingFunctionIsNotAPaintGap pins that the declaration is not
// mistaken for an unsupported paint property or a deferred phase.
func TestKeyframeTimingFunctionIsNotAPaintGap(t *testing.T) {
	if _, ok := propertyClass("animation-timing-function"); ok {
		t.Fatal("the timing function was classified as a paint property")
	}
	if phase, ok := deferredPhase("animation-timing-function"); ok {
		t.Fatalf("the timing function was deferred to phase %d", phase)
	}
	if !keyframeTimingFunction(layout.KeyframeDeclaration{Property: "animation-timing-function"}) {
		t.Fatal("the keyframe parser did not recognise the timing function")
	}
}

// TestKeyframeTimingFunctionReachesTheEffectTiming checks the end-to-end path a
// stylesheet takes: the rule, the animation record, and the bound effect all
// agree on the segment easing.
func TestKeyframeTimingFunctionReachesTheEffectTiming(t *testing.T) {
	rules := []layout.KeyframesRule{{Name: "fade", Keyframes: keyframeTimingSource()}}
	frames := keyframes(rules, "fade", 1, 200, 100)
	wants := keyframeTimingWants()
	if len(frames) != len(wants) {
		t.Fatalf("frames = %d, want %d", len(frames), len(wants))
	}
	for i, want := range wants {
		if frames[i].TimingFunction != want {
			t.Fatalf("frame %d timing function = %q, want %q", i, frames[i].TimingFunction, want)
		}
	}
}

// tolerance is the room a colour conversion and the 8-bit source share.
const tolerance = 1e-6

func closeColor(got, want Color) bool {
	return math.Abs(got.R-want.R) < tolerance && math.Abs(got.G-want.G) < tolerance &&
		math.Abs(got.B-want.B) < tolerance && math.Abs(got.A-want.A) < tolerance &&
		got.Current == want.Current
}

func TestParseColorForms(t *testing.T) {
	tests := []struct {
		value string
		want  Color
	}{
		{"#f00", Color{R: 1, A: 1}},
		{"#0f08", Color{G: 1, A: 0x88 / 255.0}},
		{"#ff0000", Color{R: 1, A: 1}},
		{"#00ff0080", Color{G: 1, A: 0x80 / 255.0}},
		{"rgb(255, 0, 0)", Color{R: 1, A: 1}},
		{"rgba(255, 0, 0, 0.5)", Color{R: 1, A: .5}},
		{"rgb(100% 0% 0% / 50%)", Color{R: 1, A: .5}},
		{"hsl(0, 100%, 50%)", Color{R: 1, A: 1}},
		{"hsl(120deg 100% 50%)", Color{G: 1, A: 1}},
		{"hsla(240, 100%, 50%, .25)", Color{B: 1, A: .25}},
		{"hsl(0.5turn, 100%, 50%)", Color{G: 1, B: 1, A: 1}},
		{"red", Color{R: 1, A: 1}},
		{"rebeccapurple", Color{R: 0x66 / 255.0, G: 0x33 / 255.0, B: 0x99 / 255.0, A: 1}},
		{"CURRENTCOLOR", Color{Current: true, A: 1}},
		{"transparent", Color{A: 0}},
	}
	for _, test := range tests {
		got, err := parseColor(test.value)
		if err != nil {
			t.Errorf("parseColor(%q) failed: %v", test.value, err)
			continue
		}
		if !closeColor(got, test.want) {
			t.Errorf("parseColor(%q) = %+v, want %+v", test.value, got, test.want)
		}
	}
}

func TestParseColorRejectsMalformedValues(t *testing.T) {
	for _, value := range []string{
		"", "#", "#12", "#12345", "#ggg", "rgb(1, 2)", "rgb(1 2 3 4)",
		"rgb(1, 2, 3, 4, 5)", "hsl(120, 50, 50%)", "hsl(120, 50%, 50)", "nonsense",
		"inherit", "initial", "unset",
	} {
		if _, err := parseColor(value); !errors.Is(err, ErrColor) {
			t.Errorf("parseColor(%q) error = %v, want ErrColor", value, err)
		}
	}
}

// TestParseColorClampsChannels pins the CSS rule that a numeric channel outside
// 0..255 and an alpha outside 0..1 clamp instead of failing.
func TestParseColorClampsChannels(t *testing.T) {
	got, err := parseColor("rgb(300, -20, 0, 1.5)")
	if err != nil {
		t.Fatalf("parseColor failed: %v", err)
	}
	want := Color{R: 1, A: 1}
	if !closeColor(got, want) {
		t.Fatalf("clamped colour = %+v, want %+v", got, want)
	}
}

// TestParseColorDefersAnUnsupportedSpace pins that a modern colour space is
// reported as deferred rather than malformed, so the binder can leave the
// declaration at base paint.
func TestParseColorDefersAnUnsupportedSpace(t *testing.T) {
	for _, value := range []string{"oklab(0.5 0 0)", "color(display-p3 1 0 0)", "lab(50% 0 0)"} {
		if _, err := parseColor(value); !errors.Is(err, ErrColorSpace) {
			t.Errorf("parseColor(%q) error = %v, want ErrColorSpace", value, err)
		}
	}
}

const radTolerance = 1e-9

func testTransformSize() transformSize {
	return transformSize{width: 200, height: 100, pointsPerPixel: 1, hasBox: true}
}

func parseTestValue(t *testing.T, value string, size transformSize) TransformValue {
	t.Helper()
	parsed, err := parseTransform(value, size)
	if err != nil {
		t.Fatalf("parseTransform(%q) = %v", value, err)
	}
	return parsed
}

func angleNear(a, b float64) bool {
	return math.Abs(math.Remainder(a-b, 2*math.Pi)) < radTolerance
}

func checkTransform(t *testing.T, want, got Transform) {
	t.Helper()
	if math.Abs(want.Tx-got.Tx) > radTolerance || math.Abs(want.Ty-got.Ty) > radTolerance {
		t.Fatalf("translate = %v,%v, want %v,%v", got.Tx, got.Ty, want.Tx, want.Ty)
	}
	if math.Abs(want.Sx-got.Sx) > radTolerance || math.Abs(want.Sy-got.Sy) > radTolerance {
		t.Fatalf("scale = %v,%v, want %v,%v", got.Sx, got.Sy, want.Sx, want.Sy)
	}
	if !angleNear(want.Rotate, got.Rotate) {
		t.Fatalf("rotate = %v, want %v", got.Rotate, want.Rotate)
	}
	if math.Abs(want.SkewX-got.SkewX) > radTolerance {
		t.Fatalf("skewX = %v, want %v", got.SkewX, want.SkewX)
	}
}

func TestParseTransformTwoDimensionalFunctions(t *testing.T) {
	size := testTransformSize()
	cases := []struct {
		value string
		want  Transform
	}{
		{"none", IdentityTransform()},
		{"translate(10px)", Transform{Tx: 10, Sx: 1, Sy: 1}},
		{"translate(10px, 20px)", Transform{Tx: 10, Ty: 20, Sx: 1, Sy: 1}},
		{"translate(10px, 0px)", Transform{Tx: 10, Sx: 1, Sy: 1}},
		{"translateX(30px)", Transform{Tx: 30, Sx: 1, Sy: 1}},
		{"translateY(40px)", Transform{Ty: 40, Sx: 1, Sy: 1}},
		{"scale(2)", Transform{Sx: 2, Sy: 2}},
		{"scale(2, 3)", Transform{Sx: 2, Sy: 3}},
		{"scaleX(2)", Transform{Sx: 2, Sy: 1}},
		{"scaleY(3)", Transform{Sx: 1, Sy: 3}},
		{"rotate(90deg)", Transform{Sx: 1, Sy: 1, Rotate: math.Pi / 2}},
		{"rotateZ(0.5turn)", Transform{Sx: 1, Sy: 1, Rotate: math.Pi}},
		{"skewX(30deg)", Transform{Sx: 1, Sy: 1, SkewX: math.Pi / 6}},
		{"matrix(2, 0, 0, 3, 10, 20)", Transform{Tx: 10, Ty: 20, Sx: 2, Sy: 3}},
		{"translate(10px, 20px) rotate(0deg) scale(2, 3)",
			Transform{Tx: 10, Ty: 20, Sx: 2, Sy: 3}},
		{"translate3d(10px, 20px, 0px)", Transform{Tx: 10, Ty: 20, Sx: 1, Sy: 1}},
		{"translate3d(10px, 20px, 0)", Transform{Tx: 10, Ty: 20, Sx: 1, Sy: 1}},
		{"scale3d(2, 3, 1)", Transform{Sx: 2, Sy: 3}},
		{"translateZ(0px)", Transform{Sx: 1, Sy: 1}},
		{"scaleZ(1)", Transform{Sx: 1, Sy: 1}},
	}
	for _, c := range cases {
		parsed := parseTestValue(t, c.value, size)
		if parsed.ThreeD {
			t.Errorf("%q marked itself 3D", c.value)
		}
		checkTransform(t, c.want, parsed.Decomposed)
	}
}

func TestParseTransformAcceptsEveryFunction(t *testing.T) {
	size := testTransformSize()
	cases := []struct {
		value string
		name  TransformFunc
	}{
		{"translate(1px, 2px)", Translate},
		{"translateX(1px)", TranslateX},
		{"translateY(1px)", TranslateY},
		{"translate3d(1px, 2px, 3px)", Translate3D},
		{"translateZ(1px)", TranslateZ},
		{"scale(2, 3)", Scale},
		{"scaleX(2)", ScaleX},
		{"scaleY(2)", ScaleY},
		{"scale3d(1, 2, 3)", Scale3D},
		{"scaleZ(2)", ScaleZ},
		{"rotate(10deg)", Rotate},
		{"rotateX(10deg)", RotateX},
		{"rotateY(10deg)", RotateY},
		{"rotateZ(10deg)", RotateZ},
		{"rotate3d(0, 0, 1, 10deg)", Rotate3D},
		{"skew(10deg, 20deg)", Skew},
		{"skewX(10deg)", SkewX},
		{"skewY(10deg)", SkewY},
		{"matrix(1, 0, 0, 1, 0, 0)", Matrix},
		{"matrix3d(1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1)", Matrix3D},
		{"perspective(100px)", Perspective},
	}
	for _, c := range cases {
		parsed := parseTestValue(t, c.value, size)
		if len(parsed.Functions) != 1 || parsed.Functions[0].Name != c.name {
			t.Errorf("%q functions = %+v, want %s", c.value, parsed.Functions, c.name)
		}
	}
}

func TestParseTransformThreeDimensionalFunctions(t *testing.T) {
	size := testTransformSize()
	cases := []string{
		"translate3d(10px, 20px, 30px)",
		"translateZ(5px)",
		"scale3d(2, 3, 4)",
		"scaleZ(2)",
		"rotateX(45deg)",
		"rotateY(1rad)",
		"rotate3d(0, 0, 1, 90deg)",
		"matrix3d(1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 10, 20, 30, 1)",
		"perspective(100px)",
	}
	for _, value := range cases {
		parsed := parseTestValue(t, value, size)
		if !parsed.ThreeD {
			t.Errorf("%q was not marked 3D", value)
		}
		if len(parsed.Functions) != 1 {
			t.Errorf("%q has %d functions, want 1", value, len(parsed.Functions))
		}
		if !parsed.Decomposed.Set || parsed.Decomposed.Sx != 1 || parsed.Decomposed.Sy != 1 {
			t.Errorf("%q decomposed = %+v, want identity scale", value, parsed.Decomposed)
		}
	}
	translate := parseTestValue(t, "translate3d(10px, 20px, 30px)", size)
	if translate.Decomposed.Tx != 10 || translate.Decomposed.Ty != 20 {
		t.Errorf("translate3d X/Y = %v,%v, want 10,20", translate.Decomposed.Tx, translate.Decomposed.Ty)
	}
	matrix := parseTestValue(t, "matrix3d(1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 4, 5, 6, 1)", size)
	if matrix.Decomposed.Tx != 4 || matrix.Decomposed.Ty != 5 {
		t.Errorf("matrix3d X/Y = %v,%v, want 4,5", matrix.Decomposed.Tx, matrix.Decomposed.Ty)
	}
}

func TestParseTransformUnits(t *testing.T) {
	size := testTransformSize()
	cases := []struct {
		value string
		want  float64
	}{
		{"translate(50%, 25%)", 100},
		{"translate(-50%, 0%)", -100},
		{"rotate(180deg)", math.Pi},
		{"rotate(200grad)", math.Pi},
		{"rotate(3.141592653589793rad)", math.Pi},
		{"rotate(0.5turn)", math.Pi},
		{"rotate(0)", 0},
		{"translate(0)", 0},
	}
	for _, c := range cases {
		parsed := parseTestValue(t, c.value, size)
		if c.value[0] == 'r' {
			if !angleNear(c.want, parsed.Decomposed.Rotate) {
				t.Errorf("%q = %v, want %v", c.value, parsed.Decomposed.Rotate, c.want)
			}
			continue
		}
		if math.Abs(parsed.Decomposed.Tx-c.want) > radTolerance {
			t.Errorf("%q = %v, want %v", c.value, parsed.Decomposed.Tx, c.want)
		}
	}
	vertical := parseTestValue(t, "translate(0%, 25%)", size)
	if vertical.Decomposed.Ty != 25 {
		t.Errorf("translate y%% = %v, want 25", vertical.Decomposed.Ty)
	}
}

func TestParseTransformScalesPointsPerPixel(t *testing.T) {
	size := testTransformSize()
	size.pointsPerPixel = 0.5
	parsed := parseTestValue(t, "translate(10px, 5%)", size)
	if parsed.Decomposed.Tx != 5 || parsed.Decomposed.Ty != 2.5 {
		t.Errorf("scaled X/Y = %v,%v, want 5,2.5", parsed.Decomposed.Tx, parsed.Decomposed.Ty)
	}
	matrix := parseTestValue(t, "matrix(1, 0, 0, 1, 10, 20)", size)
	if matrix.Decomposed.Tx != 5 || matrix.Decomposed.Ty != 10 {
		t.Errorf("scaled matrix X/Y = %v,%v, want 5,10", matrix.Decomposed.Tx, matrix.Decomposed.Ty)
	}
}

func TestParseTransformRejectsMalformedLists(t *testing.T) {
	size := testTransformSize()
	for _, value := range []string{
		"",
		"translate(10px",
		"translate(10px) rotate(20deg",
		"bogus(1)",
		"scale()",
		"matrix(1, 2, 3)",
		"matrix3d(1, 2, 3)",
		"rotate(10px)",
		"rotate(10)",
		"translate(5)",
		"translate(2rem)",
		"translate(10px), rotate(1deg)",
		"translate3d(10px, 20px, 30%)",
		"perspective(-10px)",
		"perspective(10%)",
	} {
		if _, err := parseTransform(value, size); !errors.Is(err, ErrTransform) {
			t.Errorf("parseTransform(%q) error = %v, want ErrTransform", value, err)
		}
	}
}

func TestParseTransformPercentageWithoutBoxIsAnError(t *testing.T) {
	size := testTransformSize()
	size.hasBox = false
	_, err := parseTransform("translate(50%, 25%)", size)
	if !errors.Is(err, ErrTransform) {
		t.Fatalf("percentage without a box = %v, want ErrTransform", err)
	}
	if _, err := parseTransform("translate(10px, 10px)", size); err != nil {
		t.Fatalf("absolute lengths without a box = %v", err)
	}
}

func TestKeyframesResolvePercentageAgainstTheBox(t *testing.T) {
	rules := []layout.KeyframesRule{{Name: "slide", Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: "transform", Value: "translate(0%, 0%)"}}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: "transform", Value: "translate(50%, 25%)"}}},
	}}}
	frames := keyframes(rules, "slide", 1, 200, 100)
	if len(frames) != 2 {
		t.Fatalf("frames = %d, want 2", len(frames))
	}
	if frames[1].X != 100 || frames[1].Y != 25 {
		t.Fatalf("percentage frame X/Y = %v,%v, want 100,25", frames[1].X, frames[1].Y)
	}
	if frames[1].X != frames[1].Transform.Decomposed.Tx {
		t.Fatal("X is not kept readable from the decomposed transform")
	}
}

func TestKeyframesKeepTheParseError(t *testing.T) {
	rules := []layout.KeyframesRule{{Name: "broken", Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: "transform", Value: "rotate(10px)"}}},
	}}}
	frames := keyframes(rules, "broken", 1, 100, 100)
	if len(frames) != 1 || !errors.Is(frames[0].Err, ErrTransform) {
		t.Fatalf("frame error = %v, want ErrTransform", frames[0].Err)
	}
}

// perspectiveTestSize is a 200x100 CSS-pixel box at the display's 0.75
// points-per-pixel.
func perspectiveTestSize() transformSize {
	return transformSize{width: 200, height: 100, pointsPerPixel: 0.75, hasBox: true}
}

func nearPerspective(a, b float64) bool { return math.Abs(a-b) <= 1e-9 }

func TestParsePerspectiveLengths(t *testing.T) {
	size := perspectiveTestSize()

	for _, test := range []struct {
		name  string
		value string
		want  float64
		fails bool
	}{
		{name: "none", value: "none", want: 0},
		{name: "NONE", value: "NONE", want: 0},
		{name: "zero", value: "0", want: 0},
		{name: "px", value: "500px", want: 375},
		{name: "fractional px", value: "2.5px", want: 1.875},
		{name: "negative", value: "-10px", fails: true},
		{name: "percentage", value: "10%", fails: true},
		{name: "unitless", value: "10", fails: true},
		{name: "garbage", value: "wide", fails: true},
		{name: "empty", value: "", fails: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := parsePerspective(test.value, size)
			if test.fails {
				if err == nil {
					t.Fatalf("parsePerspective(%q) = %v, want an error", test.value, got)
				}

				return
			}
			if err != nil {
				t.Fatalf("parsePerspective(%q): %v", test.value, err)
			}
			if !nearPerspective(got, test.want) {
				t.Fatalf("parsePerspective(%q) = %v, want %v", test.value, got, test.want)
			}
		})
	}
}

func TestParsePerspectiveOriginPositions(t *testing.T) {
	size := perspectiveTestSize()

	for _, test := range []struct {
		name  string
		value string
		x, y  float64
		fails bool
	}{
		{name: "percent pair", value: "50% 50%", x: 75, y: 37.5},
		{name: "quarter", value: "25% 75%", x: 37.5, y: 56.25},
		{name: "keywords", value: "top left", x: 0, y: 0},
		{name: "keywords reversed", value: "left top", x: 0, y: 0},
		{name: "bottom right", value: "right bottom", x: 150, y: 75},
		{name: "center top", value: "center top", x: 75, y: 0},
		{name: "top center", value: "top center", x: 75, y: 0},
		{name: "px pair", value: "20px 10px", x: 15, y: 7.5},
		{name: "unitless zeros", value: "0 0", x: 0, y: 0},
		{name: "single left", value: "left", x: 0, y: 37.5},
		{name: "single bottom", value: "bottom", x: 75, y: 75},
		{name: "single center", value: "center", x: 75, y: 37.5},
		{name: "single percent", value: "10%", x: 15, y: 37.5},
		{name: "length and keyword", value: "10px bottom", x: 7.5, y: 75},
		{name: "keyword and length", value: "right 10px", x: 150, y: 7.5},
		{name: "two horizontal keywords", value: "left right", fails: true},
		{name: "two vertical keywords", value: "top bottom", fails: true},
		{name: "three values", value: "left 10px top", x: 7.5, y: 0},
		{name: "empty", value: "", fails: true},
		{name: "unitless", value: "10 20", fails: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			x, y, err := parsePerspectiveOrigin(test.value, size)
			if test.fails {
				if err == nil {
					t.Fatalf("parsePerspectiveOrigin(%q) = %v,%v, want an error", test.value, x, y)
				}

				return
			}
			if err != nil {
				t.Fatalf("parsePerspectiveOrigin(%q): %v", test.value, err)
			}
			if !nearPerspective(x, test.x) || !nearPerspective(y, test.y) {
				t.Fatalf("parsePerspectiveOrigin(%q) = %v,%v, want %v,%v", test.value, x, y, test.x, test.y)
			}
		})
	}
}

func TestKeyframesReadPerspectiveDeclarations(t *testing.T) {
	rules := []layout.KeyframesRule{{
		Name: "camera",
		Keyframes: []layout.Keyframe{
			{Offset: 0, Declarations: []layout.KeyframeDeclaration{
				{Property: "perspective", Value: "none"},
			}},
			{Offset: 0.5, Declarations: []layout.KeyframeDeclaration{
				{Property: "perspective", Value: "500px"},
				{Property: "perspective-origin", Value: "top left"},
			}},
			{Offset: 1, Declarations: []layout.KeyframeDeclaration{
				{Property: "perspective", Value: "1000px"},
			}},
		},
	}}

	frames := keyframes(rules, "camera", 0.75, 200, 100)
	if len(frames) != 3 {
		t.Fatalf("frames = %d, want 3", len(frames))
	}

	first := frames[0].Perspective
	if !first.HasDistance || first.Distance != 0 {
		t.Fatalf("first perspective = %+v, want a declared none", first)
	}
	if first.HasOrigin {
		t.Fatalf("first perspective origin = %+v, want undeclared", first)
	}
	if !nearPerspective(first.OriginX, 75) || !nearPerspective(first.OriginY, 37.5) {
		t.Fatalf("first default origin = %v,%v, want 75,37.5", first.OriginX, first.OriginY)
	}

	middle := frames[1].Perspective
	if !middle.HasDistance || !nearPerspective(middle.Distance, 375) {
		t.Fatalf("middle perspective = %+v, want 500px as 375 points", middle)
	}
	if !middle.HasOrigin || middle.OriginX != 0 || middle.OriginY != 0 {
		t.Fatalf("middle origin = %+v, want the declared top left", middle)
	}

	last := frames[2].Perspective
	if !last.HasDistance || !nearPerspective(last.Distance, 750) || last.HasOrigin {
		t.Fatalf("last perspective = %+v, want 1000px and the undeclared default origin", last)
	}
}

func TestKeyframesReportInvalidPerspective(t *testing.T) {
	for _, declaration := range []layout.KeyframeDeclaration{
		{Property: "perspective", Value: "-10px"},
		{Property: "perspective-origin", Value: "left right"},
	} {
		rules := []layout.KeyframesRule{{Name: "bad", Keyframes: []layout.Keyframe{{
			Offset:       0,
			Declarations: []layout.KeyframeDeclaration{declaration},
		}}}}
		frames := keyframes(rules, "bad", 0.75, 200, 100)
		if frames[0].Err == nil {
			t.Fatalf("%s: %q parsed without an error", declaration.Property, declaration.Value)
		}
		if !strings.Contains(frames[0].Err.Error(), "perspective") {
			t.Fatalf("%s error = %v, want the perspective context", declaration.Property, frames[0].Err)
		}
	}
}

// staticCamera is the perspective metadata a box with `perspective: 800px;
// perspective-origin: 10px 20px` carries. The test box is 200x100 CSS pixels at
// one point per pixel, so the resolved origin stays 10,20 points.
func staticCamera() layout.TransformInfo {
	return layout.TransformInfo{
		Parent:             layout.IdentityMatrix(),
		World:              layout.IdentityMatrix(),
		HasPerspective:     true,
		Perspective:        800,
		PerspectiveOriginX: 10,
		PerspectiveOriginY: 20,
	}
}

// cameraRule builds one running animation whose keyframes are the caller's.
func cameraRule(frames ...layout.Keyframe) []layout.KeyframesRule {
	return []layout.KeyframesRule{{Name: "camera", Keyframes: frames}}
}

// cameraDisplay builds one 200x100 target box with one fill and one animation.
func cameraDisplay(transform layout.TransformInfo, rules []layout.KeyframesRule) *layout.Display {
	return &layout.Display{
		Width: 200, Height: 100, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect,
			X: 0, Y: 0, W: 200, H: 100, R: .1, G: .2, B: .3, Alpha: 1}},
		OpOwners: []int{0},
		Boxes: []layout.Box{{
			ID: "target", X: 0, Y: 0, W: 200, H: 100, ParentIndex: -1,
			OperationIDs: []uint64{1},
			Animations: []layout.Animation{{
				Name: "camera", DurationMS: 1000, IterationCount: 1,
				FillMode: "both", PlayState: "running",
			}},
			Transform: transform,
		}},
		Keyframes: rules,
	}
}

// sampleCamera binds the display and samples one 1 s animation at fraction.
func sampleCamera(t *testing.T, d *layout.Display, fraction float64) []RestampValue {
	t.Helper()

	start := time.Unix(200, 0)
	bounds, _, err := BindBox(d, d.Boxes[0], 0, start)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if len(bounds) != 1 {
		t.Fatalf("players = %d, want one", len(bounds))
	}
	bounds[0].Player.Step(start.Add(time.Duration(fraction * float64(time.Second))))

	return bounds[0].Player.RestampValues()
}

// oneCameraSample binds and samples once, requiring one restamp request.
func oneCameraSample(t *testing.T, d *layout.Display, fraction float64) RestampValue {
	t.Helper()

	values := sampleCamera(t, d, fraction)
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want one", len(values))
	}

	return values[0]
}

func TestParsePerspectiveOriginEdgeOffsets(t *testing.T) {
	size := perspectiveTestSize()

	for _, test := range []struct {
		name  string
		value string
		x, y  float64
	}{
		{name: "four values", value: "left 20px top 10px", x: 15, y: 7.5},
		{name: "four values from the far edges", value: "right 10% bottom 20px", x: 135, y: 60},
		{name: "four values vertical first", value: "bottom 20px right 10%", x: 135, y: 60},
		{name: "three values horizontal edge", value: "left 20px top", x: 15, y: 0},
		{name: "three values vertical edge", value: "right 10% bottom", x: 135, y: 75},
		{name: "three values vertical first", value: "top 10px left 20px", x: 15, y: 7.5},
		{name: "zero offsets", value: "left 0 top 0", x: 0, y: 0},
		{name: "percentage edge offset", value: "right 25% top", x: 112.5, y: 0},
		{name: "negative offset", value: "left -10px bottom 5px", x: -7.5, y: 71.25},
	} {
		t.Run(test.name, func(t *testing.T) {
			x, y, err := parsePerspectiveOrigin(test.value, size)
			if err != nil {
				t.Fatalf("parsePerspectiveOrigin(%q): %v", test.value, err)
			}
			if !near(x, test.x) || !near(y, test.y) {
				t.Fatalf("parsePerspectiveOrigin(%q) = %v,%v, want %v,%v",
					test.value, x, y, test.x, test.y)
			}
		})
	}
}

func TestParsePerspectiveOriginRejectsMalformedEdgeOffsets(t *testing.T) {
	size := perspectiveTestSize()

	for _, value := range []string{
		"left top 10px",        // offset must follow its edge
		"left 10px center",     // center takes no offset
		"center 10px top",      // center takes no offset
		"left 10px left 20px",  // both edges name the horizontal axis
		"left 10px 20px",       // missing the opposite edge
		"10px 20px 30px",       // a length is not an edge
		"left 10px top auto",   // auto is not an origin offset
		"left 10px top 20",     // unitless lengths are not CSS lengths
		"left 10px top 20px x", // too many components
	} {
		t.Run(value, func(t *testing.T) {
			if x, y, err := parsePerspectiveOrigin(value, size); err == nil {
				t.Fatalf("parsePerspectiveOrigin(%q) = %v,%v, want an error", value, x, y)
			}
		})
	}
}

// TestImplicitPerspectiveEndpointUsesTheStaticCamera is the regression for a
// partial keyframe list on an element that already has a camera: 0% holds the
// static 800px, not none.
func TestImplicitPerspectiveEndpointUsesTheStaticCamera(t *testing.T) {
	d := cameraDisplay(staticCamera(), cameraRule(layout.Keyframe{
		Offset: 0.5,
		Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective", Value: "500px"},
		},
	}))

	start := oneCameraSample(t, d, 0)
	if !start.HasPerspective || !near(start.Perspective, 800) {
		t.Fatalf("0%% perspective = %+v, want the static 800", start)
	}

	middle := oneCameraSample(t, cameraDisplay(staticCamera(), d.Keyframes), 0.5)
	if !middle.HasPerspective || !near(middle.Perspective, 500) {
		t.Fatalf("50%% perspective = %+v, want the declared 500", middle)
	}
}

// TestImplicitPerspectiveOriginUsesTheStaticOrigin checks the origin half of
// the same rule: a partial origin list starts at the element's static point.
func TestImplicitPerspectiveOriginUsesTheStaticOrigin(t *testing.T) {
	d := cameraDisplay(staticCamera(), cameraRule(layout.Keyframe{
		Offset: 0.5,
		Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective-origin", Value: "100px 50px"},
		},
	}))

	start := oneCameraSample(t, d, 0)
	if !start.HasPerspectiveOrigin || !near(start.PerspectiveOriginX, 10) ||
		!near(start.PerspectiveOriginY, 20) {
		t.Fatalf("0%% origin = %+v, want the static 10,20", start)
	}
	if start.HasPerspective {
		t.Fatalf("an origin-only endpoint changed the distance: %+v", start)
	}

	middle := oneCameraSample(t, cameraDisplay(staticCamera(), d.Keyframes), 0.5)
	if !near(middle.PerspectiveOriginX, 100) || !near(middle.PerspectiveOriginY, 50) {
		t.Fatalf("50%% origin = %v,%v, want the declared 100,50",
			middle.PerspectiveOriginX, middle.PerspectiveOriginY)
	}
}

// TestExplicitPerspectiveEndpointWins checks that a declared 0% or 100% stop
// is not replaced by the static base.
func TestExplicitPerspectiveEndpointWins(t *testing.T) {
	rules := cameraRule(
		layout.Keyframe{Offset: 0, Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective", Value: "500px"},
			{Property: "perspective-origin", Value: "0px 0px"},
		}},
		layout.Keyframe{Offset: 1, Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective", Value: "1000px"},
			{Property: "perspective-origin", Value: "100px 50px"},
		}},
	)

	start := oneCameraSample(t, cameraDisplay(staticCamera(), rules), 0)
	if !near(start.Perspective, 500) {
		t.Fatalf("0%% perspective = %v, want the explicit 500", start.Perspective)
	}
	if !start.HasPerspectiveOrigin || !near(start.PerspectiveOriginX, 0) ||
		!near(start.PerspectiveOriginY, 0) {
		t.Fatalf("0%% origin = %+v, want the explicit 0,0", start)
	}

	end := oneCameraSample(t, cameraDisplay(staticCamera(), rules), 1)
	if !near(end.Perspective, 1000) {
		t.Fatalf("100%% perspective = %v, want the explicit 1000", end.Perspective)
	}
	if !near(end.PerspectiveOriginX, 100) || !near(end.PerspectiveOriginY, 50) {
		t.Fatalf("100%% origin = %v,%v, want the explicit 100,50",
			end.PerspectiveOriginX, end.PerspectiveOriginY)
	}
}

// TestStaticNoneImplicitPerspectiveStaysNone pins that a missing camera is not
// invented from the box default.
func TestStaticNoneImplicitPerspectiveStaysNone(t *testing.T) {
	none := layout.TransformInfo{Parent: layout.IdentityMatrix(), World: layout.IdentityMatrix()}
	d := cameraDisplay(none, cameraRule(layout.Keyframe{
		Offset: 0.5,
		Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective", Value: "500px"},
		},
	}))

	start := oneCameraSample(t, d, 0)
	if !start.HasPerspective || !near(start.Perspective, 0) {
		t.Fatalf("0%% perspective = %+v, want a none endpoint", start)
	}

	// Before the midpoint CSS is discrete across none.
	early := oneCameraSample(t, cameraDisplay(none, d.Keyframes), 0.1)
	if !early.HasPerspective || !near(early.Perspective, 0) {
		t.Fatalf("10%% perspective = %+v, want the none endpoint", early)
	}
}

// TestMalformedPerspectiveOriginErrorsOnce pins that a bad declaration is a
// single bind error, not one per keyframe.
func TestMalformedPerspectiveOriginErrorsOnce(t *testing.T) {
	d := cameraDisplay(staticCamera(), cameraRule(
		layout.Keyframe{Offset: 0, Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective-origin", Value: "left 10px center"},
		}},
		layout.Keyframe{Offset: 1, Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective-origin", Value: "left 10px center"},
		}},
	))

	_, _, err := BindBox(d, d.Boxes[0], 0, time.Unix(200, 0))
	if err == nil {
		t.Fatal("a malformed perspective-origin bound without an error")
	}
	if !errors.Is(err, ErrPerspective) {
		t.Fatalf("error = %v, want ErrPerspective", err)
	}
	if got := strings.Count(err.Error(), "invalid perspective value"); got != 1 {
		t.Fatalf("error carries the perspective failure %d times, want one: %v", got, err)
	}
}
