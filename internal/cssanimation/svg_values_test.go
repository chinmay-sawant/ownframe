package cssanimation

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

const svgParseTolerance = 1e-9

// svgStart is the epoch every SVG paint bind starts at.
func svgStart() time.Time { return time.Unix(300, 0) }

func svgNear(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > svgParseTolerance {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

// svgRule builds a two-stop rule that declares one property at each edge.
func svgRule(name, property, from, to string) []layout.KeyframesRule {
	return []layout.KeyframesRule{{Name: name, Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: property, Value: from}}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: property, Value: to}}},
	}}}
}

// bindSVG binds one box with one operation so the parser and the player share
// the same fixture.
func bindSVG(t *testing.T, rules []layout.KeyframesRule, name string) (*layout.Display, *animation.Player) {
	t.Helper()
	d := &layout.Display{
		Width: 24, Height: 24, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{{Kind: layout.DisplayOpFillRect, W: 24, H: 24, R: 1, Alpha: 1}},
		Boxes: []layout.Box{{ID: "icon", W: 24, H: 24,
			Animations: []layout.Animation{{
				Name: name, DurationMS: 1000, IterationCount: 1,
				FillMode: "both", PlayState: "running",
			}},
		}},
		Keyframes: rules,
	}
	initial := cssmotion.Read(&d.Ops[0])
	player, err := Bind(d, d.Boxes[0], []int{0}, []cssmotion.Paint{initial}, svgStart())
	if err != nil {
		t.Fatalf("bind: %v", err)
	}

	return d, player
}

func TestParseSVGPaintKeyframes(t *testing.T) {
	rules := []layout.KeyframesRule{{Name: "icon", Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{
			{Property: "fill", Value: "#000000"},
			{Property: "fill-opacity", Value: "0.2"},
			{Property: "stroke", Value: "rgb(18, 52, 86)"},
			{Property: "stroke-opacity", Value: "40%"},
			{Property: "stroke-width", Value: "1px"},
			{Property: "stroke-dasharray", Value: "2, 4"},
			{Property: "stroke-dashoffset", Value: "0px"},
		}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{
			{Property: "fill", Value: "#ffffff"},
			{Property: "fill-opacity", Value: "0.8"},
			{Property: "stroke", Value: "red"},
			{Property: "stroke-opacity", Value: "0.9"},
			{Property: "stroke-width", Value: "5px"},
			{Property: "stroke-dasharray", Value: "4 8"},
			{Property: "stroke-dashoffset", Value: "10px"},
		}},
	}}}

	frames := keyframes(rules, "icon", 1, 24, 24)
	if len(frames) != 2 {
		t.Fatalf("frames = %d, want two", len(frames))
	}
	for i := range frames {
		if frames[i].Err != nil {
			t.Fatalf("frame %d error: %v", i, frames[i].Err)
		}
	}

	first := frames[0].SVG
	if !first.HasFill || first.Fill.R != 0 || first.Fill.G != 0 || first.Fill.B != 0 {
		t.Fatalf("first fill = %+v, want #000000", first)
	}
	if !first.HasFillOpacity {
		t.Fatal("fill-opacity is missing")
	}
	svgNear(t, "fill-opacity", first.FillOpacity, .2)
	if !first.HasStroke {
		t.Fatal("stroke is missing")
	}
	svgNear(t, "stroke red", first.Stroke.R, 18.0/255)
	svgNear(t, "stroke green", first.Stroke.G, 52.0/255)
	svgNear(t, "stroke blue", first.Stroke.B, 86.0/255)
	svgNear(t, "stroke-opacity", first.StrokeOpacity, .4)
	if !first.HasStrokeWidth {
		t.Fatal("stroke-width is missing")
	}
	svgNear(t, "stroke-width", first.StrokeWidth, 1)
	if !first.HasStrokeDashArray || len(first.StrokeDashArray) != 2 {
		t.Fatalf("dash array = %v, want two lengths", first.StrokeDashArray)
	}
	svgNear(t, "dash first", first.StrokeDashArray[0], 2)
	svgNear(t, "dash second", first.StrokeDashArray[1], 4)
	if !first.HasStrokeDashOffset {
		t.Fatal("stroke-dashoffset is missing")
	}
	svgNear(t, "dash offset", first.StrokeDashOffset, 0)

	last := frames[1].SVG
	if !last.HasFill || last.Fill.R != 1 || last.Fill.G != 1 || last.Fill.B != 1 {
		t.Fatalf("last fill = %+v, want #ffffff", last)
	}
	svgNear(t, "last stroke-width", last.StrokeWidth, 5)
	svgNear(t, "last dash offset", last.StrokeDashOffset, 10)
}

func TestSVGPaintLengthsUseLayoutPoints(t *testing.T) {
	frames := keyframes(svgRule("icon", "stroke-width", "8px", "8px"), "icon", .75, 24, 24)

	svgNear(t, "stroke width in points", frames[0].SVG.StrokeWidth, 6)
}

func TestSVGPaintNoneIsPreserved(t *testing.T) {
	frames := keyframes(svgRule("icon", "fill", "none", "#ff0000"), "icon", 1, 24, 24)

	if !frames[0].SVG.HasFill || !frames[0].SVG.Fill.None {
		t.Fatalf("fill none = %+v, want a none fill", frames[0].SVG)
	}
	if frames[1].SVG.Fill.None {
		t.Fatal("red parsed as none")
	}
}

