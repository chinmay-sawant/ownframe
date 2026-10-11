//go:build android || ios

package mobile

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/lg-remote/remote"
)

func start() error {
	app, err := remote.New(remote.WithPhone(true))
	if err != nil {
		return fmt.Errorf("mobile: remote: %w", err)
	}

	setAndroidPageProfile(app.Page())

	inputPage = app.Page()

	if err := ownframe.BindMobile(context.Background(), app.Page()); err != nil {
		return fmt.Errorf("mobile: bind: %w", err)
	}

	return nil
}
