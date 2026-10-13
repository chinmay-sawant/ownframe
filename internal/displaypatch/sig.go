package displaypatch

import "image"

// RasterSignature captures fields that decide whether a reused element chunk
// can be blitted instead of re-rasterized.
type RasterSignature struct {
	Content string
	Face    string
	Color   uint32
	Bounds  image.Rectangle
	AnimID  int
	TextOp  bool
	Text    string
	Alpha   float64
	Rotate  float64
}

// Equal reports whether two signatures match.
func (a RasterSignature) Equal(b RasterSignature) bool {
	return a == b
}

// DebugLog records patched vs added counts for a frame.
type DebugLog struct {
	Patched, Added, Removed int
	Enabled                 bool
}

// Record updates the log.
func (d *DebugLog) Record(res PatchResult) {
	if d == nil {
		return
	}
	d.Patched = len(res.Changed) + res.Unchanged
	d.Added = len(res.Added)
	d.Removed = len(res.Removed)
}

// String renders a short debug line.
func (d DebugLog) String() string {
	if !d.Enabled {
		return ""
	}
	return "patched=" + itoa(d.Patched) + " added=" + itoa(d.Added) + " removed=" + itoa(d.Removed)
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	buf := []byte{}
	for v > 0 {
		buf = append([]byte{byte('0' + v%10)}, buf...)
		v /= 10
	}
	if neg {
		return "-" + string(buf)
	}
	return string(buf)
}
