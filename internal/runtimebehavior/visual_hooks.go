package runtimebehavior

import (
	"bytes"
	"image"
)

// CompareImages compares two byte slices and reports equality.
func (v *VisualComparer) CompareImages(full, partial []byte) bool {
	if v != nil {
		v.SetImages(full, partial)
	}
	return v.Compare()
}

// Run compares two byte slices.
func (v *VisualComparer) Run(full, partial []byte) bool {
	return v.CompareImages(full, partial)
}

// RunCompare compares two byte slices.
func (v *VisualComparer) RunCompare(full, partial []byte) bool {
	return v.CompareImages(full, partial)
}

// IsEqual reports the last comparison result.
func (v *VisualComparer) IsEqual() bool {
	if v == nil {
		return true
	}
	return v.Equal
}

// Matches is an alias for IsEqual.
func (v *VisualComparer) Matches() bool {
	return v.IsEqual()
}

// SetMismatchHook registers a mismatch callback.
func (v *VisualComparer) SetMismatchHook(fn func()) {
	if v == nil {
		return
	}
	v.OnMismatch = fn
}

// SetOnMismatch registers a mismatch callback.
func (v *VisualComparer) SetOnMismatch(fn func()) {
	v.SetMismatchHook(fn)
}

// CompareImage compares two image.Image values by bounds and pixels.
func (v *VisualComparer) CompareImage(a, b image.Image) bool {
	if a == nil || b == nil {
		eq := a == b
		if v != nil {
			v.Equal = eq
			v.Mismatch = !eq
		}
		return eq
	}
	if !a.Bounds().Eq(b.Bounds()) {
		if v != nil {
			v.Equal = false
			v.Mismatch = true
		}
		return false
	}
	bds := a.Bounds()
	for y := bds.Min.Y; y < bds.Max.Y; y++ {
		for x := bds.Min.X; x < bds.Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				if v != nil {
					v.Equal = false
					v.Mismatch = true
				}
				return false
			}
		}
	}
	if v != nil {
		v.Equal = true
		v.Mismatch = false
	}
	return true
}

// CompareVisual compares two byte slices without state.
func CompareVisual(full, partial []byte) bool {
	return bytes.Equal(full, partial)
}

// VisualEqual compares two byte slices.
func VisualEqual(full, partial []byte) bool {
	return bytes.Equal(full, partial)
}
