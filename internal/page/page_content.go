package page

import "image"

// pageContent caches the painted rect. Embedded in Page so promotion
// keeps p.content working and page_types.go stays in the size limit.
type pageContent struct {
	content   image.Rectangle
	contentW  int
	contentH  int
	contentOK bool
}
