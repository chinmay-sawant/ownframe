//go:build android || ios

package mobile

import (
	"context"
	"fmt"
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/app"
)

func start(path string) error {
	a, err := app.New(path)
	if err != nil {
		return fmt.Errorf("pet care: %w", err)
	}
	application = a
	return ownframe.BindMobile(context.Background(), a.Page())
}
