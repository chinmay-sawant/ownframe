package page

// SupportsPDF reports whether this page can write a PDF. It always returns
// false: blinkless hands back a drawing list and has no document writer, so
// PDF, WritePDF, SavePDF, and Print keep returning ErrNoPDF. Check it before
// calling any of those, and fall back to PNG. The receiver is never read,
// so a nil page reports false instead of panicking.
func (p *Page) SupportsPDF() bool {
	return false
}
