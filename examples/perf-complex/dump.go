package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// dumpFilePerm keeps measurement dumps owner-only.
const dumpFilePerm = 0o600

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), dumpFilePerm)
}

func writeDump(out, mode string, d dump, p *page.Page) error {
	if err := writeJSON(filepath.Join(out, mode+".json"), d); err != nil {
		return err
	}
	// Initial samples own fresh pages; materialize the original for geometry.
	if p.Display() == nil {
		if err := redraw(p); err != nil {
			return err
		}
	}
	display := p.Display()
	type op struct {
		Kind       int
		X, Y, W, H float64
		Text       string
	}
	ops := make([]op, 0, len(display.Ops))
	for _, o := range display.Ops {
		ops = append(ops, op{int(o.Kind), o.X, o.Y, o.W, o.H, o.Text})
	}
	return writeLayout(filepath.Join(out, mode+"-layout.json.gz"), map[string]any{
		"width": display.Width, "height": display.Height, "boxes": p.Boxes(), "ops": ops,
	})
}
