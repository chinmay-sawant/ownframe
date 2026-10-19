package csspreview

func blend(a, b paint, t float64) paint {
	f := func(x, y float64) float64 { return x + (y-x)*t }
	return paint{x: f(a.x, b.x), y: f(a.y, b.y), w: f(a.w, b.w), h: f(a.h, b.h),
		r: f(a.r, b.r), g: f(a.g, b.g), b: f(a.b, b.b), a: f(a.a, b.a), opacity: f(a.opacity, b.opacity)}
}
