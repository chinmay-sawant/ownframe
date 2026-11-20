package motion

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"slices"
	"testing"
)

// TestIndexSelectMatchesDirect pins the shared-index contract: every
// selection through a per-display Index returns exactly what the direct scan
// returns, across owner, subtree, fallback, anonymous, and out-of-range
// boxes, with and without an owner map.
func TestIndexSelectMatchesDirect(t *testing.T) {
	kinds := []Kind{KindFill, KindInk, KindStroke, KindSubtree}
	ownersVariants := map[string][]int{
		"owned":      {0, 0, 0, 0, 1},
		"unowned":    nil,
		"mismatched": {0, 0},
	}
	for name, owners := range ownersVariants {
		d := selectDisplay(owners)
		idx := NewIndex(d)
		for bi := range d.Boxes {
			for _, boxIndex := range []int{bi, -1, 99} {
				for _, kind := range kinds {
					want := Select(d, d.Boxes[bi], boxIndex, kind)
					got := idx.Select(d.Boxes[bi], boxIndex, kind)
					if !slices.Equal(got, want) {
						t.Fatalf("%s box %d index %d kind %d = %v, want %v",
							name, bi, boxIndex, kind, got, want)
					}
				}
			}
		}
	}
}

// TestIndexSelectGeometricMatchesDirect covers boxes that predate operation
// ids: the geometric fallback must agree with the direct scan too.
func TestIndexSelectGeometricMatchesDirect(t *testing.T) {
	d := selectDisplay([]int{0, 0, 0, 0, 1})
	geo := layout.Box{ID: "geo", X: 1, Y: 1, W: 8, H: 8}
	idx := NewIndex(d)
	for _, kind := range []Kind{KindFill, KindInk, KindStroke, KindSubtree} {
		want := Select(d, geo, 0, kind)
		if got := idx.Select(geo, 0, kind); !slices.Equal(got, want) {
			t.Fatalf("geometric kind %d = %v, want %v", kind, got, want)
		}
	}
}

// TestIndexSelectNilDisplayMatchesDirect checks that an empty index selects
// nothing, like Select on nil.
func TestIndexSelectNilDisplayMatchesDirect(t *testing.T) {
	idx := NewIndex(nil)
	box := layout.Box{ID: "x", OperationIDs: []uint64{1}}
	for _, kind := range []Kind{KindFill, KindInk, KindStroke, KindSubtree} {
		if got := idx.Select(box, 0, kind); got != nil {
			t.Fatalf("nil index kind %d = %v, want nil", kind, got)
		}
	}
	if got := idx.ImageIndex(box, 0); got != nil {
		t.Fatalf("nil image index = %v, want nil", got)
	}
}

// imageDisplay is one SVG root with a background fill before its raster plus
// an unrelated trailing image, so the tests can check the last-image rule.
func imageDisplay(owners []int) *layout.Display {
	return &layout.Display{
		PointsPerPixel: 1,
		Ops: []layout.DisplayOp{
			{ID: 1, Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 8, H: 8},
			{ID: 2, Kind: layout.DisplayOpImage, X: 0, Y: 0, W: 8, H: 8},
			{ID: 3, Kind: layout.DisplayOpImage, X: 20, Y: 20, W: 8, H: 8},
		},
		OpOwners: owners,
		Boxes: []layout.Box{
			{ID: "icon", Tag: "svg", X: 0, Y: 0, W: 8, H: 8,
				OperationIDs: []uint64{1, 2}, ParentIndex: -1},
		},
	}
}

