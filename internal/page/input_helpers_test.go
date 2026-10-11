package page

// runOp returns the text operation that paints want, or nil.
func runOp(p *Page, want string) *DisplayOp {
	for i := range p.display.Ops {
		op := &p.display.Ops[i]
		if op.Kind == DisplayOpText && op.Text == want {
			return op
		}
	}

	return nil
}

// runSpan returns the run's left edge and width in CSS pixels.
func runSpan(p *Page, op *DisplayOp) (float64, float64) {
	pp := p.display.PointsPerPixel

	return op.X / pp, op.W / pp
}

// runBase returns the run's baseline in CSS pixels.
func runBase(p *Page, op *DisplayOp) float64 {
	return op.Y / p.display.PointsPerPixel
}
