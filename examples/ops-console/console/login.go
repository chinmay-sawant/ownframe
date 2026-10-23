package console

import (
	"context"
	"crypto/subtle"

	"github.com/chinmay-sawant/ownframe"
)

// Demo credentials only. No auth boundary; do not copy into a real app.
const (
	operatorName = "dock"
	operatorPIN  = "1947"
	cmpEqual     = 1
)

func (a *App) onChange(context.Context, ownframe.Box) error {
	a.refreshReady()

	return nil
}

func (a *App) onSubmit(_ context.Context) error {
	if a.view.Screen == screenLogin {
		a.signIn()
	}

	if a.view.Screen == screenDetail {
		a.saveNote()
	}

	a.syncView()
	a.page.SetData(&a.view)

	return nil
}

func (a *App) ready() bool {
	return a.view.Name != "" && a.view.PIN.Reveal() != ""
}

func (a *App) refreshReady() {
	ready := a.ready()
	if a.view.Ready == ready {
		return
	}

	a.view.Ready = ready
	a.page.SetData(&a.view)
}

func pinMatch(got string) bool {
	want := []byte(operatorPIN)
	have := []byte(got)
	if len(have) != len(want) {
		return false
	}

	return subtle.ConstantTimeCompare(have, want) == cmpEqual
}
