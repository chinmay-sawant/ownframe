//go:build android || ios

package mobile

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/eyes/eyes"
)

func start() error {
	a, err := eyes.New()
	if err != nil {
		return fmt.Errorf("mobile: eyes: %w", err)
	}
	a.BindTouch()
	setAndroidPageProfile(a.Page())
	inputPage = a.Page()
	if err := ownframe.BindMobile(context.Background(), a.Page()); err != nil {
		return fmt.Errorf("mobile: bind: %w", err)
	}
	return nil
}
