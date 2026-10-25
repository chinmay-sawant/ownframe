package cssanimation

import (
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/animation"
)

func testMatrices() []mat2 {
	return []mat2{
		identity2D(),
		translate2D(10, 20),
		scale2D(2, 3),
		scale2D(-1, 1),
		rotate2D(math.Pi / 3),
		skewX2D(math.Pi / 6),
		translate2D(4, 5).mul(rotate2D(math.Pi / 4)).mul(scale2D(2, 3)),
		translate2D(-7, 2).mul(rotate2D(-1.2)).mul(skewX2D(.3)).mul(scale2D(-2, .5)),
	}
}

func TestDecomposeRecomposeRoundTrip(t *testing.T) {
	for i, want := range testMatrices() {
		decomposed := decompose2D(want)
		got := recompose2D(decomposed)
		for _, field := range []struct {
			name string
			a, b float64
		}{
			{"a", want.a, got.a}, {"b", want.b, got.b},
			{"c", want.c, got.c}, {"d", want.d, got.d},
			{"e", want.e, got.e}, {"f", want.f, got.f},
		} {
			if math.Abs(field.a-field.b) > 1e-9 {
				t.Errorf("matrix %d %s = %v, want %v", i, field.name, field.b, field.a)
			}
		}
	}
}

func TestDecomposeTwoDimensionalFields(t *testing.T) {
	cases := []struct {
		name string
		m    mat2
		want animation.Transform
	}{
		{"translate", translate2D(3, 4), animation.Transform{Tx: 3, Ty: 4, Sx: 1, Sy: 1}},
		{"scale", scale2D(2, 3), animation.Transform{Sx: 2, Sy: 3}},
		{"rotate", rotate2D(math.Pi / 2), animation.Transform{Sx: 1, Sy: 1, Rotate: math.Pi / 2}},
		{"skewX", skewX2D(math.Pi / 6), animation.Transform{Sx: 1, Sy: 1, SkewX: math.Pi / 6}},
	}
	for _, c := range cases {
		got := decompose2D(c.m)
		if math.Abs(got.Tx-c.want.Tx) > 1e-9 || math.Abs(got.Ty-c.want.Ty) > 1e-9 ||
			math.Abs(got.Sx-c.want.Sx) > 1e-9 || math.Abs(got.Sy-c.want.Sy) > 1e-9 ||
			math.Abs(got.Rotate-c.want.Rotate) > 1e-9 || math.Abs(got.SkewX-c.want.SkewX) > 1e-9 {
			t.Errorf("%s decomposed = %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestComposeTwoDimensionalListOrder(t *testing.T) {
	functions := []animation.TransformFunction{
		{Name: animation.Translate, Args: []float64{10, 20}},
		{Name: animation.Scale, Args: []float64{2, 3}},
	}
	got := compose2D(functions)
	want := translate2D(10, 20).mul(scale2D(2, 3))
	if got != want {
		t.Fatalf("compose2D = %+v, want %+v", got, want)
	}
}

func TestComposeThreeDimensionalTranslation(t *testing.T) {
	functions := []animation.TransformFunction{
		{Name: animation.Translate3D, Args: []float64{10, 20, 30}},
	}
	m := compose3D(functions)
	if m[matrix3DTx] != 10 || m[matrix3DTy] != 20 || m[matrix3DTz] != 30 {
		t.Fatalf("translate3d column = %v,%v,%v, want 10,20,30",
			m[matrix3DTx], m[matrix3DTy], m[matrix3DTz])
	}
	rotated := compose3D([]animation.TransformFunction{
		{Name: animation.RotateZ, Args: []float64{math.Pi / 2}},
		{Name: animation.TranslateX, Args: []float64{10}},
	})
	if math.Abs(rotated[matrix3DTx]) > 1e-9 || math.Abs(rotated[matrix3DTy]-10) > 1e-9 {
		t.Fatalf("rotate then translate = %v,%v, want 0,10", rotated[matrix3DTx], rotated[matrix3DTy])
	}
}

func TestPerspectiveMatrix(t *testing.T) {
	m := compose3D([]animation.TransformFunction{{Name: animation.Perspective, Args: []float64{100}}})
	if m[perspectiveW] != -1.0/100 {
		t.Fatalf("perspective slot = %v, want %v", m[perspectiveW], -1.0/100)
	}
}
