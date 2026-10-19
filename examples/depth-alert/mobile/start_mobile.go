//go:build android || ios

package mobile

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/depth-alert/depthalert"
)

func start() error {
	a, err := depthalert.New()
	if err != nil {
		return fmt.Errorf("mobile: depth alert: %w", err)
	}
	app = a
	if err := ownframe.BindMobile(context.Background(), a.Page()); err != nil {
		return fmt.Errorf("mobile: bind: %w", err)
	}
	return nil
}
