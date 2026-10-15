package host

import "context"

// PDFExporter is a screen that renders PDF bytes for options O. O is the
// page's own options type, so host stays below page: web asserts
// PDFExporter[page.PDFOptions] without host importing page.
type PDFExporter[O any] interface {
	PDF(ctx context.Context, opts O) ([]byte, error)
}
