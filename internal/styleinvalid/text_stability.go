package styleinvalid

import "hash/fnv"

// TextRunStable reports whether a key is usable for cached reuse.
func TextRunStable(k TextRunKey) bool { return k.Face != "" && k.Content != "" }

// NormalizeTextRunKey returns a cache key with stable subpixel alignment.
func NormalizeTextRunKey(k TextRunKey) TextRunKey {
	if k.SubpixelX < 0 {
		k.SubpixelX = 0
	}
	return k
}

// Signature returns a compact hash for the key.
func (k TextRunKey) Signature() uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(k.Face))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(k.Content))
	return h.Sum64()
}

// Stable reports whether this key is usable for reuse.
func (k TextRunKey) Stable() bool { return TextRunStable(k) }