func TestSVGPaintFillMidpointFromBind(t *testing.T) {
	d, player := bindSVG(t, svgRule("paint", "fill", "#000000", "#ffffff"), "paint")
	player.Step(svgStart().Add(500 * time.Millisecond))

	values := player.SVGPaintValues()
	if len(values) != 1 {
		t.Fatalf("SVGPaintValues = %d values, want one", len(values))
	}
	svgNear(t, "fill red", values[0].Fill.R, .5)
	svgNear(t, "fill green", values[0].Fill.G, .5)
	svgNear(t, "fill blue", values[0].Fill.B, .5)
	if d.Ops[0].R != 1 || d.Ops[0].G != 0 || d.Ops[0].B != 0 {
		t.Fatalf("operation fill = %v,%v,%v, want the captured red left alone",
			d.Ops[0].R, d.Ops[0].G, d.Ops[0].B)
	}
}

func TestSVGPaintStrokeWidthAndDashOffsetFromBind(t *testing.T) {
	rules := []layout.KeyframesRule{{Name: "dash", Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{
			{Property: "stroke-width", Value: "1px"},
			{Property: "stroke-dashoffset", Value: "0px"},
		}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{
			{Property: "stroke-width", Value: "5px"},
			{Property: "stroke-dashoffset", Value: "10px"},
		}},
	}}}
	_, player := bindSVG(t, rules, "dash")
	player.Step(svgStart().Add(500 * time.Millisecond))

	values := player.SVGPaintValues()
	if len(values) != 1 {
		t.Fatalf("SVGPaintValues = %d values, want one", len(values))
	}
	svgNear(t, "stroke width", values[0].StrokeWidth, 3)
	svgNear(t, "dash offset", values[0].StrokeDashOffset, 5)
}

func TestSVGPaintDashArrayMismatchFromBind(t *testing.T) {
	_, player := bindSVG(t, svgRule("dash", "stroke-dasharray", "2 4", "4"), "dash")

	player.Step(svgStart().Add(250 * time.Millisecond))
	values := player.SVGPaintValues()
	if len(values) != 1 || len(values[0].StrokeDashArray) != 2 {
		t.Fatalf("dash array before the midpoint = %+v, want the two-length list", values)
	}
	svgNear(t, "dash first", values[0].StrokeDashArray[0], 2)
	svgNear(t, "dash second", values[0].StrokeDashArray[1], 4)

	player.Step(svgStart().Add(750 * time.Millisecond))
	values = player.SVGPaintValues()
	if len(values) != 1 || len(values[0].StrokeDashArray) != 1 {
		t.Fatalf("dash array after the midpoint = %+v, want the one-length list", values)
	}
	svgNear(t, "dash after", values[0].StrokeDashArray[0], 4)
}

func TestSVGPaintNoneToColourFromBind(t *testing.T) {
	_, player := bindSVG(t, svgRule("paint", "fill", "none", "#ff0000"), "paint")

	player.Step(svgStart().Add(250 * time.Millisecond))
	values := player.SVGPaintValues()
	if len(values) != 1 || !values[0].Fill.None {
		t.Fatalf("fill before the midpoint = %+v, want none", values)
	}

	player.Step(svgStart().Add(750 * time.Millisecond))
	values = player.SVGPaintValues()
	if len(values) != 1 || values[0].Fill.None {
		t.Fatalf("fill after the midpoint = %+v, want red", values)
	}
	svgNear(t, "fill red", values[0].Fill.R, 1)
}

func TestSVGPaintServerIsDeferred(t *testing.T) {
	if _, err := svgColor("url(#gradient)", Color{}, false); !errors.Is(err, ErrPaintServerDeferred) {
		t.Fatalf("svgColor error = %v, want ErrPaintServerDeferred", err)
	}

	frames := keyframes(svgRule("paint", "fill", "url(#gradient)", "#ff0000"), "paint", 1, 24, 24)
	if frames[0].Err != nil {
		t.Fatalf("url fill failed the bind: %v", frames[0].Err)
	}
	if frames[0].SVG.HasFill {
		t.Fatal("url fill guessed a colour")
	}
	if !strings.Contains(frames[0].SVG.Deferred, "url(#gradient)") {
		t.Fatalf("deferred reason = %q, want the paint server named", frames[0].SVG.Deferred)
	}
}

func TestSVGPaintWithTransformSamplesBoth(t *testing.T) {
	rules := []layout.KeyframesRule{{Name: "move", Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{
			{Property: "transform", Value: "translate(0px)"},
			{Property: "fill", Value: "#000000"},
		}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{
			{Property: "transform", Value: "translate(100px)"},
			{Property: "fill", Value: "#ffffff"},
		}},
	}}}
	_, player := bindSVG(t, rules, "move")
	player.Step(svgStart().Add(500 * time.Millisecond))

	values := player.SVGPaintValues()
	if len(values) != 1 {
		t.Fatalf("SVGPaintValues = %d values, want one", len(values))
	}
	svgNear(t, "fill red", values[0].Fill.R, .5)

	current, _, ok := player.TargetPaint(0)
	if !ok {
		t.Fatal("target paint is missing")
	}
	if !current.HasMatrix {
		t.Fatal("transform did not write a matrix")
	}
	svgNear(t, "matrix translate x", current.Matrix.E, 50)
}
