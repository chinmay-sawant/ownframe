package studio

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	osclip "github.com/chinmay-sawant/ownframe/internal/clipboard"
)

// onClick routes a click by control id. Click draws the page after this
// handler returns, so a theme change lands without an extra Redraw.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	switch box.ID {
	case "btn-apply":
		return a.Apply(ctx, a.page.FormValue("css"))
	case "btn-copy":
		css := a.page.FormValue("css")
		osclip.Write(css)
		a.view.Status = fmt.Sprintf("copied %d bytes", len(css))
		a.page.SetData(&a.view)

		return nil
	case "btn-reset":
		css, err := loadTheme(a.path)
		if err != nil {
			return err
		}

		a.page.SetFormValue("css", css)

		return a.Apply(ctx, css)
	}

	name, ok := strings.CutPrefix(box.ID, "preset-")
	if !ok {
		return nil
	}

	for _, p := range Presets() {
		if p.Name != name {
			continue
		}

		a.page.SetFormValue("css", p.CSS)
		if err := a.Apply(ctx, p.CSS); err != nil {
			return err
		}

		a.view.Preset = p.Name
		a.page.SetData(&a.view)

		return nil
	}

	return nil
}

// Apply stores css as the page theme, writes the same bytes to the theme
// file, and refreshes the editor, the token strip, and the status line.
// It does not draw: Click redraws after its handler.
func (a *App) Apply(_ context.Context, css string) error {
	if err := a.page.SetTheme(css); err != nil {
		return err
	}

	if err := saveTheme(a.path, css); err != nil {
		return err
	}

	a.view.Editor = css
	a.view.Tokens = parseTokens(css)
	a.view.Status = fmt.Sprintf("applied, %d tokens", len(a.view.Tokens))
	a.page.SetData(&a.view)

	return nil
}
