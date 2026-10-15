package page

import "image"

// pageContent caches the painted rect. Embedded in Page so promotion
// keeps p.content working and page_types.go stays in the size limit.
// Writer: the box cache stores and invalidates. Lock: the page loop
// serializes all writes; readers run on the same loop.
type pageContent struct {
	content   image.Rectangle
	contentW  int
	contentH  int
	contentOK bool
}
