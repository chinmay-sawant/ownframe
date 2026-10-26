package page

import (
	"context"
	"testing"
)

func TestPageWheelHandler(t *testing.T) {
	t.Parallel()

	p := &Page{}
	if handled, err := p.Wheel(context.Background(), 0, 0, 0, 1); handled || err != nil {
		t.Fatalf("no handler = %v %v", handled, err)
	}

	var got struct {
		x, y, dx, dy float64
	}

	p.handlers.Wheel = func(_ context.Context, x, y, dx, dy float64) (bool, error) {
		got.x, got.y, got.dx, got.dy = x, y, dx, dy

		return true, nil
	}

	handled, err := p.Wheel(context.Background(), 12, 34, 5, 6)
	if err != nil || !handled {
		t.Fatalf("handler = %v %v", handled, err)
	}

	if got.x != 12 || got.y != 34 || got.dx != 5 || got.dy != 6 {
		t.Fatalf("args = %+v", got)
	}
}
