//go:build android || ios

package mobile

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/telegram/telegram"
)

func start() error {
	a, err := telegram.New()
	if err != nil {
		return fmt.Errorf("mobile: telegram: %w", err)
	}

	app = a
	setAndroidPageProfile(a.Page())

	inputPage = a.Page()

	if err := ownframe.BindMobile(context.Background(), a.Page()); err != nil {
		return fmt.Errorf("mobile: bind: %w", err)
	}

	return nil
}
