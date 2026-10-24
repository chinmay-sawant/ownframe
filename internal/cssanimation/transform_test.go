package cssanimation

import (
	"errors"
	"math"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
)

const radTolerance = 1e-9

func testTransformSize() transformSize {
	return transformSize{width: 200, height: 100, pointsPerPixel: 1, hasBox: true}
}

func parseTestValue(t *testing.T, value string, size transformSize) animation.TransformValue {
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

func checkTransform(t *testing.T, want, got animation.Transform) {
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
		want  animation.Transform
	}{
		{"none", animation.IdentityTransform()},
		{"translate(10px)", animation.Transform{Tx: 10, Sx: 1, Sy: 1}},
		{"translate(10px, 20px)", animation.Transform{Tx: 10, Ty: 20, Sx: 1, Sy: 1}},
		{"translate(10px, 0px)", animation.Transform{Tx: 10, Sx: 1, Sy: 1}},
		{"translateX(30px)", animation.Transform{Tx: 30, Sx: 1, Sy: 1}},
		{"translateY(40px)", animation.Transform{Ty: 40, Sx: 1, Sy: 1}},
		{"scale(2)", animation.Transform{Sx: 2, Sy: 2}},
		{"scale(2, 3)", animation.Transform{Sx: 2, Sy: 3}},
		{"scaleX(2)", animation.Transform{Sx: 2, Sy: 1}},
		{"scaleY(3)", animation.Transform{Sx: 1, Sy: 3}},
		{"rotate(90deg)", animation.Transform{Sx: 1, Sy: 1, Rotate: math.Pi / 2}},
		{"rotateZ(0.5turn)", animation.Transform{Sx: 1, Sy: 1, Rotate: math.Pi}},
		{"skewX(30deg)", animation.Transform{Sx: 1, Sy: 1, SkewX: math.Pi / 6}},
		{"matrix(2, 0, 0, 3, 10, 20)", animation.Transform{Tx: 10, Ty: 20, Sx: 2, Sy: 3}},
		{"translate(10px, 20px) rotate(0deg) scale(2, 3)",
			animation.Transform{Tx: 10, Ty: 20, Sx: 2, Sy: 3}},
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
		name  animation.TransformFunc
	}{
		{"translate(1px, 2px)", animation.Translate},
		{"translateX(1px)", animation.TranslateX},
		{"translateY(1px)", animation.TranslateY},
		{"translate3d(1px, 2px, 3px)", animation.Translate3D},
		{"translateZ(1px)", animation.TranslateZ},
		{"scale(2, 3)", animation.Scale},
		{"scaleX(2)", animation.ScaleX},
		{"scaleY(2)", animation.ScaleY},
		{"scale3d(1, 2, 3)", animation.Scale3D},
		{"scaleZ(2)", animation.ScaleZ},
		{"rotate(10deg)", animation.Rotate},
		{"rotateX(10deg)", animation.RotateX},
		{"rotateY(10deg)", animation.RotateY},
		{"rotateZ(10deg)", animation.RotateZ},
		{"rotate3d(0, 0, 1, 10deg)", animation.Rotate3D},
		{"skew(10deg, 20deg)", animation.Skew},
		{"skewX(10deg)", animation.SkewX},
		{"skewY(10deg)", animation.SkewY},
		{"matrix(1, 0, 0, 1, 0, 0)", animation.Matrix},
		{"matrix3d(1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1)", animation.Matrix3D},
		{"perspective(100px)", animation.Perspective},
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
