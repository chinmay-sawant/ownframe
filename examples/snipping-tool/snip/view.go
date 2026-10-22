package snip

import (
	"bytes"
	"context"
	_ "embed"
	"image"
	"strconv"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed template.html
var templateHTML string

const (
	// DefaultWidth and DefaultHeight size a new window.
	DefaultWidth  = 720
	DefaultHeight = 560
)

// shotSrc is the image name the template shows.
const shotSrc = "shot"

// View is the template data: the shot flag and size, the selection
// rect, the status line, and the annotation count from annotate.go.
type View struct {
	HasImage bool
	Size     string
	SelX     int
	SelY     int
	SelW     int
	SelH     int
	Status   string
	Notes    int
}

// View returns a copy of the template data.
func (a *App) View() View { return a.view }

// SetForm stores a control value for tests. It needs a Redraw first.
func (a *App) SetForm(id, value string) { a.page.Form().SetFormValue(id, value) }

// SetSaveDir overrides the save directory, for tests and -data.
func (a *App) SetSaveDir(dir string) { a.saveDir = dir }

// SetWeb marks the browser still preview, which has no OS clipboard.
func (a *App) SetWeb(web bool) { a.web = web }

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box { return a.page.Boxes() }

// Click hit-tests the page and activates the control under the point.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}

// describe renders "WxH format" for shot bytes, or "" without any.
func describe(raw []byte) string {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return ""
	}
	return strconv.Itoa(cfg.Width) + "x" + strconv.Itoa(cfg.Height) + " " + format
}
