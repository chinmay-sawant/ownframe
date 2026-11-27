package page_test

import (
	"bytes"
	"context"
	"errors"
	"html"
	"io"
	"strconv"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestConfigRender(t *testing.T) {
	t.Parallel()

	if _, err := page.New(page.Config{
		HTML: "<p>conflict</p>",
		Render: func(w io.Writer, _ any) error {
			_, err := io.WriteString(w, "<p>ok</p>")
			return err
		},
		Width:  400,
		Height: 300,
	}); !errors.Is(err, page.ErrBadSource) {
		t.Fatalf("Render+HTML err = %v, want ErrBadSource", err)
	}

	p, err := page.New(page.Config{
		Render: func(w io.Writer, data any) error {
			label, _ := data.(string)
			_, err := io.WriteString(w, "<!DOCTYPE html><html><body><div id=\"msg\">"+html.EscapeString(label)+"</div></body></html>")
			return err
		},
		Width:  400,
		Height: 300,
	})
	if err != nil {
		t.Fatalf("New with Render failed: %v", err)
	}

	ctx := context.Background()
	p.SetData("first")
	if err := p.Redraw(ctx); err != nil {
		t.Fatalf("first Redraw failed: %v", err)
	}
	if !hasTextOp(p.Display().Ops, "first") {
		t.Fatal("first render text not found in Display.Ops")
	}

	p.SetData("second")
	if err := p.Redraw(ctx); err != nil {
		t.Fatalf("second Redraw failed: %v", err)
	}
	if !hasTextOp(p.Display().Ops, "second") {
		t.Fatal("second render text not found in Display.Ops")
	}
}

func hasTextOp(ops []layout.DisplayOp, want string) bool {
	for i := range ops {
		if ops[i].Text == want {
			return true
		}
	}

	return false
}

func writeScaleRowsHTML(w io.Writer, rows []scaleRow) error {
	var buf *bytes.Buffer
	if b, ok := w.(*bytes.Buffer); ok {
		buf = b
	}
	write := func(s string) error {
		if buf != nil {
			buf.WriteString(s)
			return nil
		}
		_, err := io.WriteString(w, s)
		return err
	}

	if err := write(`<!DOCTYPE html><html><head><style>` +
		`html,body{margin:0;font-family:sans-serif;color:#1c1915;background:#f4f1ea}` +
		`#head{height:40px;background:#1c1915;color:#ffffff}` +
		`.row{display:flex;box-sizing:border-box;height:40px;gap:12px;padding:8px 16px;border-bottom:1px solid #ddd6c7}` +
		`.num{width:64px}.title{width:200px;font-weight:bold}.body{color:#4a463c}` +
		`</style></head><body><div id="head">Scale benchmark</div>`); err != nil {
		return err
	}

	var numBuf [20]byte
	for i := range rows {
		r := &rows[i]
		num := string(strconv.AppendInt(numBuf[:0], int64(r.Num), 10))
		if err := write(`<div id="row-`); err != nil {
			return err
		}
		if err := write(num); err != nil {
			return err
		}
		if err := write(`" class="row"><span class="num">`); err != nil {
			return err
		}
		if err := write(num); err != nil {
			return err
		}
		if err := write(`</span><span class="title">`); err != nil {
			return err
		}
		if err := write(html.EscapeString(r.Title)); err != nil {
			return err
		}
		if err := write(`</span><span class="body">`); err != nil {
			return err
		}
		if err := write(html.EscapeString(r.Body)); err != nil {
			return err
		}
		if err := write(`</span></div>`); err != nil {
			return err
		}
	}

	return write(`</body></html>`)
}

func BenchmarkScaleRenderFunc(b *testing.B) {
	rows := makeScaleRows(1000)
	var buf bytes.Buffer
	buf.Grow(160 * 1024)
	b.ReportAllocs()
	for b.Loop() {
		buf.Reset()
		if err := writeScaleRowsHTML(&buf, rows); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*len(rows)), "ns/row")
}