// TestIndexImageMatchesTiers checks the raster lookup across the owner,
// id-set, geometric, and empty tiers.
func TestIndexImageMatchesTiers(t *testing.T) {
	d := imageDisplay([]int{0, 0, 1})
	idx := NewIndex(d)
	if got := idx.ImageIndex(d.Boxes[0], 0); !slices.Equal(got, []int{1}) {
		t.Fatalf("owner image = %v, want the box raster", got)
	}

	d = imageDisplay(nil)
	idx = NewIndex(d)
	if got := idx.ImageIndex(d.Boxes[0], 0); !slices.Equal(got, []int{1}) {
		t.Fatalf("id-set image = %v, want the box raster", got)
	}

	geo := layout.Box{ID: "geo", X: 0, Y: 0, W: 8, H: 8}
	if got := idx.ImageIndex(geo, -1); !slices.Equal(got, []int{1}) {
		t.Fatalf("geometric image = %v, want the last image in the area", got)
	}

	bare := &layout.Display{PointsPerPixel: 1,
		Ops:      []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect}},
		OpOwners: []int{0},
		Boxes:    []layout.Box{{ID: "bare", OperationIDs: []uint64{1}}},
	}
	if got := NewIndex(bare).ImageIndex(bare.Boxes[0], 0); len(got) != 0 {
		t.Fatalf("imageless image = %v, want empty", got)
	}
}

// TestIndexEmptyAllocZero checks that NewIndex, NewKeyframesIndex, and EmptyIndex
// return the shared empty index with 0 allocations.
func TestIndexEmptyAllocZero(t *testing.T) {
	if EmptyIndex() != emptyIndex {
		t.Fatal("EmptyIndex() != emptyIndex")
	}
	if got := NewIndex(nil); got != emptyIndex {
		t.Fatalf("NewIndex(nil) = %p, want emptyIndex %p", got, emptyIndex)
	}
	emptyDisplay := &layout.Display{}
	if got := NewIndex(emptyDisplay); got != emptyIndex {
		t.Fatalf("NewIndex(empty) = %p, want emptyIndex %p", got, emptyIndex)
	}
	if got := NewKeyframesIndex(nil, emptyDisplay); got != emptyIndex {
		t.Fatalf("NewKeyframesIndex(nil) = %p, want emptyIndex %p", got, emptyIndex)
	}

	allocsNil := testing.AllocsPerRun(100, func() {
		idx := NewIndex(nil)
		if idx == nil {
			t.Fatal("unexpected nil")
		}
	})
	if allocsNil != 0 {
		t.Fatalf("NewIndex(nil) allocs = %v, want 0", allocsNil)
	}

	allocsEmpty := testing.AllocsPerRun(100, func() {
		idx := NewIndex(emptyDisplay)
		if idx == nil {
			t.Fatal("unexpected nil")
		}
	})
	if allocsEmpty != 0 {
		t.Fatalf("NewIndex(empty) allocs = %v, want 0", allocsEmpty)
	}

	allocsKeyframes := testing.AllocsPerRun(100, func() {
		idx := NewKeyframesIndex(nil, emptyDisplay)
		if idx == nil {
			t.Fatal("unexpected nil")
		}
	})
	if allocsKeyframes != 0 {
		t.Fatalf("NewKeyframesIndex(nil) allocs = %v, want 0", allocsKeyframes)
	}
}

func TestOperationIndicesUsesBlinklessOwnership(t *testing.T) {
	d := &layout.Display{Ops: []layout.DisplayOp{
		{ID: 11, X: 10, Y: 10, W: 8, H: 8},
		{ID: 22, X: 10, Y: 10, W: 8, H: 8},
		{ID: 33, X: 10, Y: 10, W: 8, H: 8},
	}}
	child := layout.Box{X: 0, Y: 0, W: 30, H: 30, OperationIDs: []uint64{22}}
	got := OperationIndices(d, child)
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("child indices = %v, want [1] despite overlapping geometry", got)
	}
	parent := child
	parent.OperationIDs = []uint64{11, 22}
	got = OperationIndices(d, parent)
	if len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("parent indices = %v, want [0 1]", got)
	}
}

func TestOperationIndicesFallsBackForLegacyBox(t *testing.T) {
	d := &layout.Display{PointsPerPixel: 1, Ops: []layout.DisplayOp{
		{ID: 11, X: 5, Y: 5, W: 2, H: 2},
		{ID: 22, X: 40, Y: 40, W: 2, H: 2},
	}}
	got := OperationIndices(d, layout.Box{X: 0, Y: 0, W: 20, H: 20})
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("legacy indices = %v, want [0] from geometry", got)
	}
}

