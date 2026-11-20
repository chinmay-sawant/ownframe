package frame

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
)

func TestOpBoxMatchesDirtyOpBounds(t *testing.T) {
	t.Parallel()

	scaled := &page.Display{PixelPerPoint: 1.5}

	scaledOp := &page.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 20, W: 30, H: 40}
	cx, cy := paintCenter(scaledOp)
	scaleAbout(2, cx, cy).write(scaledOp)

	rotatedStroke := &page.DisplayOp{Kind: layout.DisplayOpStrokeRect, X: 10, Y: 20, W: 30, H: 40, Width: 4}
	rx, ry := paintCenter(rotatedStroke)
	rotateAbout(45, rx, ry).write(rotatedStroke)

	translatedText := &page.DisplayOp{Kind: layout.DisplayOpText, X: 10, Y: 30, W: 50, Size: 12, InkDescent: 4}
	mat6{1, 0, 0, 1, 15, -5}.write(translatedText)

	cases := []struct {
		name      string
		d         *page.Display
		op        *page.DisplayOp
		wantExact bool
		want      image.Rectangle
	}{
		{
			name:      "fill",
			d:         scaled,
			op:        &page.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 20, W: 30, H: 40},
			wantExact: true,
			want:      image.Rect(14, 29, 61, 91),
		},
		{
			name:      "stroke",
			d:         scaled,
			op:        &page.DisplayOp{Kind: layout.DisplayOpStrokeRect, X: 10, Y: 20, W: 30, H: 40, Width: 4},
			wantExact: true,
			// local box grows by Width/2 = 2 -> [8, 18]..[42, 62]; at 1.5x -> [12, 27]..[63, 93] +/- 1px.
			want: image.Rect(11, 26, 64, 94),
		},
		{
			name: "stroke-thin",
			d:    scaled,
			op:   &page.DisplayOp{Kind: layout.DisplayOpStrokeRect, X: 10, Y: 20, W: 30, H: 40, Width: 0},
		},
		{
			name:      "line",
			d:         scaled,
			op:        &page.DisplayOp{Kind: layout.DisplayOpLine, X: 10, Y: 20, W: 30, H: 0, Width: 6},
			wantExact: true,
			// local box grows by Width/2 = 3 -> [7, 17]..[43, 23]; at 1.5x -> [10.5, 25.5]..[64.5, 34.5] -> int +/- 1px.
			want: image.Rect(9, 24, 65, 35),
		},
		{
			name:      "text",
			d:         scaled,
			op:        &page.DisplayOp{Kind: layout.DisplayOpText, X: 10, Y: 30, W: 50, Size: 12, InkDescent: 4},
			wantExact: true,
			// local box y is [30-12, 30+4] = [18, 34]; at 1.5x -> [15, 27]..[90, 51] +/- 1px.
			want: image.Rect(14, 26, 91, 52),
		},
		{
			name: "bullet",
			d:    scaled,
			op:   &page.DisplayOp{Kind: layout.DisplayOpBullet, X: 10, Y: 30, W: 8, Size: 10, InkDescent: 2},
		},
		{
			name: "transformed-scale",
			d:    scaled,
			op:   scaledOp,
		},
		{
			name: "transformed-rotate-stroke",
			d:    scaled,
			op:   rotatedStroke,
		},
		{
			name: "transformed-translate-text",
			d:    scaled,
			op:   translatedText,
		},
		{
			name:      "nil-display",
			d:         nil,
			op:        &page.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 20, W: 30, H: 40},
			wantExact: true,
			want:      image.Rect(9, 19, 41, 61),
		},
		{
			name:      "nil-op",
			d:         scaled,
			op:        nil,
			wantExact: true,
			want:      image.Rectangle{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotFrame := OpBox(tc.d, tc.op)
			gotDirty := dirty.OpBounds(tc.op, tc.d)
			if gotFrame != gotDirty {
				t.Fatalf("OpBox = %v, dirty.OpBounds = %v", gotFrame, gotDirty)
			}
			if tc.wantExact && gotFrame != tc.want {
				t.Fatalf("bounds = %v, want %v", gotFrame, tc.want)
			}
		})
	}
}
