package page_test

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// TestSupportsPDFFalseOnEverySink checks the probe against each PDF sink:
// none of them can succeed, and the direct calls keep refusing.
func TestSupportsPDFFalseOnEverySink(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newReport(t)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if p.SupportsPDF() {
		t.Fatal("SupportsPDF = true, want false")
	}

	if _, err := p.PDF(ctx, page.PDFOptions{}); !errors.Is(err, page.ErrNoPDF) {
		t.Fatalf("PDF err = %v, want ErrNoPDF", err)
	}

	if err := p.WritePDF(ctx, &bytes.Buffer{}, page.PDFOptions{}); !errors.Is(err, page.ErrNoPDF) {
		t.Fatalf("WritePDF err = %v, want ErrNoPDF", err)
	}

	path := filepath.Join(t.TempDir(), "probe.pdf")
	if err := p.SavePDF(ctx, path, page.PDFOptions{}); !errors.Is(err, page.ErrNoPDF) {
		t.Fatalf("SavePDF err = %v, want ErrNoPDF", err)
	}

	err := p.Print(ctx, page.PDFOptions{})
	if !errors.Is(err, page.ErrNoPDF) && !errors.Is(err, page.ErrNoPrinter) {
		t.Fatalf("Print err = %v, want ErrNoPDF or ErrNoPrinter", err)
	}
}

// TestSupportsPDFNilPage checks the probe on a nil page.
func TestSupportsPDFNilPage(t *testing.T) {
	t.Parallel()

	var p *page.Page
	if p.SupportsPDF() {
		t.Fatal("SupportsPDF = true on nil page, want false")
	}
}
