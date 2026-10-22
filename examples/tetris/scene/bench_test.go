package scene

import (
	"context"
	"testing"
	"time"
)

// benchScene builds a drawn scene for the benchmarks.
func benchScene(b *testing.B, fallback bool) (*Scene, *testClock, *fakeStepper) {
	b.Helper()

	step := &fakeStepper{}
	clock := &testClock{at: time.Unix(1000, 0)}
	opts := Options{Stepper: step, Now: clock.now}
	m := &fakeModel{}

	var (
		s   *Scene
		err error
	)

	if fallback {
		s, err = newFromHTML(m, nil, opts, pageHTML(fallbackSrc))
	} else {
		s, err = New(m, nil, opts)
	}

	if err != nil {
		b.Fatal(err)
	}

	if err := s.Redraw(context.Background()); err != nil {
		b.Fatal(err)
	}

	return s, clock, step
}

// BenchmarkTickPaint measures one active frame: one fixed step, the board
// and text paint, and the frame diff.
func BenchmarkTickPaint(b *testing.B) {
	s, clock, step := benchScene(b, false)
	m, ok := s.model.(*fakeModel)
	if !ok {
		b.Fatal("scene model is not a fakeModel")
	}
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.f.Score = i
		step.advance = 1
		clock.add(time.Second / 60)

		if err := s.Tick(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRedraw measures one full rebuild: template execution, layout,
// and a fresh display list.
func BenchmarkRedraw(b *testing.B) {
	s, _, _ := benchScene(b, false)
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := s.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