func TestOperationIndicesRespectsExplicitEmptyOwnership(t *testing.T) {
	d := &layout.Display{PointsPerPixel: 1, Ops: []layout.DisplayOp{
		{ID: 11, X: 5, Y: 5, W: 2, H: 2},
	}}
	box := layout.Box{X: 0, Y: 0, W: 20, H: 20, OperationIDs: []uint64{}}
	if got := OperationIndices(d, box); len(got) != 0 {
		t.Fatalf("explicit empty ownership indices = %v, want none", got)
	}
}

func TestReadCarriesTheBakedMatrix(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Y: 5, W: 6, H: 7, Alpha: 1}
	want := layout.Matrix2D{A: 0, B: 1, C: -1, D: 0, E: 10, F: 20}
	op.SetXform(want)

	paint := Read(op)
	if !paint.HasMatrix {
		t.Fatal("a transformed operation was read without a matrix")
	}
	if paint.Matrix != want {
		t.Fatalf("matrix = %+v, want %+v", paint.Matrix, want)
	}
}

func TestReadWithoutTransformReportsNoMatrix(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Y: 5, W: 6, H: 7, Alpha: 1}

	paint := Read(op)
	if paint.HasMatrix {
		t.Fatal("an untransformed operation was read with a matrix")
	}
	if paint.Matrix != layout.IdentityMatrix() {
		t.Fatalf("matrix = %+v, want the identity", paint.Matrix)
	}
}

func TestWriteInstallsAMatrixOnlyWhenSet(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Y: 5, W: 6, H: 7, Alpha: 1}
	static := layout.Matrix2D{A: 1, B: 0, C: 0, D: 1, E: 3, F: 2}
	op.SetXform(static)

	paint := Paint{W: 20, H: 10, Alpha: 1, Opacity: 1}
	Write(op, paint)
	if got := op.Transform(); got != static {
		t.Fatalf("a paint without a matrix replaced the transform with %+v", got)
	}

	want := layout.Matrix2D{A: 2, B: 0, C: 0, D: 2, E: 8, F: 6}
	paint.Matrix, paint.HasMatrix = want, true
	Write(op, paint)
	if got := op.Transform(); got != want {
		t.Fatalf("matrix = %+v, want %+v", got, want)
	}
	if op.W != 20 || op.H != 10 {
		t.Fatalf("paint = %vx%v, want 20x10", op.W, op.H)
	}
}

// TestWriteLeavesAlphaAndSizeAlone pins the phase 03 write rule: a keyframed
// paint owns no size or alpha field, so a masked write never changes them. A
// translucent operation must keep its alpha while its transform animates.
func TestWriteLeavesAlphaAndSizeAlone(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Y: 5, W: 6, H: 7, R: .2, G: .3, B: .4, Alpha: .4}

	paint := Paint{X: 10, Y: 11, W: 99, H: 99, Alpha: 1, Opacity: .5,
		Matrix: layout.Matrix2D{A: 1, D: 1, E: 3}, HasMatrix: true, Mask: MaskX | MaskY | MaskMatrix}
	Write(op, paint)
	if op.Alpha != .4 {
		t.Fatalf("write changed alpha to %v, want the translucent .4", op.Alpha)
	}
	if op.W != 6 || op.H != 7 {
		t.Fatalf("write changed the size to %vx%v, want 6x7", op.W, op.H)
	}
	if op.X != 10 || op.Y != 11 {
		t.Fatalf("masked position write = %v,%v, want 10,11", op.X, op.Y)
	}
	if op.R != .2 || op.G != .3 || op.B != .4 {
		t.Fatalf("a geometry write changed the colour to %v,%v,%v", op.R, op.G, op.B)
	}
	if got := Read(op).Opacity; got != 1 {
		t.Fatalf("a geometry write changed the element opacity to %v, want the untouched 1", got)
	}
}

// TestReadKeepsTheCompositeOpacityOutOfThePaint checks that a translucent
// operation round-trips: Read reports the element opacity alone, so writing it
// back does not multiply the alpha a second time.
func TestReadKeepsTheCompositeOpacityOutOfThePaint(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 1, Y: 2, W: 3, H: 4, Alpha: .5}
	op.SetPaintOpacity(.5)

	paint := Read(op)
	if paint.Alpha != .5 {
		t.Fatalf("read alpha = %v, want the operation's .5", paint.Alpha)
	}
	if paint.Opacity != .5 {
		t.Fatalf("read opacity = %v, want the element's .5, not the composited .25", paint.Opacity)
	}
	Write(op, paint)
	if got := op.Opacity(); got != .25 {
		t.Fatalf("round trip composite opacity = %v, want .25", got)
	}
}

