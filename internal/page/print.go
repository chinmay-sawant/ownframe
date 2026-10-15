package page

import (
	"context"
	"errors"
	"io"

	"github.com/chinmay-sawant/ownframe/internal/print"
)

// PDFOptions names a page size, a margin, and a PDF profile. blinkless does
// not write PDF, so PDF, WritePDF, SavePDF, and Print return ErrNoPDF and
// these fields are unused.
type PDFOptions struct {
	// PageSize is a page name such as "A4" or "Letter".
	PageSize string
	// Margin is one width in millimetres applied to all four sides.
	Margin float64
	// Profile is a PDF conformance profile such as "PDF/A-4" or "PDF/UA-1".
	Profile string
}

// ErrNoPrinter means the system has no print path. Print wraps it and names
// SavePDF as the fallback.
var ErrNoPrinter = print.ErrNoPrinter

// ErrNoPDF means the layout engine does not write a PDF. blinkless returns a
// drawing list and stops there.
var ErrNoPDF = errors.New("ownframe: blinkless does not write PDF")

// PDF reports that the page cannot be written as a PDF.
func (p *Page) PDF(ctx context.Context, _ PDFOptions) ([]byte, error) {
	if err := pdfReady(ctx, p); err != nil {
		return nil, err
	}

	return nil, ErrNoPDF
}

// WritePDF reports that the page cannot be written as a PDF.
func (p *Page) WritePDF(ctx context.Context, _ io.Writer, _ PDFOptions) error {
	if err := pdfReady(ctx, p); err != nil {
		return err
	}

	return ErrNoPDF
}

// SavePDF validates path, then writes the last template output to it.
func (p *Page) SavePDF(ctx context.Context, path string, opts PDFOptions) error {
	if err := pdfReady(ctx, p); err != nil {
		return err
	}

	if err := print.ValidPath(path); err != nil {
		return err
	}

	return ErrNoPDF
}
