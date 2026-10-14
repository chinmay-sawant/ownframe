//go:build android || ios

package mobile

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/digital-wellbeing/dashboard"
)

func start() error {
	var err error
	app, err = dashboard.New()
	if err != nil {
		return fmt.Errorf("mobile: dashboard: %w", err)
	}
	app.Page().SetTouchZoomAllowed(false)
	inputPage = app.Page()
	return ownframe.BindMobile(context.Background(), app.Page())
}

func init() {
	if err := start(); err != nil {
		panic(err)
	}
}
