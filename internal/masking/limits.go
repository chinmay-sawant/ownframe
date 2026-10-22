package masking

const maxMaskPixels = 16 * 1024 * 1024

// MaxImagePixels caps one decoded image, mirroring maxMaskPixels.
const MaxImagePixels = 16 * 1024 * 1024

// PixelsTooBig reports an image too large to decode safely.
func PixelsTooBig(w, h int) bool {
	if w <= 0 || h <= 0 {
		return false
	}
	if w > MaxImagePixels || h > MaxImagePixels {
		return true
	}
	return int64(w)*int64(h) > MaxImagePixels
}
