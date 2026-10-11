package window

import "context"

type pointerRouteScreen struct {
	fakeScreen
	presses, clicks, releases, drags int
}

func (f *pointerRouteScreen) Press(context.Context, float64, float64) error {
	f.presses++
	return nil
}
func (f *pointerRouteScreen) Click(context.Context, float64, float64) error {
	f.clicks++
	return nil
}
func (f *pointerRouteScreen) Release(context.Context) error {
	f.releases++
	return nil
}
func (f *pointerRouteScreen) SelectAt(context.Context, float64, float64) error { return nil }
func (f *pointerRouteScreen) Drag(context.Context, float64, float64) error {
	f.drags++
	return nil
}
func (f *pointerRouteScreen) SelectWordAt(context.Context, float64, float64) error { return nil }
func (f *pointerRouteScreen) SelectLineAt(context.Context, float64, float64) error { return nil }
