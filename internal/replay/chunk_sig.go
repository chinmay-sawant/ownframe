package replay

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

// fnvOffsetBasis is the FNV-1a 64-bit offset basis starting a signature.
const fnvOffsetBasis uint64 = 14695981039346656037

// fnvPrime is the FNV-1a 64-bit prime folding each word.
const fnvPrime uint64 = 1099511628211

// chunkSig returns the content signature and the page box of one run. An
// unbounded member or a box past the caps gives an empty box, so the caller
// keeps the run direct.
func chunkSig(d *layout.Display, c *chunkRun, content image.Rectangle) (uint64, image.Rectangle) {
	h := fnvOffsetBasis
	box := image.Rectangle{}

	for _, op := range c.ops {
		b, ok := opBounds(op, d.PointsPerPixel)
		if !ok || b.Empty() {
			return 0, image.Rectangle{}
		}

		box = box.Union(b)
		h = sigOp(h, op)
	}

	box = box.Intersect(content)
	if box.Empty() || !chunkFits(box) {
		return 0, image.Rectangle{}
	}

	return h, box
}

// runWriteGen is the cheap op-write generation for one run: the tick-writable
// geometry, colour, opacity, and transform fields. Motion and frame anims
// patch only these fields in place on the same display; structural payloads
// (text, grid segments, path commands, image bytes, fonts) change only on a
// redraw, which changes the display fingerprint and rebuilds the plan with
// fresh signatures. Equality of the write generation implies the cached
// signature is still valid, so a repaint reuses it instead of rehashing.
func runWriteGen(c *chunkRun) uint64 {
	h := fnvOffsetBasis
	h = mix64(h, uint64(len(c.ops)))

	for _, op := range c.ops {
		h = mix64(h, uint64(op.Kind))
		h = mixF(h, op.X)
		h = mixF(h, op.Y)
		h = mixF(h, op.W)
		h = mixF(h, op.H)
		h = mixF(h, op.R)
		h = mixF(h, op.G)
		h = mixF(h, op.B)
		h = mixF(h, op.Alpha)
		h = mixF(h, op.Opacity())
		h = mixF(h, op.Width)
		h = mixF(h, float64(op.RotateDeg))
		h = mix64(h, bit(op.Fixed))
		h = mix64(h, bit(op.XformSet))
		h = sigXformWrite(h, op)
	}

	return h
}

// cachedChunkSig returns the run's cached signature when its write generation
// still matches, and recomputes both signature and box otherwise. The slow
// path returns fresh values without mutating the cached ones when called
// concurrently; the owning loop refreshes the cache via refreshRunSig.
func cachedChunkSig(d *layout.Display, c *chunkRun, content image.Rectangle) (uint64, image.Rectangle, bool) {
	if c.hasSig && runWriteGen(c) == c.writeGen {
		return c.sig, c.box, true
	}

	sig, box := chunkSig(d, c, content)
	if box.Empty() {
		return 0, image.Rectangle{}, false
	}

	return sig, box, false
}

// refreshRunSig records a freshly computed signature after a write-generation
// miss. Call on the owning Ebiten loop only.
func refreshRunSig(c *chunkRun, sig uint64, box image.Rectangle) {
	c.sig = sig
	c.box = box
	c.writeGen = runWriteGen(c)
	c.hasSig = true
}

// mix64 folds one word into an FNV-1a style signature.
func mix64(h, v uint64) uint64 {
	return (h ^ v) * fnvPrime
}

// mixF folds one float.
func mixF(h uint64, v float64) uint64 {
	return mix64(h, math.Float64bits(v))
}

// mixStr folds a string.
func mixStr(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		h = mix64(h, uint64(s[i]))
	}

	return h
}

// bit packs one bool into a word.
func bit(v bool) uint64 {
	if v {
		return 1
	}

	return 0
}
