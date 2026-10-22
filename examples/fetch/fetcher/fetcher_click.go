package fetcher

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// onClick runs the request for the clicked button. A failure lands in
// View.Status through Get or Post, so the click itself does not fail.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	switch box.ID {
	case "get":
		_ = a.Get(ctx, a.page.Form().FormValue("url"))
	case "post":
		_ = a.Post(ctx, a.page.Form().FormValue("url"))
	case "bad":
		_ = a.Get(ctx, "ftp://x")
	}

	return nil
}
