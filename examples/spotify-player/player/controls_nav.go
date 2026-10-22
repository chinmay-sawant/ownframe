package player

import (
	"context"
	"strings"
)

// clickNav applies one navigation action, reporting whether it consumed
// it. The caller still redraws and logs the click.
func (a *App) clickNav(ctx context.Context, action string) bool {
	switch {
	case action == "back":
		a.moveNav(ctx, -1)
	case action == "forward":
		a.moveNav(ctx, 1)
	case strings.HasPrefix(action, "list-"):
		a.selectList(ctx, slot(action, "list-"))
	case action == "queue":
		a.navigate(ctx, "queue")
	case strings.HasPrefix(action, "nav-"):
		a.navigate(ctx, strings.TrimPrefix(action, "nav-"))
	default:
		return false
	}

	return true
}
