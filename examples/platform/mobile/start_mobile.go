//go:build android || ios

package mobile

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/platform/platform"
)

func start() error {
	app, err := platform.New()
	if err != nil {
		return fmt.Errorf("mobile: platform: %w", err)
	}

	inputPage = app.Page()

	if err := ownframe.BindMobile(context.Background(), app.Page()); err != nil {
		return fmt.Errorf("mobile: bind: %w", err)
	}

	return nil
}