func TestBlendKeepsTheDestinationMatrix(t *testing.T) {
	from := Paint{X: 0, Alpha: 1, Opacity: 1}
	to := Paint{X: 10, Alpha: 1, Opacity: 1,
		Matrix: layout.Matrix2D{A: 1, D: 1, E: 20, F: 30}, HasMatrix: true}

	got := Blend(from, to, .5)
	if !got.HasMatrix || got.Matrix != to.Matrix {
		t.Fatalf("blended matrix = %+v (set %v), want the destination %+v", got.Matrix, got.HasMatrix, to.Matrix)
	}
	if got.X != 5 {
		t.Fatalf("blended X = %v, want 5", got.X)
	}
}

// selectDisplay is one parent with a fill, a label, and a border around a child
// card. The owners are explicit so the test does not depend on geometry.
func selectDisplay(owners []int) *layout.Display {
	return &layout.Display{
		PointsPerPixel: 1,
		Ops: []layout.DisplayOp{
			{ID: 11, Kind: layout.DisplayOpFillRect, X: 1, Y: 1, W: 8, H: 8},
			{ID: 12, Kind: layout.DisplayOpText, X: 2, Y: 4, W: 4, H: 4},
			{ID: 13, Kind: layout.DisplayOpStrokeRect, X: 1, Y: 1, W: 8, H: 8},
			{ID: 14, Kind: layout.DisplayOpLine, X: 1, Y: 9, W: 8, H: 1},
			{ID: 21, Kind: layout.DisplayOpFillRect, X: 3, Y: 3, W: 4, H: 4},
		},
		OpOwners: owners,
		Boxes: []layout.Box{
			{ID: "parent", OperationIDs: []uint64{11, 12, 13, 14, 21}, ParentIndex: -1},
			{ID: "child", OperationIDs: []uint64{21}, ParentIndex: 0},
		},
	}
}

func TestSelectKeepsColourOnTheNearestOwner(t *testing.T) {
	d := selectDisplay([]int{0, 0, 0, 0, 1})
	parent := d.Boxes[0]
	tests := []struct {
		kind Kind
		want []int
	}{
		{KindFill, []int{0}},
		{KindInk, []int{1}},
		{KindStroke, []int{2, 3}},
		{KindSubtree, []int{0, 1, 2, 3, 4}},
	}
	for _, test := range tests {
		got := Select(d, parent, 0, test.kind)
		if !slices.Equal(got, test.want) {
			t.Errorf("Select kind %d = %v, want %v", test.kind, got, test.want)
		}
	}
}

// TestSelectFallsBackWithoutOwners checks the documented fallback: an older
// engine that supplies no OpOwners has the owner kinds filtered from the box's
// subtree, so a parent's colour can still touch a child's fill there.
func TestSelectFallsBackWithoutOwners(t *testing.T) {
	d := selectDisplay(nil)
	parent := d.Boxes[0]
	if got := Select(d, parent, 0, KindFill); !slices.Equal(got, []int{0, 4}) {
		t.Fatalf("fill fallback = %v, want the subtree fills", got)
	}
	if got := Select(d, parent, -1, KindInk); !slices.Equal(got, []int{1}) {
		t.Fatalf("ink fallback = %v, want the subtree text", got)
	}
}

// TestSelectFallsBackForAnAnonymousBox checks that a box without a trusted
// display index uses the subtree even when ownership is available.
func TestSelectFallsBackForAnAnonymousBox(t *testing.T) {
	d := selectDisplay([]int{0, 0, 0, 0, 1})
	parent := d.Boxes[0]
	if got := Select(d, parent, -1, KindFill); !slices.Equal(got, []int{0, 4}) {
		t.Fatalf("anonymous fill = %v, want the subtree fills", got)
	}
}
